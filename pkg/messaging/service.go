package messaging

import (
	"fmt"
	"log/slog"
	"sort"
	"strings"
	"time"

	"github.com/patrickishaf/property-mgt/common"
	"github.com/patrickishaf/property-mgt/config"
	"github.com/patrickishaf/property-mgt/internal/db/model"
	"github.com/patrickishaf/property-mgt/internal/db/repository"
	"github.com/patrickishaf/property-mgt/providers/cpaas"
	"gorm.io/gorm"
)

const (
	dateLayout      = "2006-01-02"
	errNotFound     = "message not found"
	systemSender    = "System"
	errInvalidRecip = "invalid recipients"
)

type service struct {
	cfg         *config.AppConfig
	db          *gorm.DB
	jobRepo     *repository.JobRepository
	leaseRepo   *repository.LeaseRepository
	logger      *slog.Logger
	messageRepo *repository.MessageRepository
}

func CreateService(
	cfg *config.AppConfig,
	logger *slog.Logger,
	db *gorm.DB,
	messageRepo *repository.MessageRepository,
	leaseRepo *repository.LeaseRepository,
	jobRepo *repository.JobRepository,
) *service {
	return &service{
		cfg:         cfg,
		db:          db,
		jobRepo:     jobRepo,
		leaseRepo:   leaseRepo,
		logger:      logger,
		messageRepo: messageRepo,
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

func (this *service) getMessageRecipients(actor *common.Actor, dto getRecipientsReq) common.ServiceResponse {
	logger := this.logger.With("caller", "messaging.service.getMessageRecipients", "actor", actor.Email)
	if !isAdmin(actor) {
		return forbidden()
	}
	ownerID, managerID := adminScope(actor)
	contacts, err := this.leaseRepo.FindActiveTenantContacts(ownerID, managerID, dto.PropertyID, time.Now())
	if err != nil {
		logger.Error("failed to find tenant contacts", "err", err.Error())
		return serverError()
	}
	items := make([]recipientOptionRes, 0, len(contacts))
	for _, c := range contacts {
		items = append(items, recipientOptionRes{
			ID:                c.TenantID,
			Email:             c.Email,
			FullName:          c.FirstName + " " + c.LastName,
			HasOverdueBalance: c.HasOverdueBalance,
			PhoneNumber:       c.PhoneNumber,
			PropertyID:        c.PropertyID,
			PropertyName:      c.PropertyName,
			UnitNumber:        c.UnitNumber,
		})
	}
	return common.ServiceResponse{StatusCode: 200, Data: items}
}

func (this *service) getTemplates(actor *common.Actor) common.ServiceResponse {
	if !isAdmin(actor) {
		return forbidden()
	}
	return common.ServiceResponse{StatusCode: 200, Data: templatesRes{
		Placeholders: []string{placeholderName, placeholderUnit, placeholderProperty},
		Templates:    templates,
	}}
}

// checkSendRequest validates a send request on its own and returns the subject
// to use. It doesn't check whether the caller may message the tenants.
func checkSendRequest(dto sendMessageReq) (string, *common.ServiceResponse) {
	subject := strings.TrimSpace(dto.Subject)
	if dto.Template != "" {
		tmpl, ok := findTemplate(dto.Template)
		if !ok {
			return "", &common.ServiceResponse{StatusCode: 400, Err: "unknown template"}
		}
		if subject == "" {
			subject = tmpl.Subject
		}
	}
	if strings.TrimSpace(dto.Body) == "" {
		return "", &common.ServiceResponse{StatusCode: 400, Err: common.ErrInvalidRequest}
	}
	if dto.Channel == model.MessageChannelEmail && subject == "" {
		return "", &common.ServiceResponse{StatusCode: 400, Err: "subject is required for email without a template"}
	}

	seen := make(map[uint]bool, len(dto.Recipients))
	for _, r := range dto.Recipients {
		if seen[r.TenantID] {
			return "", &common.ServiceResponse{StatusCode: 400, Err: "duplicate recipient"}
		}
		seen[r.TenantID] = true
		if dto.Channel == model.MessageChannelEmail && strings.TrimSpace(r.Email) == "" {
			return "", &common.ServiceResponse{StatusCode: 400, Err: "email is required for every recipient when sending by email"}
		}
		if dto.Channel == model.MessageChannelSMS && strings.TrimSpace(r.PhoneNumber) == "" {
			return "", &common.ServiceResponse{StatusCode: 400, Err: "phone number is required for every recipient when sending by SMS"}
		}
	}
	return subject, nil
}

// Sender identifies who a message is from; a nil ID is the system.
type Sender struct {
	ID   *uint
	Name string
}

// Recipient is a tenant a message is sent to, with the contact details and
// property/unit used for delivery and personalisation.
type Recipient struct {
	Email        string
	FullName     string
	PhoneNumber  string
	PropertyID   uint
	PropertyName string
	TenantID     uint
	UnitNumber   string
}

// Send sends a message on one channel for another module (e.g. maintenance),
// sharing personalisation and job queueing with the HTTP endpoint. Callers are
// responsible for choosing recipients they may message. It returns the new
// message's ID.
func (this *service) Send(from Sender, channel string, templateKey string, subject string, body string, recipients []Recipient) (uint, error) {
	copies := make([]model.MessageRecipient, 0, len(recipients))
	for _, r := range recipients {
		copies = append(copies, model.MessageRecipient{
			Body:         personalise(body, r.FullName, r.UnitNumber, r.PropertyName),
			Email:        r.Email,
			FullName:     r.FullName,
			PhoneNumber:  r.PhoneNumber,
			PropertyID:   int(r.PropertyID),
			PropertyName: r.PropertyName,
			TenantID:     int(r.TenantID),
			UnitNumber:   r.UnitNumber,
		})
	}
	message, err := this.send(from, channel, templateKey, subject, body, copies)
	if err != nil {
		return 0, err
	}
	return message.ID, nil
}

// send saves a message and its recipients and, for SMS and email, queues one
// job per recipient, all in one transaction. Recipients must already be
// validated. Other modules (e.g. maintenance) can reuse it with a system sender.
func (this *service) send(from Sender, channel string, templateKey string, subject string, body string, recipients []model.MessageRecipient) (*model.Message, error) {
	now := time.Now()
	message := model.Message{
		Body:               body,
		Channel:            channel,
		NumberOfRecipients: len(recipients),
		SenderID:           from.ID,
		SenderName:         from.Name,
		SentAt:             &now,
		Status:             model.MessageStatusQueued,
		Subject:            subject,
		Template:           templateKey,
	}
	deliveryStatus := model.DeliveryStatusQueued
	if channel == model.MessageChannelInApp {
		message.Status = model.MessageStatusSent
		message.DeliveredAt = &now
		deliveryStatus = model.DeliveryStatusDelivered
	}

	err := this.db.Transaction(func(trx *gorm.DB) error {
		if err := this.messageRepo.SaveOneTRX(&message, trx); err != nil {
			return err
		}
		for i := range recipients {
			recipients[i].MessageID = message.ID
			recipients[i].DeliveryStatus = deliveryStatus
		}
		if err := this.messageRepo.SaveRecipientsTRX(recipients, trx); err != nil {
			return err
		}
		if channel == model.MessageChannelInApp {
			return nil
		}
		for _, r := range recipients {
			var job *model.Job
			var err error
			if channel == model.MessageChannelEmail {
				job, err = cpaas.QueueEmail(this.jobRepo, trx, cpaas.EmailJob{
					MessageRecipientID: r.ID,
					Payload: map[string]string{
						"body":    r.Body,
						"name":    r.FullName,
						"subject": personalise(subject, r.FullName, r.UnitNumber, r.PropertyName),
					},
					Recipients: []cpaas.EmailRecipient{{Address: r.Email, Name: r.FullName}},
					Subject:    personalise(subject, r.FullName, r.UnitNumber, r.PropertyName),
					Type:       cpaas.EmailTypeGeneralNotice,
				})
			} else {
				job, err = cpaas.QueueSMS(this.jobRepo, trx, cpaas.SMSJob{
					Body:               r.Body,
					MessageRecipientID: r.ID,
					PhoneNumber:        r.PhoneNumber,
				})
			}
			if err != nil {
				return err
			}
			if err := this.messageRepo.UpdateRecipientJobTRX(r.ID, job.ID, trx); err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return &message, nil
}

func (this *service) sendMessage(actor *common.Actor, dto sendMessageReq) common.ServiceResponse {
	logger := this.logger.With("caller", "messaging.service.sendMessage", "actor", actor.Email)
	if !isAdmin(actor) {
		return forbidden()
	}
	subject, errRes := checkSendRequest(dto)
	if errRes != nil {
		return *errRes
	}

	ownerID, managerID := adminScope(actor)
	contacts, err := this.leaseRepo.FindActiveTenantContacts(ownerID, managerID, 0, time.Now())
	if err != nil {
		logger.Error("failed to find tenant contacts", "err", err.Error())
		return serverError()
	}
	known := make(map[uint]repository.TenantContact, len(contacts))
	for _, c := range contacts {
		known[c.TenantID] = c
	}

	var invalid []string
	recipients := make([]model.MessageRecipient, 0, len(dto.Recipients))
	for _, r := range dto.Recipients {
		contact, ok := known[r.TenantID]
		if !ok {
			invalid = append(invalid, fmt.Sprint(r.TenantID))
			continue
		}
		// Contact details come from the request; property and unit fall back to
		// the tenant's active lease when the UI doesn't send them.
		propertyID, propertyName, unit := r.PropertyID, r.PropertyName, r.UnitNumber
		if propertyID == 0 {
			propertyID = contact.PropertyID
		}
		if propertyName == "" {
			propertyName = contact.PropertyName
		}
		if unit == "" {
			unit = contact.UnitNumber
		}
		recipients = append(recipients, model.MessageRecipient{
			Body:         personalise(dto.Body, r.FullName, unit, propertyName),
			Email:        strings.TrimSpace(r.Email),
			FullName:     r.FullName,
			PhoneNumber:  strings.TrimSpace(r.PhoneNumber),
			PropertyID:   int(propertyID),
			PropertyName: propertyName,
			TenantID:     int(r.TenantID),
			UnitNumber:   unit,
		})
	}
	if len(invalid) > 0 {
		sort.Strings(invalid)
		return common.ServiceResponse{StatusCode: 400, Err: errInvalidRecip + ": " + strings.Join(invalid, ", ")}
	}

	senderID := actor.ID
	message, err := this.send(
		Sender{
			ID:   &senderID,
			Name: actor.FirstName + " " + actor.LastName,
		},
		dto.Channel,
		dto.Template,
		subject,
		dto.Body,
		recipients,
	)
	if err != nil {
		logger.Error("failed to send message", "err", err.Error())
		return serverError()
	}
	return this.messageDetail(message, 201, logger)
}

func (this *service) messageDetail(message *model.Message, statusCode int, logger *slog.Logger) common.ServiceResponse {
	recipients, err := this.messageRepo.FindRecipients(message.ID)
	if err != nil {
		logger.Error("failed to find recipients", "err", err.Error())
		return serverError()
	}
	res := messageDetailRes{messageRes: toMessageRes(*message), Recipients: make([]recipientRes, 0, len(recipients))}
	for _, r := range recipients {
		res.Recipients = append(res.Recipients, toRecipientRes(r))
	}
	return common.ServiceResponse{StatusCode: statusCode, Data: res}
}

func (this *service) getMessages(actor *common.Actor, dto getMessagesReq) common.ServiceResponse {
	logger := this.logger.With("caller", "messaging.service.getMessages", "actor", actor.Email)
	filter := repository.MessageFilter{
		Channel:  dto.Channel,
		From:     parseDate(dto.From),
		Template: dto.Template,
	}
	if to := parseDate(dto.To); to != nil {
		end := to.AddDate(0, 0, 1)
		filter.To = &end
	}
	page, limit := dto.values()

	if actor.Role == common.ActorRoleTenant {
		items, total, err := this.messageRepo.FindInbox(actor.ID, dto.Unread, filter, page, limit)
		if err != nil {
			logger.Error("failed to find inbox", "err", err.Error())
			return serverError()
		}
		res := make([]inboxRes, 0, len(items))
		for _, i := range items {
			res = append(res, inboxRes{
				ID:         i.MessageID,
				Body:       i.Body,
				Channel:    i.Channel,
				ReadAt:     i.ReadAt,
				SenderName: i.SenderName,
				SentAt:     i.SentAt,
				Subject:    personalise(i.Subject, i.FullName, i.UnitNumber, i.PropertyName),
				Template:   i.Template,
			})
		}
		return common.ServiceResponse{StatusCode: 200, Data: listRes[inboxRes]{Items: res, Limit: limit, Page: page, Total: total}}
	}
	if !isAdmin(actor) {
		return forbidden()
	}

	ownerID, managerID := adminScope(actor)
	messages, total, err := this.messageRepo.FindSent(ownerID, managerID, filter, page, limit)
	if err != nil {
		logger.Error("failed to find messages", "err", err.Error())
		return serverError()
	}
	res := make([]messageRes, 0, len(messages))
	for _, m := range messages {
		res = append(res, toMessageRes(m))
	}
	return common.ServiceResponse{StatusCode: 200, Data: listRes[messageRes]{Items: res, Limit: limit, Page: page, Total: total}}
}

func (this *service) getMessageInfo(actor *common.Actor, id uint) common.ServiceResponse {
	logger := this.logger.With("caller", "messaging.service.getMessageInfo", "actor", actor.Email)
	message, err := this.messageRepo.FindOneById(id)
	if err != nil {
		logger.Error("failed to find message", "err", err.Error())
		return serverError()
	}
	if message == nil {
		return common.ServiceResponse{StatusCode: 404, Err: errNotFound}
	}

	if actor.Role == common.ActorRoleTenant {
		recipient, err := this.messageRepo.FindRecipient(message.ID, actor.ID)
		if err != nil {
			logger.Error("failed to find recipient", "err", err.Error())
			return serverError()
		}
		if recipient == nil {
			return common.ServiceResponse{StatusCode: 404, Err: errNotFound}
		}
		if recipient.ReadAt == nil {
			now := time.Now()
			if err := this.messageRepo.MarkRead(recipient.ID, now); err != nil {
				logger.Error("failed to mark message read", "err", err.Error())
				return serverError()
			}
			recipient.ReadAt = &now
		}
		return common.ServiceResponse{StatusCode: 200, Data: inboxRes{
			ID:         message.ID,
			Body:       recipient.Body,
			Channel:    message.Channel,
			ReadAt:     recipient.ReadAt,
			SenderName: message.SenderName,
			SentAt:     message.SentAt,
			Subject:    personalise(message.Subject, recipient.FullName, recipient.UnitNumber, recipient.PropertyName),
			Template:   message.Template,
		}}
	}
	if !isAdmin(actor) {
		return forbidden()
	}

	ownerID, managerID := adminScope(actor)
	visible, err := this.messageRepo.IsVisibleToAdmin(message.ID, ownerID, managerID)
	if err != nil {
		logger.Error("failed to check message access", "err", err.Error())
		return serverError()
	}
	if !visible {
		return common.ServiceResponse{StatusCode: 404, Err: errNotFound}
	}
	return this.messageDetail(message, 200, logger)
}
