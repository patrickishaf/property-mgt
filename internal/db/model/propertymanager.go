package model

import "gorm.io/gorm"

type PropertyManager struct {
	gorm.Model
	Email       string
	FirstName   string
	LastName    string
	PhoneNumber string
}
