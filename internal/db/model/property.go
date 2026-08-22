package model

import (
	"time"

	"github.com/google/uuid"
)

type Property struct {
	ID        uuid.UUID
	Address   any
	CreatedAt time.Time
	ManagerID string
	Name      string
	Type      string
	UpdatedAt time.Time
}

type PropertyUnit struct {
	id         uuid.UUID
	CreatedAt  time.Time
	Floor      int
	Label      string
	PropertyID uuid.UUID
	UpdatedAt  time.Time
}
