package model

import (
	"time"

	"gorm.io/gorm"
)

type Job struct {
	gorm.Model
	AttemptCount    int        `json:"attempt_count"`
	DueDate         *time.Time `json:"due_date"`
	Error           *string    `json:"error"`
	LastScheduledAt time.Time  `json:"last_scheduled_at"`
	Payload         string     `json:"payload"`
	Status          string     `json:"status"`
	Type            string     `json:"type"`
}

type JobType string

const (
	JobTypeEmail            JobType = "EMAIL"
	JobTypePushNotification JobType = "PUSH_NOTIFICATION"
	JobTypeSMS              JobType = "SMS"
)

func (this JobType) String() string {
	return string(this)
}
