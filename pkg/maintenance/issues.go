package maintenance

import (
	"errors"
	"fmt"
	"log/slog"
	"sort"
	"strings"

	"github.com/patrickishaf/property-mgt/common"
	"github.com/patrickishaf/property-mgt/internal/db/model"
	"github.com/patrickishaf/property-mgt/internal/db/repository"
	"github.com/patrickishaf/property-mgt/pkg/messaging"
	"gorm.io/gorm"
)

var statusLabels = map[string]string{
	model.MaintenanceStatusFixed:      "Fixed",
	model.MaintenanceStatusInProgress: "In progress",
	model.MaintenanceStatusOpen:       "Open",
}

// invalidComplaintsError lists complaints that can't be linked to an issue.
type invalidComplaintsError struct {
	ids []uint
}

func (this invalidComplaintsError) Error() string {
	parts := make([]string, 0, len(this.ids))
	for _, id := range this.ids {
		parts = append(parts, fmt.Sprint(id))
	}
	return "invalid complaints: " + strings.Join(parts, ", ")
}

// checkLinkableComplaints returns the requested IDs that can't be linked to an
// issue on the given property: unknown, on another property, or already linked
// (to any issue). Repeated IDs are reported once.
func checkLinkableComplaints(ids []uint, found []model.Complaint, propertyID int) []uint {
	byID := make(map[uint]model.Complaint, len(found))
	for _, c := range found {
		byID[c.ID] = c
	}
	seen := make(map[uint]bool, len(ids))
	var invalid []uint
	for _, id := range ids {
		if seen[id] {
			continue
		}
		seen[id] = true
		c, ok := byID[id]
		if !ok || c.PropertyID != propertyID || c.IssueID != nil {
			invalid = append(invalid, id)
		}
	}
	sort.Slice(invalid, func(i, j int) bool { return invalid[i] < invalid[j] })
	return invalid
}

// notifyChannels returns the channels to notify on: in-app when none are given,
// with duplicates removed and order kept.
func notifyChannels(requested []string) []string {
	if len(requested) == 0 {
		return []string{model.MessageChannelInApp}
	}
	seen := make(map[string]bool, len(requested))
	channels := make([]string, 0, len(requested))
	for _, c := range requested {
		if !seen[c] {
			seen[c] = true
			channels = append(channels, c)
		}
	}
	return channels
}

func uniqueIDs(ids []uint) []uint {
	seen := make(map[uint]bool, len(ids))
	out := make([]uint, 0, len(ids))
	for _, id := range ids {
		if !seen[id] {
			seen[id] = true
			out = append(out, id)
		}
	}
	return out
}

// findAdminIssue loads an issue on a property the admin owns or manages.
func (this *service) findAdminIssue(actor *common.Actor, id uint, logger *slog.Logger) (*model.Issue, *model.Property, *common.ServiceResponse) {
	if !isAdmin(actor) {
		res := forbidden()
		return nil, nil, &res
	}
	issue, err := this.issueRepo.FindOneById(id)
	if err != nil {
		logger.Error("failed to find issue", "err", err.Error())
		res := serverError()
		return nil, nil, &res
	}
	if issue == nil {
		return nil, nil, &common.ServiceResponse{StatusCode: 404, Err: errIssueNotFound}
	}
	property, err := this.propertyRepo.FindOneById(uint(issue.PropertyID))
	if err != nil {
		logger.Error("failed to find property", "err", err.Error())
		res := serverError()
		return nil, nil, &res
	}
	if !property.AccessibleBy(actor) {
		return nil, nil, &common.ServiceResponse{StatusCode: 404, Err: errIssueNotFound}
	}
	return issue, property, nil
}

// linkComplaintsTRX locks and validates the complaints, links them to the issue
// with its status, and adds an "under investigation" entry to each timeline.
func (this *service) linkComplaintsTRX(actor *common.Actor, issue *model.Issue, ids []uint, trx *gorm.DB) error {
	ids = uniqueIDs(ids)
	found, err := this.complaintRepo.FindByIdsForUpdateTRX(ids, trx)
	if err != nil {
		return err
	}
	if invalid := checkLinkableComplaints(ids, found, issue.PropertyID); len(invalid) > 0 {
		return invalidComplaintsError{ids: invalid}
	}
	if err := this.complaintRepo.LinkToIssueTRX(ids, issue.ID, issue.Status, trx); err != nil {
		return err
	}
	updates := make([]model.ComplaintUpdate, 0, len(ids))
	for _, id := range ids {
		updates = append(updates, model.ComplaintUpdate{
			ComplaintID:     id,
			CreatorFullName: fullName(actor),
			CreatorID:       int(actor.ID),
			Message:         fmt.Sprintf("Your complaint is being investigated as part of the issue %q.", issue.Title),
			Status:          issue.Status,
		})
	}
	return this.complaintUpdateRepo.SaveManyTRX(updates, trx)
}

