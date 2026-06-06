CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

CREATE TABLE IF NOT EXISTS roles (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name VARCHAR(50) NOT NULL UNIQUE,
    description TEXT,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS users (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    role_id UUID NOT NULL REFERENCES roles(id),
    name VARCHAR(150) NOT NULL,
    email VARCHAR(150) NOT NULL UNIQUE,
    password_hash TEXT NOT NULL,
    team VARCHAR(100),
    position VARCHAR(100),
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMP NULL
);

CREATE TABLE IF NOT EXISTS projects (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    project_code VARCHAR(50) NOT NULL UNIQUE,
    project_name VARCHAR(150) NOT NULL,
    client_name VARCHAR(150),
    status VARCHAR(50) NOT NULL DEFAULT 'active',
    start_date DATE,
    end_date DATE,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMP NULL
);

CREATE TABLE IF NOT EXISTS sprints (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    project_id UUID NOT NULL REFERENCES projects(id),
    sprint_name VARCHAR(150) NOT NULL,
    start_date DATE NOT NULL,
    end_date DATE NOT NULL,
    status VARCHAR(50) NOT NULL DEFAULT 'active',
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMP NULL
);

CREATE TABLE IF NOT EXISTS task_statuses (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    status_name VARCHAR(100) NOT NULL UNIQUE,
    color_name VARCHAR(30) NOT NULL,
    color_hex VARCHAR(7) NOT NULL,
    status_order INT NOT NULL DEFAULT 0,
    is_done BOOLEAN NOT NULL DEFAULT FALSE,
    is_qa_status BOOLEAN NOT NULL DEFAULT FALSE,
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS tasks (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    project_id UUID NOT NULL REFERENCES projects(id),
    sprint_id UUID NOT NULL REFERENCES sprints(id),
    developer_id UUID NOT NULL REFERENCES users(id),
    status_id UUID NOT NULL REFERENCES task_statuses(id),
    ticket_number VARCHAR(100),
    task_title VARCHAR(255) NOT NULL,
    task_description TEXT,
    task_type VARCHAR(100),
    priority VARCHAR(50) NOT NULL DEFAULT 'medium',
    estimated_point NUMERIC(10,2),
    actual_point NUMERIC(10,2),
    start_date DATE,
    due_date DATE,
    completed_date TIMESTAMP NULL,
    qa_checked_date TIMESTAMP NULL,
    created_by UUID REFERENCES users(id),
    updated_by UUID REFERENCES users(id),
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMP NULL
);

CREATE TABLE IF NOT EXISTS task_histories (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    task_id UUID NOT NULL REFERENCES tasks(id),
    old_status_id UUID REFERENCES task_statuses(id),
    new_status_id UUID NOT NULL REFERENCES task_statuses(id),
    changed_by UUID NOT NULL REFERENCES users(id),
    changed_at TIMESTAMP NOT NULL DEFAULT NOW(),
    note TEXT
);

CREATE TABLE IF NOT EXISTS comments (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    task_id UUID NOT NULL REFERENCES tasks(id),
    user_id UUID NOT NULL REFERENCES users(id),
    comment TEXT NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMP NULL
);

CREATE TABLE IF NOT EXISTS attachments (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    task_id UUID NOT NULL REFERENCES tasks(id),
    uploaded_by UUID NOT NULL REFERENCES users(id),
    file_name VARCHAR(255) NOT NULL,
    file_path TEXT NOT NULL,
    mime_type VARCHAR(100),
    file_size BIGINT,
    created_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS kpi_snapshots (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    developer_id UUID NOT NULL REFERENCES users(id),
    sprint_id UUID NOT NULL REFERENCES sprints(id),
    total_assigned INT NOT NULL DEFAULT 0,
    total_done INT NOT NULL DEFAULT 0,
    total_ready_to_check INT NOT NULL DEFAULT 0,
    total_qa_checked INT NOT NULL DEFAULT 0,
    delayed_task_count INT NOT NULL DEFAULT 0,
    completion_rate NUMERIC(5,2) NOT NULL DEFAULT 0,
    qa_pass_rate NUMERIC(5,2) NOT NULL DEFAULT 0,
    total_estimated_point NUMERIC(10,2) NOT NULL DEFAULT 0,
    total_actual_point NUMERIC(10,2) NOT NULL DEFAULT 0,
    calculated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_users_role_id ON users(role_id);
CREATE INDEX IF NOT EXISTS idx_users_deleted_at ON users(deleted_at);
CREATE INDEX IF NOT EXISTS idx_task_statuses_is_active ON task_statuses(is_active);
CREATE INDEX IF NOT EXISTS idx_tasks_project_id ON tasks(project_id);
CREATE INDEX IF NOT EXISTS idx_tasks_sprint_id ON tasks(sprint_id);
CREATE INDEX IF NOT EXISTS idx_tasks_developer_id ON tasks(developer_id);
CREATE INDEX IF NOT EXISTS idx_tasks_status_id ON tasks(status_id);
CREATE INDEX IF NOT EXISTS idx_task_histories_task_id ON task_histories(task_id);
CREATE INDEX IF NOT EXISTS idx_kpi_snapshots_developer_sprint ON kpi_snapshots(developer_id, sprint_id);

INSERT INTO roles (name, description) VALUES
('admin', 'System administrator'),
('project_manager', 'Project manager'),
('developer', 'Developer'),
('qa', 'Quality assurance'),
('management', 'Management viewer')
ON CONFLICT (name) DO NOTHING;

INSERT INTO task_statuses (status_name, color_name, color_hex, status_order, is_done, is_qa_status, is_active) VALUES
('Todo', 'gray', '#6B7280', 1, false, false, true),
('In Progress', 'yellow', '#F59E0B', 2, false, false, true),
('Ready to Check', 'blue', '#3B82F6', 3, false, true, true),
('Checked by QA', 'orange', '#F97316', 4, false, true, true),
('Done', 'green', '#22C55E', 5, true, false, true),
('Blocked', 'red', '#EF4444', 6, false, false, true)
ON CONFLICT (status_name) DO UPDATE SET
color_name = EXCLUDED.color_name,
color_hex = EXCLUDED.color_hex,
status_order = EXCLUDED.status_order,
is_done = EXCLUDED.is_done,
is_qa_status = EXCLUDED.is_qa_status,
is_active = EXCLUDED.is_active;
