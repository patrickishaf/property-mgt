package repository

import (
	"errors"
	"log/slog"
	"sync"

	"github.com/patrickishaf/property-mgt/internal/db/model"
	"gorm.io/gorm"
)

type TenantRepository struct {
	db     *gorm.DB
	logger *slog.Logger
	mu     sync.Mutex
}

// TenantFilter scopes a tenant listing. OwnerID and ManagerID restrict results to
// tenants on properties owned or managed by that admin; zero values are ignored.
type TenantFilter struct {
	ManagerID  uint
	OwnerID    uint
	PropertyID uint
}

func CreateTenantRepository(db *gorm.DB, logger *slog.Logger) *TenantRepository {
	return &TenantRepository{
		db:     db,
		logger: logger,
	}
}

func (this *TenantRepository) FindOneByEmail(email string) (*model.Tenant, error) {
	var a model.Tenant
	err := this.db.Where("email = ?", email).First(&a).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &a, nil
}

func (this *TenantRepository) FindOneById(id uint) (*model.Tenant, error) {
	var a model.Tenant
	err := this.db.Where("id = ?", id).First(&a).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &a, nil
}

func (this *TenantRepository) FindMany(filter TenantFilter, page int, limit int) ([]model.Tenant, int64, error) {
	query := this.db.Model(&model.Tenant{})
	if filter.PropertyID != 0 {
		query = query.Where("property_id = ?", filter.PropertyID)
	}
	if filter.OwnerID != 0 {
		query = query.Where("property_id IN (?)", this.db.Model(&model.Property{}).Select("id").Where("owner_id = ?", filter.OwnerID))
	}
	if filter.ManagerID != 0 {
		query = query.Where("property_id IN (?)", this.db.Model(&model.Property{}).Select("id").Where("manager_id = ?", filter.ManagerID))
	}

	query = query.Session(&gorm.Session{})

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var tenants []model.Tenant
	err := query.Order("id DESC").Offset((page - 1) * limit).Limit(limit).Find(&tenants).Error
	if err != nil {
		return nil, 0, err
	}
	return tenants, total, nil
}

func (this *TenantRepository) SaveOneTRX(tenant *model.Tenant, trx *gorm.DB) error {
	return trx.Create(tenant).Error
}

func (this *TenantRepository) UpdateLeaseIDTRX(id uint, leaseID uint, trx *gorm.DB) error {
	return trx.Model(&model.Tenant{}).Where("id = ?", id).Update("lease_id", leaseID).Error
}

func (this *TenantRepository) UpdatePasswordHash(id uint, hash string) error {
	return this.db.Model(&model.Tenant{}).Where("id = ?", id).Update("password_hash", hash).Error
}

func (this *TenantRepository) UpdateProfile(id uint, fields map[string]any) error {
	return this.db.Model(&model.Tenant{}).Where("id = ?", id).Updates(fields).Error
}
