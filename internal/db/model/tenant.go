package model

import (
	"fmt"

	"gorm.io/gorm"
)

type Tenant struct {
	gorm.Model
	Email                 string `gorm:"uniqueIndex"`
	ExpectedAnnualRent    float64
	FirstName             string
	LastName              string
	LeaseID               int
	OutstandingAnnualRent float64
	PasswordHash          string
	PhoneNumber           string
	PropertyID            int
	UnitNumber            string
}

func (t *Tenant) SafeMap() map[string]string {
	return map[string]string{
		"id":                    fmt.Sprint(t.ID),
		"createdAt":             t.CreatedAt.Local().String(),
		"email":                 t.Email,
		"expectedAnnualRent":    fmt.Sprint(t.ExpectedAnnualRent),
		"firstName":             t.FirstName,
		"lastName":              t.LastName,
		"leaseId":               fmt.Sprint(t.LeaseID),
		"outstandingAnnualRent": fmt.Sprint(t.OutstandingAnnualRent),
		"phoneNumber":           t.PhoneNumber,
		"propertyId":            fmt.Sprint(t.PropertyID),
		"unitNumber":            t.UnitNumber,
	}
}
