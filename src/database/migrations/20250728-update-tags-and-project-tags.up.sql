-- Migration: Update tags table and project tag stored procedures
-- Date: 2024-07-28

-- Create tag_color ENUM
DO $$ BEGIN
    IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'tag_color') THEN
        CREATE TYPE tag_color AS ENUM (
            'red',
            'blue',
            'green',
            'yellow',
            'purple',
            'orange',
            'pink',
            'brown',
            'black',
            'white'
        );
    END IF;
END $$;

-- Table: tags
CREATE TABLE IF NOT EXISTS tags (
  id      BIGSERIAL PRIMARY KEY,
  name    TEXT NOT NULL,
  user_id BIGINT NOT NULL,
  color tag_color NOT NULL DEFAULT 'blue',
  FOREIGN KEY (user_id) REFERENCES users (id)
);

-- Table: project_tags
CREATE TABLE IF NOT EXISTS project_tags (
  project_id BIGINT,
  tag_id     BIGINT,
  PRIMARY KEY (project_id, tag_id),
  FOREIGN KEY (project_id) REFERENCES projects (id),
  FOREIGN KEY (tag_id) REFERENCES tags (id) ON DELETE CASCADE
);
-- Table: task_tags
CREATE TABLE IF NOT EXISTS task_tags (
  task_id BIGINT,
  tag_id  BIGINT,
  PRIMARY KEY (task_id, tag_id),
  FOREIGN KEY (task_id) REFERENCES items (id),
  FOREIGN KEY (tag_id) REFERENCES tags (id) ON DELETE CASCADE
);

-- 2. Create tag
DROP FUNCTION IF EXISTS sp_insert_tag(TEXT, BIGINT, BIGINT);
CREATE OR REPLACE FUNCTION sp_insert_tag(_name TEXT, _user_id BIGINT, _project_id BIGINT)
RETURNS BIGINT AS $$
DECLARE _id BIGINT;
BEGIN
    INSERT INTO tags (name, user_id)
    VALUES (_tag_name, _user_id)
    RETURNING id INTO _new_tag_id;

    INSERT INTO project_tags (project_id, tag_id)
    VALUES (_project_id, _id);

    RETURN _id;
END; $$ LANGUAGE plpgsql;

-- 3. Delete tag
DROP FUNCTION IF EXISTS sp_delete_tag(BIGINT, BIGINT);
CREATE OR REPLACE FUNCTION sp_delete_tag(_id BIGINT)
RETURNS VOID AS $$
BEGIN
    DELETE FROM project_tags WHERE project_tags.tag_id = _id; 
    DELETE FROM tags WHERE tags.id = _id;
END; $$ LANGUAGE plpgsql;

-- 4. Get tag by id
DROP FUNCTION IF EXISTS sp_get_tag(BIGINT);
CREATE OR REPLACE FUNCTION sp_get_tag(_id BIGINT)
RETURNS TABLE(id BIGINT, name TEXT) AS $$
BEGIN
    RETURN QUERY SELECT 
        tags.id,
        tags.name 
    FROM tags 
    WHERE tags.id = _id;
END; $$ LANGUAGE plpgsql;

-- 5. Get tags by project id
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

-- 6. Update tags by project id
DROP FUNCTION IF EXISTS sp_update_tag(BIGINT, TEXT, TEXT);
CREATE OR REPLACE FUNCTION sp_update_tag(_id BIGINT, _name TEXT, _color TEXT)
RETURNS VOID AS $$
BEGIN
    UPDATE tags SET name = _name, color = _color, project = _project WHERE id = _id;
END; $$ LANGUAGE plpgsql;
