package property

import (
	"log/slog"
	"time"

	"github.com/patrickishaf/property-mgt/common"
	"github.com/patrickishaf/property-mgt/config"
	"github.com/patrickishaf/property-mgt/internal/db/model"
	"github.com/patrickishaf/property-mgt/internal/db/repository"
	"gorm.io/gorm"
)

const (
	defaultPageSize    = 20
	errManagerNotFound = "manager not found"
	errNotFound        = "property not found"
	errUnitNotFound    = "unit not found"
)

type service struct {
	adminRepo    *repository.AdminRepository
	cfg          *config.AppConfig
	db           *gorm.DB
	jobRepo      *repository.JobRepository
	leaseRepo    *repository.LeaseRepository
	logger       *slog.Logger
	propertyRepo *repository.PropertyRepository
	unitRepo     *repository.UnitRepository
}

func CreateService(
	cfg *config.AppConfig,
	logger *slog.Logger,
	db *gorm.DB,
	propertyRepo *repository.PropertyRepository,
	unitRepo *repository.UnitRepository,
	leaseRepo *repository.LeaseRepository,
	adminRepo *repository.AdminRepository,
	jobRepo *repository.JobRepository,
) *service {
	return &service{
		adminRepo:    adminRepo,
		cfg:          cfg,
		db:           db,
		jobRepo:      jobRepo,
		leaseRepo:    leaseRepo,
		logger:       logger,
		propertyRepo: propertyRepo,
		unitRepo:     unitRepo,
	}
}

func isOwner(actor *common.Actor) bool {
	return actor.Role == common.AdminRoleOwner.ToString()
}

func isAdmin(actor *common.Actor) bool {
	return isOwner(actor) || actor.Role == common.AdminRoleManager.ToString()
}

func forbidden() common.ServiceResponse {
	return common.ServiceResponse{StatusCode: 403, Err: common.ErrInvalidCredentials}
}

// findAccessibleProperty loads a property the actor owns or manages. With
// ownerOnly, managers of the property are refused with 403.
func (this *service) findAccessibleProperty(actor *common.Actor, id uint, ownerOnly bool, logger *slog.Logger) (*model.Property, *common.ServiceResponse) {
	if !isAdmin(actor) {
		res := forbidden()
		return nil, &res
	}
	property, err := this.propertyRepo.FindOneById(id)
	if err != nil {
		logger.Error("failed to find property by id", "err", err.Error())
		return nil, &common.ServiceResponse{StatusCode: 500, Err: common.ErrGeneric}
	}
	if !property.AccessibleBy(actor) {
		return nil, &common.ServiceResponse{StatusCode: 404, Err: errNotFound}
	}
	if ownerOnly && !property.OwnedBy(actor) {
		res := forbidden()
		return nil, &res
	}
	return property, nil
}

func (this *service) propertyResponse(property *model.Property, statusCode int, logger *slog.Logger) common.ServiceResponse {
	totals, err := this.leaseRepo.RentTotalsByProperty([]uint{property.ID})
	if err != nil {
		logger.Error("failed to sum rent for property", "err", err.Error())
		return common.ServiceResponse{StatusCode: 500, Err: common.ErrGeneric}
	}
	return common.ServiceResponse{StatusCode: statusCode, Data: toPropertyRes(*property, totals[property.ID])}
}

func (this *service) createProperty(actor *common.Actor, dto createPropertyReq) common.ServiceResponse {
	logger := this.logger.With("caller", "property.service.createProperty", "actor", actor.Email)
	if !isOwner(actor) {
		return forbidden()
	}

	seen := make(map[string]bool, len(dto.Units))
	for _, u := range dto.Units {
		if seen[u.Number] {
			return common.ServiceResponse{StatusCode: 400, Err: "duplicate unit number"}
		}
		seen[u.Number] = true
	}

	property := model.Property{
		ActivityStatus: dto.ActivityStatus,
		City:           dto.City,
		Country:        dto.Country,
		Location:       dto.Location,
		Name:           dto.Name,
		NumberOfUnits:  len(dto.Units),
		OwnerID:        int(actor.ID),
		State:          dto.State,
		Type:           dto.Type,
	}
	err := this.db.Transaction(func(trx *gorm.DB) error {
		if err := this.propertyRepo.SaveOneTRX(&property, trx); err != nil {
			return err
		}
		units := make([]model.PropertyUnit, 0, len(dto.Units))
		for _, u := range dto.Units {
			units = append(units, model.PropertyUnit{
				Floor:      u.Floor,
				Label:      u.Label,
				Number:     u.Number,
				PropertyID: int(property.ID),
			})
		}
		return this.unitRepo.SaveManyTRX(units, trx)
	})
	if err != nil {
		logger.Error("failed to create property", "err", err.Error())
		return common.ServiceResponse{StatusCode: 500, Err: common.ErrGeneric}
	}
	return common.ServiceResponse{StatusCode: 201, Data: toPropertyRes(property, repository.RentTotals{})}
}

