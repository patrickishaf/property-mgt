package messaging

import (
	"time"

	"github.com/patrickishaf/property-mgt/internal/db/model"
)

const defaultPageSize = 10

type pageReq struct {
	Limit int `form:"limit" binding:"omitempty,min=1,max=100"`
	Page  int `form:"page" binding:"omitempty,min=1"`
}

func (this pageReq) values() (int, int) {
	page, limit := this.Page, this.Limit
	if page == 0 {
		page = 1
	}
	if limit == 0 {
		limit = defaultPageSize
	}
	return page, limit
}

type getRecipientsReq struct {
	PropertyID uint `form:"propertyId"`
}

type recipientReq struct {
	Email        string `json:"email" binding:"omitempty,email"`
	FullName     string `json:"fullName" binding:"required"`
	PhoneNumber  string `json:"phoneNumber"`
	PropertyID   uint   `json:"propertyId"`
	PropertyName string `json:"propertyName"`
	TenantID     uint   `json:"tenantId" binding:"required"`
	UnitNumber   string `json:"unitNumber"`
}

type sendMessageReq struct {
	Body       string         `json:"body" binding:"required"`
	Channel    string         `json:"channel" binding:"required,oneof=in_app sms email"`
	Recipients []recipientReq `json:"recipients" binding:"required,min=1,dive"`
	Subject    string         `json:"subject"`
	Template   string         `json:"template"`
}

type getMessagesReq struct {
	pageReq
	Channel  string `form:"channel" binding:"omitempty,oneof=in_app sms email"`
	From     string `form:"from" binding:"omitempty,datetime=2006-01-02"`
	Template string `form:"template"`
	To       string `form:"to" binding:"omitempty,datetime=2006-01-02"`
	Unread   bool   `form:"unread"`
}

type listRes[T any] struct {
	Items []T   `json:"items"`
	Limit int   `json:"limit"`
	Page  int   `json:"page"`
	Total int64 `json:"total"`
}

type recipientOptionRes struct {
	ID                uint   `json:"id"`
	Email             string `json:"email"`
	FullName          string `json:"fullName"`
	HasOverdueBalance bool   `json:"hasOverdueBalance"`
	PhoneNumber       string `json:"phoneNumber"`
	PropertyID        uint   `json:"propertyId"`
	PropertyName      string `json:"propertyName"`
	UnitNumber        string `json:"unitNumber"`
}

type templatesRes struct {
	Placeholders []string   `json:"placeholders"`
	Templates    []template `json:"templates"`
}

type messageRes struct {
	ID                 uint       `json:"id"`
	Body               string     `json:"body"`
	Channel            string     `json:"channel"`
	CreatedAt          time.Time  `json:"createdAt"`
	DeliveredAt        *time.Time `json:"deliveredAt"`
	NumberOfRecipients int        `json:"numberOfRecipients"`
	SenderName         string     `json:"senderName"`
	SentAt             *time.Time `json:"sentAt"`
	Status             string     `json:"status"`
	Subject            string     `json:"subject"`
	Template           string     `json:"template"`
}

type recipientRes struct {
	Body           string     `json:"body"`
	DeliveryStatus string     `json:"deliveryStatus"`
	Email          string     `json:"email"`
	FullName       string     `json:"fullName"`
	PhoneNumber    string     `json:"phoneNumber"`
	PropertyID     int        `json:"propertyId"`
	PropertyName   string     `json:"propertyName"`
	ReadAt         *time.Time `json:"readAt"`
	TenantID       int        `json:"tenantId"`
	UnitNumber     string     `json:"unitNumber"`
}

type messageDetailRes struct {
	messageRes
	Recipients []recipientRes `json:"recipients"`
}

// inboxRes is a tenant's copy of a message; Body is personalised for them.
type inboxRes struct {
	ID         uint       `json:"id"`
	Body       string     `json:"body"`
	Channel    string     `json:"channel"`
	ReadAt     *time.Time `json:"readAt"`
	SenderName string     `json:"senderName"`
	SentAt     *time.Time `json:"sentAt"`
	Subject    string     `json:"subject"`
	Template   string     `json:"template"`
}

func toMessageRes(m model.Message) messageRes {
	return messageRes{
		ID:                 m.ID,
		Body:               m.Body,
		Channel:            m.Channel,
		CreatedAt:          m.CreatedAt,
		DeliveredAt:        m.DeliveredAt,
		NumberOfRecipients: m.NumberOfRecipients,
		SenderName:         m.SenderName,
		SentAt:             m.SentAt,
		Status:             m.Status,
		Subject:            m.Subject,
		Template:           m.Template,
	}
}

func toRecipientRes(r model.MessageRecipient) recipientRes {
	return recipientRes{
		Body:           r.Body,
		DeliveryStatus: r.DeliveryStatus,
		Email:          r.Email,
		FullName:       r.FullName,
		PhoneNumber:    r.PhoneNumber,
		PropertyID:     r.PropertyID,
		PropertyName:   r.PropertyName,
		ReadAt:         r.ReadAt,
		TenantID:       r.TenantID,
		UnitNumber:     r.UnitNumber,
	}
}
