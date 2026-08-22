package model

import (
	"time"

	"github.com/google/uuid"
)

type Lease struct {
	ID                   uuid.UUID
	Balance              float64
	CreatedAt            time.Time
	DaysUntilDue         float64
	MonthlyRent          float64
	NextDueDate          time.Time
	PropertyID           uuid.UUID
	Status               string
	TenantID             uuid.UUID
	TotalPaidInThisCycle float64
	unitID               uuid.UUID
}
