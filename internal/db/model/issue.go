package model

import "gorm.io/gorm"

// Issue is a maintenance job opened by an owner or manager. One or more tenant
// complaints are linked to it, and its updates are shared with those tenants.
type Issue struct {
	gorm.Model
	Category        string
	CreatorFullName string
	CreatorID       int
	Message         string
	PropertyID      int `gorm:"index"`
	Status          string
	Title           string
	Urgency         string
}

type IssueUpdate struct {
	gorm.Model
	CreatorFullName string
	CreatorID       int
	IssueID         uint `gorm:"index"`
	Message         string
	Status          string
}
