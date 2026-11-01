package repository

import (
	"errors"
	"log/slog"
	"time"

	"github.com/patrickishaf/property-mgt/internal/db/model"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type InstallmentRepository struct {
	db     *gorm.DB
	logger *slog.Logger
}

// InstallmentFilter narrows an installment listing. Zero values are ignored.
// Status may also be "overdue".
type InstallmentFilter struct {
	DueBefore  *time.Time
	LeaseID    uint
	ManagerID  uint
	OwnerID    uint
	PropertyID uint
	Status     string
	TenantID   uint
}

const InstallmentFilterOverdue = "overdue"

// DueTotals summarises rent falling due in a period: the total amount, how
// many items still owe money, and how many of those are overdue (and by how much).
type DueTotals struct {
	Expected      float64
	OpenCount     int64
	OverdueAmount float64
	OverdueCount  int64
}

func CreateInstallmentRepository(db *gorm.DB, logger *slog.Logger) *InstallmentRepository {
	return &InstallmentRepository{
		db:     db,
		logger: logger,
	}
}

func (this *InstallmentRepository) FindOneById(id uint) (*model.Installment, error) {
	var i model.Installment
	err := this.db.Where("id = ?", id).First(&i).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &i, nil
}

// FindByLeaseForUpdateTRX returns the lease's installments in due-date order and
// locks them for the rest of the transaction.
func (this *InstallmentRepository) FindByLeaseForUpdateTRX(leaseID uint, trx *gorm.DB) ([]model.Installment, error) {
	var installments []model.Installment
	err := trx.Clauses(clause.Locking{Strength: "UPDATE"}).
		Where("lease_id = ?", leaseID).Order("due_date ASC, id ASC").Find(&installments).Error
	return installments, err
}

func (this *InstallmentRepository) FindByLease(leaseID uint) ([]model.Installment, error) {
	var installments []model.Installment
	err := this.db.Where("lease_id = ?", leaseID).Order("due_date ASC, id ASC").Find(&installments).Error
	return installments, err
}

func (this *InstallmentRepository) FindMany(filter InstallmentFilter, now time.Time, page int, limit int) ([]model.Installment, int64, error) {
	query := scopeToProperties(this.db, this.db.Model(&model.Installment{}), filter.OwnerID, filter.ManagerID)
	if filter.LeaseID != 0 {
		query = query.Where("lease_id = ?", filter.LeaseID)
	}
	if filter.TenantID != 0 {
		query = query.Where("tenant_id = ?", filter.TenantID)
	}
	if filter.PropertyID != 0 {
		query = query.Where("property_id = ?", filter.PropertyID)
	}
	switch filter.Status {
	case "":
	case InstallmentFilterOverdue:
		query = query.Where("due_date < ? AND outstanding_amount > 0", now)
	default:
		query = query.Where("status = ?", filter.Status)
	}
	if filter.DueBefore != nil {
		query = query.Where("due_date < ?", *filter.DueBefore)
	}
	query = query.Session(&gorm.Session{})

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var installments []model.Installment
	err := query.Order("due_date ASC, id ASC").Offset((page - 1) * limit).Limit(limit).Find(&installments).Error
	if err != nil {
		return nil, 0, err
	}
	return installments, total, nil
}

func (this *InstallmentRepository) SaveManyTRX(installments []model.Installment, trx *gorm.DB) error {
	if len(installments) == 0 {
		return nil
	}
	return trx.Create(&installments).Error
}

// DeleteUnpaidTRX hard-deletes the lease's installments that have received no money.
func (this *InstallmentRepository) DeleteUnpaidTRX(leaseID uint, trx *gorm.DB) error {
	return trx.Unscoped().Where("lease_id = ? AND amount_paid = 0", leaseID).Delete(&model.Installment{}).Error
}

func (this *InstallmentRepository) UpdateTRX(id uint, fields map[string]any, trx *gorm.DB) error {
	return trx.Model(&model.Installment{}).Where("id = ?", id).Updates(fields).Error
}

func (this *InstallmentRepository) UpdateDueDate(id uint, dueDate time.Time) error {
	return this.db.Model(&model.Installment{}).Where("id = ?", id).Update("due_date", dueDate).Error
}

// CountOpenByLease returns how many of the lease's installments still have money
// owed, and how many of those are overdue.
func (this *InstallmentRepository) CountOpenByLease(leaseID uint, now time.Time) (open int64, overdue int64, err error) {
	var row struct {
		Open    int64
		Overdue int64
	}
	err = this.db.Model(&model.Installment{}).
		Select("COUNT(*) FILTER (WHERE outstanding_amount > 0) AS open, COUNT(*) FILTER (WHERE outstanding_amount > 0 AND due_date < ?) AS overdue", now).
		Where("lease_id = ?", leaseID).Scan(&row).Error
	return row.Open, row.Overdue, err
}

// YearTotals sums the scoped installments due in [start, end). Overdue means
// due before now with money still owed.
func (this *InstallmentRepository) YearTotals(ownerID uint, managerID uint, propertyID uint, start time.Time, end time.Time, now time.Time) (DueTotals, error) {
	query := scopeToProperties(this.db, this.db.Model(&model.Installment{}), ownerID, managerID).
		Where("due_date >= ? AND due_date < ?", start, end)
	if propertyID != 0 {
		query = query.Where("property_id = ?", propertyID)
	}
	var totals DueTotals
	err := query.Select(
		"COALESCE(SUM(amount_expected), 0) AS expected, "+
			"COUNT(*) FILTER (WHERE outstanding_amount > 0) AS open_count, "+
			"COUNT(*) FILTER (WHERE outstanding_amount > 0 AND due_date < ?) AS overdue_count, "+
			"COALESCE(SUM(outstanding_amount) FILTER (WHERE outstanding_amount > 0 AND due_date < ?), 0) AS overdue_amount",
		now, now,
	).Scan(&totals).Error
	return totals, err
}

// FindOpen returns the scoped installments that still have money owed, in
// due-date order. A non-zero tenantID limits them to that tenant.
func (this *InstallmentRepository) FindOpen(ownerID uint, managerID uint, propertyID uint, tenantID uint) ([]model.Installment, error) {
	query := scopeToProperties(this.db, this.db.Model(&model.Installment{}), ownerID, managerID).
		Where("outstanding_amount > 0")
	if propertyID != 0 {
		query = query.Where("property_id = ?", propertyID)
	}
	if tenantID != 0 {
		query = query.Where("tenant_id = ?", tenantID)
	}
	var installments []model.Installment
	err := query.Order("due_date ASC, id ASC").Find(&installments).Error
	return installments, err
}
