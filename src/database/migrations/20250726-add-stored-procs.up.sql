-- Migration: Add stored procedures for all main tables
-- Date: 2024-06-14

-- USERS
DROP FUNCTION IF EXISTS sp_insert_user(TEXT, TEXT, TEXT, TIMESTAMP);
CREATE OR REPLACE FUNCTION sp_insert_user(_email TEXT, _full_name TEXT, _password_hash TEXT, _create_date TIMESTAMP WITHOUT TIME ZONE)
RETURNS BIGINT AS $$
DECLARE _id BIGINT;
BEGIN
    INSERT INTO users (email, full_name, password_hash, create_date)
    VALUES (_email, _full_name, _password_hash, _create_date)
    RETURNING id INTO _id;
    RETURN _id;
END; $$ LANGUAGE plpgsql;

DROP FUNCTION IF EXISTS sp_update_user_details(BIGINT, TEXT);
CREATE OR REPLACE FUNCTION sp_update_user_details(_id BIGINT, _full_name TEXT)
RETURNS VOID AS $$
BEGIN
    UPDATE users SET full_name = _full_name WHERE users.id = _id;
END; $$ LANGUAGE plpgsql;

DROP FUNCTION IF EXISTS sp_update_user_password(BIGINT, TEXT);
CREATE OR REPLACE FUNCTION sp_update_user_password(_id BIGINT, _password_hash TEXT)
RETURNS VOID AS $$
BEGIN
    UPDATE users SET password_hash = _password_hash WHERE users.id = _id;
END; $$ LANGUAGE plpgsql;

DROP FUNCTION IF EXISTS sp_update_user_session(BIGINT, TEXT, TIMESTAMP);
CREATE OR REPLACE FUNCTION sp_update_user_session(_id BIGINT, _session_id TEXT, _session_start_time TIMESTAMP WITHOUT TIME ZONE)
RETURNS VOID AS $$
BEGIN
    UPDATE users SET session_id = _session_id, session_start_time = _session_start_time WHERE users.id = _id;
END; $$ LANGUAGE plpgsql;

DROP FUNCTION IF EXISTS sp_update_user_is_admin(BIGINT, BOOLEAN);
CREATE OR REPLACE FUNCTION sp_update_user_is_admin(_id BIGINT, _is_admin BOOLEAN)
RETURNS VOID AS $$
BEGIN
    UPDATE users SET is_admin = _is_admin WHERE users.id = _id;
END; $$ LANGUAGE plpgsql;

DROP FUNCTION IF EXISTS sp_delete_user(BIGINT);
CREATE OR REPLACE FUNCTION sp_delete_user(_id BIGINT)
RETURNS VOID AS $$
BEGIN
    UPDATE users SET is_deleted = TRUE WHERE users.id = _id;
END; $$ LANGUAGE plpgsql;

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

DROP FUNCTION IF EXISTS sp_get_user_password(BIGINT);
CREATE OR REPLACE FUNCTION sp_get_user_password(_id BIGINT)
RETURNS TEXT AS $$
DECLARE _password TEXT;
BEGIN
    SELECT users.password_hash INTO _password FROM users WHERE users.id = _id;
    RETURN _password;
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

DROP FUNCTION IF EXISTS sp_update_user_is_enabled(_id BIGINT, _is_enabled BOOLEAN);
CREATE OR REPLACE FUNCTION sp_update_user_is_enabled(_id BIGINT, _is_enabled BOOLEAN)
RETURNS TEXT AS $$
DECLARE _email TEXT;
BEGIN
    UPDATE users SET is_enabled = _is_enabled WHERE users.id = _id;
    SELECT users.email INTO _email FROM users WHERE users.id = _id;
    RETURN _email;
END; $$ LANGUAGE plpgsql;

DROP FUNCTION IF EXISTS sp_logout_user(BIGINT);
CREATE OR REPLACE FUNCTION sp_logout_user(_id BIGINT)
RETURNS VOID AS $$
BEGIN
    UPDATE users SET session_id = '', session_start_time = NULL WHERE users.id = _id;
END; $$ LANGUAGE plpgsql;

