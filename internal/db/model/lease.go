package model

import (
	"time"

	"gorm.io/gorm"
)

type Lease struct {
	gorm.Model
	AmountPaid      float64
	DurationType    string // annual or monthly
	EndDate         time.Time
	ExpectedRent    float64
	OutstandingRent float64
	PropertyID      int
	StartDate       time.Time
	Status          string
	TenantID        int
	UnitNumber      string
}
