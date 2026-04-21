CREATE TABLE job_queue (
    id BIGSERIAL PRIMARY KEY,
    recurrence_rule_id BIGINT NOT NULL REFERENCES recurrence_rules(id) ON DELETE CASCADE,
    status TEXT NOT NULL,  -- 'pending', 'processing', 'success', 'failed', 'retry'
    scheduled_at TIMESTAMPTZ NOT NULL,  -- когда должна была выполниться
    started_at TIMESTAMPTZ,
    completed_at TIMESTAMPTZ,
    error_message TEXT,
    retry_count INT NOT NULL DEFAULT 0,
    max_retries INT NOT NULL DEFAULT 3,
    locked_by VARCHAR(255),  -- worker ID
    locked_until TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Критичные индексы для scheduler
CREATE INDEX idx_job_queue_pending ON job_queue(status, scheduled_at) 
WHERE status IN ('pending', 'retry');

CREATE INDEX idx_job_queue_locked ON job_queue(locked_by, locked_until);

-- Для выбора незалоченных пендинг джобов
CREATE INDEX idx_job_queue_next_to_process ON job_queue(scheduled_at) 
WHERE status IN ('pending', 'retry') AND locked_until IS NULL;
