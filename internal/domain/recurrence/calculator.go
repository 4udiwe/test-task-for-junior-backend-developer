package recurrence

import (
	"fmt"
	"time"
)

// DateCalculator вычисляет следующие даты для периодических задач
type DateCalculator struct{}

// CalculateNextOccurrence вычисляет следующую дату событие на основе правила периодичности
// currentDate - последняя дата создания задачи
func (c *DateCalculator) CalculateNextOccurrence(rule *Rule, currentDate time.Time) (time.Time, error) {
	if !rule.IsEnabled {
		return time.Time{}, fmt.Errorf("recurrence rule is disabled")
	}

	if err := rule.Params.Validate(rule.RecurrenceType); err != nil {
		return time.Time{}, err
	}

	switch rule.RecurrenceType {
	case TypeDaily:
		return c.nextDailyDate(currentDate, rule.Params), nil

	case TypeMonthly:
		return c.nextMonthlyDate(currentDate, rule.Params), nil

	case TypeSpecific:
		nextDate, err := c.nextSpecificDate(currentDate, rule.Params)
		if err != nil {
			return time.Time{}, err
		}
		return nextDate, nil

	case TypeEvenDays:
		return c.nextEvenDay(currentDate), nil

	case TypeOddDays:
		return c.nextOddDay(currentDate), nil

	default:
		return time.Time{}, fmt.Errorf("unknown recurrence type: %v", rule.RecurrenceType)
	}
}

// nextDailyDate вычисляет следующий день для ежедневной периодичности
func (c *DateCalculator) nextDailyDate(current time.Time, params RecurrenceParams) time.Time {
	interval := 1
	if params.IntervalDays != nil && *params.IntervalDays > 0 {
		interval = *params.IntervalDays
	}

	return current.AddDate(0, 0, interval)
}

// nextMonthlyDate вычисляет следующую дату для ежемесячной периодичности
func (c *DateCalculator) nextMonthlyDate(current time.Time, params RecurrenceParams) time.Time {
	if len(params.MonthDays) == 0 {
		return current.AddDate(0, 1, 0)
	}

	now := time.Now().UTC()
	year, month, _ := current.Date()

	// Проверяем дни в текущем месяце
	for _, day := range params.MonthDays {
		if day > 31 {
			continue // пропускаем невалидные дни
		}

		// Пытаемся создать дату в текущем месяце
		candidate := time.Date(year, month, day, 0, 0, 0, 0, time.UTC)

		// Если дата в будущем и она >= current, используем её
		if candidate.After(current) && candidate.After(now.Add(-1*time.Hour)) {
			return candidate
		}
	}

	// Все дни текущего месяца прошли, переходим на следующий месяц
	nextMonth := current.AddDate(0, 1, 0)
	year, month, _ = nextMonth.Date()

	// Берем минимальный день из списка
	minDay := params.MonthDays[0]
	for _, day := range params.MonthDays {
		if day < minDay {
			minDay = day
		}
	}

	return time.Date(year, month, minDay, 0, 0, 0, 0, time.UTC)
}

// nextSpecificDate вычисляет следующую дату из списка конкретных дат
func (c *DateCalculator) nextSpecificDate(current time.Time, params RecurrenceParams) (time.Time, error) {
	if len(params.SpecificDates) == 0 {
		return time.Time{}, fmt.Errorf("no specific dates provided")
	}

	now := time.Now().UTC()

	for _, dateStr := range params.SpecificDates {
		t, err := time.Parse("2006-01-02", dateStr)
		if err != nil {
			return time.Time{}, fmt.Errorf("invalid date format: %s", dateStr)
		}

		// Устанавливаем время в начало дня UTC
		t = time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, time.UTC)

		// Если дата в будущем и >= current, используем её
		if t.After(current) && t.After(now.Add(-1*time.Hour)) {
			return t, nil
		}
	}

	return time.Time{}, fmt.Errorf("no future dates found in specific dates list")
}

// nextEvenDay вычисляет следующее четное число месяца
func (c *DateCalculator) nextEvenDay(current time.Time) time.Time {
	now := time.Now().UTC()
	candidate := current

	// Ищем следующее четное число в течение 60 дней
	for i := 1; i <= 60; i++ {
		candidate = current.AddDate(0, 0, i)
		day := candidate.Day()

		// Четный день и в будущем
		if day%2 == 0 && candidate.After(now.Add(-1*time.Hour)) {
			return candidate
		}
	}

	// На случай если не нашли в течение 60 дней (маловероятно)
	return current.AddDate(0, 1, 0)
}

// nextOddDay вычисляет следующее нечетное число месяца
func (c *DateCalculator) nextOddDay(current time.Time) time.Time {
	now := time.Now().UTC()
	candidate := current

	// Ищем следующее нечетное число в течение 60 дней
	for i := 1; i <= 60; i++ {
		candidate = current.AddDate(0, 0, i)
		day := candidate.Day()

		// Нечетный день и в будущем
		if day%2 == 1 && candidate.After(now.Add(-1*time.Hour)) {
			return candidate
		}
	}

	// На случай если не нашли в течение 60 дней (маловероятно)
	return current.AddDate(0, 1, 0)
}
