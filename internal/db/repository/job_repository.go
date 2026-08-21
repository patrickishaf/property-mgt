package repository

import (
	"context"
	"errors"
	"log/slog"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/patrickishaf/property-mgt/common"
	"github.com/patrickishaf/property-mgt/config"
	"github.com/patrickishaf/property-mgt/internal/db/model"
)

type JobRepository struct {
	cfg      *config.AppConfig
	connPool *pgxpool.Pool
	logger   *slog.Logger
	mu       sync.Mutex
}

func CreateJobRepository(cfg *config.AppConfig, connPool *pgxpool.Pool, logger *slog.Logger) *JobRepository {
	return &JobRepository{
		cfg:      cfg,
		connPool: connPool,
		logger:   logger,
	}
}

func (this *JobRepository) DeleteJob(ctx context.Context, jobID uuid.UUID) error {
	logger := this.logger.With("caller", "JobRepository.DeleteJob")
	tag, err := this.connPool.Exec(ctx, `DELETE FROM jobs WHERE id = $1`, jobID)
	if err != nil {
		logger.Error("failed to delete job", "err", err.Error())
		return err
	}
	if tag.RowsAffected() == 0 {
		logger.Error("no job found with id", "job_id", jobID)
		return errors.New(common.ErrDBRecordNotFound)
	}
	return nil
}

func (this *JobRepository) FindNextPendingJob(ctx context.Context) (*model.Job, error) {
	this.mu.Lock()
	defer this.mu.Unlock()
	logger := this.logger.With("caller", "JobRepository.GetNextPendingJob")

	var job model.Job
	query := `
	SELECT id, created_at, updated_at, attempt_count, due_date, error, last_scheduled_at, status, type, payload
	FROM jobs
	WHERE status = $1 AND attempt_count < $2 AND due_date <= $3
	ORDER BY last_scheduled_at ASC
	LIMIT 1`
	err := this.connPool.QueryRow(
		ctx,
		query,
		common.ProcessingStatusPending,
		this.cfg.MaxJobAttemptCount,
		time.Now(),
	).Scan(
		&job.ID,
		&job.CreatedAt,
		&job.UpdatedAt,
		&job.AttemptCount,
		&job.DueDate,
		&job.Error,
		&job.LastScheduledAt,
		&job.Status,
		&job.Type,
		&job.Payload,
	)
	if err != nil {
		if err.Error() == common.ErrNoRowsInResultSet {
			return nil, errors.New(common.ErrNoRowsInResultSet)
		}
		logger.Error("failed to find next pending job", "err", err.Error())
		return nil, errors.New(common.ErrDBOperationFailed)
	}
	return &job, nil
}

func (this *JobRepository) MarkJobAsProcessing(ctx context.Context, jobID uuid.UUID, attemptCount int) error {
	this.mu.Lock()
	defer this.mu.Unlock()
	logger := this.logger.With("caller", "JobRepository.MarkJobAsProcessing")
	query := `
	UPDATE jobs
	SET status = $1, attempt_count = $2
	WHERE id = $3
	`
	tag, err := this.connPool.Exec(ctx, query, common.ProcessingStatusProcessing, attemptCount, jobID)
	if err != nil {
		logger.Error("failed to scan query result", "err", err.Error())
		return errors.New(common.ErrDBOperationFailed)
	}
	if tag.RowsAffected() == 0 {
		logger.Error("failed to mark job as processing", "err", "no matching rows to be affected")
		return errors.New(common.ErrNoRowsInResultSet)
	}
	return nil
}

func (this *JobRepository) SaveDefaultJob(ctx context.Context, job *model.Job) (*model.Job, error) {
	logger := this.logger.With("caller", "JobRepository.SaveDefaultJob")
	query := `
	INSERT INTO jobs(
	payload,
	type
	)
	VALUES ($1, $2)
	RETURNING id, status
	`
	err := this.connPool.QueryRow(
		ctx,
		query,
		&job.Payload,
		&job.Type,
	).Scan(&job.ID, &job.Status)
	if err != nil {
		logger.Error("failed to save default job", "err", err.Error())
		return nil, errors.New(common.ErrDBOperationFailed)
	}
	return job, nil
}

