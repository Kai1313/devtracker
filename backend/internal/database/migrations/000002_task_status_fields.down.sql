DROP INDEX IF EXISTS idx_task_statuses_is_active;

ALTER TABLE task_statuses
DROP COLUMN IF EXISTS is_active,
DROP COLUMN IF EXISTS color_hex;

DO $$
BEGIN
    IF EXISTS (
        SELECT 1 FROM information_schema.columns
        WHERE table_name = 'task_statuses' AND column_name = 'status_name'
    ) AND NOT EXISTS (
        SELECT 1 FROM information_schema.columns
        WHERE table_name = 'task_statuses' AND column_name = 'name'
    ) THEN
        ALTER TABLE task_statuses RENAME COLUMN status_name TO name;
    END IF;
END $$;

DO $$
BEGIN
    IF EXISTS (
        SELECT 1 FROM information_schema.columns
        WHERE table_name = 'task_statuses' AND column_name = 'color_name'
    ) AND NOT EXISTS (
        SELECT 1 FROM information_schema.columns
        WHERE table_name = 'task_statuses' AND column_name = 'color'
    ) THEN
        ALTER TABLE task_statuses RENAME COLUMN color_name TO color;
    END IF;
END $$;

