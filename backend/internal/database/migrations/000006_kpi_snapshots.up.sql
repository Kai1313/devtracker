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

ALTER TABLE kpi_snapshots
ADD COLUMN IF NOT EXISTS total_ready_to_check INT NOT NULL DEFAULT 0,
ADD COLUMN IF NOT EXISTS total_qa_checked INT NOT NULL DEFAULT 0,
ADD COLUMN IF NOT EXISTS delayed_task_count INT NOT NULL DEFAULT 0,
ADD COLUMN IF NOT EXISTS qa_pass_rate NUMERIC(5,2) NOT NULL DEFAULT 0;

CREATE INDEX IF NOT EXISTS idx_kpi_snapshots_developer_sprint ON kpi_snapshots(developer_id, sprint_id);
CREATE INDEX IF NOT EXISTS idx_kpi_snapshots_sprint_id ON kpi_snapshots(sprint_id);
