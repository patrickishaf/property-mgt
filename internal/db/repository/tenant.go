package repository

import (
	"log/slog"
	"reflect"
	"sync"

	"github.com/patrickishaf/property-mgt/internal/db/model"
	"gorm.io/gorm"
)

type TenantRepository struct {
	db     *gorm.DB
	logger *slog.Logger
	mu     sync.Mutex
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
		return nil, err
	}
	if reflect.DeepEqual(a, model.Tenant{}) {
		return nil, nil
	}
	return &a, nil
}

func (this *TenantRepository) FindOneById(id uint) (*model.Tenant, error) {
	var a model.Tenant
	err := this.db.Where("id = ?", id).First(&a).Error
	if err != nil {
		return nil, err
	}
	if reflect.DeepEqual(a, model.Tenant{}) {
		return nil, nil
	}
	return &a, nil
}
