package scheduler

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"time"

	recurrence "example.com/taskservice/internal/domain/recurrence"
	scheduler "example.com/taskservice/internal/domain/scheduler"
	taskdomain "example.com/taskservice/internal/domain/task"
	taskrepo "example.com/taskservice/internal/usecase/task"
)

// Scheduler управляет создание экземпляров периодических задач
type Scheduler struct {
	logger              *slog.Logger
	jobRepo             taskrepo.JobRepository
	recurrenceRuleRepo  taskrepo.RecurrenceRuleRepository
	taskRepo            taskrepo.Repository
	dateCalculator      *recurrence.DateCalculator
	pollInterval        time.Duration
	lockDuration        time.Duration
	maxConcurrentJobs   int
	workerID            string
	isRunning           bool
	stopChan            chan struct{}
}

// NewScheduler создает новый экземпляр планировщика
func NewScheduler(
	logger *slog.Logger,
	jobRepo taskrepo.JobRepository,
	recurrenceRuleRepo taskrepo.RecurrenceRuleRepository,
	taskRepo taskrepo.Repository,
	pollInterval time.Duration,
	lockDuration time.Duration,
) *Scheduler {
	if pollInterval == 0 {
		pollInterval = 1 * time.Hour
	}
	if lockDuration == 0 {
		lockDuration = 5 * time.Minute
	}

	return &Scheduler{
		logger:             logger,
		jobRepo:            jobRepo,
		recurrenceRuleRepo: recurrenceRuleRepo,
		taskRepo:           taskRepo,
		dateCalculator:     &recurrence.DateCalculator{},
		pollInterval:       pollInterval,
		lockDuration:       lockDuration,
		maxConcurrentJobs:  10,
		workerID:           fmt.Sprintf("scheduler-%d-%d", os.Getpid(), time.Now().Unix()),
		stopChan:           make(chan struct{}),
	}
}

// Start запускает планировщик в background goroutine с graceful shutdown поддержкой
func (s *Scheduler) Start(ctx context.Context) {
	if s.isRunning {
		s.logger.Warn("scheduler already running")
		return
	}

	s.isRunning = true
	s.logger.Info("scheduler started", "worker_id", s.workerID, "poll_interval", s.pollInterval)

	go s.run(ctx)
}

// run - основной loop планировщика
func (s *Scheduler) run(ctx context.Context) {
	defer func() {
		s.isRunning = false
		s.logger.Info("scheduler stopped")
	}()

	ticker := time.NewTicker(s.pollInterval)
	defer ticker.Stop()

	// Сразу обработаем пропущенные таски при старте
	if err := s.processPendingJobs(ctx); err != nil {
		s.logger.Error("initial processing failed", "error", err)
	}

	for {
		select {
		case <-ctx.Done():
			s.logger.Info("scheduler context cancelled, stopping")
			return

		case <-s.stopChan:
			s.logger.Info("scheduler stop signal received")
			return

		case <-ticker.C:
			if err := s.processPendingJobs(ctx); err != nil {
				s.logger.Error("process pending jobs failed", "error", err)
			}
		}
	}
}

// Stop отправляет сигнал к остановке планировщика (graceful shutdown)
func (s *Scheduler) Stop() {
	if s.isRunning {
		close(s.stopChan)
	}
}

// processPendingJobs обрабатывает все задания которые нужно выполнить
func (s *Scheduler) processPendingJobs(ctx context.Context) error {
	// Получаем пендинг задания
	jobs, err := s.jobRepo.GetPendingJobs(ctx, s.maxConcurrentJobs)
	if err != nil {
		return fmt.Errorf("get pending jobs: %w", err)
	}

	if len(jobs) == 0 {
		return nil // Нет задании для обработки
	}

	s.logger.Info("processing pending jobs", "count", len(jobs))

	for _, job := range jobs {
		// Пытаемся захватить лок
		acquired, err := s.jobRepo.AcquireLock(ctx, job.ID, s.workerID, int64(s.lockDuration.Seconds()))
		if err != nil {
			s.logger.Error("failed to acquire lock", "job_id", job.ID, "error", err)
			continue
		}

		if !acquired {
			// Уже залочено другим worker'ом
			continue
		}

		// Обработаем задание
		if err := s.processJob(ctx, job); err != nil {
			s.logger.Error("failed to process job", "job_id", job.ID, "error", err)
			s.markJobFailed(ctx, job.ID, err.Error())
		} else {
			// Успешно обработано
			s.markJobSuccess(ctx, job.ID)
		}

		// Освобождаем лок
		if err := s.jobRepo.ReleaseLock(ctx, job.ID); err != nil {
			s.logger.Error("failed to release lock", "job_id", job.ID, "error", err)
		}
	}

	return nil
}

