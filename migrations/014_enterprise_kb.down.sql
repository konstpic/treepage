DROP TABLE IF EXISTS document_chunks;
DROP TABLE IF EXISTS document_view_stats;
DROP TABLE IF EXISTS search_query_log;
ALTER TABLE documents DROP COLUMN IF EXISTS workflow_state;
DROP TABLE IF EXISTS document_comments;
DROP TABLE IF EXISTS page_acl_rules;
