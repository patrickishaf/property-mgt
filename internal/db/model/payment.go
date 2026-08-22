package model

import (
	"time"

	"github.com/google/uuid"
)

type Payment struct {
	id         uuid.UUID
	Amount     float64
	CreatedAt  time.Time
	Date       time.Time
	Method     string
	RecordedBy string
	Reference  string
}
