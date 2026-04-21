package task

import (
	"context"
	"fmt"
	"strings"
	"time"

	recurrence "example.com/taskservice/internal/domain/recurrence"
	scheduler "example.com/taskservice/internal/domain/scheduler"
	taskdomain "example.com/taskservice/internal/domain/task"
)

type Service struct {
	repo                    Repository
	recurrenceRuleRepo      RecurrenceRuleRepository
	jobRepo                 JobRepository
	now                     func() time.Time
	dateCalculator          *recurrence.DateCalculator
}

func NewService(repo Repository) *Service {
	return &Service{
		repo:           repo,
		now:            func() time.Time { return time.Now().UTC() },
		dateCalculator: &recurrence.DateCalculator{},
	}
}

// NewServiceWithRecurrence создает сервис с поддержкой периодических задач
func NewServiceWithRecurrence(repo Repository, recurrenceRepo RecurrenceRuleRepository, jobRepo JobRepository) *Service {
	return &Service{
		repo:               repo,
		recurrenceRuleRepo: recurrenceRepo,
		jobRepo:            jobRepo,
		now:                func() time.Time { return time.Now().UTC() },
		dateCalculator:     &recurrence.DateCalculator{},
	}
}

func (s *Service) Create(ctx context.Context, input CreateInput) (*taskdomain.Task, error) {
	normalized, err := validateCreateInput(input)
	if err != nil {
		return nil, err
	}

	model := &taskdomain.Task{
		Title:       normalized.Title,
		Description: normalized.Description,
		Status:      normalized.Status,
	}
	now := s.now()
	model.CreatedAt = now
	model.UpdatedAt = now

	created, err := s.repo.Create(ctx, model)
	if err != nil {
		return nil, err
	}

	return created, nil
}

func (s *Service) GetByID(ctx context.Context, id int64) (*taskdomain.Task, error) {
	if id <= 0 {
		return nil, fmt.Errorf("%w: id must be positive", ErrInvalidInput)
	}

	return s.repo.GetByID(ctx, id)
}

func (s *Service) Update(ctx context.Context, id int64, input UpdateInput) (*taskdomain.Task, error) {
	if id <= 0 {
		return nil, fmt.Errorf("%w: id must be positive", ErrInvalidInput)
	}

	normalized, err := validateUpdateInput(input)
	if err != nil {
		return nil, err
	}

	model := &taskdomain.Task{
		ID:          id,
		Title:       normalized.Title,
		Description: normalized.Description,
		Status:      normalized.Status,
		UpdatedAt:   s.now(),
	}

	updated, err := s.repo.Update(ctx, model)
	if err != nil {
		return nil, err
	}

	return updated, nil
}

func (s *Service) Delete(ctx context.Context, id int64) error {
	if id <= 0 {
		return fmt.Errorf("%w: id must be positive", ErrInvalidInput)
	}

	return s.repo.Delete(ctx, id)
}

func (s *Service) List(ctx context.Context) ([]taskdomain.Task, error) {
	return s.repo.List(ctx)
}

func validateCreateInput(input CreateInput) (CreateInput, error) {
	input.Title = strings.TrimSpace(input.Title)
	input.Description = strings.TrimSpace(input.Description)

	if input.Title == "" {
		return CreateInput{}, fmt.Errorf("%w: title is required", ErrInvalidInput)
	}

	if input.Status == "" {
		input.Status = taskdomain.StatusNew
	}

	if !input.Status.Valid() {
		return CreateInput{}, fmt.Errorf("%w: invalid status", ErrInvalidInput)
	}

	return input, nil
}

func validateUpdateInput(input UpdateInput) (UpdateInput, error) {
	input.Title = strings.TrimSpace(input.Title)
	input.Description = strings.TrimSpace(input.Description)

	if input.Title == "" {
		return UpdateInput{}, fmt.Errorf("%w: title is required", ErrInvalidInput)
	}

	if !input.Status.Valid() {
		return UpdateInput{}, fmt.Errorf("%w: invalid status", ErrInvalidInput)
	}

	return input, nil
}

// CreateWithRecurrence создает периодическую задачу и первое задание в очереди
type CreateWithRecurrenceInput struct {
	Title         string
	Description   string
	RecurrenceRule *recurrence.Rule
}

