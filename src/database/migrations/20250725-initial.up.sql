-- INITIAL MIGRATION
-- Date: 2024-07-25

-- Table: users
CREATE TABLE IF NOT EXISTS users (
  id                  BIGSERIAL PRIMARY KEY,
  email               TEXT UNIQUE,
  full_name           TEXT DEFAULT '',
  session_id          TEXT DEFAULT '',
  session_start_time  TIMESTAMP DEFAULT NULL,
  password_hash       TEXT DEFAULT '',
  is_deleted          BOOLEAN DEFAULT FALSE,
  is_admin            BOOLEAN DEFAULT FALSE,
  is_enabled          BOOLEAN NOT NULL DEFAULT FALSE,
  access_granted_by   BIGINT,
  create_date         TIMESTAMP NOT NULL DEFAULT NOW(),
  FOREIGN KEY (access_granted_by) REFERENCES users (id)
);

-- Table: roles
CREATE TABLE IF NOT EXISTS roles (
  id          BIGSERIAL PRIMARY KEY,
  name        TEXT DEFAULT '' NOT NULL,
  description TEXT DEFAULT '' NOT NULL
);

-- Insert into roles
INSERT INTO roles (id, name, description)
VALUES
  (1, 'Big boss', 'Project Administrator.'),
  (2, 'Little boss', 'Can create, assign and complete tasks as well as add/remove users from the project.'),
  (3, 'Grunt', 'Can create, assign and complete tasks.'),
  (4, 'Pleb', 'Can complete tasks.')
ON CONFLICT (id) DO NOTHING;

-- Table: projects
CREATE TABLE IF NOT EXISTS projects (
  id            BIGSERIAL PRIMARY KEY,
  name          TEXT DEFAULT '' NOT NULL,
  description   TEXT,
  owner         BIGINT,
  start_date    TIMESTAMP DEFAULT NULL, 
  due_date      TIMESTAMP DEFAULT NULL, 
  created_by    BIGINT,
  created_date  TIMESTAMP NOT NULL DEFAULT NOW(), 
  modified_by   BIGINT,
  modified_date TIMESTAMP DEFAULT NULL, 
  is_complete   BOOLEAN DEFAULT FALSE,
  is_deleted    BOOLEAN DEFAULT FALSE,
  FOREIGN KEY (owner) REFERENCES users (id),
  FOREIGN KEY (created_by) REFERENCES users (id),
  FOREIGN KEY (modified_by) REFERENCES users (id)
);

-- Table: tasks (representing tasks based on foreign keys)
CREATE TABLE IF NOT EXISTS tasks (
  id            BIGSERIAL PRIMARY KEY,
  name          TEXT DEFAULT '' NOT NULL,
  description   TEXT,
  assigned_to   BIGINT,
  project_id    BIGINT,
  status        INTEGER DEFAULT 0,
  complete_date TIMESTAMP DEFAULT NULL, 
  due_date      TIMESTAMP DEFAULT NULL, 
  created_by    BIGINT,
  created_date  TIMESTAMP NOT NULL DEFAULT NOW(), 
  modified_by   BIGINT,
  modified_date TIMESTAMP DEFAULT NULL, 
  is_deleted    BOOLEAN DEFAULT FALSE,
  FOREIGN KEY (assigned_to) REFERENCES users (id),
  FOREIGN KEY (created_by) REFERENCES users (id),
  FOREIGN KEY (modified_by) REFERENCES users (id),
  FOREIGN KEY (project_id) REFERENCES projects (id)
);

-- Table: project_users
CREATE TABLE IF NOT EXISTS project_users (
  project_id BIGINT,
  user_id    BIGINT,
  role_id    BIGINT,
  PRIMARY KEY (project_id, user_id),
  FOREIGN KEY (project_id) REFERENCES projects (id),
  FOREIGN KEY (user_id) REFERENCES users (id),
  FOREIGN KEY (role_id) REFERENCES roles (id)
);
