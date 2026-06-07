DROP INDEX IF EXISTS idx_audit_logs_entity_id;

ALTER TABLE audit_logs
DROP COLUMN IF EXISTS user_agent,
DROP COLUMN IF EXISTS entity_id;

ALTER TABLE audit_logs
ALTER COLUMN old_value TYPE TEXT USING old_value::text,
ALTER COLUMN new_value TYPE TEXT USING new_value::text;
