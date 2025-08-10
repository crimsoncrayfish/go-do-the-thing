-- Migration: Remove analytics functions for admin dashboard
-- Date: 2025-08-10

-- Drop all analytics functions
DROP FUNCTION IF EXISTS sp_get_user_stats();
DROP FUNCTION IF EXISTS sp_get_project_stats();
DROP FUNCTION IF EXISTS sp_get_task_stats();
DROP FUNCTION IF EXISTS sp_get_recent_activity();
DROP FUNCTION IF EXISTS sp_get_daily_activity();
DROP FUNCTION IF EXISTS sp_search_users(TEXT);