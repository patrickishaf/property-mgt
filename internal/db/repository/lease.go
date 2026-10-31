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
