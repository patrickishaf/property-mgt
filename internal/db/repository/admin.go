package repository

import (
	"log/slog"
	"sync"

	"github.com/patrickishaf/property-mgt/common"
	"github.com/patrickishaf/property-mgt/internal/db/model"
	"gorm.io/gorm"
)

type AdminRepository struct {
	db     *gorm.DB
	logger *slog.Logger
	mu     sync.Mutex
}

func CreateAdminRepository(db *gorm.DB, logger *slog.Logger) *AdminRepository {
	return &AdminRepository{
		db:     db,
		logger: logger,
	}
}

func (this *AdminRepository) FindOneById(id uint) (*model.Admin, error) {
	var a model.Admin
	err := this.db.Where("id = ?", id).First(&a).Error
	if err != nil {
		if err.Error() == "record not found" {
			return nil, nil
		}
		return nil, err
	}
	return &a, nil
}

func (this *AdminRepository) FindOneByEmail(email string, role common.AdminRole) (*model.Admin, error) {
	var a model.Admin
	err := this.db.Where(&model.Admin{Email: email, Role: role.ToString()}).First(&a).Error
	if err != nil {
		if err.Error() == "record not found" {
			return nil, nil
		}
		return nil, err
	}
	return &a, nil
}

func (this *AdminRepository) SaveOne(admin *model.Admin) error {
	return this.db.Create(admin).Error
}
