package model

import (
	"time"

	"gorm.io/gorm"
)

const (
	InstallmentStatusPaid          = "paid"
	InstallmentStatusPartiallyPaid = "partially_paid"
	InstallmentStatusUnpaid        = "unpaid"
)

type Installment struct {
	gorm.Model
	AmountExpected    float64
	AmountPaid        float64
	DueDate           *time.Time
	LastPaymentDate   *time.Time
	LeaseID           int `gorm:"index"`
	OutstandingAmount float64
	PropertyID        int `gorm:"index"`
	Status            string
	TenantID          int `gorm:"index"`
}

// IsOverdue reports whether the installment is past due with money still owed.
func (i *Installment) IsOverdue(now time.Time) bool {
	return i.DueDate != nil && i.DueDate.Before(now) && i.OutstandingAmount > 0
}
