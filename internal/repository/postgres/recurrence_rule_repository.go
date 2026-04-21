package postgres

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	recurrence "example.com/taskservice/internal/domain/recurrence"
)

type RecurrenceRuleRepository struct {
	pool *pgxpool.Pool
}

func NewRecurrenceRuleRepository(pool *pgxpool.Pool) *RecurrenceRuleRepository {
	return &RecurrenceRuleRepository{pool: pool}
}

func (r *RecurrenceRuleRepository) Create(ctx context.Context, rule *recurrence.Rule) (*recurrence.Rule, error) {
	paramsJSON, err := json.Marshal(rule.Params)
	if err != nil {
		return nil, fmt.Errorf("marshal params: %w", err)
	}

	query := `
		INSERT INTO recurrence_rules (name, recurrence_type, params, next_occurrence_date, is_enabled, created_at, updated_at)
		VALUES ($1, $2, $3::jsonb, $4, $5, NOW(), NOW())
		RETURNING id, name, recurrence_type, params, next_occurrence_date, is_enabled, created_at, updated_at
	`

	row := r.pool.QueryRow(ctx, query, rule.Name, rule.RecurrenceType.String(), paramsJSON, rule.NextOccurrenceDate, rule.IsEnabled)
	return scanRecurrenceRule(row)
}

func (r *RecurrenceRuleRepository) GetByID(ctx context.Context, id int64) (*recurrence.Rule, error) {
	query := `
		SELECT id, name, recurrence_type, params, next_occurrence_date, is_enabled, created_at, updated_at
		FROM recurrence_rules
		WHERE id = $1
	`

	row := r.pool.QueryRow(ctx, query, id)
	found, err := scanRecurrenceRule(row)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("recurrence rule not found")
		}
		return nil, err
	}

	return found, nil
}

func (r *RecurrenceRuleRepository) Update(ctx context.Context, rule *recurrence.Rule) (*recurrence.Rule, error) {
	paramsJSON, err := json.Marshal(rule.Params)
	if err != nil {
		return nil, fmt.Errorf("marshal params: %w", err)
	}

	query := `
		UPDATE recurrence_rules
		SET name = $1,
		    recurrence_type = $2,
		    params = $3::jsonb,
		    next_occurrence_date = $4,
		    is_enabled = $5,
		    updated_at = NOW()
		WHERE id = $6
		RETURNING id, name, recurrence_type, params, next_occurrence_date, is_enabled, created_at, updated_at
	`

	row := r.pool.QueryRow(ctx, query, rule.Name, rule.RecurrenceType.String(), paramsJSON, rule.NextOccurrenceDate, rule.IsEnabled, rule.ID)
	updated, err := scanRecurrenceRule(row)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("recurrence rule not found")
		}
		return nil, err
	}

	return updated, nil
}

func (r *RecurrenceRuleRepository) List(ctx context.Context, enabled *bool) ([]recurrence.Rule, error) {
	query := `
		SELECT id, name, recurrence_type, params, next_occurrence_date, is_enabled, created_at, updated_at
		FROM recurrence_rules
	`

	args := []interface{}{}
	if enabled != nil {
		query += ` WHERE is_enabled = $1`
		args = append(args, *enabled)
	}
	query += ` ORDER BY id DESC`

	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	rules := make([]recurrence.Rule, 0)
	for rows.Next() {
		rule, err := scanRecurrenceRule(rows)
		if err != nil {
			return nil, err
		}
		rules = append(rules, *rule)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return rules, nil
}

func (r *RecurrenceRuleRepository) Delete(ctx context.Context, id int64) error {
	query := `DELETE FROM recurrence_rules WHERE id = $1`

	result, err := r.pool.Exec(ctx, query, id)
	if err != nil {
		return err
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("recurrence rule not found")
	}

	return nil
}

type recurrenceScanner interface {
	Scan(dest ...interface{}) error
}

func scanRecurrenceRule(scanner recurrenceScanner) (*recurrence.Rule, error) {
	var (
		rule      recurrence.Rule
		typeStr   string
		paramsRaw []byte
	)

	if err := scanner.Scan(
		&rule.ID,
		&rule.Name,
		&typeStr,
		&paramsRaw,
		&rule.NextOccurrenceDate,
		&rule.IsEnabled,
		&rule.CreatedAt,
		&rule.UpdatedAt,
	); err != nil {
		return nil, err
	}

	recType, err := recurrence.FromString(typeStr)
	if err != nil {
		return nil, err
	}
	rule.RecurrenceType = recType

	if err := json.Unmarshal(paramsRaw, &rule.Params); err != nil {
		return nil, fmt.Errorf("unmarshal params: %w", err)
	}

	return &rule, nil
}
