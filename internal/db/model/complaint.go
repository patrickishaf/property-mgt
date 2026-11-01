package model

import "gorm.io/gorm"

// Maintenance categories shared by complaints and issues.
const (
	MaintenanceCategoryElectrical = "electrical"
	MaintenanceCategoryLighting   = "lighting"
	MaintenanceCategoryOther      = "other"
	MaintenanceCategoryPlumbing   = "plumbing"
	MaintenanceCategorySecurity   = "security"
	MaintenanceCategoryStructural = "structural"
)

// Maintenance statuses shared by complaints and issues. A complaint linked to
// an issue follows the issue's status.
const (
	MaintenanceStatusFixed      = "fixed"
	MaintenanceStatusInProgress = "in_progress"
	MaintenanceStatusOpen       = "open"
)

// Maintenance urgency levels shared by complaints and issues.
const (
	MaintenanceUrgencyLow    = "low"
	MaintenanceUrgencyNormal = "normal"
	MaintenanceUrgencyUrgent = "urgent"
)

// Complaint is a maintenance problem reported by a tenant. Owners and managers
// group related complaints under an Issue.
type Complaint struct {
	gorm.Model
	Category        string
	CreatorFullName string
	CreatorID       int   `gorm:"index"`
	IssueID         *uint `gorm:"index"`
	Message         string
	PropertyID      int `gorm:"index"`
	Status          string
	UnitNumber      string
	Urgency         string
}

type ComplaintUpdate struct {
	gorm.Model
	ComplaintID     uint `gorm:"index"`
	CreatorFullName string
	CreatorID       int
	Status          string
	Message         string
}