func (s *Service) CreateWithRecurrence(ctx context.Context, input CreateWithRecurrenceInput) (*taskdomain.Task, error) {
	if s.recurrenceRuleRepo == nil || s.jobRepo == nil {
		return nil, fmt.Errorf("%w: recurrence not supported in this service instance", ErrInvalidInput)
	}

	// Валидация правила
	if input.RecurrenceRule == nil {
		return nil, fmt.Errorf("%w: recurrence rule is required", ErrInvalidInput)
	}

	if err := input.RecurrenceRule.Params.Validate(input.RecurrenceRule.RecurrenceType); err != nil {
		return nil, fmt.Errorf("%w: %v", ErrInvalidInput, err)
	}

	// Создаем задачу на основе правила
	now := s.now()
	task := &taskdomain.Task{
		Title:            input.Title,
		Description:      input.Description,
		Status:           taskdomain.StatusNew,
		RecurrenceRuleID: &input.RecurrenceRule.ID,
		CreatedAt:        now,
		UpdatedAt:        now,
	}

	created, err := s.repo.Create(ctx, task)
	if err != nil {
		return nil, fmt.Errorf("create task: %w", err)
	}

	// Создаем первое задание для планировщика
	job := &scheduler.Job{
		RecurrenceRuleID: input.RecurrenceRule.ID,
		Status:           scheduler.JobStatusPending,
		ScheduledAt:      input.RecurrenceRule.NextOccurrenceDate,
		MaxRetries:       3,
		CreatedAt:        now,
		UpdatedAt:        now,
	}

	if _, err := s.jobRepo.Create(ctx, job); err != nil {
		return nil, fmt.Errorf("create job: %w", err)
	}

	return created, nil
}

// CreateRecurrenceRule создает новое правило периодичности
func (s *Service) CreateRecurrenceRule(ctx context.Context, rule *recurrence.Rule) (*recurrence.Rule, error) {
	if s.recurrenceRuleRepo == nil {
		return nil, fmt.Errorf("%w: recurrence not supported in this service instance", ErrInvalidInput)
	}

	if rule.Name == "" {
		return nil, fmt.Errorf("%w: rule name is required", ErrInvalidInput)
	}

	if err := rule.Params.Validate(rule.RecurrenceType); err != nil {
		return nil, fmt.Errorf("%w: %v", ErrInvalidInput, err)
	}

	now := s.now()
	rule.CreatedAt = now
	rule.UpdatedAt = now
	rule.IsEnabled = true

	// Вычисляем первую дату события
	nextDate, err := s.dateCalculator.CalculateNextOccurrence(rule, now)
	if err != nil {
		return nil, fmt.Errorf("calculate next occurrence: %w", err)
	}

	rule.NextOccurrenceDate = nextDate

	return s.recurrenceRuleRepo.Create(ctx, rule)
}

// UpdateRecurrenceRule обновляет существующее правило
func (s *Service) UpdateRecurrenceRule(ctx context.Context, rule *recurrence.Rule) (*recurrence.Rule, error) {
	if s.recurrenceRuleRepo == nil {
		return nil, fmt.Errorf("%w: recurrence not supported in this service instance", ErrInvalidInput)
	}

	if rule.ID <= 0 {
		return nil, fmt.Errorf("%w: invalid rule id", ErrInvalidInput)
	}

	if err := rule.Params.Validate(rule.RecurrenceType); err != nil {
		return nil, fmt.Errorf("%w: %v", ErrInvalidInput, err)
	}

	rule.UpdatedAt = s.now()

	return s.recurrenceRuleRepo.Update(ctx, rule)
}

// GetRecurrenceRule получает правило по ID
func (s *Service) GetRecurrenceRule(ctx context.Context, id int64) (*recurrence.Rule, error) {
	if s.recurrenceRuleRepo == nil {
		return nil, fmt.Errorf("%w: recurrence not supported in this service instance", ErrInvalidInput)
	}

	if id <= 0 {
		return nil, fmt.Errorf("%w: invalid rule id", ErrInvalidInput)
	}

	return s.recurrenceRuleRepo.GetByID(ctx, id)
}

// ListRecurrenceRules получает все правила (опционально фильтруя по enabled статусу)
func (s *Service) ListRecurrenceRules(ctx context.Context, enabled *bool) ([]recurrence.Rule, error) {
	if s.recurrenceRuleRepo == nil {
		return nil, fmt.Errorf("%w: recurrence not supported in this service instance", ErrInvalidInput)
	}

	return s.recurrenceRuleRepo.List(ctx, enabled)
}

// DeleteRecurrenceRule удаляет правило
func (s *Service) DeleteRecurrenceRule(ctx context.Context, id int64) error {
	if s.recurrenceRuleRepo == nil {
		return fmt.Errorf("%w: recurrence not supported in this service instance", ErrInvalidInput)
	}

	if id <= 0 {
		return fmt.Errorf("%w: invalid rule id", ErrInvalidInput)
	}

	return s.recurrenceRuleRepo.Delete(ctx, id)
}