-- PROJECTS
DROP FUNCTION IF EXISTS sp_insert_project(TEXT, TEXT, BIGINT, DATE, DATE, BIGINT, TIMESTAMP, BIGINT, TIMESTAMP, BOOLEAN, BOOLEAN);
CREATE OR REPLACE FUNCTION sp_insert_project(_name TEXT, _description TEXT, _owner BIGINT, _start_date DATE, _due_date DATE, _created_by BIGINT, _created_date TIMESTAMP WITHOUT TIME ZONE, _modified_by BIGINT, _modified_date TIMESTAMP WITHOUT TIME ZONE, _is_complete BOOLEAN, _is_deleted BOOLEAN)
RETURNS BIGINT AS $$
DECLARE _id BIGINT;
BEGIN
    INSERT INTO projects (name, description, owner, start_date, due_date, created_by, created_date, modified_by, modified_date, is_complete, is_deleted)
    VALUES (_name, _description, _owner, _start_date, _due_date, _created_by, _created_date, _modified_by, _modified_date, _is_complete, _is_deleted)
    RETURNING id INTO _id;
    RETURN _id;
END; $$ LANGUAGE plpgsql;

DROP FUNCTION IF EXISTS sp_update_project(BIGINT, TEXT, TEXT, BIGINT, DATE, DATE, BIGINT, TIMESTAMP);
CREATE OR REPLACE FUNCTION sp_update_project(_id BIGINT, _name TEXT, _description TEXT, _owner BIGINT, _start_date DATE, _due_date DATE, _modified_by BIGINT, _modified_date TIMESTAMP WITHOUT TIME ZONE)
RETURNS VOID AS $$
BEGIN
    UPDATE projects SET name = _name, description = _description, owner = _owner, start_date = _start_date, due_date = _due_date, modified_by = _modified_by, modified_date = _modified_date WHERE projects.id = _id;
END; $$ LANGUAGE plpgsql;

DROP FUNCTION IF EXISTS sp_delete_project(BIGINT, BIGINT, TIMESTAMP);
CREATE OR REPLACE FUNCTION sp_delete_project(_id BIGINT, _modified_by BIGINT, _modified_date TIMESTAMP WITHOUT TIME ZONE)
RETURNS VOID AS $$
BEGIN
    UPDATE projects SET is_deleted = TRUE, modified_by = _modified_by, modified_date = _modified_date WHERE projects.id = _id;
END; $$ LANGUAGE plpgsql;

DROP FUNCTION IF EXISTS sp_get_projects_by_user(BIGINT);
CREATE OR REPLACE FUNCTION sp_get_projects_by_user(_user_id BIGINT)
RETURNS TABLE(id BIGINT, name TEXT, description TEXT, owner BIGINT, start_date TIMESTAMP WITHOUT TIME ZONE, due_date TIMESTAMP WITHOUT TIME ZONE, created_by BIGINT, created_date TIMESTAMP WITHOUT TIME ZONE, modified_by BIGINT, modified_date TIMESTAMP WITHOUT TIME ZONE, is_complete BOOLEAN, is_deleted BOOLEAN) AS $$
BEGIN
    RETURN QUERY SELECT projects.id, projects.name, projects.description, projects.owner, projects.start_date, projects.due_date, projects.created_by, projects.created_date, projects.modified_by, projects.modified_date, projects.is_complete, projects.is_deleted FROM projects JOIN project_users ON projects.id = project_users.project_id WHERE project_users.user_id = _user_id AND projects.is_deleted = FALSE;
END; $$ LANGUAGE plpgsql;

DROP FUNCTION IF EXISTS sp_get_project(BIGINT);
CREATE OR REPLACE FUNCTION sp_get_project(_id BIGINT)
RETURNS TABLE(id BIGINT, name TEXT, description TEXT, owner BIGINT, start_date TIMESTAMP WITHOUT TIME ZONE, due_date TIMESTAMP WITHOUT TIME ZONE, created_by BIGINT, created_date TIMESTAMP WITHOUT TIME ZONE, modified_by BIGINT, modified_date TIMESTAMP WITHOUT TIME ZONE, is_complete BOOLEAN, is_deleted BOOLEAN) AS $$
BEGIN
    RETURN QUERY SELECT projects.id, projects.name, projects.description, projects.owner, projects.start_date, projects.due_date, projects.created_by, projects.created_date, projects.modified_by, projects.modified_date, projects.is_complete, projects.is_deleted FROM projects WHERE projects.id = _id;
