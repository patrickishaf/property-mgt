package model

import (
	"gorm.io/gorm"
)

type Tenant struct {
	gorm.Model
	ExpectedAnnualRent    float64
	FirstName             string
	LastName              string
	LeaseID               int
	OutstandingAnnualRent float64
	PhoneNumber           string
	PropertyID            string
	UnitNumber            string
}
