-- Migration: Revert tags table and project tag stored procedures
-- Date: 2024-07-28

-- 1. Drop Stored Procedures (reverse order of creation/modification)
DROP FUNCTION IF EXISTS sp_update_tag(BIGINT, TEXT, TEXT);
DROP FUNCTION IF EXISTS sp_get_project_tags(BIGINT);
DROP FUNCTION IF EXISTS sp_get_tag(BIGINT);
DROP FUNCTION IF EXISTS sp_delete_tag(BIGINT);
DROP FUNCTION IF EXISTS sp_insert_tag(TEXT, BIGINT, BIGINT);

-- 2. Drop junction tables
DROP TABLE IF EXISTS project_tags;
DROP TABLE IF EXISTS task_tags;

DROP TABLE IF EXISTS task;

-- 4. Drop the tag_color ENUM type
DROP TYPE IF EXISTS tag_color;
