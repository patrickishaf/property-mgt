package repository

import (
	"errors"
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
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &a, nil
}

func (this *AdminRepository) FindOneByEmail(email string) (*model.Admin, error) {
	var a model.Admin
	err := this.db.Where(&model.Admin{Email: email}).First(&a).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &a, nil
}

func (this *AdminRepository) FindOneByEmailAndRole(email string, role common.AdminRole) (*model.Admin, error) {
	var a model.Admin
	err := this.db.Where(&model.Admin{Email: email, Role: role.ToString()}).First(&a).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &a, nil
}

func (this *AdminRepository) SaveOne(admin *model.Admin) error {
	return this.db.Create(admin).Error
}

func (this *AdminRepository) SaveOneTRX(admin *model.Admin, trx *gorm.DB) error {
	return trx.Create(admin).Error
}

func (this *AdminRepository) FindManagersByOwner(ownerID uint, page int, limit int) ([]model.Admin, int64, error) {
	query := this.db.Model(&model.Admin{}).
		Where("owner_id = ? AND role = ?", ownerID, common.AdminRoleManager.ToString()).
		Session(&gorm.Session{})

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var managers []model.Admin
	err := query.Order("id DESC").Offset((page - 1) * limit).Limit(limit).Find(&managers).Error
	if err != nil {
		return nil, 0, err
	}
	return managers, total, nil
}

func (this *AdminRepository) UpdatePasswordHash(id uint, hash string) error {
	return this.db.Model(&model.Admin{}).Where("id = ?", id).Update("password_hash", hash).Error
}

func (this *AdminRepository) UpdateProfile(id uint, fields map[string]any) error {
	return this.db.Model(&model.Admin{}).Where("id = ?", id).Updates(fields).Error
}

func (this *AdminRepository) DeleteTRX(id uint, trx *gorm.DB) error {
	return trx.Where("id = ?", id).Delete(&model.Admin{}).Error
}
