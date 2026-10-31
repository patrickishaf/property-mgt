package model

import (
	"time"

	"gorm.io/gorm"
)

const (
	PaymentStatusExpired       = "expired"
	PaymentStatusFailed        = "failed"
	PaymentStatusPaid          = "paid"
	PaymentStatusPartiallyPaid = "partially_paid"
	PaymentStatusUnpaid        = "unpaid"

	PaymentPurposeInstallments = "installments"
	PaymentPurposeLease        = "lease"
	PaymentPurposePartial      = "partial"

	PaymentSourceCheckout = "checkout"
	PaymentSourceOffline  = "offline"

	PaymentMethodCheckout = "checkout"
)

type Payment struct {
	gorm.Model
	Amount         float64
	CheckoutID     string
	CheckoutURL    string
	Currency       string
	ExpiresAt      *time.Time
	LeaseID        int `gorm:"index"`
	Method         string
	Note           string
	PaidAt         *time.Time
	PropertyID     int `gorm:"index"`
	PropertyName   string
	Purpose        string
	RecordedBy     *uint
	Reference      string `gorm:"uniqueIndex"`
	Source         string
	Status         string `gorm:"index"`
	TargetAmount   float64
	TenantFullName string
	TenantID       int `gorm:"index"`
	UnitNumber     string
	// InstallmentIDs is a comma-separated list of the installments a checkout was
	// started for; allocation re-reads them when the checkout completes.
	InstallmentIDs string
}

// PaymentAllocation records how much of a payment went to an installment. A nil
// InstallmentID means the amount was applied to a lease without installments.
type PaymentAllocation struct {
	gorm.Model
	Amount        float64
	InstallmentID *uint
	PaymentID     uint `gorm:"index"`
}
