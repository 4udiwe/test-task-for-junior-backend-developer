package recurrence

import (
	"encoding/json"
	"fmt"
)

// RecurrenceType определяет тип периодичности задачи
type RecurrenceType int

const (
	// TypeDaily - каждый N-й день
	TypeDaily RecurrenceType = iota
	// TypeMonthly - конкретные дни месяца (1-30)
	TypeMonthly
	// TypeSpecific - конкретные даты
	TypeSpecific
	// TypeEvenDays - только четные числа месяца
	TypeEvenDays
	// TypeOddDays - только нечетные числа месяца
	TypeOddDays
)

// String возвращает строковое представление типа периодичности
func (r RecurrenceType) String() string {
	switch r {
	case TypeDaily:
		return "daily"
	case TypeMonthly:
		return "monthly"
	case TypeSpecific:
		return "specific"
	case TypeEvenDays:
		return "even"
	case TypeOddDays:
		return "odd"
	default:
		return "unknown"
	}
}

// FromString преобразует строку в RecurrenceType
func FromString(s string) (RecurrenceType, error) {
	switch s {
	case "daily":
		return TypeDaily, nil
	case "monthly":
		return TypeMonthly, nil
	case "specific":
		return TypeSpecific, nil
	case "even":
		return TypeEvenDays, nil
	case "odd":
		return TypeOddDays, nil
	default:
		return -1, fmt.Errorf("unknown recurrence type: %s", s)
	}
}

// RecurrenceParams содержит параметры периодичности в зависимости от типа
type RecurrenceParams struct {
	// Daily: каждый N-й день
	IntervalDays *int `json:"interval_days,omitempty"`

	// Monthly: конкретные дни месяца (1-30)
	MonthDays []int `json:"month_days,omitempty"`

	// Specific: конкретные даты (ISO 8601 format: "2025-05-01")
	SpecificDates []string `json:"specific_dates,omitempty"`

	// Common параметры
	StartDate string `json:"start_date,omitempty"` // ISO 8601 format
	EndDate   *string `json:"end_date,omitempty"`  // ISO 8601 format (null = бесконечно)
}

// UnmarshalJSON custom unmarshal для работы с JSONB из БД
func (rp *RecurrenceParams) UnmarshalJSON(data []byte) error {
	type Alias RecurrenceParams
	aux := &struct {
		*Alias
	}{
		Alias: (*Alias)(rp),
	}
	return json.Unmarshal(data, &aux)
}

// MarshalJSON custom marshal для сохранения в JSONB
func (rp RecurrenceParams) MarshalJSON() ([]byte, error) {
	type Alias RecurrenceParams
	return json.Marshal((*Alias)(&rp))
}

// Validate проверяет корректность параметров
func (rp RecurrenceParams) Validate(recType RecurrenceType) error {
	switch recType {
	case TypeDaily:
		if rp.IntervalDays == nil || *rp.IntervalDays < 1 {
			return fmt.Errorf("interval_days must be >= 1 for daily recurrence")
		}

	case TypeMonthly:
		if len(rp.MonthDays) == 0 {
			return fmt.Errorf("month_days must not be empty for monthly recurrence")
		}
		for _, day := range rp.MonthDays {
			if day < 1 || day > 30 {
				return fmt.Errorf("month_days must be between 1 and 30, got %d", day)
			}
		}

	case TypeSpecific:
		if len(rp.SpecificDates) == 0 {
			return fmt.Errorf("specific_dates must not be empty for specific recurrence")
		}

	case TypeEvenDays, TypeOddDays:
		// Нет дополнительных параметров, просто проверяем существование

	default:
		return fmt.Errorf("unknown recurrence type: %v", recType)
	}

	return nil
}
