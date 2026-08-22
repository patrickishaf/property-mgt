package model

import (
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Property struct {
	gorm.Model
	Address   string
	ManagerID string
	Name      string
	Type      string
}

type PropertyUnit struct {
	gorm.Model
	Floor      int
	Label      string
	PropertyID uuid.UUID
}
