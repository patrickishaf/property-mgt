package property

import (
	"time"

	"github.com/patrickishaf/property-mgt/internal/db/model"
	"github.com/patrickishaf/property-mgt/internal/db/repository"
)

type pageReq struct {
	Limit int `form:"limit" binding:"omitempty,min=1,max=100"`
	Page  int `form:"page" binding:"omitempty,min=1"`
}

func (this pageReq) values() (int, int) {
	page, limit := this.Page, this.Limit
	if page == 0 {
		page = 1
	}
	if limit == 0 {
		limit = defaultPageSize
	}
	return page, limit
}

type getPropertiesReq struct {
	pageReq
	IncludeArchived bool `form:"includeArchived"`
}

type createUnitReq struct {
	Floor  int    `json:"floor"`
	Label  string `json:"label"`
	Number string `json:"number" binding:"required"`
}

type createPropertyReq struct {
	ActivityStatus string          `json:"activityStatus"`
	City           string          `json:"city" binding:"required"`
	Country        string          `json:"country" binding:"required"`
	Location       string          `json:"location"`
	Name           string          `json:"name" binding:"required"`
	State          string          `json:"state" binding:"required"`
	Type           string          `json:"type" binding:"required"`
	Units          []createUnitReq `json:"units" binding:"omitempty,dive"`
}

type updatePropertyReq struct {
	ActivityStatus *string `json:"activityStatus"`
	City           *string `json:"city" binding:"omitempty,min=1"`
	Country        *string `json:"country" binding:"omitempty,min=1"`
	Location       *string `json:"location"`
	Name           *string `json:"name" binding:"omitempty,min=1"`
	State          *string `json:"state" binding:"omitempty,min=1"`
	Type           *string `json:"type" binding:"omitempty,min=1"`
}

type assignManagerReq struct {
	ManagerID uint `json:"managerId" binding:"required"`
}

type updateUnitReq struct {
	Floor *int    `json:"floor"`
	Label *string `json:"label"`
}

type createManagerReq struct {
	Email       string `json:"email" binding:"required,email"`
	FirstName   string `json:"firstName" binding:"required"`
	LastName    string `json:"lastName" binding:"required"`
	PhoneNumber string `json:"phoneNumber" binding:"required"`
}

type updateManagerReq struct {
	FirstName   *string `json:"firstName" binding:"omitempty,min=1"`
	LastName    *string `json:"lastName" binding:"omitempty,min=1"`
	PhoneNumber *string `json:"phoneNumber" binding:"omitempty,min=1"`
}

type listRes[T any] struct {
	Items []T   `json:"items"`
	Limit int   `json:"limit"`
	Page  int   `json:"page"`
	Total int64 `json:"total"`
}

type propertyRes struct {
	ID                    uint       `json:"id"`
	ActivityStatus        string     `json:"activityStatus"`
	ArchivedAt            *time.Time `json:"archivedAt"`
	City                  string     `json:"city"`
	Country               string     `json:"country"`
	Location              string     `json:"location"`
	ManagerID             *uint      `json:"managerId"`
	Name                  string     `json:"name"`
	NumberOfOccupiedUnits int        `json:"numberOfOccupiedUnits"`
	NumberOfUnits         int        `json:"numberOfUnits"`
	OutstandingRent       float64    `json:"outstandingRent"`
	OwnerID               uint       `json:"ownerId"`
	RentCollected         float64    `json:"rentCollected"`
	RentExpected          float64    `json:"rentExpected"`
	State                 string     `json:"state"`
	Type                  string     `json:"type"`
}

type unitRes struct {
	ID       uint   `json:"id"`
	Floor    int    `json:"floor"`
	Label    string `json:"label"`
	Number   string `json:"number"`
	Occupied bool   `json:"occupied"`
}

const (
	managerStatusActive  = "active"
	managerStatusInvited = "invited"
)

type managerRes struct {
	ID          uint   `json:"id"`
	Email       string `json:"email"`
	FirstName   string `json:"firstName"`
	LastName    string `json:"lastName"`
	PhoneNumber string `json:"phoneNumber"`
	Status      string `json:"status"`
}

func toPropertyRes(p model.Property, totals repository.RentTotals) propertyRes {
	res := propertyRes{
		ID:                    p.ID,
		ActivityStatus:        p.ActivityStatus,
		ArchivedAt:            p.ArchivedAt,
		City:                  p.City,
		Country:               p.Country,
		Location:              p.Location,
		Name:                  p.Name,
		NumberOfOccupiedUnits: p.NumberOfOccupiedUnits,
		NumberOfUnits:         p.NumberOfUnits,
		OutstandingRent:       totals.Outstanding,
		OwnerID:               uint(p.OwnerID),
		RentCollected:         totals.Collected,
		RentExpected:          totals.Expected,
		State:                 p.State,
		Type:                  p.Type,
	}
	if p.ManagerID > 0 {
		managerID := uint(p.ManagerID)
		res.ManagerID = &managerID
	}
	return res
}

func toUnitRes(u model.PropertyUnit, occupied bool) unitRes {
	return unitRes{
		ID:       u.ID,
		Floor:    u.Floor,
		Label:    u.Label,
		Number:   u.Number,
		Occupied: occupied,
	}
}

func toManagerRes(a model.Admin) managerRes {
	status := managerStatusInvited
	if a.PasswordHash != "" {
		status = managerStatusActive
	}
	return managerRes{
		ID:          a.ID,
		Email:       a.Email,
		FirstName:   a.FirstName,
		LastName:    a.LastName,
		PhoneNumber: a.PhoneNumber,
		Status:      status,
	}
}