func (this *service) issueDetail(issue *model.Issue, property *model.Property, statusCode int, logger *slog.Logger) common.ServiceResponse {
	counts, err := this.complaintRepo.CountByIssues([]uint{issue.ID})
	if err != nil {
		logger.Error("failed to count linked complaints", "err", err.Error())
		return serverError()
	}
	res := issueDetailRes{issueRes: toIssueRes(*issue, counts[issue.ID])}
	if property != nil {
		res.PropertyName = property.Name
	}
	return common.ServiceResponse{StatusCode: statusCode, Data: res}
}

func (this *service) createIssue(actor *common.Actor, dto createIssueReq) common.ServiceResponse {
	logger := this.logger.With("caller", "maintenance.service.createIssue", "actor", actor.Email)
	if !isAdmin(actor) {
		return forbidden()
	}
	property, err := this.propertyRepo.FindOneById(dto.PropertyID)
	if err != nil {
		logger.Error("failed to find property", "err", err.Error())
		return serverError()
	}
	if !property.AccessibleBy(actor) {
		return common.ServiceResponse{StatusCode: 404, Err: "property not found"}
	}

	issue := model.Issue{
		Category:        dto.Category,
		CreatorFullName: fullName(actor),
		CreatorID:       int(actor.ID),
		Message:         dto.Message,
		PropertyID:      int(property.ID),
		Status:          model.MaintenanceStatusOpen,
		Title:           dto.Title,
		Urgency:         dto.Urgency,
	}
	err = this.db.Transaction(func(trx *gorm.DB) error {
		if err := this.issueRepo.SaveOneTRX(&issue, trx); err != nil {
			return err
		}
		if len(dto.ComplaintIDs) == 0 {
			return nil
		}
		return this.linkComplaintsTRX(actor, &issue, dto.ComplaintIDs, trx)
	})
	var invalid invalidComplaintsError
	if errors.As(err, &invalid) {
		return common.ServiceResponse{StatusCode: 400, Err: invalid.Error()}
	}
	if err != nil {
		logger.Error("failed to create issue", "err", err.Error())
		return serverError()
	}
	return this.issueDetail(&issue, property, 201, logger)
}

func (this *service) linkComplaints(actor *common.Actor, issueID uint, dto linkComplaintsReq) common.ServiceResponse {
	logger := this.logger.With("caller", "maintenance.service.linkComplaints", "actor", actor.Email)
	issue, property, errRes := this.findAdminIssue(actor, issueID, logger)
	if errRes != nil {
		return *errRes
	}
	err := this.db.Transaction(func(trx *gorm.DB) error {
		locked, err := this.issueRepo.FindOneByIdForUpdateTRX(issue.ID, trx)
		if err != nil {
			return err
		}
		issue = locked
		return this.linkComplaintsTRX(actor, locked, dto.ComplaintIDs, trx)
	})
	var invalid invalidComplaintsError
	if errors.As(err, &invalid) {
		return common.ServiceResponse{StatusCode: 400, Err: invalid.Error()}
	}
	if err != nil {
		logger.Error("failed to link complaints", "err", err.Error())
		return serverError()
	}
	return this.issueDetail(issue, property, 200, logger)
}

func (this *service) postIssueUpdate(actor *common.Actor, issueID uint, dto postIssueUpdateReq) common.ServiceResponse {
	logger := this.logger.With("caller", "maintenance.service.postIssueUpdate", "actor", actor.Email)
	issue, property, errRes := this.findAdminIssue(actor, issueID, logger)
	if errRes != nil {
		return *errRes
	}

	var update model.IssueUpdate
	var complaints []model.Complaint
	err := this.db.Transaction(func(trx *gorm.DB) error {
		locked, err := this.issueRepo.FindOneByIdForUpdateTRX(issue.ID, trx)
		if err != nil {
			return err
		}
		status := locked.Status
		if dto.Status != "" {
			status = dto.Status
		}
		update = model.IssueUpdate{
			CreatorFullName: fullName(actor),
			CreatorID:       int(actor.ID),
			IssueID:         locked.ID,
			Message:         dto.Message,
			Status:          status,
		}
		if err := this.issueUpdateRepo.SaveOneTRX(&update, trx); err != nil {
			return err
		}
		if err := this.issueRepo.UpdateStatusTRX(locked.ID, status, trx); err != nil {
			return err
		}
		if err := this.complaintRepo.UpdateStatusByIssueTRX(locked.ID, status, trx); err != nil {
			return err
		}
		complaints, err = this.complaintRepo.FindByIssueTRX(locked.ID, trx)
		if err != nil {
			return err
		}
		mirrored := make([]model.ComplaintUpdate, 0, len(complaints))
		for _, c := range complaints {
			mirrored = append(mirrored, model.ComplaintUpdate{
				ComplaintID:     c.ID,
				CreatorFullName: update.CreatorFullName,
				CreatorID:       update.CreatorID,
				Message:         update.Message,
				Status:          status,
			})
		}
		issue = locked
		issue.Status = status
		return this.complaintUpdateRepo.SaveManyTRX(mirrored, trx)
	})
	if err != nil {
		logger.Error("failed to post issue update", "err", err.Error())
		return serverError()
	}

	notifications := this.notifyTenants(actor, issue, property, complaints, update, notifyChannels(dto.NotifyChannels), logger)
	return common.ServiceResponse{StatusCode: 201, Data: postIssueUpdateRes{
		Notifications: notifications,
		Update:        toIssueUpdateRes(update),
	}}
}

