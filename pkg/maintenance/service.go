package maintenance

import (
	"log/slog"

	"github.com/patrickishaf/property-mgt/common"
	"github.com/patrickishaf/property-mgt/config"
	"github.com/patrickishaf/property-mgt/internal/db/model"
	"github.com/patrickishaf/property-mgt/internal/db/repository"
	"github.com/patrickishaf/property-mgt/pkg/messaging"
	"gorm.io/gorm"
)

const (
	errComplaintNotFound = "complaint not found"
	errIssueNotFound     = "issue not found"
)

// messenger sends messages to tenants; implemented by the messaging service.
type messenger interface {
	Send(from messaging.Sender, channel string, templateKey string, subject string, body string, recipients []messaging.Recipient) (uint, error)
}

type service struct {
	cfg                 *config.AppConfig
	complaintRepo       *repository.ComplaintRepository
	complaintUpdateRepo *repository.ComplaintUpdateRepository
	db                  *gorm.DB
	issueRepo           *repository.IssueRepository
	issueUpdateRepo     *repository.IssueUpdateRepository
	leaseRepo           *repository.LeaseRepository
	logger              *slog.Logger
	messenger           messenger
	propertyRepo        *repository.PropertyRepository
	tenantRepo          *repository.TenantRepository
}

func CreateService(
	cfg *config.AppConfig,
	logger *slog.Logger,
	db *gorm.DB,
	complaintRepo *repository.ComplaintRepository,
	complaintUpdateRepo *repository.ComplaintUpdateRepository,
	issueRepo *repository.IssueRepository,
	issueUpdateRepo *repository.IssueUpdateRepository,
	leaseRepo *repository.LeaseRepository,
	tenantRepo *repository.TenantRepository,
	propertyRepo *repository.PropertyRepository,
	messenger messenger,
) *service {
	return &service{
		cfg:                 cfg,
		complaintRepo:       complaintRepo,
		complaintUpdateRepo: complaintUpdateRepo,
		db:                  db,
		issueRepo:           issueRepo,
		issueUpdateRepo:     issueUpdateRepo,
		leaseRepo:           leaseRepo,
		logger:              logger,
		messenger:           messenger,
		propertyRepo:        propertyRepo,
		tenantRepo:          tenantRepo,
	}
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

func fullName(actor *common.Actor) string {
	return actor.FirstName + " " + actor.LastName
}

func (this *service) createComplaint(actor *common.Actor, dto createComplaintReq) common.ServiceResponse {
	logger := this.logger.With("caller", "maintenance.service.createComplaint", "actor", actor.Email)
	if actor.Role != common.ActorRoleTenant {
		return forbidden()
	}
	lease, err := this.leaseRepo.FindActiveByTenant(actor.ID)
	if err != nil {
		logger.Error("failed to find active lease", "err", err.Error())
		return serverError()
	}
	if lease == nil {
		return common.ServiceResponse{StatusCode: 404, Err: "no active lease"}
	}

	complaint := model.Complaint{
		Category:        dto.Category,
		CreatorFullName: fullName(actor),
		CreatorID:       int(actor.ID),
		Message:         dto.Message,
		PropertyID:      lease.PropertyID,
		Status:          model.MaintenanceStatusOpen,
		UnitNumber:      lease.UnitNumber,
		Urgency:         dto.Urgency,
	}
	if err := this.complaintRepo.SaveOne(&complaint); err != nil {
		logger.Error("failed to save complaint", "err", err.Error())
		return serverError()
	}
	return common.ServiceResponse{StatusCode: 201, Data: complaintDetailRes{complaintRes: toComplaintRes(complaint)}}
}

// findVisibleComplaint loads a complaint the actor may see: their own as a
// tenant, or one on a property they own or manage as an admin.
func (this *service) findVisibleComplaint(actor *common.Actor, id uint, logger *slog.Logger) (*model.Complaint, *common.ServiceResponse) {
	complaint, err := this.complaintRepo.FindOneById(id)
	if err != nil {
		logger.Error("failed to find complaint", "err", err.Error())
		res := serverError()
		return nil, &res
	}
	notFound := &common.ServiceResponse{StatusCode: 404, Err: errComplaintNotFound}
	if complaint == nil {
		return nil, notFound
	}
	switch {
	case actor.Role == common.ActorRoleTenant:
		if uint(complaint.CreatorID) != actor.ID {
			return nil, notFound
		}
	case isAdmin(actor):
		property, err := this.propertyRepo.FindOneById(uint(complaint.PropertyID))
		if err != nil {
			logger.Error("failed to find property", "err", err.Error())
			res := serverError()
			return nil, &res
		}
		if !property.AccessibleBy(actor) {
			return nil, notFound
		}
	default:
		res := forbidden()
		return nil, &res
	}
	return complaint, nil
}

func (this *service) getComplaints(actor *common.Actor, dto getComplaintsReq) common.ServiceResponse {
	logger := this.logger.With("caller", "maintenance.service.getComplaints", "actor", actor.Email)
	filter := repository.ComplaintFilter{
		Category:   dto.Category,
		Linked:     dto.Linked,
		PropertyID: dto.PropertyID,
		Status:     dto.Status,
		Urgency:    dto.Urgency,
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
	complaints, total, err := this.complaintRepo.FindMany(filter, page, limit)
	if err != nil {
		logger.Error("failed to find complaints", "err", err.Error())
		return serverError()
	}
	items := make([]complaintRes, 0, len(complaints))
	for _, c := range complaints {
		items = append(items, toComplaintRes(c))
	}
	return common.ServiceResponse{StatusCode: 200, Data: listRes[complaintRes]{Items: items, Limit: limit, Page: page, Total: total}}
}

func (this *service) getComplaintInfo(actor *common.Actor, id uint) common.ServiceResponse {
	logger := this.logger.With("caller", "maintenance.service.getComplaintInfo", "actor", actor.Email)
	complaint, errRes := this.findVisibleComplaint(actor, id, logger)
	if errRes != nil {
		return *errRes
	}
	res := complaintDetailRes{complaintRes: toComplaintRes(*complaint)}
	if complaint.IssueID != nil {
		issue, err := this.issueRepo.FindOneById(*complaint.IssueID)
		if err != nil {
			logger.Error("failed to find issue", "err", err.Error())
			return serverError()
		}
		if issue != nil {
			res.IssueTitle = issue.Title
		}
	}
	return common.ServiceResponse{StatusCode: 200, Data: res}
}

func (this *service) getComplaintUpdates(actor *common.Actor, id uint) common.ServiceResponse {
	logger := this.logger.With("caller", "maintenance.service.getComplaintUpdates", "actor", actor.Email)
	complaint, errRes := this.findVisibleComplaint(actor, id, logger)
	if errRes != nil {
		return *errRes
	}
	updates, err := this.complaintUpdateRepo.FindByComplaint(complaint.ID)
	if err != nil {
		logger.Error("failed to find complaint updates", "err", err.Error())
		return serverError()
	}
	items := make([]updateRes, 0, len(updates))
	for _, u := range updates {
		items = append(items, toComplaintUpdateRes(u))
	}
	return common.ServiceResponse{StatusCode: 200, Data: items}
}

func (this *service) getLinkedComplaints(actor *common.Actor, dto getLinkedComplaintsReq) common.ServiceResponse {
	logger := this.logger.With("caller", "maintenance.service.getLinkedComplaints", "actor", actor.Email)
	issue, _, errRes := this.findAdminIssue(actor, dto.IssueID, logger)
	if errRes != nil {
		return *errRes
	}
	complaints, err := this.complaintRepo.FindByIssue(issue.ID)
	if err != nil {
		logger.Error("failed to find linked complaints", "err", err.Error())
		return serverError()
	}
	items := make([]complaintRes, 0, len(complaints))
	for _, c := range complaints {
		items = append(items, toComplaintRes(c))
	}
	return common.ServiceResponse{StatusCode: 200, Data: items}
}
