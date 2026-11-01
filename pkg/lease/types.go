package lease

import (
	"time"

	"github.com/patrickishaf/property-mgt/internal/db/model"
)

const (
	dateLayout      = "2006-01-02"
	defaultPageSize = 10
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

type getLeasesReq struct {
	pageReq
	EndingBefore string `form:"endingBefore" binding:"omitempty,datetime=2006-01-02"`
	HasBalance   *bool  `form:"hasBalance"`
	PropertyID   uint   `form:"propertyId"`
	Status       string `form:"status" binding:"omitempty,oneof=active ended terminated"`
	TenantID     uint   `form:"tenantId"`
	UnitNumber   string `form:"unitNumber"`
}

type createLeaseReq struct {
	AllowPartialPayment  bool    `json:"allowPartialPayment"`
	DurationType         string  `json:"durationType" binding:"required,oneof=annual monthly"`
	EndDate              string  `json:"endDate" binding:"required,datetime=2006-01-02"`
	ExpectedRent         float64 `json:"expectedRent" binding:"required,gt=0"`
	MinimumPartialAmount float64 `json:"minimumPartialAmount" binding:"gte=0"`
	PropertyID           uint    `json:"propertyId" binding:"required"`
	StartDate            string  `json:"startDate" binding:"required,datetime=2006-01-02"`
	TenantID             uint    `json:"tenantId" binding:"required"`
	UnitNumber           string  `json:"unitNumber" binding:"required"`
}

type updateLeaseReq struct {
	AllowPartialPayment  *bool    `json:"allowPartialPayment"`
	EndDate              *string  `json:"endDate" binding:"omitempty,datetime=2006-01-02"`
	ExpectedRent         *float64 `json:"expectedRent" binding:"omitempty,gt=0"`
	MinimumPartialAmount *float64 `json:"minimumPartialAmount" binding:"omitempty,gte=0"`
	Status               *string  `json:"status" binding:"omitempty,oneof=ended terminated"`
}

type getInstallmentsReq struct {
	pageReq
	DueBefore  string `form:"dueBefore" binding:"omitempty,datetime=2006-01-02"`
	LeaseID    uint   `form:"leaseId"`
	PropertyID uint   `form:"propertyId"`
	Status     string `form:"status" binding:"omitempty,oneof=unpaid partially_paid paid overdue"`
	TenantID   uint   `form:"tenantId"`
}

type installmentItemReq struct {
	Amount  float64 `json:"amount" binding:"required,gt=0"`
	DueDate string  `json:"dueDate" binding:"required,datetime=2006-01-02"`
}

type setInstallmentsReq struct {
	Items []installmentItemReq `json:"items" binding:"required,min=1,dive"`
}

type updateInstallmentReq struct {
	DueDate string `json:"dueDate" binding:"required,datetime=2006-01-02"`
}

type listRes[T any] struct {
	Items []T   `json:"items"`
	Limit int   `json:"limit"`
	Page  int   `json:"page"`
	Total int64 `json:"total"`
}

type leaseRes struct {
	ID                   uint       `json:"id"`
	AllowPartialPayment  bool       `json:"allowPartialPayment"`
	AmountPaid           float64    `json:"amountPaid"`
	DurationType         string     `json:"durationType"`
	EndDate              string     `json:"endDate"`
	EndedAt              *time.Time `json:"endedAt"`
	ExpectedRent         float64    `json:"expectedRent"`
	MinimumPartialAmount float64    `json:"minimumPartialAmount"`
	OutstandingRent      float64    `json:"outstandingRent"`
	PropertyID           int        `json:"propertyId"`
	StartDate            string     `json:"startDate"`
	Status               string     `json:"status"`
	TenantID             int        `json:"tenantId"`
	UnitNumber           string     `json:"unitNumber"`
}

type leaseDetailRes struct {
	leaseRes
	OpenInstallmentCount    int64  `json:"openInstallmentCount"`
	OverdueInstallmentCount int64  `json:"overdueInstallmentCount"`
	PropertyName            string `json:"propertyName"`
	TenantFullName          string `json:"tenantFullName"`
}

type installmentRes struct {
	ID                uint       `json:"id"`
	AmountExpected    float64    `json:"amountExpected"`
	AmountPaid        float64    `json:"amountPaid"`
	DueDate           string     `json:"dueDate"`
	LastPaymentDate   *time.Time `json:"lastPaymentDate"`
	LeaseID           int        `json:"leaseId"`
	OutstandingAmount float64    `json:"outstandingAmount"`
	Overdue           bool       `json:"overdue"`
	PropertyID        int        `json:"propertyId"`
	Status            string     `json:"status"`
	TenantID          int        `json:"tenantId"`
}

func toLeaseRes(l model.Lease) leaseRes {
	return leaseRes{
		ID:                   l.ID,
		AllowPartialPayment:  l.AllowPartialPayment,
		AmountPaid:           l.AmountPaid,
		DurationType:         l.DurationType,
		EndDate:              l.EndDate.Format(dateLayout),
		EndedAt:              l.EndedAt,
		ExpectedRent:         l.ExpectedRent,
		MinimumPartialAmount: l.MinimumPartialAmount,
		OutstandingRent:      l.OutstandingRent,
		PropertyID:           l.PropertyID,
		StartDate:            l.StartDate.Format(dateLayout),
		Status:               l.Status,
		TenantID:             l.TenantID,
		UnitNumber:           l.UnitNumber,
	}
}

func toInstallmentRes(i model.Installment, now time.Time) installmentRes {
	res := installmentRes{
		ID:                i.ID,
		AmountExpected:    i.AmountExpected,
		AmountPaid:        i.AmountPaid,
		LastPaymentDate:   i.LastPaymentDate,
		LeaseID:           i.LeaseID,
		OutstandingAmount: i.OutstandingAmount,
		Overdue:           i.IsOverdue(now),
		PropertyID:        i.PropertyID,
		Status:            i.Status,
		TenantID:          i.TenantID,
	}
	if i.DueDate != nil {
		res.DueDate = i.DueDate.Format(dateLayout)
	}
	return res
}
