CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

CREATE TABLE roles (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name VARCHAR(50) NOT NULL UNIQUE,
    description TEXT,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE TABLE users (
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

CREATE TABLE permissions (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name VARCHAR(100) NOT NULL UNIQUE,
    description TEXT,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE TABLE role_permissions (
    role_id UUID NOT NULL REFERENCES roles(id) ON DELETE CASCADE,
    permission_id UUID NOT NULL REFERENCES permissions(id) ON DELETE CASCADE,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    PRIMARY KEY (role_id, permission_id)
);

CREATE TABLE user_roles (
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    role_id UUID NOT NULL REFERENCES roles(id) ON DELETE CASCADE,
    assigned_at TIMESTAMP NOT NULL DEFAULT NOW(),
    PRIMARY KEY (user_id, role_id)
);

CREATE TABLE projects (
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

CREATE TABLE sprints (
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

CREATE TABLE task_statuses (
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

CREATE TABLE tasks (
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

CREATE TABLE task_histories (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    task_id UUID NOT NULL REFERENCES tasks(id),
    old_status_id UUID REFERENCES task_statuses(id),
    new_status_id UUID NOT NULL REFERENCES task_statuses(id),
    changed_by UUID NOT NULL REFERENCES users(id),
    changed_at TIMESTAMP NOT NULL DEFAULT NOW(),
    note TEXT
);

CREATE TABLE comments (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    task_id UUID NOT NULL REFERENCES tasks(id),
    user_id UUID NOT NULL REFERENCES users(id),
    comment TEXT NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMP NULL
);

CREATE TABLE attachments (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    task_id UUID NOT NULL REFERENCES tasks(id),
    uploaded_by UUID NOT NULL REFERENCES users(id),
    file_name VARCHAR(255) NOT NULL,
    file_path TEXT NOT NULL,
    mime_type VARCHAR(100),
    file_size BIGINT,
    created_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE TABLE kpi_snapshots (
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

CREATE INDEX idx_users_role_id ON users(role_id);
CREATE INDEX idx_users_deleted_at ON users(deleted_at);
CREATE INDEX idx_role_permissions_permission_id ON role_permissions(permission_id);
CREATE INDEX idx_user_roles_role_id ON user_roles(role_id);
CREATE INDEX idx_task_statuses_is_active ON task_statuses(is_active);
CREATE INDEX idx_tasks_project_id ON tasks(project_id);
CREATE INDEX idx_tasks_sprint_id ON tasks(sprint_id);
CREATE INDEX idx_tasks_developer_id ON tasks(developer_id);
CREATE INDEX idx_tasks_status_id ON tasks(status_id);
CREATE INDEX idx_task_histories_task_id ON task_histories(task_id);
CREATE INDEX idx_kpi_snapshots_developer_sprint ON kpi_snapshots(developer_id, sprint_id);

INSERT INTO roles (name, description) VALUES
('admin', 'System administrator'),
('project_manager', 'Project manager'),
('developer', 'Developer'),
('qa', 'Quality assurance'),
('management', 'Management viewer');

INSERT INTO permissions (name, description) VALUES
('manage_users', 'Manage user accounts and role assignments'),
('manage_projects', 'Create, update, and delete projects'),
('manage_sprints', 'Create, update, close, and delete sprints'),
('manage_tasks', 'Create, update, and delete tasks'),
('manage_task_statuses', 'Create, update, and delete task statuses'),
('view_assigned_tasks', 'View assigned tasks'),
('update_own_task_status', 'Update own task status'),
('update_qa_status', 'Update QA task status'),
('view_dashboard', 'View dashboard summaries'),
('view_kpi', 'View KPI dashboards'),
('view_reports', 'View reports');

INSERT INTO role_permissions (role_id, permission_id)
SELECT roles.id, permissions.id
FROM roles
CROSS JOIN permissions
WHERE roles.name = 'admin';

INSERT INTO role_permissions (role_id, permission_id)
SELECT roles.id, permissions.id
FROM roles
JOIN permissions ON permissions.name IN (
    'manage_projects',
    'manage_sprints',
    'manage_tasks',
    'view_kpi'
)
WHERE roles.name = 'project_manager';

INSERT INTO role_permissions (role_id, permission_id)
SELECT roles.id, permissions.id
FROM roles
JOIN permissions ON permissions.name IN (
    'view_assigned_tasks',
    'update_own_task_status'
)
WHERE roles.name = 'developer';

INSERT INTO role_permissions (role_id, permission_id)
SELECT roles.id, permissions.id
FROM roles
JOIN permissions ON permissions.name IN (
    'view_assigned_tasks',
    'update_qa_status'
)
WHERE roles.name = 'qa';

INSERT INTO role_permissions (role_id, permission_id)
SELECT roles.id, permissions.id
FROM roles
JOIN permissions ON permissions.name IN (
    'view_dashboard',
    'view_kpi',
    'view_reports'
)
WHERE roles.name = 'management';

INSERT INTO task_statuses (status_name, color_name, color_hex, status_order, is_done, is_qa_status, is_active) VALUES
('Todo', 'gray', '#6B7280', 1, false, false, true),
('In Progress', 'yellow', '#F59E0B', 2, false, false, true),
('Ready to Check', 'blue', '#3B82F6', 3, false, true, true),
('Checked by QA', 'orange', '#F97316', 4, false, true, true),
('Done', 'green', '#22C55E', 5, true, false, true),
('Blocked', 'red', '#EF4444', 6, false, false, true);
