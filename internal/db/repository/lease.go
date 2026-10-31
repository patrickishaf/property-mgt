package repository

import (
	"context"
	"errors"
	"log/slog"

	"github.com/patrickishaf/property-mgt/common"
	"github.com/patrickishaf/property-mgt/internal/db/model"
	"gorm.io/gorm"
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
