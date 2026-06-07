DELETE FROM permissions
WHERE name = 'view_ready_to_check_tasks';

DELETE FROM role_permissions
USING roles
WHERE role_permissions.role_id = roles.id
AND roles.name IN ('project_manager', 'qa');

INSERT INTO role_permissions (role_id, permission_id)
SELECT roles.id, permissions.id
FROM roles
JOIN permissions ON permissions.name IN (
    'manage_projects',
    'manage_sprints',
    'manage_tasks',
    'view_kpi'
)
WHERE roles.name = 'project_manager'
ON CONFLICT DO NOTHING;

INSERT INTO role_permissions (role_id, permission_id)
SELECT roles.id, permissions.id
FROM roles
JOIN permissions ON permissions.name IN (
    'view_assigned_tasks',
    'update_qa_status'
)
WHERE roles.name = 'qa'
ON CONFLICT DO NOTHING;
