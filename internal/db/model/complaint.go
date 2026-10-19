package model

import "gorm.io/gorm"

type Complaint struct {
	gorm.Model
	CreatorFullName string
	CreatorID       int
	Message         string
	Status          string
}

type ComplaintUpdate struct {
	gorm.Model
	CreatorFullName string
	CreatorID       int
	Status          string
	Message         string
}
