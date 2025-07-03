-- Migration: Update users to have an is_enabled flag and add user-registrations table
-- Date: 2024-07-25

DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM information_schema.columns
        WHERE table_name='users' AND column_name='is_enabled'
    ) THEN
        ALTER TABLE users ADD COLUMN is_enabled BOOLEAN NOT NULL DEFAULT FALSE;
    END IF;
END$$;

DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM information_schema.columns
        WHERE table_name='users' AND column_name='access_granted_by'
    ) THEN
        ALTER TABLE users ADD COLUMN access_granted_by INTEGER;
    END IF;
END$$;

DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM pg_constraint
        WHERE conname = 'fk_access_granted_by'
          AND conrelid = 'users'::regclass
    ) THEN
        ALTER TABLE users ADD CONSTRAINT fk_access_granted_by
            FOREIGN KEY (access_granted_by) REFERENCES users (id);
    END IF;
END$$;

DROP FUNCTION sp_get_user_by_email(TEXT);
CREATE OR REPLACE FUNCTION sp_get_user_by_email(_email TEXT)
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
    access_granted_by INTEGER
) AS $$
BEGIN
    RETURN QUERY
    SELECT
        users.id,
        users.email,
        users.full_name,
        users.session_id,
        users.session_start_time,
        users.is_admin,
        users.is_enabled,
        users.is_deleted,
        users.create_date,
        users.access_granted_by
    FROM users
    WHERE users.email = _email;
END; $$ LANGUAGE plpgsql;

DROP FUNCTION IF EXISTS sp_get_user_by_id(BIGINT);
CREATE OR REPLACE FUNCTION sp_get_user_by_id(_id BIGINT)
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
    access_granted_by INTEGER
) AS $$
BEGIN
    RETURN QUERY
    SELECT
        users.id,
        users.email,
        users.full_name,
        users.session_id,
        users.session_start_time,
        users.is_admin,
        users.is_enabled,
        users.is_deleted,
        users.create_date,
        users.access_granted_by
    FROM users
    WHERE users.id = _id;
END; $$ LANGUAGE plpgsql;

DROP FUNCTION IF EXISTS sp_get_users_not_deleted();
CREATE OR REPLACE FUNCTION sp_get_users_not_deleted()
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
    access_granted_by INTEGER
) AS $$
BEGIN
    RETURN QUERY
    SELECT
        users.id,
        users.email,
        users.full_name,
        users.session_id,
        users.session_start_time,
        users.is_admin,
        users.is_enabled,
        users.is_deleted,
        users.create_date,
        users.access_granted_by
    FROM users
    WHERE users.is_deleted = FALSE;
END; $$ LANGUAGE plpgsql;

CREATE OR REPLACE FUNCTION sp_update_user_is_enabled(_id BIGINT, _is_enabled BOOLEAN)
RETURNS VOID AS $$
BEGIN
    UPDATE users SET is_enabled = _is_enabled WHERE users.id = _id;
END; $$ LANGUAGE plpgsql;
