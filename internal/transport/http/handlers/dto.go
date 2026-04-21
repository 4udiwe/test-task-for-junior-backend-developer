package handlers

import (
	"time"

	recurrence "example.com/taskservice/internal/domain/recurrence"
	taskdomain "example.com/taskservice/internal/domain/task"
)

type taskMutationDTO struct {
	Title       string            `json:"title"`
	Description string            `json:"description"`
	Status      taskdomain.Status `json:"status"`
}

type taskDTO struct {
	ID               int64             `json:"id"`
	Title            string            `json:"title"`
	Description      string            `json:"description"`
	Status           taskdomain.Status `json:"status"`
	RecurrenceRuleID *int64            `json:"recurrence_rule_id,omitempty"`
	ParentTaskID     *int64            `json:"parent_task_id,omitempty"`
	CreatedAt        time.Time         `json:"created_at"`
	UpdatedAt        time.Time         `json:"updated_at"`
}

func newTaskDTO(task *taskdomain.Task) taskDTO {
	return taskDTO{
		ID:               task.ID,
		Title:            task.Title,
		Description:      task.Description,
		Status:           task.Status,
		RecurrenceRuleID: task.RecurrenceRuleID,
		ParentTaskID:     task.ParentTaskID,
		CreatedAt:        task.CreatedAt,
		UpdatedAt:        task.UpdatedAt,
	}
}

// Recurrence Rule DTOs
type RecurrenceParamsDTO struct {
	IntervalDays  *int     `json:"interval_days,omitempty"`
	MonthDays     []int    `json:"month_days,omitempty"`
	SpecificDates []string `json:"specific_dates,omitempty"`
	StartDate     string   `json:"start_date,omitempty"`
	EndDate       *string  `json:"end_date,omitempty"`
}

type CreateRecurrenceRuleDTO struct {
	Name           string                `json:"name"`
	RecurrenceType string                `json:"recurrence_type"` // "daily", "monthly", "specific", "even", "odd"
	Params         RecurrenceParamsDTO   `json:"params"`
}

type RecurrenceRuleDTO struct {
	ID                 int64               `json:"id"`
	Name               string              `json:"name"`
	RecurrenceType     string              `json:"recurrence_type"`
	Params             RecurrenceParamsDTO `json:"params"`
	NextOccurrenceDate time.Time           `json:"next_occurrence_date"`
	IsEnabled          bool                `json:"is_enabled"`
	CreatedAt          time.Time           `json:"created_at"`
	UpdatedAt          time.Time           `json:"updated_at"`
}

func newRecurrenceRuleDTO(rule *recurrence.Rule) RecurrenceRuleDTO {
	params := RecurrenceParamsDTO{
		IntervalDays:  rule.Params.IntervalDays,
		MonthDays:     rule.Params.MonthDays,
		SpecificDates: rule.Params.SpecificDates,
		StartDate:     rule.Params.StartDate,
		EndDate:       rule.Params.EndDate,
	}

	return RecurrenceRuleDTO{
		ID:                 rule.ID,
		Name:               rule.Name,
		RecurrenceType:     rule.RecurrenceType.String(),
		Params:             params,
		NextOccurrenceDate: rule.NextOccurrenceDate,
		IsEnabled:          rule.IsEnabled,
		CreatedAt:          rule.CreatedAt,
		UpdatedAt:          rule.UpdatedAt,
	}
}

// Task with Recurrence DTOs
type CreateTaskWithRecurrenceDTO struct {
	Title              string                  `json:"title"`
	Description        string                  `json:"description"`
	RecurrenceRuleID   *int64                  `json:"recurrence_rule_id,omitempty"`
	RecurrenceRuleData *CreateRecurrenceRuleDTO `json:"recurrence_rule_data,omitempty"`
}

