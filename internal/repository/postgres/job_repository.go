package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	scheduler "example.com/taskservice/internal/domain/scheduler"
)

type JobRepository struct {
	pool *pgxpool.Pool
}

func NewJobRepository(pool *pgxpool.Pool) *JobRepository {
	return &JobRepository{pool: pool}
}

func (r *JobRepository) Create(ctx context.Context, job *scheduler.Job) (*scheduler.Job, error) {
	query := `
		INSERT INTO job_queue (recurrence_rule_id, status, scheduled_at, started_at, completed_at, error_message, retry_count, max_retries, locked_by, locked_until, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, NOW(), NOW())
		RETURNING id, recurrence_rule_id, status, scheduled_at, started_at, completed_at, error_message, retry_count, max_retries, locked_by, locked_until, created_at, updated_at
	`

	row := r.pool.QueryRow(
		ctx,
		query,
		job.RecurrenceRuleID,
		job.Status,
		job.ScheduledAt,
		job.StartedAt,
		job.CompletedAt,
		job.ErrorMessage,
		job.RetryCount,
		job.MaxRetries,
		job.LockedBy,
		job.LockedUntil,
	)

	return scanJob(row)
}

func (r *JobRepository) GetByID(ctx context.Context, id int64) (*scheduler.Job, error) {
	query := `
		SELECT id, recurrence_rule_id, status, scheduled_at, started_at, completed_at, error_message, retry_count, max_retries, locked_by, locked_until, created_at, updated_at
		FROM job_queue
		WHERE id = $1
	`

	row := r.pool.QueryRow(ctx, query, id)
	found, err := scanJob(row)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("job not found")
		}
		return nil, err
	}

	return found, nil
}

func (r *JobRepository) Update(ctx context.Context, job *scheduler.Job) (*scheduler.Job, error) {
	query := `
		UPDATE job_queue
		SET recurrence_rule_id = $1,
		    status = $2,
		    scheduled_at = $3,
		    started_at = $4,
		    completed_at = $5,
		    error_message = $6,
		    retry_count = $7,
		    max_retries = $8,
		    locked_by = $9,
		    locked_until = $10,
		    updated_at = NOW()
		WHERE id = $11
		RETURNING id, recurrence_rule_id, status, scheduled_at, started_at, completed_at, error_message, retry_count, max_retries, locked_by, locked_until, created_at, updated_at
	`

	row := r.pool.QueryRow(
		ctx,
		query,
		job.RecurrenceRuleID,
		job.Status,
		job.ScheduledAt,
		job.StartedAt,
		job.CompletedAt,
		job.ErrorMessage,
		job.RetryCount,
		job.MaxRetries,
		job.LockedBy,
		job.LockedUntil,
		job.ID,
	)

	updated, err := scanJob(row)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("job not found")
		}
		return nil, err
	}

	return updated, nil
}

func (r *JobRepository) GetPendingJobs(ctx context.Context, limit int) ([]scheduler.Job, error) {
	query := `
		SELECT id, recurrence_rule_id, status, scheduled_at, started_at, completed_at, error_message, retry_count, max_retries, locked_by, locked_until, created_at, updated_at
		FROM job_queue
		WHERE status IN ('pending', 'retry')
		AND scheduled_at <= NOW()
		AND (locked_until IS NULL OR locked_until < NOW())
		ORDER BY scheduled_at ASC
		LIMIT $1
	`

	rows, err := r.pool.Query(ctx, query, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	jobs := make([]scheduler.Job, 0, limit)
	for rows.Next() {
		job, err := scanJob(rows)
		if err != nil {
			return nil, err
		}
		jobs = append(jobs, *job)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return jobs, nil
}

func (r *JobRepository) AcquireLock(ctx context.Context, jobID int64, workerID string, lockDuration int64) (bool, error) {
	// lockDuration в секундах
	query := `
		UPDATE job_queue
		SET status = $1,
		    locked_by = $2,
		    locked_until = NOW() + INTERVAL '1 second' * $3,
		    started_at = NOW(),
		    updated_at = NOW()
		WHERE id = $4
		AND (locked_until IS NULL OR locked_until < NOW())
		RETURNING id
	`

	row := r.pool.QueryRow(ctx, query, scheduler.JobStatusProcessing, workerID, lockDuration, jobID)
	var resultID int64
	if err := row.Scan(&resultID); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return false, nil // Already locked by another worker
		}
		return false, err
	}

	return true, nil
}

func (r *JobRepository) ReleaseLock(ctx context.Context, jobID int64) error {
	query := `
		UPDATE job_queue
		SET locked_by = NULL,
		    locked_until = NULL,
		    updated_at = NOW()
		WHERE id = $1
	`

	_, err := r.pool.Exec(ctx, query, jobID)
	return err
}

type jobScanner interface {
	Scan(dest ...interface{}) error
}

func scanJob(scanner jobScanner) (*scheduler.Job, error) {
	var (
		job    scheduler.Job
		status string
	)

	if err := scanner.Scan(
		&job.ID,
		&job.RecurrenceRuleID,
		&status,
		&job.ScheduledAt,
		&job.StartedAt,
		&job.CompletedAt,
		&job.ErrorMessage,
		&job.RetryCount,
		&job.MaxRetries,
		&job.LockedBy,
		&job.LockedUntil,
		&job.CreatedAt,
		&job.UpdatedAt,
	); err != nil {
		return nil, err
	}

	job.Status = scheduler.JobStatus(status)

	return &job, nil
}
