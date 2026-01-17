-- Rollback: Restore original sp_get_tasks_by_search without status/tag filtering

DROP FUNCTION IF EXISTS sp_get_tasks_by_search(BIGINT, TEXT, BIGINT, TEXT, BIGINT[]);
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