END; $$ LANGUAGE plpgsql;

DROP FUNCTION IF EXISTS sp_get_project_count(BIGINT);
CREATE OR REPLACE FUNCTION sp_get_project_count(_user_id BIGINT)
RETURNS BIGINT AS $$
DECLARE _count BIGINT;
BEGIN
    SELECT COUNT(tasks.id) INTO _count FROM tasks WHERE tasks.is_deleted = FALSE AND tasks.assigned_to = _user_id;
    RETURN _count;
END; $$ LANGUAGE plpgsql;

-- TASKS (TASKS)
DROP FUNCTION IF EXISTS sp_insert_task(TEXT, TEXT, INTEGER, BIGINT, DATE, BIGINT, TIMESTAMP, BIGINT, TIMESTAMP, BIGINT);
CREATE OR REPLACE FUNCTION sp_insert_task(_name TEXT, _description TEXT, _status INTEGER, _assigned_to BIGINT, _due_date DATE, _created_by BIGINT, _created_date TIMESTAMP WITHOUT TIME ZONE, _modified_by BIGINT, _modified_date TIMESTAMP WITHOUT TIME ZONE, _project_id BIGINT)
RETURNS BIGINT AS $$
DECLARE _id BIGINT;
BEGIN
    INSERT INTO tasks (name, description, status, assigned_to, due_date, created_by, created_date, modified_by, modified_date, project_id)
    VALUES (_name, _description, _status, _assigned_to, _due_date, _created_by, _created_date, _modified_by, _modified_date, _project_id)
    RETURNING id INTO _id;
    RETURN _id;
END; $$ LANGUAGE plpgsql;

DROP FUNCTION IF EXISTS sp_update_task(BIGINT, TEXT, TEXT, BIGINT, DATE, BIGINT);
CREATE OR REPLACE FUNCTION sp_update_task(_id BIGINT, _name TEXT, _description TEXT, _assigned_to BIGINT, _due_date DATE, _project_id BIGINT)
RETURNS VOID AS $$
BEGIN
    UPDATE tasks 
    SET name = _name, description = _description, assigned_to = _assigned_to, due_date = _due_date, project_id = _project_id 
    WHERE tasks.id = _id;
END; $$ LANGUAGE plpgsql;

DROP FUNCTION IF EXISTS sp_update_task_status(BIGINT, INTEGER, TIMESTAMP, BIGINT, TIMESTAMP);
CREATE OR REPLACE FUNCTION sp_update_task_status(_id BIGINT, _status INTEGER, _complete_date TIMESTAMP WITHOUT TIME ZONE, _modified_by BIGINT, _modified_date TIMESTAMP WITHOUT TIME ZONE)
RETURNS VOID AS $$
BEGIN
    UPDATE tasks 
    SET status = _status, complete_date = _complete_date, modified_by = _modified_by, modified_date = _modified_date 
    WHERE tasks.id = _id;
END; $$ LANGUAGE plpgsql;

DROP FUNCTION IF EXISTS sp_delete_task(BIGINT, BIGINT, TIMESTAMP);
CREATE OR REPLACE FUNCTION sp_delete_task(_id BIGINT, _modified_by BIGINT, _modified_date TIMESTAMP WITHOUT TIME ZONE)
RETURNS VOID AS $$
BEGIN
    UPDATE tasks SET is_deleted = TRUE, modified_by = _modified_by, modified_date = _modified_date WHERE tasks.id = _id;
END; $$ LANGUAGE plpgsql;

DROP FUNCTION IF EXISTS sp_restore_task(BIGINT, BIGINT, TIMESTAMP);
CREATE OR REPLACE FUNCTION sp_restore_task(_id BIGINT, _modified_by BIGINT, _modified_date TIMESTAMP WITHOUT TIME ZONE)
RETURNS VOID AS $$
BEGIN
    UPDATE tasks SET is_deleted = FALSE, modified_by = _modified_by, modified_date = _modified_date WHERE tasks.id = _id;
