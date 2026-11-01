package lease

import (
	"errors"
	"time"

	"github.com/patrickishaf/property-mgt/common"
	"github.com/patrickishaf/property-mgt/internal/db/model"
	"github.com/patrickishaf/property-mgt/internal/db/repository"
	"gorm.io/gorm"
)

const errInstallmentNotFound = "installment not found"

type scheduleMismatchError struct {
	expected float64
}

func (this scheduleMismatchError) Error() string {
	return "installments must add up to the lease's outstanding rent of " + describeAmount(this.expected)
}

func toInstallmentList(installments []model.Installment, now time.Time) []installmentRes {
	items := make([]installmentRes, 0, len(installments))
	for _, i := range installments {
		items = append(items, toInstallmentRes(i, now))
	}
	return items
}

func (this *service) getInstallments(actor *common.Actor, dto getInstallmentsReq) common.ServiceResponse {
	logger := this.logger.With("caller", "lease.service.getInstallments", "actor", actor.Email)
	filter := repository.InstallmentFilter{
		DueBefore:  parseDate(dto.DueBefore),
		LeaseID:    dto.LeaseID,
		PropertyID: dto.PropertyID,
		Status:     dto.Status,
		TenantID:   dto.TenantID,
	}
	switch {
	case actor.Role == common.ActorRoleTenant:
		filter.TenantID = actor.ID
	case isAdmin(actor):
		filter.OwnerID, filter.ManagerID = adminScope(actor)
	default:
		return forbidden()
	}

	now := time.Now()
	page, limit := dto.values()
	installments, total, err := this.installmentRepo.FindMany(filter, now, page, limit)
	if err != nil {
		logger.Error("failed to find installments", "err", err.Error())
		return serverError()
	}
	return common.ServiceResponse{StatusCode: 200, Data: listRes[installmentRes]{
		Items: toInstallmentList(installments, now),
		Limit: limit,
		Page:  page,
		Total: total,
	}}
}

func (this *service) getLeaseInstallments(actor *common.Actor, leaseID uint) common.ServiceResponse {
	logger := this.logger.With("caller", "lease.service.getLeaseInstallments", "actor", actor.Email)
	lease, errRes := this.findVisibleLease(actor, leaseID, logger)
	if errRes != nil {
		return *errRes
	}
	installments, err := this.installmentRepo.FindByLease(lease.ID)
	if err != nil {
		logger.Error("failed to find installments", "err", err.Error())
		return serverError()
	}
	return common.ServiceResponse{StatusCode: 200, Data: toInstallmentList(installments, time.Now())}
}

// createLeaseInstallment replaces the lease's unpaid installments with a new
// schedule. Installments that have received money are kept; the new items plus
// what is still owed on the kept ones must equal the lease's outstanding rent.
func (this *service) createLeaseInstallment(actor *common.Actor, leaseID uint, dto setInstallmentsReq) common.ServiceResponse {
	logger := this.logger.With("caller", "lease.service.createLeaseInstallment", "actor", actor.Email)
	lease, errRes := this.findAdminLease(actor, leaseID, logger)
	if errRes != nil {
		return *errRes
	}
	if lease.Status != model.LeaseStatusActive {
		return common.ServiceResponse{StatusCode: 409, Err: errLeaseNotActive}
	}

	err := this.db.Transaction(func(trx *gorm.DB) error {
		locked, err := this.leaseRepo.FindOneByIdForUpdateTRX(lease.ID, trx)
		if err != nil {
			return err
		}
		existing, err := this.installmentRepo.FindByLeaseForUpdateTRX(lease.ID, trx)
		if err != nil {
			return err
		}
		total := 0.0
		for _, inst := range existing {
			if inst.AmountPaid > 0 {
				total += inst.OutstandingAmount
			}
		}
		items := make([]model.Installment, 0, len(dto.Items))
		for _, item := range dto.Items {
			dueDate, _ := time.Parse(dateLayout, item.DueDate)
			amount := round2(item.Amount)
			total += amount
			items = append(items, model.Installment{
				AmountExpected:    amount,
				DueDate:           &dueDate,
				LeaseID:           int(lease.ID),
				OutstandingAmount: amount,
				PropertyID:        lease.PropertyID,
				Status:            model.InstallmentStatusUnpaid,
				TenantID:          lease.TenantID,
			})
		}
		if round2(total) != round2(locked.OutstandingRent) {
			return scheduleMismatchError{expected: locked.OutstandingRent}
		}
		if err := this.installmentRepo.DeleteUnpaidTRX(lease.ID, trx); err != nil {
			return err
		}
		return this.installmentRepo.SaveManyTRX(items, trx)
	})
	var mismatch scheduleMismatchError
	if errors.As(err, &mismatch) {
		return common.ServiceResponse{StatusCode: 400, Err: mismatch.Error()}
	}
	if err != nil {
		logger.Error("failed to set installments", "err", err.Error())
		return serverError()
	}

	installments, err := this.installmentRepo.FindByLease(lease.ID)
	if err != nil {
		logger.Error("failed to find installments", "err", err.Error())
		return serverError()
	}
	return common.ServiceResponse{StatusCode: 201, Data: toInstallmentList(installments, time.Now())}
}

func (this *service) updateInstallment(actor *common.Actor, id uint, dto updateInstallmentReq) common.ServiceResponse {
	logger := this.logger.With("caller", "lease.service.updateInstallment", "actor", actor.Email)
	if !isAdmin(actor) {
		return forbidden()
	}
	installment, err := this.installmentRepo.FindOneById(id)
	if err != nil {
		logger.Error("failed to find installment", "err", err.Error())
		return serverError()
	}
	if installment == nil {
		return common.ServiceResponse{StatusCode: 404, Err: errInstallmentNotFound}
	}
	property, err := this.propertyRepo.FindOneById(uint(installment.PropertyID))
	if err != nil {
		logger.Error("failed to find property", "err", err.Error())
		return serverError()
	}
	if !property.AccessibleBy(actor) {
		return common.ServiceResponse{StatusCode: 404, Err: errInstallmentNotFound}
	}

	dueDate, _ := time.Parse(dateLayout, dto.DueDate)
	if err := this.installmentRepo.UpdateDueDate(installment.ID, dueDate); err != nil {
		logger.Error("failed to update installment", "err", err.Error())
		return serverError()
	}
	installment.DueDate = &dueDate
	return common.ServiceResponse{StatusCode: 200, Data: toInstallmentRes(*installment, time.Now())}
}
