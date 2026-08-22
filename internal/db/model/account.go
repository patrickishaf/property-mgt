package model

import (
	"time"

	"github.com/google/uuid"
)

type Account struct {
	ID        uuid.UUID
	CreatedAt time.Time
	Email     string
	UpdatedAt time.Time
	Name      string
	Password  string
	Phone     string
	Role      string
	TenantID  *string
	ManagerID *string
}
