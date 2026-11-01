package repository

import (
	"log/slog"

	"github.com/patrickishaf/property-mgt/internal/db/model"
	"gorm.io/gorm"
)

type ComplaintUpdateRepository struct {
	db     *gorm.DB
	logger *slog.Logger
}

func CreateComplaintUpdateRepository(db *gorm.DB, logger *slog.Logger) *ComplaintUpdateRepository {
	return &ComplaintUpdateRepository{
		db:     db,
		logger: logger,
	}
}

func (this *ComplaintUpdateRepository) SaveManyTRX(updates []model.ComplaintUpdate, trx *gorm.DB) error {
	if len(updates) == 0 {
		return nil
	}
	return trx.Create(&updates).Error
}

// FindByComplaint returns the complaint's timeline, oldest first.
func (this *ComplaintUpdateRepository) FindByComplaint(complaintID uint) ([]model.ComplaintUpdate, error) {
	var updates []model.ComplaintUpdate
	err := this.db.Where("complaint_id = ?", complaintID).Order("id ASC").Find(&updates).Error
	return updates, err
}
