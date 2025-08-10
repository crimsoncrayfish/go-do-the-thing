-- Migration: Revert tag queries
-- Date: 2025-08-04

-- 1. Get tags by project id
DROP FUNCTION IF EXISTS sp_get_tag(BIGINT);
DROP FUNCTION IF EXISTS sp_get_tag(_id BIGINT, _project_id BIGINT);

DROP FUNCTION IF EXISTS sp_insert_tag(_name TEXT, color tag_color, _user_id BIGINT, _project_id BIGINT);
DROP FUNCTION IF EXISTS sp_get_project_tags(_project_id BIGINT);

ALTER TABLE tags ALTER COLUMN color DROP DEFAULT;
ALTER TABLE tags ALTER COLUMN color TYPE TEXT;
DROP TYPE tag_color;
CREATE TYPE tag_color AS ENUM (
    'red',
    'orange',
    'amber',
    'green',
    'cyan',
    'blue',
    'indigo',
    'purple',
    'fuchsia',
    'rose'
);
ALTER TABLE tags
    ALTER COLUMN color TYPE tag_color
    USING color::tag_color;
ALTER TABLE tags
    ALTER COLUMN color SET DEFAULT 'red';

CREATE OR REPLACE FUNCTION sp_get_tag(_id BIGINT, _project_id BIGINT)
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
        project_tags AS pt ON t.id = pt.tag_id
    WHERE
        pt.project_id = _project_id
    AND
        t.id = _id;
END; $$ LANGUAGE plpgsql;

DROP FUNCTION IF EXISTS sp_get_project_tags(BIGINT);
CREATE OR REPLACE FUNCTION sp_get_project_tags(_project_id BIGINT)
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
        project_tags AS pt ON t.id = pt.tag_id
    WHERE
        pt.project_id = _project_id;
END; $$ LANGUAGE plpgsql;

DROP FUNCTION IF EXISTS sp_insert_tag(TEXT, BIGINT, BIGINT);
CREATE OR REPLACE FUNCTION sp_insert_tag(_name TEXT, _color tag_color, _user_id BIGINT, _project_id BIGINT)
RETURNS BIGINT AS $$
DECLARE
    _id BIGINT;
BEGIN
    INSERT INTO tags (name, color, user_id)
    VALUES (_name, _color, _user_id) 
    RETURNING id INTO _id;

    INSERT INTO project_tags (project_id, tag_id)
    VALUES (_project_id, _id);

    RETURN _id;
END; $$ LANGUAGE plpgsql;

DROP FUNCTION IF EXISTS sp_update_tag(BIGINT, TEXT, TEXT);
CREATE OR REPLACE FUNCTION sp_update_tag(_id BIGINT, _name TEXT, _color tag_color)
RETURNS VOID AS $$
BEGIN
    UPDATE tags SET name = _name, color = _color WHERE id = _id;
END; $$ LANGUAGE plpgsql;