END; $$ LANGUAGE plpgsql;

DROP FUNCTION IF EXISTS sp_get_tasks_by_user(BIGINT);
CREATE OR REPLACE FUNCTION sp_get_tasks_by_user(_user_id BIGINT)
RETURNS TABLE(id BIGINT, name TEXT, description TEXT, status INTEGER, assigned_to BIGINT, due_date TIMESTAMP WITHOUT TIME ZONE, created_by BIGINT, created_date TIMESTAMP WITHOUT TIME ZONE, modified_by BIGINT, modified_date TIMESTAMP WITHOUT TIME ZONE, is_deleted BOOLEAN, project_id BIGINT, complete_date TIMESTAMP WITHOUT TIME ZONE) AS $$
BEGIN
    RETURN QUERY SELECT tasks.id, tasks.name, tasks.description, tasks.status, tasks.assigned_to, tasks.due_date, tasks.created_by, tasks.created_date, tasks.modified_by, tasks.modified_date, tasks.is_deleted, tasks.project_id, tasks.complete_date FROM tasks WHERE tasks.is_deleted = FALSE AND tasks.assigned_to = _user_id;
END; $$ LANGUAGE plpgsql;

DROP FUNCTION IF EXISTS sp_get_tasks_by_user_and_project(BIGINT, BIGINT);
CREATE OR REPLACE FUNCTION sp_get_tasks_by_user_and_project(_user_id BIGINT, _project_id BIGINT)
RETURNS TABLE(id BIGINT, name TEXT, description TEXT, status INTEGER, assigned_to BIGINT, due_date TIMESTAMP WITHOUT TIME ZONE, created_by BIGINT, created_date TIMESTAMP WITHOUT TIME ZONE, modified_by BIGINT, modified_date TIMESTAMP WITHOUT TIME ZONE, is_deleted BOOLEAN, project_id BIGINT, complete_date TIMESTAMP WITHOUT TIME ZONE) AS $$
BEGIN
    RETURN QUERY SELECT tasks.id, tasks.name, tasks.description, tasks.status, tasks.assigned_to, tasks.due_date, tasks.created_by, tasks.created_date, tasks.modified_by, tasks.modified_date, tasks.is_deleted, tasks.project_id, tasks.complete_date FROM tasks WHERE tasks.is_deleted = FALSE AND tasks.assigned_to = _user_id AND tasks.project_id = _project_id;
END; $$ LANGUAGE plpgsql;

DROP FUNCTION IF EXISTS sp_get_task(BIGINT);
CREATE OR REPLACE FUNCTION sp_get_task(_id BIGINT)
RETURNS TABLE(id BIGINT, name TEXT, description TEXT, status INTEGER, assigned_to BIGINT, due_date TIMESTAMP WITHOUT TIME ZONE, created_by BIGINT, created_date TIMESTAMP WITHOUT TIME ZONE, modified_by BIGINT, modified_date TIMESTAMP WITHOUT TIME ZONE, is_deleted BOOLEAN, project_id BIGINT, complete_date TIMESTAMP WITHOUT TIME ZONE) AS $$
BEGIN
    RETURN QUERY SELECT tasks.id, tasks.name, tasks.description, tasks.status, tasks.assigned_to, tasks.due_date, tasks.created_by, tasks.created_date, tasks.modified_by, tasks.modified_date, tasks.is_deleted, tasks.project_id, tasks.complete_date FROM tasks WHERE tasks.id = _id;
END; $$ LANGUAGE plpgsql;

DROP FUNCTION IF EXISTS sp_get_tasks_count(BIGINT);
CREATE OR REPLACE FUNCTION sp_get_tasks_count(_user_id BIGINT)
RETURNS BIGINT AS $$
DECLARE _count BIGINT;
BEGIN
    SELECT COUNT(tasks.id) INTO _count FROM tasks WHERE tasks.is_deleted = FALSE AND tasks.assigned_to = _user_id;
    RETURN _count;
END; $$ LANGUAGE plpgsql;

