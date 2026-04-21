package recurrence

import "time"

// Rule представляет конфигурацию периодичности для задач
type Rule struct {
	ID                 int64
	Name               string              // "Ежедневный обход", "Еженедельный отчет"
	RecurrenceType     RecurrenceType
	Params             RecurrenceParams
	NextOccurrenceDate time.Time           // когда создать следующую задачу
	IsEnabled          bool
	CreatedAt          time.Time
	UpdatedAt          time.Time
}
