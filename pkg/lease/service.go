package lease

import (
	"errors"
	"fmt"
	"log/slog"
	"math"
	"time"

	"github.com/patrickishaf/property-mgt/common"
	"github.com/patrickishaf/property-mgt/config"
	"github.com/patrickishaf/property-mgt/internal/db/model"
	"github.com/patrickishaf/property-mgt/internal/db/repository"
	"gorm.io/gorm"
)

const (
	errLeaseNotActive = "lease is not active"
	errLeaseNotFound  = "lease not found"
)

var errNotActive = errors.New(errLeaseNotActive)

type service struct {
	cfg             *config.AppConfig
	db              *gorm.DB
	installmentRepo *repository.InstallmentRepository
	leaseRepo       *repository.LeaseRepository
	logger          *slog.Logger
	propertyRepo    *repository.PropertyRepository
	tenantRepo      *repository.TenantRepository
	unitRepo        *repository.UnitRepository
}

func CreateService(
	cfg *config.AppConfig,
	logger *slog.Logger,
	db *gorm.DB,
	leaseRepo *repository.LeaseRepository,
	installmentRepo *repository.InstallmentRepository,
	tenantRepo *repository.TenantRepository,
	propertyRepo *repository.PropertyRepository,
	unitRepo *repository.UnitRepository,
) *service {
	return &service{
		cfg:             cfg,
		db:              db,
		installmentRepo: installmentRepo,
		leaseRepo:       leaseRepo,
		logger:          logger,
		propertyRepo:    propertyRepo,
		tenantRepo:      tenantRepo,
		unitRepo:        unitRepo,
	}
}

// round2 rounds a money amount to 2 decimal places.
func round2(amount float64) float64 {
	return math.Round(amount*100) / 100
}

func describeAmount(amount float64) string {
	return fmt.Sprintf("%.2f", round2(amount))
}

func isAdmin(actor *common.Actor) bool {
	return actor.Role == common.AdminRoleOwner.ToString() || actor.Role == common.AdminRoleManager.ToString()
}

func forbidden() common.ServiceResponse {
	return common.ServiceResponse{StatusCode: 403, Err: common.ErrInvalidCredentials}
}

func serverError() common.ServiceResponse {
	return common.ServiceResponse{StatusCode: 500, Err: common.ErrGeneric}
}

// adminScope returns the owner/manager IDs that limit queries to the actor's properties.
func adminScope(actor *common.Actor) (ownerID uint, managerID uint) {
	if actor.Role == common.AdminRoleOwner.ToString() {
		return actor.ID, 0
	}
	return 0, actor.ID
}

func parseDate(value string) *time.Time {
	if value == "" {
		return nil
	}
	t, err := time.Parse(dateLayout, value)
	if err != nil {
		return nil
	}
	return &t
}

// findAdminLease loads a lease on a property the admin owns or manages.
func (this *service) findAdminLease(actor *common.Actor, leaseID uint, logger *slog.Logger) (*model.Lease, *common.ServiceResponse) {
	if !isAdmin(actor) {
		res := forbidden()
		return nil, &res
	}
	lease, err := this.leaseRepo.FindOneById(leaseID)
	if err != nil {
		logger.Error("failed to find lease", "err", err.Error())
		res := serverError()
		return nil, &res
	}
	if lease == nil {
		return nil, &common.ServiceResponse{StatusCode: 404, Err: errLeaseNotFound}
	}
	property, err := this.propertyRepo.FindOneById(uint(lease.PropertyID))
	if err != nil {
		logger.Error("failed to find property", "err", err.Error())
		res := serverError()
		return nil, &res
	}
	if !property.AccessibleBy(actor) {
		return nil, &common.ServiceResponse{StatusCode: 404, Err: errLeaseNotFound}
	}
	return lease, nil
}

