-- TASK EVENTS MIGRATION ROLLBACK
-- Date: 2026-01-17

-- Drop stored procedures
DROP FUNCTION IF EXISTS sp_is_task_in_progress(BIGINT);
DROP FUNCTION IF EXISTS sp_get_task_time_spent(BIGINT);
DROP FUNCTION IF EXISTS sp_insert_task_event(BIGINT, BIGINT, TEXT, JSONB);

-- Drop indexes
DROP INDEX IF EXISTS idx_task_events_task_type;
DROP INDEX IF EXISTS idx_task_events_event_time;
DROP INDEX IF EXISTS idx_task_events_event_type;
DROP INDEX IF EXISTS idx_task_events_user_id;
DROP INDEX IF EXISTS idx_task_events_task_id;

-- Drop table
DROP TABLE IF EXISTS task_events;
