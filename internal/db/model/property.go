package model

import (
	"gorm.io/gorm"
)

type Property struct {
	gorm.Model
	ActivityStatus          string
	AnnualRentExpected      float64
	City                    string
	Country                 string
	ManagerID               int
	Name                    string
	NumberOfUnits           int
	NumberOfOccupiedUnits   int
	OwnerID                 int
	RentCollected           int
	RentExpected            float64
	State                   string
	TotalOutstandingPayment int
	Type                    string
}

type PropertyUnit struct {
	gorm.Model
	Floor      int
	Label      string
	Number     string
	PropertyID int
}
