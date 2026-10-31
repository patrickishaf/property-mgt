package repository

import (
	"errors"
	"log/slog"
	"time"

	"github.com/patrickishaf/property-mgt/internal/db/model"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type PaymentRepository struct {
	db     *gorm.DB
	logger *slog.Logger
}

// PaymentFilter narrows a payment listing. Zero values are ignored. OwnerID and
// ManagerID scope results to that admin's properties.
type PaymentFilter struct {
	From       *time.Time
	LeaseID    uint
	ManagerID  uint
	Method     string
	OwnerID    uint
	PropertyID uint
	Purpose    string
	Source     string
	Status     string
	TenantID   uint
	To         *time.Time
}

func CreatePaymentRepository(db *gorm.DB, logger *slog.Logger) *PaymentRepository {
	return &PaymentRepository{
		db:     db,
		logger: logger,
	}
}

func (this *PaymentRepository) SaveOneTRX(payment *model.Payment, trx *gorm.DB) error {
	return trx.Create(payment).Error
}

func (this *PaymentRepository) FindOneById(id uint) (*model.Payment, error) {
	var p model.Payment
	err := this.db.Where("id = ?", id).First(&p).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &p, nil
}

// FindOneByReferenceForUpdateTRX loads a payment by reference and locks its row
// for the rest of the transaction.
func (this *PaymentRepository) FindOneByReferenceForUpdateTRX(reference string, trx *gorm.DB) (*model.Payment, error) {
	var p model.Payment
	err := trx.Clauses(clause.Locking{Strength: "UPDATE"}).Where("reference = ?", reference).First(&p).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &p, nil
}

func (this *PaymentRepository) ExistsByReference(reference string) (bool, error) {
	var count int64
	err := this.db.Model(&model.Payment{}).Where("reference = ?", reference).Count(&count).Error
	return count > 0, err
}

// FindOpenCheckoutForLease returns an unpaid checkout on the lease that has not
// expired yet, if any.
func (this *PaymentRepository) FindOpenCheckoutForLease(leaseID uint, now time.Time) (*model.Payment, error) {
	var p model.Payment
	err := this.db.Where("lease_id = ? AND source = ? AND status = ? AND expires_at > ?",
		leaseID, model.PaymentSourceCheckout, model.PaymentStatusUnpaid, now).First(&p).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &p, nil
}

// ExpireStaleCheckouts marks unpaid checkouts past their expiry as expired.
func (this *PaymentRepository) ExpireStaleCheckouts(now time.Time) error {
	return this.db.Model(&model.Payment{}).
		Where("source = ? AND status = ? AND expires_at <= ?", model.PaymentSourceCheckout, model.PaymentStatusUnpaid, now).
		Update("status", model.PaymentStatusExpired).Error
}

func (this *PaymentRepository) UpdateTRX(id uint, fields map[string]any, trx *gorm.DB) error {
	return trx.Model(&model.Payment{}).Where("id = ?", id).Updates(fields).Error
}

func (this *PaymentRepository) FindMany(filter PaymentFilter, page int, limit int) ([]model.Payment, int64, error) {
	query := scopeToProperties(this.db, this.db.Model(&model.Payment{}), filter.OwnerID, filter.ManagerID)
	if filter.PropertyID != 0 {
		query = query.Where("property_id = ?", filter.PropertyID)
	}
	if filter.TenantID != 0 {
		query = query.Where("tenant_id = ?", filter.TenantID)
	}
	if filter.LeaseID != 0 {
		query = query.Where("lease_id = ?", filter.LeaseID)
	}
	for column, value := range map[string]string{
		"status": filter.Status, "method": filter.Method, "source": filter.Source, "purpose": filter.Purpose,
	} {
		if value != "" {
			query = query.Where(column+" = ?", value)
		}
	}
	if filter.From != nil {
		query = query.Where("COALESCE(paid_at, created_at) >= ?", *filter.From)
	}
	if filter.To != nil {
		query = query.Where("COALESCE(paid_at, created_at) < ?", *filter.To)
	}
	query = query.Session(&gorm.Session{})

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var payments []model.Payment
	err := query.Order("id DESC").Offset((page - 1) * limit).Limit(limit).Find(&payments).Error
	if err != nil {
		return nil, 0, err
	}
	return payments, total, nil
}

func (this *PaymentRepository) SaveAllocationsTRX(allocations []model.PaymentAllocation, trx *gorm.DB) error {
	if len(allocations) == 0 {
		return nil
	}
	return trx.Create(&allocations).Error
}

func (this *PaymentRepository) FindAllocations(paymentID uint) ([]model.PaymentAllocation, error) {
	var allocations []model.PaymentAllocation
	err := this.db.Where("payment_id = ?", paymentID).Order("id ASC").Find(&allocations).Error
	return allocations, err
}

func (this *PaymentRepository) SaveReceiptTRX(receipt *model.Receipt, trx *gorm.DB) error {
	return trx.Create(receipt).Error
}

func (this *PaymentRepository) FindReceipt(paymentID uint) (*model.Receipt, error) {
	var r model.Receipt
	err := this.db.Where("payment_id = ?", paymentID).First(&r).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &r, nil
}