DROP FUNCTION IF EXISTS sp_get_total_project_tasks(BIGINT);
CREATE OR REPLACE FUNCTION sp_get_total_project_tasks(_project_id BIGINT)
RETURNS BIGINT AS $$
BEGIN
    RETURN (SELECT COUNT(*) FROM tasks WHERE project_id = _project_id AND is_deleted = FALSE);
END;
$$ LANGUAGE plpgsql;

DROP FUNCTION IF EXISTS sp_get_completed_project_tasks(BIGINT);
CREATE OR REPLACE FUNCTION sp_get_completed_project_tasks(_project_id BIGINT)
RETURNS BIGINT AS $$
BEGIN
    RETURN (SELECT COUNT(*) FROM tasks WHERE project_id = _project_id AND is_deleted = FALSE AND status = 1);
END;
$$ LANGUAGE plpgsql;

-- PROJECT_USERS
DROP FUNCTION IF EXISTS sp_insert_project_user(BIGINT, BIGINT, BIGINT);
CREATE OR REPLACE FUNCTION sp_insert_project_user(_project_id BIGINT, _user_id BIGINT, _role_id BIGINT)
RETURNS VOID AS $$
BEGIN
    INSERT INTO project_users (project_id, user_id, role_id) VALUES (_project_id, _user_id, _role_id);
END; $$ LANGUAGE plpgsql;

DROP FUNCTION IF EXISTS sp_update_project_user(BIGINT, BIGINT, BIGINT);
CREATE OR REPLACE FUNCTION sp_update_project_user(_project_id BIGINT, _user_id BIGINT, _role_id BIGINT)
RETURNS VOID AS $$
BEGIN
    UPDATE project_users SET role_id = _role_id WHERE project_users.project_id = _project_id AND project_users.user_id = _user_id;
END; $$ LANGUAGE plpgsql;

DROP FUNCTION IF EXISTS sp_delete_project_user(BIGINT, BIGINT);
CREATE OR REPLACE FUNCTION sp_delete_project_user(_project_id BIGINT, _user_id BIGINT)
RETURNS VOID AS $$
BEGIN
    DELETE FROM project_users WHERE project_users.project_id = _project_id AND project_users.user_id = _user_id;
END; $$ LANGUAGE plpgsql;

DROP FUNCTION IF EXISTS sp_get_all_project_users_for_project(BIGINT);
CREATE OR REPLACE FUNCTION sp_get_all_project_users_for_project(_project_id BIGINT)
RETURNS TABLE(project_id BIGINT, user_id BIGINT, role_id BIGINT) AS $$
BEGIN
    RETURN QUERY SELECT project_users.project_id, project_users.user_id, project_users.role_id FROM project_users WHERE project_users.project_id = _project_id;
END; $$ LANGUAGE plpgsql;

DROP FUNCTION IF EXISTS sp_get_all_project_users_for_user(BIGINT);
CREATE OR REPLACE FUNCTION sp_get_all_project_users_for_user(_user_id BIGINT)
RETURNS TABLE(project_id BIGINT, user_id BIGINT, role_id BIGINT) AS $$
BEGIN
    RETURN QUERY SELECT project_users.project_id, project_users.user_id, project_users.role_id FROM project_users WHERE project_users.user_id = _user_id;
END; $$ LANGUAGE plpgsql;

DROP FUNCTION IF EXISTS sp_get_project_user_roles(BIGINT, BIGINT);
CREATE OR REPLACE FUNCTION sp_get_project_user_roles(_user_id BIGINT, _project_id BIGINT)
RETURNS TABLE(role_id BIGINT) AS $$
BEGIN
    RETURN QUERY SELECT project_users.role_id FROM project_users WHERE project_users.user_id = _user_id AND project_users.project_id = _project_id;
END; $$ LANGUAGE plpgsql;

-- ROLES
DROP FUNCTION IF EXISTS sp_get_all_roles();
CREATE OR REPLACE FUNCTION sp_get_all_roles()
RETURNS TABLE(id BIGINT, name TEXT, description TEXT) AS $$
BEGIN
    RETURN QUERY SELECT roles.id, roles.name, roles.description FROM roles;
END; $$ LANGUAGE plpgsql; 
