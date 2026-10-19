package model

import (
	"time"

	"gorm.io/gorm"
)

type Message struct {
	gorm.Model
	Body               string
	Channel            string
	DeliveredAt        *time.Time
	NumberOfRecipients int
	SentAt             *time.Time
	SenderName         string
	Subject            string
}
