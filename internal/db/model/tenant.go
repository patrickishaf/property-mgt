package model

import (
	"time"

	"github.com/google/uuid"
)

type Tenant struct {
	ID            uuid.UUID
	AnnualRent    float64
	CreatedAt     time.Time
	DueDayOfMonth int
	Email         string
	LeaseStart    string
	MonthlyRent   float64
	Name          string
	Phone         string
	PropertyID    uuid.UUID
	UnitID        uuid.UUID
}