// notifyTenants messages each tenant with a linked complaint once per channel,
// using their stored contact details. Tenants without the contact detail a
// channel needs are skipped for it. Failures are reported, not returned, so a
// saved update is never undone by a notification problem.
func (this *service) notifyTenants(actor *common.Actor, issue *model.Issue, property *model.Property, complaints []model.Complaint, update model.IssueUpdate, channels []string, logger *slog.Logger) []notificationRes {
	var recipients []messaging.Recipient
	seen := make(map[int]bool, len(complaints))
	for _, c := range complaints {
		if seen[c.CreatorID] {
			continue
		}
		seen[c.CreatorID] = true
		tenant, err := this.tenantRepo.FindOneById(uint(c.CreatorID))
		if err != nil {
			logger.Error("failed to find tenant to notify", "tenant_id", c.CreatorID, "err", err.Error())
			continue
		}
		if tenant == nil {
			continue
		}
		recipients = append(recipients, messaging.Recipient{
			Email:        tenant.Email,
			FullName:     tenant.FirstName + " " + tenant.LastName,
			PhoneNumber:  tenant.PhoneNumber,
			PropertyID:   uint(issue.PropertyID),
			PropertyName: property.Name,
			TenantID:     tenant.ID,
			UnitNumber:   c.UnitNumber,
		})
	}

	subject := "Update on: " + issue.Title
	body := fmt.Sprintf("Dear {name}, there is an update on the maintenance issue %q at {property}: %s Status: %s.",
		issue.Title, update.Message, statusLabels[update.Status])
	senderID := actor.ID
	from := messaging.Sender{ID: &senderID, Name: fullName(actor)}

	results := make([]notificationRes, 0, len(channels))
	for _, channel := range channels {
		var reachable []messaging.Recipient
		for _, r := range recipients {
			if (channel == model.MessageChannelSMS && r.PhoneNumber == "") ||
				(channel == model.MessageChannelEmail && r.Email == "") {
				continue
			}
			reachable = append(reachable, r)
		}
		result := notificationRes{Channel: channel, Recipients: len(reachable), Skipped: len(recipients) - len(reachable)}
		if len(reachable) > 0 {
			messageID, err := this.messenger.Send(from, channel, "", subject, body, reachable)
			if err != nil {
				logger.Error("failed to notify tenants", "channel", channel, "err", err.Error())
				result.Error = "notification failed"
			} else {
				result.MessageID = &messageID
			}
		}
		results = append(results, result)
	}
	return results
}

func (this *service) getIssues(actor *common.Actor, dto getIssuesReq) common.ServiceResponse {
	logger := this.logger.With("caller", "maintenance.service.getIssues", "actor", actor.Email)
	if !isAdmin(actor) {
		return forbidden()
	}
	filter := repository.IssueFilter{
		Category:   dto.Category,
		PropertyID: dto.PropertyID,
		Status:     dto.Status,
		Urgency:    dto.Urgency,
	}
	filter.OwnerID, filter.ManagerID = adminScope(actor)
	page, limit := dto.values()
	issues, total, err := this.issueRepo.FindMany(filter, page, limit)
	if err != nil {
		logger.Error("failed to find issues", "err", err.Error())
		return serverError()
	}
	ids := make([]uint, 0, len(issues))
	for _, i := range issues {
		ids = append(ids, i.ID)
	}
	counts, err := this.complaintRepo.CountByIssues(ids)
	if err != nil {
		logger.Error("failed to count linked complaints", "err", err.Error())
		return serverError()
	}
	items := make([]issueRes, 0, len(issues))
	for _, i := range issues {
		items = append(items, toIssueRes(i, counts[i.ID]))
	}
	return common.ServiceResponse{StatusCode: 200, Data: listRes[issueRes]{Items: items, Limit: limit, Page: page, Total: total}}
}

func (this *service) getIssueInfo(actor *common.Actor, id uint) common.ServiceResponse {
	logger := this.logger.With("caller", "maintenance.service.getIssueInfo", "actor", actor.Email)
	issue, property, errRes := this.findAdminIssue(actor, id, logger)
	if errRes != nil {
		return *errRes
	}
	return this.issueDetail(issue, property, 200, logger)
}

func (this *service) getIssueUpdates(actor *common.Actor, id uint) common.ServiceResponse {
	logger := this.logger.With("caller", "maintenance.service.getIssueUpdates", "actor", actor.Email)
	issue, _, errRes := this.findAdminIssue(actor, id, logger)
	if errRes != nil {
		return *errRes
	}
	updates, err := this.issueUpdateRepo.FindByIssue(issue.ID)
	if err != nil {
		logger.Error("failed to find issue updates", "err", err.Error())
		return serverError()
	}
	items := make([]updateRes, 0, len(updates))
	for _, u := range updates {
		items = append(items, toIssueUpdateRes(u))
	}
	return common.ServiceResponse{StatusCode: 200, Data: items}
}
