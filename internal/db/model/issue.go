package model

import "gorm.io/gorm"

type Issue struct {
	gorm.Model
	Message string
	Status  string
}
