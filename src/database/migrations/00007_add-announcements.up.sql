-- Migration: Add announcements system
-- Date: 2025-08-10

-- 1. Create announcements table
CREATE TABLE announcements (
    id BIGSERIAL PRIMARY KEY,
    title TEXT NOT NULL,
    created_by BIGINT NOT NULL REFERENCES users(id),
    created_date TIMESTAMP WITHOUT TIME ZONE DEFAULT NOW(),
    expires_date TIMESTAMP WITHOUT TIME ZONE NOT NULL,
    is_active BOOLEAN DEFAULT true,
    is_deleted BOOLEAN DEFAULT false
);

-- 2. Create user announcement dismissals table
CREATE TABLE user_announcement_dismissals (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL REFERENCES users(id),
    announcement_id BIGINT NOT NULL REFERENCES announcements(id),
    dismissed_date TIMESTAMP WITHOUT TIME ZONE DEFAULT NOW(),
    UNIQUE(user_id, announcement_id)
);

-- 3. Create indexes for performance
CREATE INDEX idx_announcements_active ON announcements(is_active, is_deleted, expires_date);
CREATE INDEX idx_announcements_created_date ON announcements(created_date);
CREATE INDEX idx_user_dismissals_user_id ON user_announcement_dismissals(user_id);
CREATE INDEX idx_user_dismissals_announcement_id ON user_announcement_dismissals(announcement_id);

-- 4. Get active announcements for a user (excluding dismissed ones)
CREATE OR REPLACE FUNCTION sp_get_user_announcements(p_user_id BIGINT, p_limit INTEGER DEFAULT 5)
RETURNS TABLE(
    id BIGINT,
    title TEXT,
    created_by BIGINT,
    created_date TIMESTAMP WITHOUT TIME ZONE,
    expires_date TIMESTAMP WITHOUT TIME ZONE,
    creator_name TEXT
) AS $$
BEGIN
    RETURN QUERY
    SELECT
        a.id,
        a.title,
        a.created_by,
        a.created_date,
        a.expires_date,
        u.full_name as creator_name
    FROM announcements a
    JOIN users u ON a.created_by = u.id
    LEFT JOIN user_announcement_dismissals uad ON a.id = uad.announcement_id AND uad.user_id = p_user_id
    WHERE a.is_active = true
    AND a.is_deleted = false
    AND a.expires_date > NOW()
    AND uad.id IS NULL
    ORDER BY a.created_date DESC
    LIMIT p_limit;
END; $$ LANGUAGE plpgsql;

-- 5. Get latest announcement for dashboard (excluding dismissed ones)
CREATE OR REPLACE FUNCTION sp_get_latest_announcement_for_user(p_user_id BIGINT)
RETURNS TABLE(
    id BIGINT,
    title TEXT,
    created_by BIGINT,
    created_date TIMESTAMP WITHOUT TIME ZONE,
    expires_date TIMESTAMP WITHOUT TIME ZONE,
    creator_name TEXT
) AS $$
BEGIN
    RETURN QUERY
    SELECT
        a.id,
        a.title,
        a.created_by,
        a.created_date,
        a.expires_date,
        u.full_name as creator_name
    FROM announcements a
    JOIN users u ON a.created_by = u.id
    LEFT JOIN user_announcement_dismissals uad ON a.id = uad.announcement_id AND uad.user_id = p_user_id
    WHERE a.is_active = true
    AND a.is_deleted = false
    AND a.expires_date > NOW()
    AND uad.id IS NULL
    ORDER BY a.created_date DESC
    LIMIT 1;
END; $$ LANGUAGE plpgsql;

-- 6. Get all announcements for admin management
CREATE OR REPLACE FUNCTION sp_get_all_announcements()
RETURNS TABLE(
    id BIGINT,
    title TEXT,
    created_by BIGINT,
    created_date TIMESTAMP WITHOUT TIME ZONE,
    expires_date TIMESTAMP WITHOUT TIME ZONE,
    is_active BOOLEAN,
    is_deleted BOOLEAN,
    creator_name TEXT
) AS $$
BEGIN
    RETURN QUERY
    SELECT
        a.id,
        a.title,
        a.created_by,
        a.created_date,
        a.expires_date,
        a.is_active,
        a.is_deleted,
        u.full_name as creator_name
    FROM announcements a
    JOIN users u ON a.created_by = u.id
    WHERE a.is_deleted = false
    ORDER BY a.created_date DESC;
END; $$ LANGUAGE plpgsql;