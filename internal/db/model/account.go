package model

import (
	"gorm.io/gorm"
)

type Account struct {
	gorm.Model
	Email     string
	Name      string
	Password  string
	Phone     string
	Role      string
	TenantID  *string
	ManagerID *string
}
