-- Migration: Add analytics functions for admin dashboard
-- Date: 2025-08-10

-- 1. Get user statistics
CREATE OR REPLACE FUNCTION sp_get_user_stats()
RETURNS TABLE(
    total_users BIGINT,
    active_users BIGINT,
    inactive_users BIGINT,
    admin_users BIGINT
) AS $$
BEGIN
    RETURN QUERY
    SELECT
        COUNT(*) as total_users,
        COUNT(*) FILTER (WHERE is_enabled = true AND is_deleted = false) as active_users,
        COUNT(*) FILTER (WHERE is_enabled = false AND is_deleted = false) as inactive_users,
        COUNT(*) FILTER (WHERE is_admin = true AND is_deleted = false) as admin_users
    FROM users
    WHERE is_deleted = false;
END; $$ LANGUAGE plpgsql;

-- 2. Get project statistics
CREATE OR REPLACE FUNCTION sp_get_project_stats()
RETURNS TABLE(
    total_projects BIGINT,
    active_projects BIGINT,
    completed_projects BIGINT,
    overdue_projects BIGINT
) AS $$
BEGIN
    RETURN QUERY
    SELECT
        COUNT(*) as total_projects,
        COUNT(*) FILTER (WHERE is_complete = false AND is_deleted = false) as active_projects,
        COUNT(*) FILTER (WHERE is_complete = true AND is_deleted = false) as completed_projects,
        COUNT(*) FILTER (WHERE due_date < NOW() AND is_complete = false AND is_deleted = false) as overdue_projects
    FROM projects
    WHERE is_deleted = false;
END; $$ LANGUAGE plpgsql;

-- 3. Get task statistics
CREATE OR REPLACE FUNCTION sp_get_task_stats()
RETURNS TABLE(
    total_tasks BIGINT,
    pending_tasks BIGINT,
    in_progress_tasks BIGINT,
    completed_tasks BIGINT,
    overdue_tasks BIGINT
) AS $$
BEGIN
    RETURN QUERY
    SELECT
        COUNT(*) as total_tasks,
        COUNT(*) FILTER (WHERE status = 0 AND is_deleted = false) as pending_tasks,
        COUNT(*) FILTER (WHERE status = 1 AND is_deleted = false) as in_progress_tasks,
        COUNT(*) FILTER (WHERE status = 2 AND is_deleted = false) as completed_tasks,
        COUNT(*) FILTER (WHERE due_date < NOW() AND status != 2 AND is_deleted = false) as overdue_tasks
    FROM tasks
    WHERE is_deleted = false;
END; $$ LANGUAGE plpgsql;

-- 4. Get recent activity (last 30 days)
CREATE OR REPLACE FUNCTION sp_get_recent_activity()
RETURNS TABLE(
    new_users_last_30_days BIGINT,
    new_projects_last_30_days BIGINT,
    new_tasks_last_30_days BIGINT,
    completed_tasks_last_30_days BIGINT
) AS $$
BEGIN
    RETURN QUERY
    SELECT
        (SELECT COUNT(*) FROM users WHERE create_date >= NOW() - INTERVAL '30 days' AND is_deleted = false) as new_users_last_30_days,
        (SELECT COUNT(*) FROM projects WHERE created_date >= NOW() - INTERVAL '30 days' AND is_deleted = false) as new_projects_last_30_days,
        (SELECT COUNT(*) FROM tasks WHERE created_date >= NOW() - INTERVAL '30 days' AND is_deleted = false) as new_tasks_last_30_days,
        (SELECT COUNT(*) FROM tasks WHERE complete_date >= NOW() - INTERVAL '30 days' AND is_deleted = false) as completed_tasks_last_30_days;
END; $$ LANGUAGE plpgsql;

-- 5. Get daily activity for the last 7 days (for trend charts)
CREATE OR REPLACE FUNCTION sp_get_daily_activity()
RETURNS TABLE(
    activity_date DATE,
    new_users BIGINT,
    new_projects BIGINT,
    new_tasks BIGINT,
    completed_tasks BIGINT
) AS $$
BEGIN
    RETURN QUERY
    WITH date_series AS (
        SELECT generate_series(
            CURRENT_DATE - INTERVAL '6 days',
            CURRENT_DATE,
            INTERVAL '1 day'
        )::DATE as activity_date
    )
    SELECT
        ds.activity_date,
        COALESCE(u.new_users, 0) as new_users,
        COALESCE(p.new_projects, 0) as new_projects,
        COALESCE(t.new_tasks, 0) as new_tasks,
        COALESCE(tc.completed_tasks, 0) as completed_tasks
    FROM date_series ds
    LEFT JOIN (
        SELECT DATE(create_date) as date, COUNT(*) as new_users
        FROM users
        WHERE create_date >= CURRENT_DATE - INTERVAL '6 days'
        AND is_deleted = false
        GROUP BY DATE(create_date)
    ) u ON ds.activity_date = u.date
    LEFT JOIN (
        SELECT DATE(created_date) as date, COUNT(*) as new_projects
        FROM projects
        WHERE created_date >= CURRENT_DATE - INTERVAL '6 days'
        AND is_deleted = false
        GROUP BY DATE(created_date)
    ) p ON ds.activity_date = p.date
    LEFT JOIN (
        SELECT DATE(created_date) as date, COUNT(*) as new_tasks
        FROM tasks
        WHERE created_date >= CURRENT_DATE - INTERVAL '6 days'
        AND is_deleted = false
        GROUP BY DATE(created_date)
    ) t ON ds.activity_date = t.date
    LEFT JOIN (
        SELECT DATE(complete_date) as date, COUNT(*) as completed_tasks
        FROM tasks
        WHERE complete_date >= CURRENT_DATE - INTERVAL '6 days'
        AND is_deleted = false
        GROUP BY DATE(complete_date)
    ) tc ON ds.activity_date = tc.date
    ORDER BY ds.activity_date;
END; $$ LANGUAGE plpgsql;

-- 6. Search users by email or full name
CREATE OR REPLACE FUNCTION sp_search_users(search_term TEXT)
RETURNS TABLE(
    id BIGINT,
    email TEXT,
    full_name TEXT,
    session_id TEXT,
    session_start_time TIMESTAMP WITHOUT TIME ZONE,
    is_admin BOOLEAN,
    is_enabled BOOLEAN,
    is_deleted BOOLEAN,
    create_date TIMESTAMP WITHOUT TIME ZONE,
    access_granted_by BIGINT
) AS $$
BEGIN
    RETURN QUERY
    SELECT
        u.id,
        u.email,
        u.full_name,
        u.session_id,
        u.session_start_time,
        u.is_admin,
        u.is_enabled,
        u.is_deleted,
        u.create_date,
        u.access_granted_by
    FROM users u
    WHERE u.is_deleted = false
    AND (
        LOWER(u.email) LIKE LOWER('%' || search_term || '%')
        OR LOWER(u.full_name) LIKE LOWER('%' || search_term || '%')
    )
    ORDER BY u.create_date DESC;
END; $$ LANGUAGE plpgsql;