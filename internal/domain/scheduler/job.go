package scheduler

import "time"

// JobStatus определяет статус выполнения периодического задания
type JobStatus string

const (
	// JobStatusPending - ждет выполнения
	JobStatusPending JobStatus = "pending"
	// JobStatusProcessing - в процессе выполнения
	JobStatusProcessing JobStatus = "processing"
	// JobStatusSuccess - выполнено успешно
	JobStatusSuccess JobStatus = "success"
	// JobStatusFailed - ошибка выполнения (без повторов)
	JobStatusFailed JobStatus = "failed"
	// JobStatusRetry - ожидание повторного попытка
	JobStatusRetry JobStatus = "retry"
)

// Job представляет отдельное задание для создания экземпляра периодической задачи
type Job struct {
	ID               int64
	RecurrenceRuleID int64
	Status           JobStatus
	ScheduledAt      time.Time // когда должна была выполниться
	StartedAt        *time.Time
	CompletedAt      *time.Time
	ErrorMessage     *string
	RetryCount       int
	MaxRetries       int
	LockedBy         *string    // worker_id если в процессе
	LockedUntil      *time.Time // когда lock истекает
	CreatedAt        time.Time
	UpdatedAt        time.Time
}
