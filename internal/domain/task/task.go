package task

import "time"

type Status string

const (
	StatusNew        Status = "new"
	StatusInProgress Status = "in_progress"
	StatusDone       Status = "done"
)

type Task struct {
	ID               int64     `json:"id"`
	Title            string    `json:"title"`
	Description      string    `json:"description"`
	Status           Status    `json:"status"`
	RecurrenceRuleID *int64    `json:"recurrence_rule_id,omitempty"`  // ID периодического правила (если задача периодическая)
	ParentTaskID     *int64    `json:"parent_task_id,omitempty"`      // ID родительской задачи (для цепочки экземпляров)
	CreatedAt        time.Time `json:"created_at"`
	UpdatedAt        time.Time `json:"updated_at"`
}

func (s Status) Valid() bool {
	switch s {
	case StatusNew, StatusInProgress, StatusDone:
		return true
	default:
		return false
	}
}
