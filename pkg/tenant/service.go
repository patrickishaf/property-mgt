package tenant

import (
	"log/slog"
	"time"

	"github.com/patrickishaf/property-mgt/common"
	"github.com/patrickishaf/property-mgt/config"
	"github.com/patrickishaf/property-mgt/internal/db/model"
	"github.com/patrickishaf/property-mgt/internal/db/repository"
	"github.com/patrickishaf/property-mgt/providers/cpaas"
	"gorm.io/gorm"
)

const (
	defaultPageSize = 20
	errNotFound     = "tenant not found"
)

type service struct {
	cfg          *config.AppConfig
	db           *gorm.DB
	jobRepo      *repository.JobRepository
	leaseRepo    *repository.LeaseRepository
	logger       *slog.Logger
	propertyRepo *repository.PropertyRepository
	tenantRepo   *repository.TenantRepository
	unitRepo     *repository.UnitRepository
}

func CreateService(
	cfg *config.AppConfig,
	logger *slog.Logger,
	db *gorm.DB,
	tenantRepo *repository.TenantRepository,
	leaseRepo *repository.LeaseRepository,
	propertyRepo *repository.PropertyRepository,
	jobRepo *repository.JobRepository,
	unitRepo *repository.UnitRepository,
) *service {
	return &service{
		cfg:          cfg,
		db:           db,
		jobRepo:      jobRepo,
		leaseRepo:    leaseRepo,
		logger:       logger,
		propertyRepo: propertyRepo,
		tenantRepo:   tenantRepo,
		unitRepo:     unitRepo,
	}
}

func isAdmin(actor *common.Actor) bool {
	return actor.Role == common.AdminRoleOwner.ToString() || actor.Role == common.AdminRoleManager.ToString()
}

// findAccessibleTenant loads a tenant the actor may see. Tenants can only see
// themselves; admins can see tenants on properties they own or manage.
func (this *service) findAccessibleTenant(actor *common.Actor, id uint, logger *slog.Logger) (*model.Tenant, *common.ServiceResponse) {
	if actor.Role == common.ActorRoleTenant && actor.ID != id {
		return nil, &common.ServiceResponse{StatusCode: 404, Err: errNotFound}
	}
	if actor.Role != common.ActorRoleTenant && !isAdmin(actor) {
		return nil, &common.ServiceResponse{StatusCode: 403, Err: common.ErrInvalidCredentials}
	}

	tenant, err := this.tenantRepo.FindOneById(id)
	if err != nil {
		logger.Error("failed to find tenant by id", "err", err.Error())
		return nil, &common.ServiceResponse{StatusCode: 500, Err: common.ErrGeneric}
	}
	if tenant == nil {
		return nil, &common.ServiceResponse{StatusCode: 404, Err: errNotFound}
	}

	if isAdmin(actor) {
		property, err := this.propertyRepo.FindOneById(uint(tenant.PropertyID))
		if err != nil {
			logger.Error("failed to find property by id", "err", err.Error())
			return nil, &common.ServiceResponse{StatusCode: 500, Err: common.ErrGeneric}
		}
		if !property.AccessibleBy(actor) {
			return nil, &common.ServiceResponse{StatusCode: 404, Err: errNotFound}
		}
	}
	return tenant, nil
}

func (this *service) createTenant(actor *common.Actor, dto createTenantReq) common.ServiceResponse {
	logger := this.logger.With("caller", "tenant.service.createTenant", "actor", actor.Email)
	if !isAdmin(actor) {
		return common.ServiceResponse{StatusCode: 403, Err: common.ErrInvalidCredentials}
	}

	startDate, _ := time.Parse(dateLayout, dto.Lease.StartDate)
	endDate, _ := time.Parse(dateLayout, dto.Lease.EndDate)
	if !endDate.After(startDate) {
		return common.ServiceResponse{StatusCode: 400, Err: "lease end date must be after start date"}
	}

	property, err := this.propertyRepo.FindOneById(dto.PropertyID)
	if err != nil {
		logger.Error("failed to find property by id", "err", err.Error())
		return common.ServiceResponse{StatusCode: 500, Err: common.ErrGeneric}
	}
	if !property.AccessibleBy(actor) {
		return common.ServiceResponse{StatusCode: 404, Err: "property not found"}
	}

	existing, err := this.tenantRepo.FindOneByEmail(dto.Email)
	if err != nil {
		logger.Error("failed to find tenant by email", "err", err.Error())
		return common.ServiceResponse{StatusCode: 500, Err: common.ErrGeneric}
	}
	if existing != nil {
		return common.ServiceResponse{StatusCode: 409, Err: "email already taken"}
	}

	if property.ArchivedAt != nil {
		return common.ServiceResponse{StatusCode: 409, Err: "property is archived"}
	}
	unit, err := this.unitRepo.FindOneByPropertyAndNumber(property.ID, dto.UnitNumber)
	if err != nil {
		logger.Error("failed to find unit by number", "err", err.Error())
		return common.ServiceResponse{StatusCode: 500, Err: common.ErrGeneric}
	}
	if unit == nil {
		return common.ServiceResponse{StatusCode: 404, Err: "unit not found"}
	}

	occupied, err := this.leaseRepo.ExistsActiveForUnit(property.ID, dto.UnitNumber)
	if err != nil {
		logger.Error("failed to check unit occupancy", "err", err.Error())
		return common.ServiceResponse{StatusCode: 500, Err: common.ErrGeneric}
	}
	if occupied {
		return common.ServiceResponse{StatusCode: 409, Err: "unit already has an active lease"}
	}

	tenant := model.Tenant{
		Email:                 dto.Email,
		ExpectedAnnualRent:    dto.Lease.ExpectedRent,
		FirstName:             dto.FirstName,
		LastName:              dto.LastName,
		OutstandingAnnualRent: dto.Lease.ExpectedRent,
		PhoneNumber:           dto.PhoneNumber,
		PropertyID:            int(property.ID),
		UnitNumber:            dto.UnitNumber,
	}
	var inviteLink string
	err = this.db.Transaction(func(trx *gorm.DB) error {
		if err := this.tenantRepo.SaveOneTRX(&tenant, trx); err != nil {
			return err
		}
		lease := model.Lease{
			DurationType:    dto.Lease.DurationType,
			EndDate:         endDate,
			ExpectedRent:    dto.Lease.ExpectedRent,
			OutstandingRent: dto.Lease.ExpectedRent,
			PropertyID:      int(property.ID),
			StartDate:       startDate,
			Status:          model.LeaseStatusActive,
			TenantID:        int(tenant.ID),
			UnitNumber:      dto.UnitNumber,
		}
		if err := this.leaseRepo.SaveOneTRX(&lease, trx); err != nil {
			return err
		}
		if err := this.tenantRepo.UpdateLeaseIDTRX(tenant.ID, lease.ID, trx); err != nil {
			return err
		}
		tenant.LeaseID = int(lease.ID)
		if err := this.propertyRepo.IncrementOccupiedUnitsTRX(property.ID, trx); err != nil {
			return err
		}

		link, err := cpaas.QueueInviteEmail(this.cfg, this.jobRepo, trx, common.Actor{
			ID:        tenant.ID,
			Email:     tenant.Email,
			FirstName: tenant.FirstName,
			Role:      common.ActorRoleTenant,
		}, tenant.FirstName+" "+tenant.LastName, cpaas.EmailTypeTenantInvite, "You've been invited to set up your tenant account")
		inviteLink = link
		return err
	})
	if err != nil {
		logger.Error("failed to create tenant", "err", err.Error())
		return common.ServiceResponse{StatusCode: 500, Err: common.ErrGeneric}
	}

	if this.cfg.Env == "development" {
		logger.Debug("tenant invite link", "tenant_id", tenant.ID, "link", inviteLink)
	}
	return common.ServiceResponse{StatusCode: 201, Data: toTenantRes(tenant)}
}

