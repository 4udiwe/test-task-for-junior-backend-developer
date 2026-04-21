-- Добавляем колонки для поддержки периодических задач
ALTER TABLE tasks 
ADD COLUMN recurrence_rule_id BIGINT REFERENCES recurrence_rules(id) ON DELETE SET NULL,
ADD COLUMN parent_task_id BIGINT REFERENCES tasks(id) ON DELETE CASCADE;

-- Индексы для быстрого поиска
CREATE INDEX idx_tasks_parent_task_id ON tasks(parent_task_id);
CREATE INDEX idx_tasks_recurrence_rule_id ON tasks(recurrence_rule_id);
