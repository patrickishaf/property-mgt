package repository

import (
	"errors"
	"log/slog"
	"time"

	"github.com/patrickishaf/property-mgt/internal/db/model"
	"gorm.io/gorm"
)

type MessageRepository struct {
	db     *gorm.DB
	logger *slog.Logger
}

// MessageFilter narrows message listings. Zero values are ignored.
type MessageFilter struct {
	Channel  string
	From     *time.Time
	Template string
	To       *time.Time
}

// InboxItem is a tenant's copy of a message.
type InboxItem struct {
	Body         string
	Channel      string
	FullName     string
	MessageID    uint
	PropertyName string
	ReadAt       *time.Time
	RecipientID  uint
	SenderName   string
	SentAt       *time.Time
	Subject      string
	Template     string
	UnitNumber   string
}

func CreateMessageRepository(db *gorm.DB, logger *slog.Logger) *MessageRepository {
	return &MessageRepository{
		db:     db,
		logger: logger,
	}
}

func (this *MessageRepository) SaveOneTRX(message *model.Message, trx *gorm.DB) error {
	return trx.Create(message).Error
}

func (this *MessageRepository) SaveRecipientsTRX(recipients []model.MessageRecipient, trx *gorm.DB) error {
	if len(recipients) == 0 {
		return nil
	}
	return trx.Create(&recipients).Error
}

func (this *MessageRepository) UpdateRecipientJobTRX(id uint, jobID uint, trx *gorm.DB) error {
	return trx.Model(&model.MessageRecipient{}).Where("id = ?", id).Update("job_id", jobID).Error
}

func (this *MessageRepository) FindOneById(id uint) (*model.Message, error) {
	var m model.Message
	err := this.db.Where("id = ?", id).First(&m).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &m, nil
}

// senderScope limits messages to those sent by the admin or, for an owner,
// by the managers they created.
func (this *MessageRepository) senderScope(query *gorm.DB, ownerID uint, managerID uint) *gorm.DB {
	if ownerID != 0 {
		return query.Where("sender_id = ? OR sender_id IN (?)", ownerID,
			this.db.Model(&model.Admin{}).Select("id").Where("owner_id = ?", ownerID))
	}
	return query.Where("sender_id = ?", managerID)
}

// IsVisibleToAdmin reports whether the message was sent by the admin or, for an
// owner, by one of their managers.
func (this *MessageRepository) IsVisibleToAdmin(messageID uint, ownerID uint, managerID uint) (bool, error) {
	var count int64
	query := this.senderScope(this.db.Model(&model.Message{}).Where("id = ?", messageID), ownerID, managerID)
	err := query.Count(&count).Error
	return count > 0, err
}

func applyMessageFilter(query *gorm.DB, prefix string, filter MessageFilter) *gorm.DB {
	if filter.Channel != "" {
		query = query.Where(prefix+"channel = ?", filter.Channel)
	}
	if filter.Template != "" {
		query = query.Where(prefix+"template = ?", filter.Template)
	}
	if filter.From != nil {
		query = query.Where(prefix+"sent_at >= ?", *filter.From)
	}
	if filter.To != nil {
		query = query.Where(prefix+"sent_at < ?", *filter.To)
	}
	return query
}

// FindSent lists messages visible to an admin, newest first.
func (this *MessageRepository) FindSent(ownerID uint, managerID uint, filter MessageFilter, page int, limit int) ([]model.Message, int64, error) {
	query := this.senderScope(this.db.Model(&model.Message{}), ownerID, managerID)
	query = applyMessageFilter(query, "", filter).Session(&gorm.Session{})

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var messages []model.Message
	err := query.Order("id DESC").Offset((page - 1) * limit).Limit(limit).Find(&messages).Error
	if err != nil {
		return nil, 0, err
	}
	return messages, total, nil
}

// FindInbox lists a tenant's copies of messages, newest first.
func (this *MessageRepository) FindInbox(tenantID uint, unreadOnly bool, filter MessageFilter, page int, limit int) ([]InboxItem, int64, error) {
	query := this.db.Table("message_recipients AS r").
		Joins("JOIN messages m ON m.id = r.message_id AND m.deleted_at IS NULL").
		Where("r.tenant_id = ? AND r.deleted_at IS NULL", tenantID)
	if unreadOnly {
		query = query.Where("r.read_at IS NULL")
	}
	query = applyMessageFilter(query, "m.", filter).Session(&gorm.Session{})

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var items []InboxItem
	err := query.Select("r.id AS recipient_id, m.id AS message_id, m.subject, r.body, m.channel, m.template, m.sender_name, m.sent_at, r.read_at, r.full_name, r.unit_number, r.property_name").
		Order("m.id DESC").Offset((page - 1) * limit).Limit(limit).Scan(&items).Error
	if err != nil {
		return nil, 0, err
	}
	return items, total, nil
}

func (this *MessageRepository) FindRecipients(messageID uint) ([]model.MessageRecipient, error) {
	var recipients []model.MessageRecipient
	err := this.db.Where("message_id = ?", messageID).Order("id ASC").Find(&recipients).Error
	return recipients, err
}

func (this *MessageRepository) FindRecipient(messageID uint, tenantID uint) (*model.MessageRecipient, error) {
	var r model.MessageRecipient
	err := this.db.Where("message_id = ? AND tenant_id = ?", messageID, tenantID).First(&r).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &r, nil
}

// MarkRead sets the recipient's read time if it isn't set yet.
func (this *MessageRepository) MarkRead(recipientID uint, at time.Time) error {
	return this.db.Model(&model.MessageRecipient{}).
		Where("id = ? AND read_at IS NULL", recipientID).Update("read_at", at).Error
}