// findVisibleLease loads a lease the actor may read: their own as a tenant, or
// one on a property they own or manage as an admin.
func (this *service) findVisibleLease(actor *common.Actor, leaseID uint, logger *slog.Logger) (*model.Lease, *common.ServiceResponse) {
	if isAdmin(actor) {
		return this.findAdminLease(actor, leaseID, logger)
	}
	if actor.Role != common.ActorRoleTenant {
		res := forbidden()
		return nil, &res
	}
	lease, err := this.leaseRepo.FindOneById(leaseID)
	if err != nil {
		logger.Error("failed to find lease", "err", err.Error())
		res := serverError()
		return nil, &res
	}
	if lease == nil || uint(lease.TenantID) != actor.ID {
		return nil, &common.ServiceResponse{StatusCode: 404, Err: errLeaseNotFound}
	}
	return lease, nil
}

func (this *service) getLeases(actor *common.Actor, dto getLeasesReq) common.ServiceResponse {
	logger := this.logger.With("caller", "lease.service.getLeases", "actor", actor.Email)
	filter := repository.LeaseFilter{
		EndingBefore: parseDate(dto.EndingBefore),
		HasBalance:   dto.HasBalance,
		PropertyID:   dto.PropertyID,
		Status:       dto.Status,
		TenantID:     dto.TenantID,
		UnitNumber:   dto.UnitNumber,
	}
	switch {
	case actor.Role == common.ActorRoleTenant:
		filter.TenantID = actor.ID
	case isAdmin(actor):
		filter.OwnerID, filter.ManagerID = adminScope(actor)
	default:
		return forbidden()
	}

	page, limit := dto.values()
	leases, total, err := this.leaseRepo.FindMany(filter, page, limit)
	if err != nil {
		logger.Error("failed to find leases", "err", err.Error())
		return serverError()
	}
	items := make([]leaseRes, 0, len(leases))
	for _, l := range leases {
		items = append(items, toLeaseRes(l))
	}
	return common.ServiceResponse{StatusCode: 200, Data: listRes[leaseRes]{Items: items, Limit: limit, Page: page, Total: total}}
}

func (this *service) leaseDetail(lease *model.Lease, statusCode int, logger *slog.Logger) common.ServiceResponse {
	res := leaseDetailRes{leaseRes: toLeaseRes(*lease)}
	if tenant, err := this.tenantRepo.FindOneById(uint(lease.TenantID)); err != nil {
		logger.Error("failed to find tenant", "err", err.Error())
		return serverError()
	} else if tenant != nil {
		res.TenantFullName = tenant.FirstName + " " + tenant.LastName
	}
	if property, err := this.propertyRepo.FindOneById(uint(lease.PropertyID)); err != nil {
		logger.Error("failed to find property", "err", err.Error())
		return serverError()
	} else if property != nil {
		res.PropertyName = property.Name
	}
	open, overdue, err := this.installmentRepo.CountOpenByLease(lease.ID, time.Now())
	if err != nil {
		logger.Error("failed to count installments", "err", err.Error())
		return serverError()
	}
	res.OpenInstallmentCount = open
	res.OverdueInstallmentCount = overdue
	return common.ServiceResponse{StatusCode: statusCode, Data: res}
}

func (this *service) getLeaseById(actor *common.Actor, id uint) common.ServiceResponse {
	logger := this.logger.With("caller", "lease.service.getLeaseById", "actor", actor.Email)
	lease, errRes := this.findVisibleLease(actor, id, logger)
	if errRes != nil {
		return *errRes
	}
	return this.leaseDetail(lease, 200, logger)
}

