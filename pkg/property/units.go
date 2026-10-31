package property

import (
	"github.com/patrickishaf/property-mgt/common"
	"github.com/patrickishaf/property-mgt/internal/db/model"
	"gorm.io/gorm"
)

const errPropertyArchived = "property is archived"

func (this *service) getUnits(actor *common.Actor, propertyID uint) common.ServiceResponse {
	logger := this.logger.With("caller", "property.service.getUnits", "actor", actor.Email)
	property, errRes := this.findAccessibleProperty(actor, propertyID, false, logger)
	if errRes != nil {
		return *errRes
	}

	units, err := this.unitRepo.FindByProperty(property.ID)
	if err != nil {
		logger.Error("failed to find units", "err", err.Error())
		return common.ServiceResponse{StatusCode: 500, Err: common.ErrGeneric}
	}
	occupied, err := this.leaseRepo.ActiveUnitNumbers(property.ID)
	if err != nil {
		logger.Error("failed to find occupied units", "err", err.Error())
		return common.ServiceResponse{StatusCode: 500, Err: common.ErrGeneric}
	}
	items := make([]unitRes, 0, len(units))
	for _, u := range units {
		items = append(items, toUnitRes(u, occupied[u.Number]))
	}
	return common.ServiceResponse{StatusCode: 200, Data: items}
}

func (this *service) createUnit(actor *common.Actor, propertyID uint, dto createUnitReq) common.ServiceResponse {
	logger := this.logger.With("caller", "property.service.createUnit", "actor", actor.Email)
	property, errRes := this.findAccessibleProperty(actor, propertyID, false, logger)
	if errRes != nil {
		return *errRes
	}
	if property.ArchivedAt != nil {
		return common.ServiceResponse{StatusCode: 409, Err: errPropertyArchived}
	}

	existing, err := this.unitRepo.FindOneByPropertyAndNumber(property.ID, dto.Number)
	if err != nil {
		logger.Error("failed to find unit by number", "err", err.Error())
		return common.ServiceResponse{StatusCode: 500, Err: common.ErrGeneric}
	}
	if existing != nil {
		return common.ServiceResponse{StatusCode: 409, Err: "unit number already exists"}
	}

	units := []model.PropertyUnit{{
		Floor:      dto.Floor,
		Label:      dto.Label,
		Number:     dto.Number,
		PropertyID: int(property.ID),
	}}
	err = this.db.Transaction(func(trx *gorm.DB) error {
		if err := this.unitRepo.SaveManyTRX(units, trx); err != nil {
			return err
		}
		return this.propertyRepo.AddUnitsTRX(property.ID, 1, trx)
	})
	if err != nil {
		logger.Error("failed to create unit", "err", err.Error())
		return common.ServiceResponse{StatusCode: 500, Err: common.ErrGeneric}
	}
	return common.ServiceResponse{StatusCode: 201, Data: toUnitRes(units[0], false)}
}

func (this *service) updateUnit(actor *common.Actor, propertyID uint, unitID uint, dto updateUnitReq) common.ServiceResponse {
	logger := this.logger.With("caller", "property.service.updateUnit", "actor", actor.Email)
	property, errRes := this.findAccessibleProperty(actor, propertyID, false, logger)
	if errRes != nil {
		return *errRes
	}
	unit, err := this.unitRepo.FindOneById(property.ID, unitID)
	if err != nil {
		logger.Error("failed to find unit", "err", err.Error())
		return common.ServiceResponse{StatusCode: 500, Err: common.ErrGeneric}
	}
	if unit == nil {
		return common.ServiceResponse{StatusCode: 404, Err: errUnitNotFound}
	}

	fields := map[string]any{}
	if dto.Floor != nil {
		fields["floor"] = *dto.Floor
		unit.Floor = *dto.Floor
	}
	if dto.Label != nil {
		fields["label"] = *dto.Label
		unit.Label = *dto.Label
	}
	if len(fields) > 0 {
		if err := this.unitRepo.Update(unit.ID, fields); err != nil {
			logger.Error("failed to update unit", "err", err.Error())
			return common.ServiceResponse{StatusCode: 500, Err: common.ErrGeneric}
		}
	}
	occupied, err := this.leaseRepo.ExistsActiveForUnit(property.ID, unit.Number)
	if err != nil {
		logger.Error("failed to check unit occupancy", "err", err.Error())
		return common.ServiceResponse{StatusCode: 500, Err: common.ErrGeneric}
	}
	return common.ServiceResponse{StatusCode: 200, Data: toUnitRes(*unit, occupied)}
}

func (this *service) deleteUnit(actor *common.Actor, propertyID uint, unitID uint) common.ServiceResponse {
	logger := this.logger.With("caller", "property.service.deleteUnit", "actor", actor.Email)
	property, errRes := this.findAccessibleProperty(actor, propertyID, false, logger)
	if errRes != nil {
		return *errRes
	}
	unit, err := this.unitRepo.FindOneById(property.ID, unitID)
	if err != nil {
		logger.Error("failed to find unit", "err", err.Error())
		return common.ServiceResponse{StatusCode: 500, Err: common.ErrGeneric}
	}
	if unit == nil {
		return common.ServiceResponse{StatusCode: 404, Err: errUnitNotFound}
	}
	occupied, err := this.leaseRepo.ExistsActiveForUnit(property.ID, unit.Number)
	if err != nil {
		logger.Error("failed to check unit occupancy", "err", err.Error())
		return common.ServiceResponse{StatusCode: 500, Err: common.ErrGeneric}
	}
	if occupied {
		return common.ServiceResponse{StatusCode: 409, Err: "unit is occupied"}
	}

	err = this.db.Transaction(func(trx *gorm.DB) error {
		if err := this.unitRepo.DeleteTRX(unit.ID, trx); err != nil {
			return err
		}
		return this.propertyRepo.AddUnitsTRX(property.ID, -1, trx)
	})
	if err != nil {
		logger.Error("failed to delete unit", "err", err.Error())
		return common.ServiceResponse{StatusCode: 500, Err: common.ErrGeneric}
	}
	return common.ServiceResponse{StatusCode: 200, Data: nil}
}
