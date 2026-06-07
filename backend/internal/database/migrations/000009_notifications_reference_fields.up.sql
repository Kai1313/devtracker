ALTER TABLE notifications
ADD COLUMN IF NOT EXISTS reference_module VARCHAR(100),
ADD COLUMN IF NOT EXISTS reference_id UUID;

DO $$
BEGIN
    IF EXISTS (
        SELECT 1 FROM information_schema.columns
        WHERE table_name = 'notifications'
        AND column_name = 'task_id'
    ) THEN
        UPDATE notifications
        SET reference_module = 'tasks',
            reference_id = task_id
        WHERE task_id IS NOT NULL
        AND reference_id IS NULL;
    END IF;
END $$;

DROP INDEX IF EXISTS idx_notifications_task_id;

ALTER TABLE notifications
DROP COLUMN IF EXISTS task_id;

CREATE INDEX IF NOT EXISTS idx_notifications_reference ON notifications(reference_module, reference_id);
CREATE INDEX IF NOT EXISTS idx_notifications_user_read_created ON notifications(user_id, is_read, created_at DESC);
