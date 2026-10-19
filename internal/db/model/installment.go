package model

import (
	"time"

	"gorm.io/gorm"
)

type Installment struct {
	gorm.Model
	AmountExpected    float64
	AmountPaid        float64
	DueDate           *time.Time
	LastPaymentDate   *time.Time
	LeaseID           int
	OutstandingAmount float64
	PropertyID        int
	Status            string
	TenantID          int
}
