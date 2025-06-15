-- INITIAL MIGRATION (Updated with Date/Time Types)

-- Table: users
CREATE TABLE IF NOT EXISTS users (
  id                  SERIAL PRIMARY KEY,
  email               TEXT UNIQUE,
  full_name           TEXT DEFAULT '',
  session_id          TEXT DEFAULT '',
  session_start_time  TIMESTAMP DEFAULT NULL,
  password_hash       TEXT DEFAULT '',
  is_deleted          BOOLEAN DEFAULT FALSE,
  is_admin            BOOLEAN DEFAULT FALSE,
  create_date         TIMESTAMP NOT NULL DEFAULT NOW()
);

-- Table: roles
CREATE TABLE IF NOT EXISTS roles (
  id          SERIAL PRIMARY KEY,
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
  id            SERIAL PRIMARY KEY,
  name          TEXT DEFAULT '' NOT NULL,
  description   TEXT,
  owner         INTEGER,
  start_date    TIMESTAMP DEFAULT NULL, 
  due_date      TIMESTAMP DEFAULT NULL, 
  created_by    INTEGER,
  created_date  TIMESTAMP NOT NULL DEFAULT NOW(), 
  modified_by   INTEGER,
  modified_date TIMESTAMP DEFAULT NULL, 
  is_complete   BOOLEAN DEFAULT FALSE,
  is_deleted    BOOLEAN DEFAULT FALSE,
  FOREIGN KEY (owner) REFERENCES users (id),
  FOREIGN KEY (created_by) REFERENCES users (id),
  FOREIGN KEY (modified_by) REFERENCES users (id)
);

-- Table: tags
CREATE TABLE IF NOT EXISTS tags (
  id      SERIAL PRIMARY KEY,
  name    TEXT,
  user_id INTEGER,
  FOREIGN KEY (user_id) REFERENCES users (id)
);

-- Table: project_tags
CREATE TABLE IF NOT EXISTS project_tags (
  project_id INTEGER,
  tag_id     INTEGER,
  PRIMARY KEY (project_id, tag_id),
  FOREIGN KEY (project_id) REFERENCES projects (id),
  FOREIGN KEY (tag_id) REFERENCES tags (id)
);

-- Table: items (representing tasks based on foreign keys)
CREATE TABLE IF NOT EXISTS items (
  id            SERIAL PRIMARY KEY,
  name          TEXT DEFAULT '' NOT NULL,
  description   TEXT,
  assigned_to   INTEGER,
  project_id    INTEGER,
  status        INTEGER DEFAULT 0,
  complete_date TIMESTAMP DEFAULT NULL, 
  due_date      TIMESTAMP DEFAULT NULL, 
  created_by    INTEGER,
  created_date  TIMESTAMP NOT NULL DEFAULT NOW(), 
  modified_by   INTEGER,
  modified_date TIMESTAMP DEFAULT NULL, 
  is_deleted    BOOLEAN DEFAULT FALSE,
  FOREIGN KEY (assigned_to) REFERENCES users (id),
  FOREIGN KEY (created_by) REFERENCES users (id),
  FOREIGN KEY (modified_by) REFERENCES users (id),
  FOREIGN KEY (project_id) REFERENCES projects (id)
);

-- Table: task_tags
CREATE TABLE IF NOT EXISTS task_tags (
  task_id INTEGER,
  tag_id  INTEGER,
  PRIMARY KEY (task_id, tag_id),
  FOREIGN KEY (task_id) REFERENCES items (id),
  FOREIGN KEY (tag_id) REFERENCES tags (id)
);

-- Table: project_users
CREATE TABLE IF NOT EXISTS project_users (
  project_id INTEGER,
  user_id    INTEGER,
  role_id    INTEGER,
  PRIMARY KEY (project_id, user_id),
  FOREIGN KEY (project_id) REFERENCES projects (id),
  FOREIGN KEY (user_id) REFERENCES users (id),
  FOREIGN KEY (role_id) REFERENCES roles (id)
);
