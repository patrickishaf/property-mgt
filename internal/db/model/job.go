package model

import (
	"time"

	"github.com/google/uuid"
)

type Job struct {
	ID              uuid.UUID  `json:"id"`
	CreatedAt       time.Time  `json:"created_at"`
	UpdatedAt       time.Time  `json:"updated_at"`
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
