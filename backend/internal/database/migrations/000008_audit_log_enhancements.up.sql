ALTER TABLE audit_logs
ADD COLUMN IF NOT EXISTS entity_id UUID,
ADD COLUMN IF NOT EXISTS user_agent TEXT;

DO $$
BEGIN
    IF EXISTS (
        SELECT 1 FROM information_schema.columns
        WHERE table_name = 'audit_logs'
        AND column_name = 'old_value'
        AND udt_name <> 'jsonb'
    ) THEN
        ALTER TABLE audit_logs
        ALTER COLUMN old_value TYPE JSONB USING NULLIF(old_value, '')::jsonb;
    END IF;
END $$;

DO $$
BEGIN
    IF EXISTS (
        SELECT 1 FROM information_schema.columns
        WHERE table_name = 'audit_logs'
        AND column_name = 'new_value'
        AND udt_name <> 'jsonb'
    ) THEN
        ALTER TABLE audit_logs
        ALTER COLUMN new_value TYPE JSONB USING NULLIF(new_value, '')::jsonb;
    END IF;
END $$;

CREATE INDEX IF NOT EXISTS idx_audit_logs_entity_id ON audit_logs(entity_id);
CREATE INDEX IF NOT EXISTS idx_audit_logs_action ON audit_logs(action);
CREATE INDEX IF NOT EXISTS idx_audit_logs_created_at ON audit_logs(created_at);
