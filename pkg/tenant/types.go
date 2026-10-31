package tenant

import "github.com/patrickishaf/property-mgt/internal/db/model"

const dateLayout = "2006-01-02"

type createLeaseReq struct {
	AllowPartialPayment  bool    `json:"allowPartialPayment"`
	DurationType         string  `json:"durationType" binding:"required,oneof=annual monthly"`
	EndDate              string  `json:"endDate" binding:"required,datetime=2006-01-02"`
	ExpectedRent         float64 `json:"expectedRent" binding:"required,gt=0"`
	MinimumPartialAmount float64 `json:"minimumPartialAmount" binding:"gte=0"`
	StartDate            string  `json:"startDate" binding:"required,datetime=2006-01-02"`
}

type createTenantReq struct {
	Email       string         `json:"email" binding:"required,email"`
	FirstName   string         `json:"firstName" binding:"required"`
	LastName    string         `json:"lastName" binding:"required"`
	Lease       createLeaseReq `json:"lease" binding:"required"`
	PhoneNumber string         `json:"phoneNumber" binding:"required"`
	PropertyID  uint           `json:"propertyId" binding:"required"`
	UnitNumber  string         `json:"unitNumber" binding:"required"`
}

type getTenantsReq struct {
	Limit      int  `form:"limit" binding:"omitempty,min=1,max=100"`
	Page       int  `form:"page" binding:"omitempty,min=1"`
	PropertyID uint `form:"propertyId"`
}

type updateTenantReq struct {
	FirstName   *string `json:"firstName" binding:"omitempty,min=1"`
	LastName    *string `json:"lastName" binding:"omitempty,min=1"`
	PhoneNumber *string `json:"phoneNumber" binding:"omitempty,min=1"`
}

type tenantRes struct {
	ID              uint    `json:"id"`
	Email           string  `json:"email"`
	ExpectedRent    float64 `json:"expectedRent"`
	FirstName       string  `json:"firstName"`
	LastName        string  `json:"lastName"`
	LeaseID         int     `json:"leaseId"`
	OutstandingRent float64 `json:"outstandingRent"`
	PhoneNumber     string  `json:"phoneNumber"`
	PropertyID      int     `json:"propertyId"`
	UnitNumber      string  `json:"unitNumber"`
}

type tenantListRes struct {
	Items []tenantRes `json:"items"`
	Limit int         `json:"limit"`
	Page  int         `json:"page"`
	Total int64       `json:"total"`
}

func toTenantRes(t model.Tenant) tenantRes {
	return tenantRes{
		ID:              t.ID,
		Email:           t.Email,
		ExpectedRent:    t.ExpectedAnnualRent,
		FirstName:       t.FirstName,
		LastName:        t.LastName,
		LeaseID:         t.LeaseID,
		OutstandingRent: t.OutstandingAnnualRent,
		PhoneNumber:     t.PhoneNumber,
		PropertyID:      t.PropertyID,
		UnitNumber:      t.UnitNumber,
	}
}
