package repository

import (
	"errors"
	"log/slog"

	"github.com/patrickishaf/property-mgt/internal/db/model"
	"gorm.io/gorm"
)

type PropertyRepository struct {
	db     *gorm.DB
	logger *slog.Logger
}

func CreatePropertyRepository(db *gorm.DB, logger *slog.Logger) *PropertyRepository {
	return &PropertyRepository{
		db:     db,
		logger: logger,
	}
}

func (this *PropertyRepository) FindOneById(id uint) (*model.Property, error) {
	var p model.Property
	err := this.db.Where("id = ?", id).First(&p).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &p, nil
}

func (this *PropertyRepository) IncrementOccupiedUnitsTRX(id uint, trx *gorm.DB) error {
	return trx.Model(&model.Property{}).Where("id = ?", id).
		Update("number_of_occupied_units", gorm.Expr("number_of_occupied_units + 1")).Error
}
