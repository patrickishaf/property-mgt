package model

import (
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Tenant struct {
	gorm.Model
	AnnualRent    float64
	DueDayOfMonth int
	Email         string
	LeaseStart    string
	MonthlyRent   float64
	Name          string
	Phone         string
	PropertyID    uuid.UUID
	UnitID        uuid.UUID
}
