-- Migration: Revert tag queries --
-- Date: 2025-08-04 --

DROP FUNCTION IF EXISTS sp_get_tag(_id BIGINT, _project_id BIGINT);
CREATE OR REPLACE FUNCTION sp_get_tag(_id BIGINT)
RETURNS TABLE(id BIGINT, name TEXT) AS $$
BEGIN
    RETURN QUERY SELECT 
        tags.id,
        tags.name 
    FROM tags 
    WHERE tags.id = _id;
END; $$ LANGUAGE plpgsql;

DROP FUNCTION IF EXISTS sp_get_project_tags(BIGINT);
CREATE OR REPLACE FUNCTION sp_get_project_tags(_project_id BIGINT)
RETURNS TABLE(id BIGINT, name TEXT, color TEXT) AS $$
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

DROP FUNCTION IF EXISTS sp_insert_tag(_name TEXT, color tag_color, _user_id BIGINT, _project_id BIGINT);
CREATE OR REPLACE FUNCTION sp_insert_tag(_name TEXT, _user_id BIGINT, _project_id BIGINT)
RETURNS BIGINT AS $$
DECLARE
    _id BIGINT;
BEGIN
    INSERT INTO tags (name, user_id)
    VALUES (_name, _user_id) 
    RETURNING id INTO _id;

    INSERT INTO project_tags (project_id, tag_id)
    VALUES (_project_id, _id);

    RETURN _id;
END; $$ LANGUAGE plpgsql;
