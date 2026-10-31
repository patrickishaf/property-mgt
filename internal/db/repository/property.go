package repository

import (
	"errors"
	"log/slog"
	"time"

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

// PropertyFilter scopes a property listing to an owner or a manager.
type PropertyFilter struct {
	IncludeArchived bool
	ManagerID       uint
	OwnerID         uint
}

func (this *PropertyRepository) SaveOneTRX(property *model.Property, trx *gorm.DB) error {
	return trx.Create(property).Error
}

func (this *PropertyRepository) FindMany(filter PropertyFilter, page int, limit int) ([]model.Property, int64, error) {
	query := this.db.Model(&model.Property{})
	if filter.OwnerID != 0 {
		query = query.Where("owner_id = ?", filter.OwnerID)
	}
	if filter.ManagerID != 0 {
		query = query.Where("manager_id = ?", filter.ManagerID)
	}
	if !filter.IncludeArchived {
		query = query.Where("archived_at IS NULL")
	}
	query = query.Session(&gorm.Session{})

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var properties []model.Property
	err := query.Order("id DESC").Offset((page - 1) * limit).Limit(limit).Find(&properties).Error
	if err != nil {
		return nil, 0, err
	}
	return properties, total, nil
}

func (this *PropertyRepository) Update(id uint, fields map[string]any) error {
	return this.db.Model(&model.Property{}).Where("id = ?", id).Updates(fields).Error
}

// SetManager assigns a manager to the property; managerID 0 clears it.
func (this *PropertyRepository) SetManager(id uint, managerID uint) error {
	return this.db.Model(&model.Property{}).Where("id = ?", id).Update("manager_id", managerID).Error
}

func (this *PropertyRepository) ClearManagerForAllTRX(managerID uint, trx *gorm.DB) error {
	return trx.Model(&model.Property{}).Where("manager_id = ?", managerID).Update("manager_id", 0).Error
}

func (this *PropertyRepository) Archive(id uint, at time.Time) error {
	return this.db.Model(&model.Property{}).Where("id = ?", id).Update("archived_at", at).Error
}

// AddUnitsTRX changes the property's unit count by delta, which may be negative.
func (this *PropertyRepository) AddUnitsTRX(id uint, delta int, trx *gorm.DB) error {
	return trx.Model(&model.Property{}).Where("id = ?", id).
		Update("number_of_units", gorm.Expr("number_of_units + ?", delta)).Error
}

// scopeToProperties limits a query on a table with a property_id column to the
// properties an owner owns or a manager manages. Zero IDs are ignored.
func scopeToProperties(db *gorm.DB, query *gorm.DB, ownerID uint, managerID uint) *gorm.DB {
	if ownerID != 0 {
		query = query.Where("property_id IN (?)", db.Model(&model.Property{}).Select("id").Where("owner_id = ?", ownerID))
	}
	if managerID != 0 {
		query = query.Where("property_id IN (?)", db.Model(&model.Property{}).Select("id").Where("manager_id = ?", managerID))
	}
	return query
}
