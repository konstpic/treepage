-- Phase 1: conflict tracking, sync job conflicts, audit index

ALTER TABLE documents ADD COLUMN IF NOT EXISTS synced_content_hash VARCHAR(64);
ALTER TABLE documents ADD COLUMN IF NOT EXISTS has_pending_changes BOOLEAN NOT NULL DEFAULT false;
ALTER TABLE documents ADD COLUMN IF NOT EXISTS last_synced_at TIMESTAMPTZ;

ALTER TABLE sync_jobs ADD COLUMN IF NOT EXISTS conflicts_skipped INT NOT NULL DEFAULT 0;

CREATE INDEX IF NOT EXISTS idx_audit_logs_created ON audit_logs(created_at DESC);
CREATE INDEX IF NOT EXISTS idx_audit_logs_action ON audit_logs(action);

-- Backfill synced hash for existing git-linked documents
UPDATE documents
SET synced_content_hash = encode(sha256(convert_to(content, 'UTF8')), 'hex'),
    last_synced_at = updated_at
WHERE repository_id IS NOT NULL AND synced_content_hash IS NULL;
