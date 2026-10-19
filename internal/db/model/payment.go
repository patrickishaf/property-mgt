package model

import (
	"time"

	"gorm.io/gorm"
)

type Payment struct {
	gorm.Model
	Amount             float64
	DueDate            *time.Time
	Method             string
	PaidAt             time.Time
	PropertyID         int
	PropertyName       string
	PropertyUnitNumber string
	Reference          string
	Status             string
	TenantFullName     string
}
