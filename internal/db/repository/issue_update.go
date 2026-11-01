package repository

import (
	"log/slog"

	"github.com/patrickishaf/property-mgt/internal/db/model"
	"gorm.io/gorm"
)

type IssueUpdateRepository struct {
	db     *gorm.DB
	logger *slog.Logger
}

func CreateIssueUpdateRepository(db *gorm.DB, logger *slog.Logger) *IssueUpdateRepository {
	return &IssueUpdateRepository{
		db:     db,
		logger: logger,
	}
}

func (this *IssueUpdateRepository) SaveOneTRX(update *model.IssueUpdate, trx *gorm.DB) error {
	return trx.Create(update).Error
}

// FindByIssue returns the issue's updates, oldest first.
func (this *IssueUpdateRepository) FindByIssue(issueID uint) ([]model.IssueUpdate, error) {
	var updates []model.IssueUpdate
	err := this.db.Where("issue_id = ?", issueID).Order("id ASC").Find(&updates).Error
	return updates, err
}
