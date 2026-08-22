package model

import (
	"time"

	"gorm.io/gorm"
)

type Payment struct {
	gorm.Model
	Amount     float64
	Date       time.Time
	Method     string
	RecordedBy string
	Reference  string
}
