DO $$
BEGIN
    IF EXISTS (
        SELECT 1 FROM information_schema.columns
        WHERE table_name = 'task_statuses' AND column_name = 'name'
    ) AND NOT EXISTS (
        SELECT 1 FROM information_schema.columns
        WHERE table_name = 'task_statuses' AND column_name = 'status_name'
    ) THEN
        ALTER TABLE task_statuses RENAME COLUMN name TO status_name;
    END IF;
END $$;

DO $$
BEGIN
    IF EXISTS (
        SELECT 1 FROM information_schema.columns
        WHERE table_name = 'task_statuses' AND column_name = 'color'
    ) AND NOT EXISTS (
        SELECT 1 FROM information_schema.columns
        WHERE table_name = 'task_statuses' AND column_name = 'color_name'
    ) THEN
        ALTER TABLE task_statuses RENAME COLUMN color TO color_name;
    END IF;
END $$;

ALTER TABLE task_statuses
ADD COLUMN IF NOT EXISTS color_hex VARCHAR(7),
ADD COLUMN IF NOT EXISTS is_active BOOLEAN NOT NULL DEFAULT TRUE;

UPDATE task_statuses
SET color_hex = CASE LOWER(color_name)
    WHEN 'gray' THEN '#6B7280'
    WHEN 'yellow' THEN '#F59E0B'
    WHEN 'blue' THEN '#3B82F6'
    WHEN 'orange' THEN '#F97316'
    WHEN 'green' THEN '#22C55E'
    WHEN 'red' THEN '#EF4444'
    ELSE '#6B7280'
END
WHERE color_hex IS NULL OR color_hex = '';

ALTER TABLE task_statuses
ALTER COLUMN color_hex SET NOT NULL;

CREATE INDEX IF NOT EXISTS idx_task_statuses_is_active ON task_statuses(is_active);

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

