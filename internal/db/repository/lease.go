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
