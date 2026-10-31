package model

import (
	"time"

	"github.com/patrickishaf/property-mgt/common"
	"gorm.io/gorm"
)

type Property struct {
	gorm.Model
	ActivityStatus        string
	ArchivedAt            *time.Time
	City                  string
	Country               string
	Location              string
	ManagerID             int
	Name                  string
	NumberOfUnits         int
	NumberOfOccupiedUnits int
	OwnerID               int
	State                 string
	Type                  string
}

// AccessibleBy reports whether the actor owns or manages the property.
func (p *Property) AccessibleBy(actor *common.Actor) bool {
	if p == nil || actor == nil {
		return false
	}
	switch actor.Role {
	case common.AdminRoleOwner.ToString():
		return p.OwnerID > 0 && uint(p.OwnerID) == actor.ID
	case common.AdminRoleManager.ToString():
		return p.ManagerID > 0 && uint(p.ManagerID) == actor.ID
	default:
		return false
	}
}

// OwnedBy reports whether the actor is the property's owner.
func (p *Property) OwnedBy(actor *common.Actor) bool {
	return p != nil && actor != nil && actor.Role == common.AdminRoleOwner.ToString() &&
		p.OwnerID > 0 && uint(p.OwnerID) == actor.ID
}

type PropertyUnit struct {
	gorm.Model
	Floor      int
	Label      string
	Number     string `gorm:"uniqueIndex:idx_property_unit_number"`
	PropertyID int    `gorm:"uniqueIndex:idx_property_unit_number"`
}
