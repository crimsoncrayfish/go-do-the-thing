-- Migration: Add status and tag filtering to task search
-- Date: 2025-01-17

DROP FUNCTION IF EXISTS sp_get_tasks_by_search(BIGINT, TEXT, BIGINT);
CREATE OR REPLACE FUNCTION sp_get_tasks_by_search(
    _user_id BIGINT,
    _search_text TEXT,
    _project_id BIGINT,
    _status TEXT,
    _tag_ids BIGINT[]
)
RETURNS TABLE(
    id BIGINT,
    name TEXT,
    description TEXT,
    status INTEGER,
    assigned_to BIGINT,
    due_date TIMESTAMP WITHOUT TIME ZONE,
    created_by BIGINT,
    created_date TIMESTAMP WITHOUT TIME ZONE,
    modified_by BIGINT,
    modified_date TIMESTAMP WITHOUT TIME ZONE,
    is_deleted BOOLEAN,
    project_id BIGINT,
    complete_date TIMESTAMP WITHOUT TIME ZONE
) AS $$
BEGIN
    RETURN QUERY 
        SELECT 
            t.id,
            t.name,
            t.description,
            t.status,
            t.assigned_to,
            t.due_date,
            t.created_by,
            t.created_date,
            t.modified_by,
            t.modified_date,
            t.is_deleted,
            t.project_id,
            t.complete_date 
        FROM tasks t
        WHERE t.is_deleted = FALSE
          AND t.assigned_to = _user_id
          -- Project filter (0 means all projects)
          AND (_project_id = 0 OR t.project_id = _project_id)
          -- Text search filter
          AND (
              _search_text = '' 
              OR t.name ILIKE '%' || _search_text || '%' 
              OR t.description ILIKE '%' || _search_text || '%'
          )
          -- Status filter (note: zero date '0001-01-01' is treated as incomplete)
          AND (
              _status = 'all'
              OR (_status = 'completed' AND t.complete_date IS NOT NULL AND t.complete_date > '0001-01-01'::timestamp)
              OR (_status = 'incomplete' AND (t.complete_date IS NULL OR t.complete_date = '0001-01-01'::timestamp))
              OR (_status = 'overdue' AND (t.complete_date IS NULL OR t.complete_date = '0001-01-01'::timestamp) AND t.due_date < NOW())
          )
          -- Tag filter (empty array means no tag filter)
          AND (
              array_length(_tag_ids, 1) IS NULL
              OR EXISTS (
                  SELECT 1 FROM task_tags tt 
                  WHERE tt.task_id = t.id 
                  AND tt.tag_id = ANY(_tag_ids)
              )
          );
END;
$$ LANGUAGE plpgsql;
