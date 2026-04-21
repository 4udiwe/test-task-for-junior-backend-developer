-- Удаляем индексы
DROP INDEX IF EXISTS idx_tasks_recurrence_rule_id;
DROP INDEX IF EXISTS idx_tasks_parent_task_id;

-- Удаляем колонки
ALTER TABLE tasks 
DROP COLUMN IF EXISTS parent_task_id,
DROP COLUMN IF EXISTS recurrence_rule_id;
