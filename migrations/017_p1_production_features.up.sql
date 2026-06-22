-- P1: Git conflict diff snapshot, pgvector for RAG scale

ALTER TABLE documents ADD COLUMN IF NOT EXISTS sync_snapshot_content TEXT NOT NULL DEFAULT '';

UPDATE documents
SET sync_snapshot_content = content
WHERE repository_id IS NOT NULL AND sync_snapshot_content = '';

DO $$
BEGIN
  CREATE EXTENSION IF NOT EXISTS vector;
EXCEPTION
  WHEN OTHERS THEN
    RAISE NOTICE 'pgvector extension unavailable: %', SQLERRM;
END $$;

DO $$
BEGIN
  IF EXISTS (SELECT 1 FROM pg_extension WHERE extname = 'vector') THEN
    ALTER TABLE document_chunks ADD COLUMN IF NOT EXISTS embedding_vector vector(768);
    CREATE INDEX IF NOT EXISTS idx_document_chunks_embedding_hnsw
      ON document_chunks USING hnsw (embedding_vector vector_cosine_ops)
      WHERE embedding_vector IS NOT NULL;
  END IF;
END $$;
