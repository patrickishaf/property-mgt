package repository

import (
	"errors"
	"log/slog"

	"github.com/patrickishaf/property-mgt/internal/db/model"
	"gorm.io/gorm"
)

type UnitRepository struct {
	db     *gorm.DB
	logger *slog.Logger
}

func CreateUnitRepository(db *gorm.DB, logger *slog.Logger) *UnitRepository {
	return &UnitRepository{
		db:     db,
		logger: logger,
	}
}

func (this *UnitRepository) SaveManyTRX(units []model.PropertyUnit, trx *gorm.DB) error {
	if len(units) == 0 {
		return nil
	}
	return trx.Create(&units).Error
}

func (this *UnitRepository) FindByProperty(propertyID uint) ([]model.PropertyUnit, error) {
	var units []model.PropertyUnit
	err := this.db.Where("property_id = ?", propertyID).Order("number ASC").Find(&units).Error
	return units, err
}

func (this *UnitRepository) FindOneById(propertyID uint, id uint) (*model.PropertyUnit, error) {
	var u model.PropertyUnit
	err := this.db.Where("property_id = ? AND id = ?", propertyID, id).First(&u).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &u, nil
}

func (this *UnitRepository) FindOneByPropertyAndNumber(propertyID uint, number string) (*model.PropertyUnit, error) {
	var u model.PropertyUnit
	err := this.db.Where("property_id = ? AND number = ?", propertyID, number).First(&u).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &u, nil
}

func (this *UnitRepository) Update(id uint, fields map[string]any) error {
	return this.db.Model(&model.PropertyUnit{}).Where("id = ?", id).Updates(fields).Error
}

// DeleteTRX hard-deletes the unit so its number can be reused.
func (this *UnitRepository) DeleteTRX(id uint, trx *gorm.DB) error {
	return trx.Unscoped().Where("id = ?", id).Delete(&model.PropertyUnit{}).Error
}
