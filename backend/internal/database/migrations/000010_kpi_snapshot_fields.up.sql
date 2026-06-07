DO $$
BEGIN
    IF EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name = 'kpi_snapshots' AND column_name = 'total_assigned')
    AND NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name = 'kpi_snapshots' AND column_name = 'total_assigned_tasks') THEN
        ALTER TABLE kpi_snapshots RENAME COLUMN total_assigned TO total_assigned_tasks;
    END IF;

    IF EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name = 'kpi_snapshots' AND column_name = 'total_done')
    AND NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name = 'kpi_snapshots' AND column_name = 'total_done_tasks') THEN
        ALTER TABLE kpi_snapshots RENAME COLUMN total_done TO total_done_tasks;
    END IF;

    IF EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name = 'kpi_snapshots' AND column_name = 'total_ready_to_check')
    AND NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name = 'kpi_snapshots' AND column_name = 'total_ready_to_check_tasks') THEN
        ALTER TABLE kpi_snapshots RENAME COLUMN total_ready_to_check TO total_ready_to_check_tasks;
    END IF;

    IF EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name = 'kpi_snapshots' AND column_name = 'total_qa_checked')
    AND NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name = 'kpi_snapshots' AND column_name = 'total_checked_by_qa_tasks') THEN
        ALTER TABLE kpi_snapshots RENAME COLUMN total_qa_checked TO total_checked_by_qa_tasks;
    END IF;

    IF EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name = 'kpi_snapshots' AND column_name = 'delayed_task_count')
    AND NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name = 'kpi_snapshots' AND column_name = 'delayed_tasks') THEN
        ALTER TABLE kpi_snapshots RENAME COLUMN delayed_task_count TO delayed_tasks;
    END IF;

    IF EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name = 'kpi_snapshots' AND column_name = 'total_estimated_point')
    AND NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name = 'kpi_snapshots' AND column_name = 'total_estimated_points') THEN
        ALTER TABLE kpi_snapshots RENAME COLUMN total_estimated_point TO total_estimated_points;
    END IF;

    IF EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name = 'kpi_snapshots' AND column_name = 'total_actual_point')
    AND NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name = 'kpi_snapshots' AND column_name = 'total_actual_points') THEN
        ALTER TABLE kpi_snapshots RENAME COLUMN total_actual_point TO total_actual_points;
    END IF;

    IF EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name = 'kpi_snapshots' AND column_name = 'calculated_at')
    AND NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name = 'kpi_snapshots' AND column_name = 'generated_at') THEN
        ALTER TABLE kpi_snapshots RENAME COLUMN calculated_at TO generated_at;
    END IF;
END $$;

ALTER TABLE kpi_snapshots
ADD COLUMN IF NOT EXISTS total_assigned_tasks INT NOT NULL DEFAULT 0,
ADD COLUMN IF NOT EXISTS total_done_tasks INT NOT NULL DEFAULT 0,
ADD COLUMN IF NOT EXISTS total_ready_to_check_tasks INT NOT NULL DEFAULT 0,
ADD COLUMN IF NOT EXISTS total_checked_by_qa_tasks INT NOT NULL DEFAULT 0,
ADD COLUMN IF NOT EXISTS delayed_tasks INT NOT NULL DEFAULT 0,
ADD COLUMN IF NOT EXISTS total_estimated_points NUMERIC(10,2) NOT NULL DEFAULT 0,
ADD COLUMN IF NOT EXISTS total_actual_points NUMERIC(10,2) NOT NULL DEFAULT 0,
ADD COLUMN IF NOT EXISTS average_completion_time_hours NUMERIC(10,2) NOT NULL DEFAULT 0,
ADD COLUMN IF NOT EXISTS generated_at TIMESTAMP NOT NULL DEFAULT NOW(),
ADD COLUMN IF NOT EXISTS created_at TIMESTAMP NOT NULL DEFAULT NOW();

ALTER TABLE kpi_snapshots
DROP COLUMN IF EXISTS qa_pass_rate;

CREATE UNIQUE INDEX IF NOT EXISTS idx_kpi_snapshots_sprint_developer ON kpi_snapshots(sprint_id, developer_id);
CREATE INDEX IF NOT EXISTS idx_kpi_snapshots_sprint_id ON kpi_snapshots(sprint_id);
