-- Migration: Create Task tag queries
-- Date: 2025-08-19 

-- 1. Create/Delete task tags 
DROP FUNCTION IF EXISTS sp_set_task_tags(BIGINT, BIGINT, BIGINT[]);
CREATE OR REPLACE FUNCTION sp_set_task_tags(_projectId BIGINT, _taskId BIGINT, _tagIds BIGINT[])
RETURNS void
LANGUAGE plpgsql
AS $$
DECLARE
    invalid_tag_count INT;
BEGIN
    IF NOT EXISTS (SELECT 1
                   FROM tasks
                   WHERE id = _taskId
                     AND project_id = _projectId) THEN
        RAISE EXCEPTION 'Task (ID: %) does not exist or does not belong to Project (ID: %)', _taskId, _projectId;
    END IF;

    SELECT COUNT(DISTINCT tid)
    INTO invalid_tag_count
    FROM unnest(_tagIds) AS tid
    WHERE NOT EXISTS (SELECT 1
                      FROM project_tags
                      WHERE project_id = _projectId
                        AND project_tags.tag_id = tid);

    IF invalid_tag_count > 0 THEN
        RAISE EXCEPTION 'One or more provided tag IDs are not associated with Project (ID: %)', _projectId;
    END IF;

	DELETE FROM task_tags
	WHERE task_id = _taskId
	AND tag_id NOT IN (SELECT unnest(_tagIds));

    INSERT INTO task_tags
    (task_id, tag_id)
    SELECT _taskId, tags.* FROM unnest(_tagIds) as tags
    ON CONFLICT (task_id, tag_id) DO NOTHING;
END;
$$;

-- 2. Get tags by task id
DROP FUNCTION IF EXISTS sp_get_task_tags(BIGINT);
CREATE OR REPLACE FUNCTION sp_get_task_tags(_taskId BIGINT)
RETURNS TABLE(id BIGINT, name TEXT, color tag_color) AS $$
BEGIN
    RETURN QUERY
    SELECT
        t.id,
        t.name,
        t.color
    FROM
        tags AS t
    INNER JOIN
        task_tags AS tt ON t.id = tt.tag_id
    WHERE
        tt.task_id = _taskId;
END; $$ LANGUAGE plpgsql;

DROP FUNCTION IF EXISTS sp_get_tasks_by_search(BIGINT, TEXT);
CREATE OR REPLACE FUNCTION sp_get_tasks_by_search(_user_id BIGINT, _search_text TEXT, _project_id BIGINT)
RETURNS TABLE(id BIGINT, name TEXT, description TEXT, status INTEGER, assigned_to BIGINT, due_date TIMESTAMP WITHOUT TIME ZONE, created_by BIGINT, created_date TIMESTAMP WITHOUT TIME ZONE, modified_by BIGINT, modified_date TIMESTAMP WITHOUT TIME ZONE, is_deleted BOOLEAN, project_id BIGINT, complete_date TIMESTAMP WITHOUT TIME ZONE) AS $$
BEGIN
    RETURN QUERY 
        SELECT tasks.id, tasks.name, tasks.description, tasks.status, tasks.assigned_to, tasks.due_date, tasks.created_by, tasks.created_date, tasks.modified_by, tasks.modified_date, tasks.is_deleted, tasks.project_id, tasks.complete_date 
        FROM tasks 
        WHERE tasks.is_deleted = FALSE
          AND tasks.assigned_to = _user_id
          AND (tasks.project_id = _project_id OR _project_id = 0)
          AND (tasks.name ILIKE '%' || _search_text || '%' OR tasks.description ILIKE '%' || _search_text || '%');
END; $$ LANGUAGE plpgsql;
