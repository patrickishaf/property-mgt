package payment

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

type createPaymentReq struct {
	Amount         float64 `json:"amount" binding:"required,gt=0"`
	InstallmentIDs []uint  `json:"installmentIds" binding:"omitempty,dive,gt=0"`
	LeaseID        uint    `json:"leaseId" binding:"required"`
	Method         string  `json:"method" binding:"required,oneof=cash bank_transfer pos cheque other"`
	Note           string  `json:"note"`
	PaidAt         string  `json:"paidAt" binding:"omitempty,datetime=2006-01-02"`
	Reference      string  `json:"reference"`
}

type getPaymentsReq struct {
	pageReq
	From       string `form:"from" binding:"omitempty,datetime=2006-01-02"`
	LeaseID    uint   `form:"leaseId"`
	Method     string `form:"method"`
	PropertyID uint   `form:"propertyId"`
	Purpose    string `form:"purpose" binding:"omitempty,oneof=lease installments partial"`
	Source     string `form:"source" binding:"omitempty,oneof=offline checkout"`
	Status     string `form:"status" binding:"omitempty,oneof=unpaid partially_paid paid failed expired"`
	TenantID   uint   `form:"tenantId"`
	To         string `form:"to" binding:"omitempty,datetime=2006-01-02"`
}

type getStatsReq struct {
	PropertyID uint `form:"propertyId"`
}

type initiateCheckoutReq struct {
	Amount         float64 `json:"amount" binding:"omitempty,gt=0"`
	InstallmentIDs []uint  `json:"installmentIds" binding:"omitempty,dive,gt=0"`
	LeaseID        uint    `json:"leaseId"`
	Type           string  `json:"type" binding:"required,oneof=lease installments partial"`
}

type listRes[T any] struct {
	Items []T   `json:"items"`
	Limit int   `json:"limit"`
	Page  int   `json:"page"`
	Total int64 `json:"total"`
}

type paymentRes struct {
	ID             uint       `json:"id"`
	Amount         float64    `json:"amount"`
	CheckoutURL    string     `json:"checkoutUrl,omitempty"`
	CreatedAt      time.Time  `json:"createdAt"`
	Currency       string     `json:"currency"`
	ExpiresAt      *time.Time `json:"expiresAt,omitempty"`
	LeaseID        int        `json:"leaseId"`
	Method         string     `json:"method"`
	Note           string     `json:"note"`
	PaidAt         *time.Time `json:"paidAt"`
	PropertyID     int        `json:"propertyId"`
	PropertyName   string     `json:"propertyName"`
	Purpose        string     `json:"purpose"`
	RecordedBy     *uint      `json:"recordedBy"`
	Reference      string     `json:"reference"`
	Source         string     `json:"source"`
	Status         string     `json:"status"`
	TargetAmount   float64    `json:"targetAmount"`
	TenantFullName string     `json:"tenantFullName"`
	TenantID       int        `json:"tenantId"`
	UnitNumber     string     `json:"unitNumber"`
}

type allocationRes struct {
	Amount        float64 `json:"amount"`
	InstallmentID *uint   `json:"installmentId"`
}

type receiptRes struct {
	ID                           uint   `json:"id"`
	AmountPaid                   string `json:"amountPaid"`
	InstallmentAmount            string `json:"installmentAmount"`
	InstallmentAmountPaid        string `json:"installmentAmountPaid"`
	InstallmentDueDate           string `json:"installmentDueDate"`
	InstallmentStatus            string `json:"installmentStatus"`
	OutstandingInstallmentAmount string `json:"outstandingInstallmentAmount"`
	PaymentDate                  string `json:"paymentDate"`
	PaymentMethod                string `json:"paymentMethod"`
	PreviousPaymentAmount        string `json:"previousPaymentAmount"`
	SenderEmail                  string `json:"senderEmail"`
	SenderFullName               string `json:"senderFullName"`
}

type paymentDetailRes struct {
	paymentRes
	Allocations []allocationRes `json:"allocations"`
	Receipt     *receiptRes     `json:"receipt"`
}

type portfolioRes struct {
	ActiveTenantCount       int64   `json:"activeTenantCount"`
	OpenInstallmentCount    int64   `json:"openInstallmentCount"`
	OutstandingRentBalance  float64 `json:"outstandingRentBalance"`
	OverdueInstallmentCount int64   `json:"overdueInstallmentCount"`
	PercentageRentCollected float64 `json:"percentageRentCollected"`
	TotalOverdueAmount      float64 `json:"totalOverdueAmount"`
	TotalRentCollected      float64 `json:"totalRentCollected"`
	TotalRentExpected       float64 `json:"totalRentExpected"`
}

func toPaymentRes(p model.Payment) paymentRes {
	return paymentRes{
		ID:             p.ID,
		Amount:         p.Amount,
		CheckoutURL:    p.CheckoutURL,
		CreatedAt:      p.CreatedAt,
		Currency:       p.Currency,
		ExpiresAt:      p.ExpiresAt,
		LeaseID:        p.LeaseID,
		Method:         p.Method,
		Note:           p.Note,
		PaidAt:         p.PaidAt,
		PropertyID:     p.PropertyID,
		PropertyName:   p.PropertyName,
		Purpose:        p.Purpose,
		RecordedBy:     p.RecordedBy,
		Reference:      p.Reference,
		Source:         p.Source,
		Status:         p.Status,
		TargetAmount:   p.TargetAmount,
		TenantFullName: p.TenantFullName,
		TenantID:       p.TenantID,
		UnitNumber:     p.UnitNumber,
	}
}

func toReceiptRes(r model.Receipt) receiptRes {
	return receiptRes{
		ID:                           r.ID,
		AmountPaid:                   r.AmountPaid,
		InstallmentAmount:            r.InstallmentAmount,
		InstallmentAmountPaid:        r.InstallmentAmountPaid,
		InstallmentDueDate:           r.InstallmentDueDate,
		InstallmentStatus:            r.InstallmentStatus,
		OutstandingInstallmentAmount: r.OutstandingInstallmentPayment,
		PaymentDate:                  r.PaymentDate,
		PaymentMethod:                r.PaymentMethod,
		PreviousPaymentAmount:        r.PreviousPaymentAmount,
		SenderEmail:                  r.SenderEmail,
		SenderFullName:               r.SenderFullName,
	}
}