func (this *service) getTenants(actor *common.Actor, dto getTenantsReq) common.ServiceResponse {
	logger := this.logger.With("caller", "tenant.service.getTenants", "actor", actor.Email)
	if !isAdmin(actor) {
		return common.ServiceResponse{StatusCode: 403, Err: common.ErrInvalidCredentials}
	}

	filter := repository.TenantFilter{PropertyID: dto.PropertyID}
	if actor.Role == common.AdminRoleOwner.ToString() {
		filter.OwnerID = actor.ID
	} else {
		filter.ManagerID = actor.ID
	}
	if dto.PropertyID != 0 {
		property, err := this.propertyRepo.FindOneById(dto.PropertyID)
		if err != nil {
			logger.Error("failed to find property by id", "err", err.Error())
			return common.ServiceResponse{StatusCode: 500, Err: common.ErrGeneric}
		}
		if !property.AccessibleBy(actor) {
			return common.ServiceResponse{StatusCode: 404, Err: "property not found"}
		}
	}

	page, limit := dto.Page, dto.Limit
	if page == 0 {
		page = 1
	}
	if limit == 0 {
		limit = defaultPageSize
	}
	tenants, total, err := this.tenantRepo.FindMany(filter, page, limit)
	if err != nil {
		logger.Error("failed to find tenants", "err", err.Error())
		return common.ServiceResponse{StatusCode: 500, Err: common.ErrGeneric}
	}

	items := make([]tenantRes, 0, len(tenants))
	for _, t := range tenants {
		items = append(items, toTenantRes(t))
	}
	return common.ServiceResponse{StatusCode: 200, Data: tenantListRes{
		Items: items,
		Limit: limit,
		Page:  page,
		Total: total,
	}}
}

func (this *service) getTenantInfo(actor *common.Actor, id uint) common.ServiceResponse {
	logger := this.logger.With("caller", "tenant.service.getTenantInfo", "actor", actor.Email)
	tenant, errRes := this.findAccessibleTenant(actor, id, logger)
	if errRes != nil {
		return *errRes
	}
	return common.ServiceResponse{StatusCode: 200, Data: toTenantRes(*tenant)}
}

func (this *service) updateTenantInfo(actor *common.Actor, id uint, dto updateTenantReq) common.ServiceResponse {
	logger := this.logger.With("caller", "tenant.service.updateTenantInfo", "actor", actor.Email)
	tenant, errRes := this.findAccessibleTenant(actor, id, logger)
	if errRes != nil {
		return *errRes
	}

	fields := map[string]any{}
	if dto.FirstName != nil {
		fields["first_name"] = *dto.FirstName
		tenant.FirstName = *dto.FirstName
	}
	if dto.LastName != nil {
		fields["last_name"] = *dto.LastName
		tenant.LastName = *dto.LastName
	}
	if dto.PhoneNumber != nil {
		fields["phone_number"] = *dto.PhoneNumber
		tenant.PhoneNumber = *dto.PhoneNumber
	}
	if len(fields) > 0 {
		if err := this.tenantRepo.UpdateProfile(tenant.ID, fields); err != nil {
			logger.Error("failed to update tenant", "err", err.Error())
			return common.ServiceResponse{StatusCode: 500, Err: common.ErrGeneric}
		}
	}
	return common.ServiceResponse{StatusCode: 200, Data: toTenantRes(*tenant)}
}
