-- Migration: Update ID columns from INTEGER to BIGINT
-- Date: 2024-06-14
-- Purpose: Ensure all ID columns are BIGINT to match Go's int64 type

-- Update users table
ALTER TABLE users ALTER COLUMN id TYPE BIGINT;
ALTER TABLE users ALTER COLUMN id SET DEFAULT nextval('users_id_seq'::regclass);

-- Update projects table
ALTER TABLE projects ALTER COLUMN id TYPE BIGINT;
ALTER TABLE projects ALTER COLUMN id SET DEFAULT nextval('projects_id_seq'::regclass);

-- Update items table
ALTER TABLE items ALTER COLUMN id TYPE BIGINT;
ALTER TABLE items ALTER COLUMN id SET DEFAULT nextval('items_id_seq'::regclass);

-- Update project_users table
ALTER TABLE project_users ALTER COLUMN project_id TYPE BIGINT;
ALTER TABLE project_users ALTER COLUMN user_id TYPE BIGINT;
ALTER TABLE project_users ALTER COLUMN role_id TYPE BIGINT;

-- Update tags table
ALTER TABLE tags ALTER COLUMN id TYPE BIGINT;
ALTER TABLE tags ALTER COLUMN id SET DEFAULT nextval('tags_id_seq'::regclass);
ALTER TABLE tags ALTER COLUMN user_id TYPE BIGINT;

-- Update project_tags table
ALTER TABLE project_tags ALTER COLUMN project_id TYPE BIGINT;
ALTER TABLE project_tags ALTER COLUMN tag_id TYPE BIGINT;

-- Update roles table
ALTER TABLE roles ALTER COLUMN id TYPE BIGINT;
ALTER TABLE roles ALTER COLUMN id SET DEFAULT nextval('roles_id_seq'::regclass);

-- Update task_tags table (if it exists)
-- ALTER TABLE task_tags ALTER COLUMN task_id TYPE BIGINT;
-- ALTER TABLE task_tags ALTER COLUMN tag_id TYPE BIGINT;

-- Thoroughly update all relevant columns to BIGINT

-- items table
ALTER TABLE items ALTER COLUMN assigned_to TYPE BIGINT;
ALTER TABLE items ALTER COLUMN project_id TYPE BIGINT;
ALTER TABLE items ALTER COLUMN created_by TYPE BIGINT;
ALTER TABLE items ALTER COLUMN modified_by TYPE BIGINT;

-- projects table
ALTER TABLE projects ALTER COLUMN owner TYPE BIGINT;
ALTER TABLE projects ALTER COLUMN created_by TYPE BIGINT;
ALTER TABLE projects ALTER COLUMN modified_by TYPE BIGINT;

-- task_tags table
ALTER TABLE task_tags ALTER COLUMN task_id TYPE BIGINT;
ALTER TABLE task_tags ALTER COLUMN tag_id TYPE BIGINT;

-- Verify the changes
SELECT 
    table_name, 
    column_name, 
    data_type 
FROM information_schema.columns 
WHERE table_schema = 'public' 
    AND column_name = 'id' 
    AND table_name IN ('users', 'projects', 'items', 'tags', 'roles')
ORDER BY table_name; 