// processJob обрабатывает одно задание - создает экземпляр задачи и следующее задание
func (s *Scheduler) processJob(ctx context.Context, job scheduler.Job) error {
	// Получаем правило периодичности
	rule, err := s.recurrenceRuleRepo.GetByID(ctx, job.RecurrenceRuleID)
	if err != nil {
		return fmt.Errorf("get recurrence rule: %w", err)
	}

	if !rule.IsEnabled {
		s.logger.Info("recurrence rule is disabled, skipping", "rule_id", rule.ID)
		return nil
	}

	// Создаем новый экземпляр задачи
	now := time.Now().UTC()
	newTask := &taskdomain.Task{
		Title:            rule.Name,
		Description:      "Generated from recurrence rule",
		Status:           taskdomain.StatusNew,
		RecurrenceRuleID: &rule.ID,
		CreatedAt:        now,
		UpdatedAt:        now,
	}

	created, err := s.taskRepo.Create(ctx, newTask)
	if err != nil {
		return fmt.Errorf("create task instance: %w", err)
	}

	s.logger.Info("task instance created", "task_id", created.ID, "rule_id", rule.ID)

	// Вычисляем следующую дату события
	nextDate, err := s.dateCalculator.CalculateNextOccurrence(rule, job.ScheduledAt)
	if err != nil {
		return fmt.Errorf("calculate next occurrence: %w", err)
	}

	// Обновляем правило с новой датой
	rule.NextOccurrenceDate = nextDate
	if _, err := s.recurrenceRuleRepo.Update(ctx, rule); err != nil {
		return fmt.Errorf("update recurrence rule: %w", err)
	}

	s.logger.Info("recurrence rule updated", "rule_id", rule.ID, "next_date", nextDate)

	// Создаем следующее задание в очереди
	nextJob := &scheduler.Job{
		RecurrenceRuleID: rule.ID,
		Status:           scheduler.JobStatusPending,
		ScheduledAt:      nextDate,
		MaxRetries:       3,
		CreatedAt:        now,
		UpdatedAt:        now,
	}

	if _, err := s.jobRepo.Create(ctx, nextJob); err != nil {
		return fmt.Errorf("create next job: %w", err)
	}

	s.logger.Info("next job created", "rule_id", rule.ID, "scheduled_at", nextDate)

	return nil
}

// markJobSuccess помечает задание как успешно выполненное
func (s *Scheduler) markJobSuccess(ctx context.Context, jobID int64) {
	now := time.Now().UTC()
	completedAt := now

	job := &scheduler.Job{
		ID:          jobID,
		Status:      scheduler.JobStatusSuccess,
		CompletedAt: &completedAt,
		UpdatedAt:   now,
	}

	if _, err := s.jobRepo.Update(ctx, job); err != nil {
		s.logger.Error("failed to mark job success", "job_id", jobID, "error", err)
	}
}

// markJobFailed помечает задание как неудачное
func (s *Scheduler) markJobFailed(ctx context.Context, jobID int64, errMsg string) {
	job := &scheduler.Job{
		ID:           jobID,
		Status:       scheduler.JobStatusFailed,
		ErrorMessage: &errMsg,
		UpdatedAt:    time.Now().UTC(),
	}

	if _, err := s.jobRepo.Update(ctx, job); err != nil {
		s.logger.Error("failed to mark job failed", "job_id", jobID, "error", err)
	}
}
