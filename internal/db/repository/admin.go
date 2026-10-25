package repository

import (
	"log/slog"
	"reflect"
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

func (this *AdminRepository) FindOneByEmail(email string, role common.AdminRole) (*model.Admin, error) {
	var a model.Admin
	err := this.db.Where(&model.Admin{Email: email, Role: role.ToString()}).Find(&a).Error
	if err != nil {
		return nil, err
	}
	if reflect.DeepEqual(a, model.Admin{}) {
		return nil, nil
	}
	return &a, nil
}

func (this *AdminRepository) SaveOne(admin *model.Admin) error {
	return this.db.Create(admin).Error
}
