package model

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Lease struct {
	gorm.Model
	Balance              float64
	DaysUntilDue         float64
	MonthlyRent          float64
	NextDueDate          time.Time
	PropertyID           uuid.UUID
	Status               string
	TenantID             uuid.UUID
	TotalPaidInThisCycle float64
	UnitID               uuid.UUID
}
