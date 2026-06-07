INSERT INTO roles (name, description) VALUES
('admin', 'System administrator'),
('project_manager', 'Project manager'),
('developer', 'Developer'),
('qa', 'Quality assurance'),
('management', 'Management viewer')
ON CONFLICT (name) DO UPDATE SET
description = EXCLUDED.description;

INSERT INTO permissions (name, description) VALUES
('manage_users', 'Manage user accounts and role assignments'),
('manage_projects', 'Create, update, and delete projects'),
('manage_sprints', 'Create, update, close, and delete sprints'),
('manage_tasks', 'Create, update, and delete tasks'),
('manage_task_statuses', 'Create, update, and delete task statuses'),
('view_assigned_tasks', 'View assigned tasks'),
('view_ready_to_check_tasks', 'View tasks ready to check'),
('update_own_task_status', 'Update own task status'),
('update_qa_status', 'Update QA task status'),
('view_dashboard', 'View dashboard summaries'),
('view_kpi', 'View KPI dashboards'),
('view_reports', 'View reports')
ON CONFLICT (name) DO UPDATE SET
description = EXCLUDED.description;

DELETE FROM role_permissions
USING roles
WHERE role_permissions.role_id = roles.id
AND roles.name IN ('admin', 'project_manager', 'developer', 'qa', 'management');

INSERT INTO role_permissions (role_id, permission_id)
SELECT roles.id, permissions.id
FROM roles
CROSS JOIN permissions
WHERE roles.name = 'admin'
ON CONFLICT DO NOTHING;

INSERT INTO role_permissions (role_id, permission_id)
SELECT roles.id, permissions.id
FROM roles
JOIN permissions ON permissions.name IN (
    'manage_projects',
    'manage_sprints',
    'manage_tasks',
    'view_dashboard',
    'view_kpi'
)
WHERE roles.name = 'project_manager'
ON CONFLICT DO NOTHING;

INSERT INTO role_permissions (role_id, permission_id)
SELECT roles.id, permissions.id
FROM roles
JOIN permissions ON permissions.name IN (
    'view_assigned_tasks',
    'update_own_task_status'
)
WHERE roles.name = 'developer'
ON CONFLICT DO NOTHING;

INSERT INTO role_permissions (role_id, permission_id)
SELECT roles.id, permissions.id
FROM roles
JOIN permissions ON permissions.name IN (
    'view_ready_to_check_tasks',
    'update_qa_status'
)
WHERE roles.name = 'qa'
ON CONFLICT DO NOTHING;

INSERT INTO role_permissions (role_id, permission_id)
SELECT roles.id, permissions.id
FROM roles
JOIN permissions ON permissions.name IN (
    'view_dashboard',
    'view_kpi',
    'view_reports'
)
WHERE roles.name = 'management'
ON CONFLICT DO NOTHING;

INSERT INTO user_roles (user_id, role_id)
SELECT id, role_id
FROM users
ON CONFLICT DO NOTHING;