func (this *service) createLease(actor *common.Actor, dto createLeaseReq) common.ServiceResponse {
	logger := this.logger.With("caller", "lease.service.createLease", "actor", actor.Email)
	if !isAdmin(actor) {
		return forbidden()
	}
	startDate, _ := time.Parse(dateLayout, dto.StartDate)
	endDate, _ := time.Parse(dateLayout, dto.EndDate)
	if !endDate.After(startDate) {
		return common.ServiceResponse{StatusCode: 400, Err: "lease end date must be after start date"}
	}
	if dto.MinimumPartialAmount > dto.ExpectedRent {
		return common.ServiceResponse{StatusCode: 400, Err: "minimum partial amount cannot exceed the lease's expected rent"}
	}

	tenant, err := this.tenantRepo.FindOneById(dto.TenantID)
	if err != nil {
		logger.Error("failed to find tenant", "err", err.Error())
		return serverError()
	}
	if tenant == nil {
		return common.ServiceResponse{StatusCode: 404, Err: "tenant not found"}
	}
	currentProperty, err := this.propertyRepo.FindOneById(uint(tenant.PropertyID))
	if err != nil {
		logger.Error("failed to find tenant's property", "err", err.Error())
		return serverError()
	}
	if !currentProperty.AccessibleBy(actor) {
		return common.ServiceResponse{StatusCode: 404, Err: "tenant not found"}
	}

	property, err := this.propertyRepo.FindOneById(dto.PropertyID)
	if err != nil {
		logger.Error("failed to find property", "err", err.Error())
		return serverError()
	}
	if !property.AccessibleBy(actor) {
		return common.ServiceResponse{StatusCode: 404, Err: "property not found"}
	}
	if property.ArchivedAt != nil {
		return common.ServiceResponse{StatusCode: 409, Err: "property is archived"}
	}
	unit, err := this.unitRepo.FindOneByPropertyAndNumber(property.ID, dto.UnitNumber)
	if err != nil {
		logger.Error("failed to find unit", "err", err.Error())
		return serverError()
	}
	if unit == nil {
		return common.ServiceResponse{StatusCode: 404, Err: "unit not found"}
	}
	occupied, err := this.leaseRepo.ExistsActiveForUnit(property.ID, dto.UnitNumber)
	if err != nil {
		logger.Error("failed to check unit occupancy", "err", err.Error())
		return serverError()
	}
	if occupied {
		return common.ServiceResponse{StatusCode: 409, Err: "unit already has an active lease"}
	}
	active, err := this.leaseRepo.FindActiveByTenant(tenant.ID)
	if err != nil {
		logger.Error("failed to find active lease", "err", err.Error())
		return serverError()
	}
	if active != nil {
		return common.ServiceResponse{StatusCode: 409, Err: "tenant already has an active lease"}
	}

	rent := round2(dto.ExpectedRent)
	lease := model.Lease{
		AllowPartialPayment:  dto.AllowPartialPayment,
		DurationType:         dto.DurationType,
		EndDate:              endDate,
		ExpectedRent:         rent,
		MinimumPartialAmount: round2(dto.MinimumPartialAmount),
		OutstandingRent:      rent,
		PropertyID:           int(property.ID),
		StartDate:            startDate,
		Status:               model.LeaseStatusActive,
		TenantID:             int(tenant.ID),
		UnitNumber:           dto.UnitNumber,
	}
	err = this.db.Transaction(func(trx *gorm.DB) error {
		if err := this.leaseRepo.SaveOneTRX(&lease, trx); err != nil {
			return err
		}
		if err := this.propertyRepo.IncrementOccupiedUnitsTRX(property.ID, trx); err != nil {
			return err
		}
		return this.tenantRepo.UpdateTRX(tenant.ID, map[string]any{
			"expected_annual_rent":    rent,
			"lease_id":                lease.ID,
			"outstanding_annual_rent": rent,
			"property_id":             property.ID,
			"unit_number":             dto.UnitNumber,
		}, trx)
	})
	if err != nil {
		logger.Error("failed to create lease", "err", err.Error())
		return serverError()
	}
	return this.leaseDetail(&lease, 201, logger)
}

// checkLeaseUpdate validates a lease update against the lease's current state.
// openInstallments is how many of its installments still have money owed.
func checkLeaseUpdate(lease model.Lease, dto updateLeaseReq, openInstallments int64) *common.ServiceResponse {
	active := lease.Status == model.LeaseStatusActive
	if (dto.EndDate != nil || dto.ExpectedRent != nil || dto.Status != nil) && !active {
		return &common.ServiceResponse{StatusCode: 409, Err: errLeaseNotActive}
	}
	if dto.EndDate != nil {
		endDate, _ := time.Parse(dateLayout, *dto.EndDate)
		if !endDate.After(lease.StartDate) {
			return &common.ServiceResponse{StatusCode: 400, Err: "lease end date must be after start date"}
		}
	}
	rent := lease.ExpectedRent
	if dto.ExpectedRent != nil {
		rent = round2(*dto.ExpectedRent)
		if openInstallments > 0 {
			return &common.ServiceResponse{StatusCode: 409, Err: "lease has unpaid installments"}
		}
		if rent < round2(lease.AmountPaid) {
			return &common.ServiceResponse{StatusCode: 400, Err: "expected rent cannot be less than the amount already paid (" + describeAmount(lease.AmountPaid) + ")"}
		}
	}
	minimum := lease.MinimumPartialAmount
	if dto.MinimumPartialAmount != nil {
		minimum = round2(*dto.MinimumPartialAmount)
	}
	if minimum > rent {
		return &common.ServiceResponse{StatusCode: 400, Err: "minimum partial amount cannot exceed the lease's expected rent"}
	}
	return nil
}

