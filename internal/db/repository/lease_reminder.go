package repository

import (
	"errors"
	"log/slog"

	"github.com/patrickishaf/property-mgt/internal/db/model"
	"gorm.io/gorm"
)

type LeaseReminderRepository struct {
	db     *gorm.DB
	logger *slog.Logger
}

func CreateLeaseReminderRepository(db *gorm.DB, logger *slog.Logger) *LeaseReminderRepository {
	return &LeaseReminderRepository{db: db, logger: logger}
}

func (this *LeaseReminderRepository) SaveOne(reminder *model.LeaseReminder) error {
	return this.db.Create(reminder).Error
}

// FindLatest returns the most recent reminder of a kind for a lease, or nil.
func (this *LeaseReminderRepository) FindLatest(leaseID uint, kind string) (*model.LeaseReminder, error) {
	var r model.LeaseReminder
	err := this.db.Where("lease_id = ? AND kind = ?", leaseID, kind).Order("sent_at DESC").First(&r).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &r, nil
}

func (this *LeaseReminderRepository) Count(leaseID uint, kind string) (int64, error) {
	var count int64
	err := this.db.Model(&model.LeaseReminder{}).Where("lease_id = ? AND kind = ?", leaseID, kind).Count(&count).Error
	return count, err
}
