ALTER TABLE notifications
ADD COLUMN IF NOT EXISTS task_id UUID REFERENCES tasks(id) ON DELETE CASCADE;

UPDATE notifications
SET task_id = reference_id
WHERE reference_module = 'tasks'
AND reference_id IS NOT NULL
AND task_id IS NULL;

DROP INDEX IF EXISTS idx_notifications_reference;

ALTER TABLE notifications
DROP COLUMN IF EXISTS reference_id,
DROP COLUMN IF EXISTS reference_module;

CREATE INDEX IF NOT EXISTS idx_notifications_task_id ON notifications(task_id);
