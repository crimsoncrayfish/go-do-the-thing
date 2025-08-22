-- Migration: Revert Task tag queries
-- Date: 2025-08-19 

DROP FUNCTION IF EXISTS sp_set_task_tags(BIGINT, BIGINT, BIGINT[]);
DROP FUNCTION IF EXISTS sp_get_task_tags(BIGINT);
