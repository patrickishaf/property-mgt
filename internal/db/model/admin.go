package model

import (
	"fmt"

	"gorm.io/gorm"
)

type Admin struct {
	gorm.Model
	Email        string
	FirstName    string
	LastName     string
	PasswordHash string
	PhoneNumber  string
	Role         string
}

func (a *Admin) Safe() map[string]string {
	return map[string]string{
		"id":          fmt.Sprint(a.ID),
		"createdAt":   a.CreatedAt.Local().String(),
		"email":       a.Email,
		"firstName":   a.FirstName,
		"lastName":    a.LastName,
		"phoneNumber": a.PhoneNumber,
		"role":        a.Role,
	}
}
