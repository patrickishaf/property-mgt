package model

import (
	"time"

	"gorm.io/gorm"
)

const (
	LeaseStatusActive     = "active"
	LeaseStatusEnded      = "ended"
	LeaseStatusTerminated = "terminated"
)

type Lease struct {
	gorm.Model
	AmountPaid           float64
	AllowPartialPayment  bool
	DurationType         string // annual or monthly
	EndDate              time.Time
	EndedAt              *time.Time
	ExpectedRent         float64
	MinimumPartialAmount float64
	OutstandingRent      float64
	PropertyID           int
	StartDate            time.Time
	Status               string
	TenantID             int
	UnitNumber           string
}

// Lease reminder kinds. Each records one automatic reminder sent to a tenant
// about a lease, so the daily reminder job never repeats itself.
const (
	LeaseReminderExpiring30  = "expiring_30"
	LeaseReminderExpiring14  = "expiring_14"
	LeaseReminderExpiring7   = "expiring_7"
	LeaseReminderOverdueRent = "overdue_rent"
	LeaseReminderExpired     = "lease_expired"
)

type LeaseReminder struct {
	gorm.Model
	Kind      string `gorm:"index:idx_lease_reminder,priority:2"`
	LeaseID   uint   `gorm:"index:idx_lease_reminder,priority:1"`
	MessageID uint
	SentAt    time.Time
}
