package repository

import (
	"context"
	"errors"
	"log/slog"
	"time"

	"github.com/patrickishaf/property-mgt/common"
	"github.com/patrickishaf/property-mgt/internal/db/model"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type LeaseRepository struct {
	db     *gorm.DB
	logger *slog.Logger
}

func CreateLeaseRepo(db *gorm.DB, logger *slog.Logger) *LeaseRepository {
	return &LeaseRepository{
		db:     db,
		logger: logger,
	}
}

func (this *LeaseRepository) SaveOne(ctx context.Context, lease *model.Lease) error {
	err := this.db.Create(lease).Error
	if err != nil {
		this.logger.ErrorContext(ctx, "failed to insert lease into the db", "err", err.Error())
		return errors.New(common.ErrDBInsertFailed)
	}
	return nil
}

func (this *LeaseRepository) SaveOneTRX(lease *model.Lease, trx *gorm.DB) error {
	return trx.Create(lease).Error
}

func (this *LeaseRepository) ExistsActiveForUnit(propertyID uint, unitNumber string) (bool, error) {
	var count int64
	err := this.db.Model(&model.Lease{}).
		Where("property_id = ? AND unit_number = ? AND status = ?", propertyID, unitNumber, model.LeaseStatusActive).
		Count(&count).Error
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

func (this *LeaseRepository) ExistsActiveForProperty(propertyID uint) (bool, error) {
	var count int64
	err := this.db.Model(&model.Lease{}).
		Where("property_id = ? AND status = ?", propertyID, model.LeaseStatusActive).
		Count(&count).Error
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

// ActiveUnitNumbers returns the unit numbers of a property that have an active lease.
func (this *LeaseRepository) ActiveUnitNumbers(propertyID uint) (map[string]bool, error) {
	var numbers []string
	err := this.db.Model(&model.Lease{}).
		Where("property_id = ? AND status = ?", propertyID, model.LeaseStatusActive).
		Pluck("unit_number", &numbers).Error
	if err != nil {
		return nil, err
	}
	occupied := make(map[string]bool, len(numbers))
	for _, n := range numbers {
		occupied[n] = true
	}
	return occupied, nil
}

type RentTotals struct {
	Collected   float64
	Expected    float64
	Outstanding float64
}

// RentTotalsByProperty sums rent across the active leases of each property.
func (this *LeaseRepository) RentTotalsByProperty(propertyIDs []uint) (map[uint]RentTotals, error) {
	totals := make(map[uint]RentTotals, len(propertyIDs))
	if len(propertyIDs) == 0 {
		return totals, nil
	}
	var rows []struct {
		PropertyID  uint
		Collected   float64
		Expected    float64
		Outstanding float64
	}
	err := this.db.Model(&model.Lease{}).
		Select("property_id, COALESCE(SUM(amount_paid), 0) AS collected, COALESCE(SUM(expected_rent), 0) AS expected, COALESCE(SUM(outstanding_rent), 0) AS outstanding").
		Where("property_id IN ? AND status = ?", propertyIDs, model.LeaseStatusActive).
		Group("property_id").
		Scan(&rows).Error
	if err != nil {
		return nil, err
	}
	for _, r := range rows {
		totals[r.PropertyID] = RentTotals{Collected: r.Collected, Expected: r.Expected, Outstanding: r.Outstanding}
	}
	return totals, nil
}

func (this *LeaseRepository) FindOneById(id uint) (*model.Lease, error) {
	var l model.Lease
	err := this.db.Where("id = ?", id).First(&l).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &l, nil
}

// FindOneByIdForUpdateTRX loads the lease and locks its row for the rest of the transaction.
func (this *LeaseRepository) FindOneByIdForUpdateTRX(id uint, trx *gorm.DB) (*model.Lease, error) {
	var l model.Lease
	err := trx.Clauses(clause.Locking{Strength: "UPDATE"}).Where("id = ?", id).First(&l).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &l, nil
}

func (this *LeaseRepository) FindActiveByTenant(tenantID uint) (*model.Lease, error) {
	var l model.Lease
	err := this.db.Where("tenant_id = ? AND status = ?", tenantID, model.LeaseStatusActive).Order("id DESC").First(&l).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &l, nil
}

func (this *LeaseRepository) Update(id uint, fields map[string]any) error {
	return this.db.Model(&model.Lease{}).Where("id = ?", id).Updates(fields).Error
}

func (this *LeaseRepository) UpdateTRX(id uint, fields map[string]any, trx *gorm.DB) error {
	return trx.Model(&model.Lease{}).Where("id = ?", id).Updates(fields).Error
}

// PortfolioTotals sums rent over active leases in the scoped properties.
type PortfolioTotals struct {
	ActiveTenantCount int64
	Collected         float64
	Expected          float64
	Outstanding       float64
}

func (this *LeaseRepository) PortfolioTotals(ownerID uint, managerID uint, propertyID uint) (PortfolioTotals, error) {
	query := scopeToProperties(this.db, this.db.Model(&model.Lease{}), ownerID, managerID).
		Where("status = ?", model.LeaseStatusActive)
	if propertyID != 0 {
		query = query.Where("property_id = ?", propertyID)
	}
	var totals PortfolioTotals
	err := query.Select(
		"COUNT(DISTINCT tenant_id) AS active_tenant_count, " +
			"COALESCE(SUM(amount_paid), 0) AS collected, " +
			"COALESCE(SUM(expected_rent), 0) AS expected, " +
			"COALESCE(SUM(outstanding_rent), 0) AS outstanding",
	).Scan(&totals).Error
	return totals, err
}

// LeaseFilter narrows a lease listing. Zero values are ignored. OwnerID and
// ManagerID scope results to that admin's properties.
type LeaseFilter struct {
	EndingBefore *time.Time
	HasBalance   *bool
	ManagerID    uint
	OwnerID      uint
	PropertyID   uint
	Status       string
	TenantID     uint
	UnitNumber   string
}

func (this *LeaseRepository) FindMany(filter LeaseFilter, page int, limit int) ([]model.Lease, int64, error) {
	query := scopeToProperties(this.db, this.db.Model(&model.Lease{}), filter.OwnerID, filter.ManagerID)
	if filter.PropertyID != 0 {
		query = query.Where("property_id = ?", filter.PropertyID)
	}
	if filter.TenantID != 0 {
		query = query.Where("tenant_id = ?", filter.TenantID)
	}
	if filter.UnitNumber != "" {
		query = query.Where("unit_number = ?", filter.UnitNumber)
	}
	if filter.Status != "" {
		query = query.Where("status = ?", filter.Status)
	}
	if filter.EndingBefore != nil {
		query = query.Where("end_date < ?", *filter.EndingBefore)
	}
	if filter.HasBalance != nil {
		if *filter.HasBalance {
			query = query.Where("outstanding_rent > 0")
		} else {
			query = query.Where("outstanding_rent <= 0")
		}
	}
	query = query.Session(&gorm.Session{})

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var leases []model.Lease
	err := query.Order("id DESC").Offset((page - 1) * limit).Limit(limit).Find(&leases).Error
	if err != nil {
		return nil, 0, err
	}
	return leases, total, nil
}

// FindActiveEndingBefore returns the IDs of active leases whose end date is before the given time.
func (this *LeaseRepository) FindActiveEndingBefore(before time.Time) ([]uint, error) {
	var ids []uint
	err := this.db.Model(&model.Lease{}).
		Where("status = ? AND end_date < ?", model.LeaseStatusActive, before).
		Order("id ASC").Pluck("id", &ids).Error
	return ids, err
}

// FindLatestWithBalanceByTenant returns the tenant's most recent lease that still has rent owed.
func (this *LeaseRepository) FindLatestWithBalanceByTenant(tenantID uint) (*model.Lease, error) {
	var l model.Lease
	err := this.db.Where("tenant_id = ? AND outstanding_rent > 0", tenantID).Order("id DESC").First(&l).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &l, nil
}

// TenantContact is a tenant with an active lease, as offered when choosing
// message recipients.
type TenantContact struct {
	Email             string
	FirstName         string
	HasOverdueBalance bool
	LastName          string
	PhoneNumber       string
	PropertyID        uint
	PropertyName      string
	TenantID          uint
	UnitNumber        string
}

// FindActiveTenantContacts returns tenants with an active lease on the scoped
// properties. A tenant has an overdue balance when their lease has an
// installment past its due date with money owed or, for a lease without
// installments, when rent is still owed after the lease start date.
func (this *LeaseRepository) FindActiveTenantContacts(ownerID uint, managerID uint, propertyID uint, now time.Time) ([]TenantContact, error) {
	query := this.db.Table("leases AS l").
		Select(`t.id AS tenant_id, t.first_name, t.last_name, t.email, t.phone_number,
			l.property_id, p.name AS property_name, l.unit_number,
			CASE WHEN EXISTS (SELECT 1 FROM installments i WHERE i.lease_id = l.id AND i.deleted_at IS NULL)
				THEN EXISTS (SELECT 1 FROM installments i WHERE i.lease_id = l.id AND i.deleted_at IS NULL
					AND i.outstanding_amount > 0 AND i.due_date < ?)
				ELSE l.outstanding_rent > 0 AND l.start_date < ?
			END AS has_overdue_balance`, now, now).
		Joins("JOIN tenants t ON t.id = l.tenant_id AND t.deleted_at IS NULL").
		Joins("JOIN properties p ON p.id = l.property_id AND p.deleted_at IS NULL").
		Where("l.status = ? AND l.deleted_at IS NULL", model.LeaseStatusActive)
	if ownerID != 0 {
		query = query.Where("p.owner_id = ?", ownerID)
	}
	if managerID != 0 {
		query = query.Where("p.manager_id = ?", managerID)
	}
	if propertyID != 0 {
		query = query.Where("l.property_id = ?", propertyID)
	}
	var contacts []TenantContact
	err := query.Order("p.name ASC, l.unit_number ASC, t.id ASC").Scan(&contacts).Error
	return contacts, err
}
