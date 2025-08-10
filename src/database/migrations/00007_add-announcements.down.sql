-- Migration: Remove announcements system
-- Date: 2025-08-10

-- Drop stored procedures
DROP FUNCTION IF EXISTS sp_get_all_announcements();
DROP FUNCTION IF EXISTS sp_get_latest_announcement_for_user(BIGINT);
DROP FUNCTION IF EXISTS sp_get_user_announcements(BIGINT, INTEGER);

-- Drop indexes
DROP INDEX IF EXISTS idx_user_dismissals_announcement_id;
DROP INDEX IF EXISTS idx_user_dismissals_user_id;
DROP INDEX IF EXISTS idx_announcements_created_date;
DROP INDEX IF EXISTS idx_announcements_active;

-- Drop tables (order matters due to foreign keys)
DROP TABLE IF EXISTS user_announcement_dismissals;
DROP TABLE IF EXISTS announcements;