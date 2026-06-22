DROP INDEX IF EXISTS idx_document_chunks_embedding_hnsw;

ALTER TABLE document_chunks DROP COLUMN IF EXISTS embedding_vector;

ALTER TABLE documents DROP COLUMN IF EXISTS sync_snapshot_content;
