DROP INDEX IF EXISTS idx_kpi_snapshots_sprint_developer;

ALTER TABLE kpi_snapshots
DROP COLUMN IF EXISTS average_completion_time_hours,
DROP COLUMN IF EXISTS created_at;
