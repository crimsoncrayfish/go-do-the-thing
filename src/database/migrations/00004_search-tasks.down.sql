-- Migration: Revert Search for tasks
-- Date: 2024-07-29

DROP FUNCTION IF EXISTS sp_get_tasks_by_search(BIGINT, TEXT);
