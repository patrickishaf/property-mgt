package repository

import (
	"errors"
	"log/slog"
	"sync"
	"time"

	"github.com/patrickishaf/property-mgt/common"
	"github.com/patrickishaf/property-mgt/config"
	"github.com/patrickishaf/property-mgt/internal/db/model"
	"gorm.io/gorm"
)

type JobRepository struct {
	cfg    *config.AppConfig
	db     *gorm.DB
	logger *slog.Logger
	mu     sync.Mutex
}

func CreateJobRepository(cfg *config.AppConfig, db *gorm.DB, logger *slog.Logger) *JobRepository {
	return &JobRepository{
		cfg:    cfg,
		db:     db,
		logger: logger,
	}
}

func (this *JobRepository) DeleteJob(jobID uint) error {
	logger := this.logger.With("caller", "JobRepository.DeleteJob")
	err := this.db.Where("id = ?", jobID).Delete(&model.Job{}).Error
	if err != nil {
		logger.Error("failed to delete job", "err", err.Error())
		return err
	}
	return nil
}

func (this *JobRepository) FindNextPendingJob() (*model.Job, error) {
	this.mu.Lock()
	defer this.mu.Unlock()
	logger := this.logger.With("caller", "JobRepository.FindNextPendingJob")

	var job model.Job
	err := this.db.Where(
		"status = ? AND attempt_count < ? AND due_date <= ?",
		common.ProcessingStatusPending,
		this.cfg.MaxJobAttemptCount,
		time.Now(),
	).Order("last_scheduled_at ASC").Limit(1).Find(&job).Error
	if err != nil {
		logger.Error("failed to find next pending job", "err", err.Error())
		return nil, errors.New(common.ErrDBRecordNotFound)
	}
	return &job, nil
}

func (this *JobRepository) MarkJobAsProcessing(jobID int, attemptCount int) error {
	this.mu.Lock()
	defer this.mu.Unlock()
	logger := this.logger.With("caller", "JobRepository.MarkJobAsProcessing")
	var j model.Job
	err := this.db.Model(&j).Where("id = ?", jobID).Updates(model.Job{
		AttemptCount: attemptCount,
		Status:       common.ProcessingStatusProcessing.String(),
	}).Error
	if err != nil {
		logger.Error("failed to mark job as processing", "err", err.Error())
		return errors.New(common.ErrDBOperationFailed)
	}
	return nil
}

func (this *JobRepository) SaveDefaultJob(job *model.Job) (*model.Job, error) {
	logger := this.logger.With("caller", "JobRepository.SaveDefaultJob")
	if err := this.db.Create(job).Error; err != nil {
		logger.Error("failed to save default job", "err", err.Error())
		return nil, errors.New(common.ErrDBOperationFailed)
	}
	return job, nil
}

func (this *JobRepository) SaveDefaultJobTRX(job *model.Job, trx *gorm.DB) (*model.Job, error) {
	logger := this.logger.With("caller", "JobRepository.SaveDefaultJobTRX")
	if err := trx.Create(job).Error; err != nil {
		logger.Error("failed to save default job", "err", err.Error())
		return nil, errors.New(common.ErrDBOperationFailed)
	}
	return job, nil
}

func (this *JobRepository) UpdateAndRescheduleJob(jobID uint, status common.ProcessingStatus, attemptCount int, nextDueDate time.Time) error {
	this.mu.Lock()
	defer this.mu.Unlock()
	logger := this.logger.With("caller", "JobRepository.MarkJobAsProcessing")
	var job model.Job
	err := this.db.Model(job).Where("id = ?", jobID).Updates(model.Job{
		Status:          status.String(),
		AttemptCount:    attemptCount,
		LastScheduledAt: time.Now(),
		DueDate:         &nextDueDate,
	}).Error
	if err != nil {
		logger.Error("failed to update and reschedule job", "err", err.Error())
		return errors.New(common.ErrDBOperationFailed)
	}
	return nil
}

func (this *JobRepository) UpdateJob(jobID uint, status common.ProcessingStatus, attemptCount int) error {
	this.mu.Lock()
	defer this.mu.Unlock()
	logger := this.logger.With("caller", "JobRepository.MarkJobAsProcessing")
	var job model.Job
	err := this.db.Model(job).Where("id = ?", jobID).Updates(model.Job{
		Status:       status.String(),
		AttemptCount: attemptCount,
	}).Error
	if err != nil {
		logger.Error("failed to update job", "err", err.Error())
		return errors.New(common.ErrDBOperationFailed)
	}
	return nil
}

func (this *JobRepository) UpdateJobStatus(jobID uint, status common.ProcessingStatus) error {
	logger := this.logger.With("caller", "JobRepository.MarkJobAsProcessing")
	var job model.Job
	err := this.db.Model(job).Where("id = ?", jobID).Updates(model.Job{
		Status: status.String(),
	}).Error
	if err != nil {
		logger.Error("failed to update job status", "err", err.Error())
		return errors.New(common.ErrDBOperationFailed)
	}
	return nil
}
