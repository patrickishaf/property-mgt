package repository

import (
	"errors"
	"log/slog"

	"github.com/patrickishaf/property-mgt/internal/db/model"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type ComplaintRepository struct {
	db     *gorm.DB
	logger *slog.Logger
}

// ComplaintFilter narrows a complaint listing. Zero values are ignored.
// OwnerID and ManagerID scope results to that admin's properties.
type ComplaintFilter struct {
	Category   string
	IssueID    uint
	Linked     *bool
	ManagerID  uint
	OwnerID    uint
	PropertyID uint
	Status     string
	TenantID   uint
	Urgency    string
}

func CreateComplaintRepository(db *gorm.DB, logger *slog.Logger) *ComplaintRepository {
	return &ComplaintRepository{
		db:     db,
		logger: logger,
	}
}

func (this *ComplaintRepository) SaveOne(complaint *model.Complaint) error {
	return this.db.Create(complaint).Error
}

func (this *ComplaintRepository) FindOneById(id uint) (*model.Complaint, error) {
	var c model.Complaint
	err := this.db.Where("id = ?", id).First(&c).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &c, nil
}

// FindByIdsForUpdateTRX loads the complaints and locks them for the rest of the transaction.
func (this *ComplaintRepository) FindByIdsForUpdateTRX(ids []uint, trx *gorm.DB) ([]model.Complaint, error) {
	var complaints []model.Complaint
	err := trx.Clauses(clause.Locking{Strength: "UPDATE"}).Where("id IN ?", ids).Order("id ASC").Find(&complaints).Error
	return complaints, err
}

func (this *ComplaintRepository) FindByIssue(issueID uint) ([]model.Complaint, error) {
	var complaints []model.Complaint
	err := this.db.Where("issue_id = ?", issueID).Order("id ASC").Find(&complaints).Error
	return complaints, err
}

func (this *ComplaintRepository) FindByIssueTRX(issueID uint, trx *gorm.DB) ([]model.Complaint, error) {
	var complaints []model.Complaint
	err := trx.Where("issue_id = ?", issueID).Order("id ASC").Find(&complaints).Error
	return complaints, err
}

func (this *ComplaintRepository) FindMany(filter ComplaintFilter, page int, limit int) ([]model.Complaint, int64, error) {
	query := scopeToProperties(this.db, this.db.Model(&model.Complaint{}), filter.OwnerID, filter.ManagerID)
	if filter.TenantID != 0 {
		query = query.Where("creator_id = ?", filter.TenantID)
	}
	if filter.PropertyID != 0 {
		query = query.Where("property_id = ?", filter.PropertyID)
	}
	if filter.IssueID != 0 {
		query = query.Where("issue_id = ?", filter.IssueID)
	}
	for column, value := range map[string]string{
		"status": filter.Status, "category": filter.Category, "urgency": filter.Urgency,
	} {
		if value != "" {
			query = query.Where(column+" = ?", value)
		}
	}
	if filter.Linked != nil {
		if *filter.Linked {
			query = query.Where("issue_id IS NOT NULL")
		} else {
			query = query.Where("issue_id IS NULL")
		}
	}
	query = query.Session(&gorm.Session{})

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var complaints []model.Complaint
	err := query.Order("id DESC").Offset((page - 1) * limit).Limit(limit).Find(&complaints).Error
	if err != nil {
		return nil, 0, err
	}
	return complaints, total, nil
}

// CountByIssues returns how many complaints are linked to each issue.
func (this *ComplaintRepository) CountByIssues(issueIDs []uint) (map[uint]int64, error) {
	counts := make(map[uint]int64, len(issueIDs))
	if len(issueIDs) == 0 {
		return counts, nil
	}
	var rows []struct {
		IssueID uint
		Count   int64
	}
	err := this.db.Model(&model.Complaint{}).Select("issue_id, COUNT(*) AS count").
		Where("issue_id IN ?", issueIDs).Group("issue_id").Scan(&rows).Error
	if err != nil {
		return nil, err
	}
	for _, r := range rows {
		counts[r.IssueID] = r.Count
	}
	return counts, nil
}

// LinkToIssueTRX links the complaints to the issue and gives them its status.
func (this *ComplaintRepository) LinkToIssueTRX(ids []uint, issueID uint, status string, trx *gorm.DB) error {
	return trx.Model(&model.Complaint{}).Where("id IN ?", ids).
		Updates(map[string]any{"issue_id": issueID, "status": status}).Error
}

// UpdateStatusByIssueTRX sets the status of every complaint linked to the issue.
func (this *ComplaintRepository) UpdateStatusByIssueTRX(issueID uint, status string, trx *gorm.DB) error {
	return trx.Model(&model.Complaint{}).Where("issue_id = ?", issueID).Update("status", status).Error
}
