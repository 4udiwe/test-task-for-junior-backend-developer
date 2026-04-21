package task

import (
	"context"

	recurrence "example.com/taskservice/internal/domain/recurrence"
	scheduler "example.com/taskservice/internal/domain/scheduler"
	taskdomain "example.com/taskservice/internal/domain/task"
)

type Repository interface {
	Create(ctx context.Context, task *taskdomain.Task) (*taskdomain.Task, error)
	GetByID(ctx context.Context, id int64) (*taskdomain.Task, error)
	Update(ctx context.Context, task *taskdomain.Task) (*taskdomain.Task, error)
	Delete(ctx context.Context, id int64) error
	List(ctx context.Context) ([]taskdomain.Task, error)
}

type RecurrenceRuleRepository interface {
	Create(ctx context.Context, rule *recurrence.Rule) (*recurrence.Rule, error)
	GetByID(ctx context.Context, id int64) (*recurrence.Rule, error)
	Update(ctx context.Context, rule *recurrence.Rule) (*recurrence.Rule, error)
	List(ctx context.Context, enabled *bool) ([]recurrence.Rule, error)
	Delete(ctx context.Context, id int64) error
}

type JobRepository interface {
	Create(ctx context.Context, job *scheduler.Job) (*scheduler.Job, error)
	GetByID(ctx context.Context, id int64) (*scheduler.Job, error)
	Update(ctx context.Context, job *scheduler.Job) (*scheduler.Job, error)
	
	// GetPendingJobs находит задания которые нужно выполнить
	// SELECT * WHERE status IN ('pending', 'retry')
	//              AND scheduled_at <= NOW()
	//              AND (locked_until IS NULL OR locked_until < NOW())
	GetPendingJobs(ctx context.Context, limit int) ([]scheduler.Job, error)
	
	// AcquireLock пытается захватить лок на задание для обработки
	// Возвращает (true, nil) если лок захвачен, (false, nil) если уже залочено
	AcquireLock(ctx context.Context, jobID int64, workerID string, lockDuration int64) (bool, error)
	
	// ReleaseLock снимает лок с задания
	ReleaseLock(ctx context.Context, jobID int64) error
}

type Usecase interface {
	Create(ctx context.Context, input CreateInput) (*taskdomain.Task, error)
	GetByID(ctx context.Context, id int64) (*taskdomain.Task, error)
	Update(ctx context.Context, id int64, input UpdateInput) (*taskdomain.Task, error)
	Delete(ctx context.Context, id int64) error
	List(ctx context.Context) ([]taskdomain.Task, error)
}

type CreateInput struct {
	Title       string
	Description string
	Status      taskdomain.Status
}

type UpdateInput struct {
	Title       string
	Description string
	Status      taskdomain.Status
}

