package payment

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
	return "installments must add up to the lease's outstanding rent of " + describeOutstanding(this.expected)
}

func (this *service) getInstallments(actor *common.Actor, dto getInstallmentsReq) common.ServiceResponse {
	logger := this.logger.With("caller", "payment.service.getInstallments", "actor", actor.Email)
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
	items := make([]installmentRes, 0, len(installments))
	for _, i := range installments {
		items = append(items, toInstallmentRes(i, now))
	}
	return common.ServiceResponse{StatusCode: 200, Data: listRes[installmentRes]{Items: items, Limit: limit, Page: page, Total: total}}
}

// setInstallments replaces the lease's unpaid installments with a new schedule.
// Installments that have received money are kept; the new items plus what is
// still owed on the kept ones must equal the lease's outstanding rent.
func (this *service) setInstallments(actor *common.Actor, dto setInstallmentsReq) common.ServiceResponse {
	logger := this.logger.With("caller", "payment.service.setInstallments", "actor", actor.Email)
	lease, _, errRes := this.findAdminLease(actor, dto.LeaseID, logger)
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
	now := time.Now()
	items := make([]installmentRes, 0, len(installments))
	for _, i := range installments {
		items = append(items, toInstallmentRes(i, now))
	}
	return common.ServiceResponse{StatusCode: 201, Data: items}
}

func (this *service) updateInstallment(actor *common.Actor, id uint, dto updateInstallmentReq) common.ServiceResponse {
	logger := this.logger.With("caller", "payment.service.updateInstallment", "actor", actor.Email)
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
	ok, err := this.canSeeRecord(actor, installment.TenantID, installment.PropertyID)
	if err != nil {
		logger.Error("failed to check installment access", "err", err.Error())
		return serverError()
	}
	if !ok {
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

func (this *service) updateLease(actor *common.Actor, id uint, dto updateLeaseReq) common.ServiceResponse {
	logger := this.logger.With("caller", "payment.service.updateLease", "actor", actor.Email)
	lease, _, errRes := this.findAdminLease(actor, id, logger)
	if errRes != nil {
		return *errRes
	}

	fields := map[string]any{}
	if dto.AllowPartialPayment != nil {
		fields["allow_partial_payment"] = *dto.AllowPartialPayment
		lease.AllowPartialPayment = *dto.AllowPartialPayment
	}
	if dto.MinimumPartialAmount != nil {
		minimum := round2(*dto.MinimumPartialAmount)
		if minimum > round2(lease.ExpectedRent) {
			return common.ServiceResponse{StatusCode: 400, Err: "minimum partial amount cannot exceed the lease's expected rent"}
		}
		fields["minimum_partial_amount"] = minimum
		lease.MinimumPartialAmount = minimum
	}
	if len(fields) > 0 {
		if err := this.leaseRepo.Update(lease.ID, fields); err != nil {
			logger.Error("failed to update lease", "err", err.Error())
			return serverError()
		}
	}
	return common.ServiceResponse{StatusCode: 200, Data: toLeaseRes(*lease)}
}
