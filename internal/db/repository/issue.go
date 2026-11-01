package repository

import (
	"errors"
	"log/slog"

	"github.com/patrickishaf/property-mgt/internal/db/model"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type IssueRepository struct {
	db     *gorm.DB
	logger *slog.Logger
}

// IssueFilter narrows an issue listing. Zero values are ignored. OwnerID and
// ManagerID scope results to that admin's properties.
type IssueFilter struct {
	Category   string
	ManagerID  uint
	OwnerID    uint
	PropertyID uint
	Status     string
	Urgency    string
}

func CreateIssueRepository(db *gorm.DB, logger *slog.Logger) *IssueRepository {
	return &IssueRepository{
		db:     db,
		logger: logger,
	}
}

func (this *IssueRepository) SaveOneTRX(issue *model.Issue, trx *gorm.DB) error {
	return trx.Create(issue).Error
}

func (this *IssueRepository) FindOneById(id uint) (*model.Issue, error) {
	var i model.Issue
	err := this.db.Where("id = ?", id).First(&i).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &i, nil
}

// FindOneByIdForUpdateTRX loads the issue and locks it for the rest of the transaction.
func (this *IssueRepository) FindOneByIdForUpdateTRX(id uint, trx *gorm.DB) (*model.Issue, error) {
	var i model.Issue
	err := trx.Clauses(clause.Locking{Strength: "UPDATE"}).Where("id = ?", id).First(&i).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &i, nil
}

func (this *IssueRepository) FindMany(filter IssueFilter, page int, limit int) ([]model.Issue, int64, error) {
	query := scopeToProperties(this.db, this.db.Model(&model.Issue{}), filter.OwnerID, filter.ManagerID)
	if filter.PropertyID != 0 {
		query = query.Where("property_id = ?", filter.PropertyID)
	}
	for column, value := range map[string]string{
		"status": filter.Status, "category": filter.Category, "urgency": filter.Urgency,
	} {
		if value != "" {
			query = query.Where(column+" = ?", value)
		}
	}
	query = query.Session(&gorm.Session{})

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var issues []model.Issue
	err := query.Order("id DESC").Offset((page - 1) * limit).Limit(limit).Find(&issues).Error
	if err != nil {
		return nil, 0, err
	}
	return issues, total, nil
}

func (this *IssueRepository) UpdateStatusTRX(id uint, status string, trx *gorm.DB) error {
	return trx.Model(&model.Issue{}).Where("id = ?", id).Update("status", status).Error
}