func (this *service) updateLease(actor *common.Actor, id uint, dto updateLeaseReq) common.ServiceResponse {
	logger := this.logger.With("caller", "lease.service.updateLease", "actor", actor.Email)
	lease, errRes := this.findAdminLease(actor, id, logger)
	if errRes != nil {
		return *errRes
	}
	var openInstallments int64
	if dto.ExpectedRent != nil {
		open, _, err := this.installmentRepo.CountOpenByLease(lease.ID, time.Now())
		if err != nil {
			logger.Error("failed to count installments", "err", err.Error())
			return serverError()
		}
		openInstallments = open
	}
	if errRes := checkLeaseUpdate(*lease, dto, openInstallments); errRes != nil {
		return *errRes
	}

	err := this.db.Transaction(func(trx *gorm.DB) error {
		locked, err := this.leaseRepo.FindOneByIdForUpdateTRX(lease.ID, trx)
		if err != nil {
			return err
		}
		if (dto.EndDate != nil || dto.ExpectedRent != nil || dto.Status != nil) && locked.Status != model.LeaseStatusActive {
			return errNotActive
		}
		fields := map[string]any{}
		if dto.AllowPartialPayment != nil {
			fields["allow_partial_payment"] = *dto.AllowPartialPayment
		}
		if dto.MinimumPartialAmount != nil {
			fields["minimum_partial_amount"] = round2(*dto.MinimumPartialAmount)
		}
		if dto.EndDate != nil {
			endDate, _ := time.Parse(dateLayout, *dto.EndDate)
			fields["end_date"] = endDate
		}
		if dto.ExpectedRent != nil {
			rent := round2(*dto.ExpectedRent)
			outstanding := round2(rent - locked.AmountPaid)
			fields["expected_rent"] = rent
			fields["outstanding_rent"] = outstanding
			tenant, err := this.tenantRepo.FindOneById(uint(locked.TenantID))
			if err != nil {
				return err
			}
			if tenant != nil && uint(tenant.LeaseID) == locked.ID {
				err := this.tenantRepo.UpdateTRX(tenant.ID, map[string]any{
					"expected_annual_rent":    rent,
					"outstanding_annual_rent": outstanding,
				}, trx)
				if err != nil {
					return err
				}
			}
		}
		if len(fields) > 0 {
			if err := this.leaseRepo.UpdateTRX(locked.ID, fields, trx); err != nil {
				return err
			}
		}
		if dto.Status != nil {
			return this.endLeaseTRX(locked, *dto.Status, time.Now(), trx)
		}
		return nil
	})
	if errors.Is(err, errNotActive) {
		return common.ServiceResponse{StatusCode: 409, Err: errLeaseNotActive}
	}
	if err != nil {
		logger.Error("failed to update lease", "err", err.Error())
		return serverError()
	}

	updated, err := this.leaseRepo.FindOneById(lease.ID)
	if err != nil || updated == nil {
		logger.Error("failed to reload lease", "lease_id", lease.ID)
		return serverError()
	}
	return this.leaseDetail(updated, 200, logger)
}

// endLeaseTRX marks a locked, active lease as ended or terminated and frees its
// unit. The outstanding balance and installments are left as they are.
func (this *service) endLeaseTRX(lease *model.Lease, status string, now time.Time, trx *gorm.DB) error {
	if lease.Status != model.LeaseStatusActive {
		return errNotActive
	}
	err := this.leaseRepo.UpdateTRX(lease.ID, map[string]any{
		"ended_at": now,
		"status":   status,
	}, trx)
	if err != nil {
		return err
	}
	return this.propertyRepo.DecrementOccupiedUnitsTRX(uint(lease.PropertyID), trx)
}
