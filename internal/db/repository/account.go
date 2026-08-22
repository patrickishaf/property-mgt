package repository

import (
	"errors"
	"log/slog"

	"github.com/patrickishaf/property-mgt/common"
	"github.com/patrickishaf/property-mgt/internal/db/model"
	"gorm.io/gorm"
)

type AccountRepository struct {
	db     *gorm.DB
	logger *slog.Logger
}

func CreateAccountRepo(db *gorm.DB, logger *slog.Logger) *AccountRepository {
	return &AccountRepository{
		db:     db,
		logger: logger,
	}
}

func (this *AccountRepository) SaveOne(acc *model.Account) error {
	logger := this.logger.With("caller", "AccountRepository.SaveOne")
	err := this.db.Create(acc).Error
	if err != nil {
		logger.Error("failed to insert account into the db", "err", err.Error())
		return errors.New(common.ErrDBInsertFailed)
	}
	return nil
}
