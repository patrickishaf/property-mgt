package property

import (
	"log/slog"

	"github.com/patrickishaf/property-mgt/common"
	"github.com/patrickishaf/property-mgt/internal/db/model"
	"github.com/patrickishaf/property-mgt/providers/cpaas"
	"gorm.io/gorm"
)

// findOwnedManager loads a manager that belongs to the owner making the request.
func (this *service) findOwnedManager(actor *common.Actor, id uint, logger *slog.Logger) (*model.Admin, *common.ServiceResponse) {
	manager, err := this.adminRepo.FindOneById(id)
	if err != nil {
		logger.Error("failed to find manager by id", "err", err.Error())
		return nil, &common.ServiceResponse{StatusCode: 500, Err: common.ErrGeneric}
	}
	if manager == nil || manager.Role != common.AdminRoleManager.ToString() ||
		manager.OwnerID == nil || *manager.OwnerID != actor.ID {
		return nil, &common.ServiceResponse{StatusCode: 404, Err: errManagerNotFound}
	}
	return manager, nil
}

func (this *service) createPropertyManager(actor *common.Actor, dto createManagerReq) common.ServiceResponse {
	logger := this.logger.With("caller", "property.service.createPropertyManager", "actor", actor.Email)
	if !isOwner(actor) {
		return forbidden()
	}

	existing, err := this.adminRepo.FindOneByEmail(dto.Email)
	if err != nil {
		logger.Error("failed to find admin by email", "err", err.Error())
		return common.ServiceResponse{StatusCode: 500, Err: common.ErrGeneric}
	}
	if existing != nil {
		return common.ServiceResponse{StatusCode: 409, Err: "email already taken"}
	}

	ownerID := actor.ID
	manager := model.Admin{
		Email:       dto.Email,
		FirstName:   dto.FirstName,
		LastName:    dto.LastName,
		OwnerID:     &ownerID,
		PhoneNumber: dto.PhoneNumber,
		Role:        common.AdminRoleManager.ToString(),
	}
	var inviteLink string
	err = this.db.Transaction(func(trx *gorm.DB) error {
		if err := this.adminRepo.SaveOneTRX(&manager, trx); err != nil {
			return err
		}
		link, err := cpaas.QueueInviteEmail(this.cfg, this.jobRepo, trx, common.Actor{
			ID:        manager.ID,
			Email:     manager.Email,
			FirstName: manager.FirstName,
			Role:      manager.Role,
		}, manager.FirstName+" "+manager.LastName, cpaas.EmailTypeManagerInvite, "You've been invited to manage properties")
		inviteLink = link
		return err
	})
	if err != nil {
		logger.Error("failed to create manager", "err", err.Error())
		return common.ServiceResponse{StatusCode: 500, Err: common.ErrGeneric}
	}

	if this.cfg.Env == "development" {
		logger.Debug("manager invite link", "manager_id", manager.ID, "link", inviteLink)
	}
	return common.ServiceResponse{StatusCode: 201, Data: toManagerRes(manager)}
}

func (this *service) getPropertyManagers(actor *common.Actor, dto pageReq) common.ServiceResponse {
	logger := this.logger.With("caller", "property.service.getPropertyManagers", "actor", actor.Email)
	if !isOwner(actor) {
		return forbidden()
	}

	page, limit := dto.values()
	managers, total, err := this.adminRepo.FindManagersByOwner(actor.ID, page, limit)
	if err != nil {
		logger.Error("failed to find managers", "err", err.Error())
		return common.ServiceResponse{StatusCode: 500, Err: common.ErrGeneric}
	}
	items := make([]managerRes, 0, len(managers))
	for _, m := range managers {
		items = append(items, toManagerRes(m))
	}
	return common.ServiceResponse{StatusCode: 200, Data: listRes[managerRes]{
		Items: items,
		Limit: limit,
		Page:  page,
		Total: total,
	}}
}

func (this *service) getPropertyManager(actor *common.Actor, id uint) common.ServiceResponse {
	logger := this.logger.With("caller", "property.service.getPropertyManager", "actor", actor.Email)
	if actor.Role == common.AdminRoleManager.ToString() {
		if actor.ID != id {
			return common.ServiceResponse{StatusCode: 404, Err: errManagerNotFound}
		}
		manager, err := this.adminRepo.FindOneById(id)
		if err != nil {
			logger.Error("failed to find manager by id", "err", err.Error())
			return common.ServiceResponse{StatusCode: 500, Err: common.ErrGeneric}
		}
		if manager == nil {
			return common.ServiceResponse{StatusCode: 404, Err: errManagerNotFound}
		}
		return common.ServiceResponse{StatusCode: 200, Data: toManagerRes(*manager)}
	}
	if !isOwner(actor) {
		return forbidden()
	}
	manager, errRes := this.findOwnedManager(actor, id, logger)
	if errRes != nil {
		return *errRes
	}
	return common.ServiceResponse{StatusCode: 200, Data: toManagerRes(*manager)}
}

func (this *service) updatePropertyManager(actor *common.Actor, id uint, dto updateManagerReq) common.ServiceResponse {
	logger := this.logger.With("caller", "property.service.updatePropertyManager", "actor", actor.Email)
	if !isOwner(actor) {
		return forbidden()
	}
	manager, errRes := this.findOwnedManager(actor, id, logger)
	if errRes != nil {
		return *errRes
	}

	fields := map[string]any{}
	if dto.FirstName != nil {
		fields["first_name"] = *dto.FirstName
		manager.FirstName = *dto.FirstName
	}
	if dto.LastName != nil {
		fields["last_name"] = *dto.LastName
		manager.LastName = *dto.LastName
	}
	if dto.PhoneNumber != nil {
		fields["phone_number"] = *dto.PhoneNumber
		manager.PhoneNumber = *dto.PhoneNumber
	}
	if len(fields) > 0 {
		if err := this.adminRepo.UpdateProfile(manager.ID, fields); err != nil {
			logger.Error("failed to update manager", "err", err.Error())
			return common.ServiceResponse{StatusCode: 500, Err: common.ErrGeneric}
		}
	}
	return common.ServiceResponse{StatusCode: 200, Data: toManagerRes(*manager)}
}

func (this *service) removePropertyManager(actor *common.Actor, id uint) common.ServiceResponse {
	logger := this.logger.With("caller", "property.service.removePropertyManager", "actor", actor.Email)
	if !isOwner(actor) {
		return forbidden()
	}
	manager, errRes := this.findOwnedManager(actor, id, logger)
	if errRes != nil {
		return *errRes
	}

	err := this.db.Transaction(func(trx *gorm.DB) error {
		if err := this.propertyRepo.ClearManagerForAllTRX(manager.ID, trx); err != nil {
			return err
		}
		return this.adminRepo.DeleteTRX(manager.ID, trx)
	})
	if err != nil {
		logger.Error("failed to remove manager", "err", err.Error())
		return common.ServiceResponse{StatusCode: 500, Err: common.ErrGeneric}
	}
	return common.ServiceResponse{StatusCode: 200, Data: nil}
}