func (this *service) getProperties(actor *common.Actor, dto getPropertiesReq) common.ServiceResponse {
	logger := this.logger.With("caller", "property.service.getProperties", "actor", actor.Email)
	if !isAdmin(actor) {
		return forbidden()
	}

	filter := repository.PropertyFilter{IncludeArchived: dto.IncludeArchived}
	if isOwner(actor) {
		filter.OwnerID = actor.ID
	} else {
		filter.ManagerID = actor.ID
	}
	page, limit := dto.values()
	properties, total, err := this.propertyRepo.FindMany(filter, page, limit)
	if err != nil {
		logger.Error("failed to find properties", "err", err.Error())
		return common.ServiceResponse{StatusCode: 500, Err: common.ErrGeneric}
	}

	ids := make([]uint, 0, len(properties))
	for _, p := range properties {
		ids = append(ids, p.ID)
	}
	totals, err := this.leaseRepo.RentTotalsByProperty(ids)
	if err != nil {
		logger.Error("failed to sum rent for properties", "err", err.Error())
		return common.ServiceResponse{StatusCode: 500, Err: common.ErrGeneric}
	}
	items := make([]propertyRes, 0, len(properties))
	for _, p := range properties {
		items = append(items, toPropertyRes(p, totals[p.ID]))
	}
	return common.ServiceResponse{StatusCode: 200, Data: listRes[propertyRes]{
		Items: items,
		Limit: limit,
		Page:  page,
		Total: total,
	}}
}

func (this *service) getPropertyInfo(actor *common.Actor, id uint) common.ServiceResponse {
	logger := this.logger.With("caller", "property.service.getPropertyInfo", "actor", actor.Email)
	property, errRes := this.findAccessibleProperty(actor, id, false, logger)
	if errRes != nil {
		return *errRes
	}
	return this.propertyResponse(property, 200, logger)
}

func (this *service) updatePropertyInfo(actor *common.Actor, id uint, dto updatePropertyReq) common.ServiceResponse {
	logger := this.logger.With("caller", "property.service.updatePropertyInfo", "actor", actor.Email)
	property, errRes := this.findAccessibleProperty(actor, id, true, logger)
	if errRes != nil {
		return *errRes
	}

	fields := map[string]any{}
	set := func(column string, value *string, target *string) {
		if value != nil {
			fields[column] = *value
			*target = *value
		}
	}
	set("activity_status", dto.ActivityStatus, &property.ActivityStatus)
	set("city", dto.City, &property.City)
	set("country", dto.Country, &property.Country)
	set("location", dto.Location, &property.Location)
	set("name", dto.Name, &property.Name)
	set("state", dto.State, &property.State)
	set("type", dto.Type, &property.Type)
	if len(fields) > 0 {
		if err := this.propertyRepo.Update(property.ID, fields); err != nil {
			logger.Error("failed to update property", "err", err.Error())
			return common.ServiceResponse{StatusCode: 500, Err: common.ErrGeneric}
		}
	}
	return this.propertyResponse(property, 200, logger)
}

func (this *service) archiveProperty(actor *common.Actor, id uint) common.ServiceResponse {
	logger := this.logger.With("caller", "property.service.archiveProperty", "actor", actor.Email)
	property, errRes := this.findAccessibleProperty(actor, id, true, logger)
	if errRes != nil {
		return *errRes
	}
	if property.ArchivedAt != nil {
		return common.ServiceResponse{StatusCode: 409, Err: "property is already archived"}
	}
	hasActiveLeases, err := this.leaseRepo.ExistsActiveForProperty(property.ID)
	if err != nil {
		logger.Error("failed to check active leases", "err", err.Error())
		return common.ServiceResponse{StatusCode: 500, Err: common.ErrGeneric}
	}
	if hasActiveLeases {
		return common.ServiceResponse{StatusCode: 409, Err: "property has active leases"}
	}

	now := time.Now()
	if err := this.propertyRepo.Archive(property.ID, now); err != nil {
		logger.Error("failed to archive property", "err", err.Error())
		return common.ServiceResponse{StatusCode: 500, Err: common.ErrGeneric}
	}
	property.ArchivedAt = &now
	return this.propertyResponse(property, 200, logger)
}

func (this *service) assignPropertyManager(actor *common.Actor, id uint, dto assignManagerReq) common.ServiceResponse {
	logger := this.logger.With("caller", "property.service.assignPropertyManager", "actor", actor.Email)
	property, errRes := this.findAccessibleProperty(actor, id, true, logger)
	if errRes != nil {
		return *errRes
	}
	manager, errRes := this.findOwnedManager(actor, dto.ManagerID, logger)
	if errRes != nil {
		return *errRes
	}

	if err := this.propertyRepo.SetManager(property.ID, manager.ID); err != nil {
		logger.Error("failed to assign manager", "err", err.Error())
		return common.ServiceResponse{StatusCode: 500, Err: common.ErrGeneric}
	}
	property.ManagerID = int(manager.ID)
	return this.propertyResponse(property, 200, logger)
}

func (this *service) unassignPropertyManager(actor *common.Actor, id uint) common.ServiceResponse {
	logger := this.logger.With("caller", "property.service.unassignPropertyManager", "actor", actor.Email)
	property, errRes := this.findAccessibleProperty(actor, id, true, logger)
	if errRes != nil {
		return *errRes
	}
	if err := this.propertyRepo.SetManager(property.ID, 0); err != nil {
		logger.Error("failed to unassign manager", "err", err.Error())
		return common.ServiceResponse{StatusCode: 500, Err: common.ErrGeneric}
	}
	property.ManagerID = 0
	return this.propertyResponse(property, 200, logger)
}
