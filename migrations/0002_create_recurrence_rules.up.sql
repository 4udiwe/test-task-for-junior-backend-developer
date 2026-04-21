CREATE TABLE recurrence_rules (
    id BIGSERIAL PRIMARY KEY,
    name TEXT NOT NULL,
    recurrence_type TEXT NOT NULL,  -- 'daily', 'monthly', 'specific', 'even', 'odd'
    params JSONB NOT NULL DEFAULT '{}',
    next_occurrence_date TIMESTAMPTZ NOT NULL,
    is_enabled BOOLEAN NOT NULL DEFAULT true,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Индексы для быстрого поиска
CREATE INDEX idx_recurrence_rules_enabled ON recurrence_rules(is_enabled);
CREATE INDEX idx_recurrence_rules_next_date ON recurrence_rules(next_occurrence_date);

