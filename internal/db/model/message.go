package model

import (
	"time"

	"gorm.io/gorm"
)

const (
	MessageChannelEmail = "email"
	MessageChannelInApp = "in_app"
	MessageChannelSMS   = "sms"

	// MessageStatusQueued means email/SMS jobs are waiting for the cpaas worker.
	MessageStatusQueued = "queued"
	// MessageStatusSent means every recipient has the message (in-app messages
	// are delivered as soon as they are saved).
	MessageStatusSent = "sent"

	DeliveryStatusDelivered = "delivered"
	DeliveryStatusQueued    = "queued"
)

// Message is a notice sent by an owner/manager (or the system) to tenants.
// Body keeps the text as written, with {name}/{unit}/{property} placeholders.
type Message struct {
	gorm.Model
	Body               string
	Channel            string
	DeliveredAt        *time.Time
	NumberOfRecipients int
	SenderID           *uint `gorm:"index"`
	SenderName         string
	SentAt             *time.Time
	Status             string
	Subject            string
	Template           string
}

// MessageRecipient is one tenant's copy of a message, with the contact details
// it was sent to and its personalised body.
type MessageRecipient struct {
	gorm.Model
	Body           string
	DeliveryStatus string
	Email          string
	FullName       string
	JobID          *uint
	MessageID      uint `gorm:"index"`
	PhoneNumber    string
	PropertyID     int
	PropertyName   string
	ReadAt         *time.Time
	TenantID       int `gorm:"index"`
	UnitNumber     string
}