func (this *JobRepository) SaveDefaultJobTRX(ctx context.Context, job *model.Job, trx *pgx.Tx) (*model.Job, error) {
	logger := this.logger.With("caller", "JobRepository.SaveDefaultJobTRX")
	query := `
	INSERT INTO jobs(
	payload,
	type
	)
	VALUES ($1, $2)
	RETURNING id, status
	`
	err := (*trx).QueryRow(
		ctx,
		query,
		&job.Payload,
		&job.Type,
	).Scan(&job.ID, &job.Status)
	if err != nil {
		logger.Error("failed to save default job", "err", err.Error())
		return nil, errors.New(common.ErrDBOperationFailed)
	}
	return job, nil
}

func (this *JobRepository) UpdateAndRescheduleJob(ctx context.Context, jobID uuid.UUID, status common.ProcessingStatus, attemptCount int, nextDueDate time.Time) error {
	this.mu.Lock()
	defer this.mu.Unlock()
	logger := this.logger.With("caller", "JobRepository.MarkJobAsProcessing")
	query := `
	UPDATE jobs
	SET status = $1, attempt_count = $2, last_scheduled_at = $3, due_date = $4
	WHERE id = $5
	RETURNING id, created_at, updated_at, attempt_count, error, last_scheduled_at, status, type, payload
	`
	var job model.Job
	err := this.connPool.QueryRow(ctx, query, status, attemptCount, time.Now(), nextDueDate, jobID).Scan(
		&job.ID,
		&job.CreatedAt,
		&job.UpdatedAt,
		&job.AttemptCount,
		&job.Error,
		&job.LastScheduledAt,
		&job.Status,
		&job.Type,
		&job.Payload,
	)
	if err != nil {
		logger.Error("failed to scan query result", "err", err.Error())
		return errors.New(common.ErrDBOperationFailed)
	}
	return nil
}

func (this *JobRepository) UpdateJob(ctx context.Context, jobID uuid.UUID, status common.ProcessingStatus, attemptCount int) error {
	this.mu.Lock()
	defer this.mu.Unlock()
	logger := this.logger.With("caller", "JobRepository.MarkJobAsProcessing")
	query := `
	UPDATE jobs
	SET status = $1, attempt_count = $2
	WHERE id = $3
	RETURNING id, created_at, updated_at, attempt_count, error, last_scheduled_at, status, type, payload
	`
	var job model.Job
	err := this.connPool.QueryRow(ctx, query, status, attemptCount, jobID).Scan(
		&job.ID,
		&job.CreatedAt,
		&job.UpdatedAt,
		&job.AttemptCount,
		&job.Error,
		&job.LastScheduledAt,
		&job.Status,
		&job.Type,
		&job.Payload,
	)
	if err != nil {
		logger.Error("failed to scan query result", "err", err.Error())
		return errors.New(common.ErrDBOperationFailed)
	}
	return nil
}

func (this *JobRepository) UpdateJobStatus(ctx context.Context, jobID uuid.UUID, status common.ProcessingStatus) error {
	logger := this.logger.With("caller", "JobRepository.MarkJobAsProcessing")
	query := `
	UPDATE jobs
	SET status = $1
	WHERE id = $2
	RETURNING id, updated_at, status
	`
	var job model.Job
	err := this.connPool.QueryRow(ctx, query, status, jobID).Scan(
		&job.ID,
		&job.UpdatedAt,
		&job.Status,
	)
	if err != nil {
		logger.Error("failed to scan query result", "err", err.Error())
		return errors.New(common.ErrDBOperationFailed)
	}
	return nil
}
