DROP INDEX IF EXISTS idx_audit_logs_action;
DROP INDEX IF EXISTS idx_audit_logs_created;

ALTER TABLE sync_jobs DROP COLUMN IF EXISTS conflicts_skipped;
ALTER TABLE documents DROP COLUMN IF EXISTS last_synced_at;
ALTER TABLE documents DROP COLUMN IF EXISTS has_pending_changes;
ALTER TABLE documents DROP COLUMN IF EXISTS synced_content_hash;
