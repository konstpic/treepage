-- RAG: vector embeddings, feedback, learned synonyms
ALTER TABLE document_chunks ADD COLUMN IF NOT EXISTS embedding jsonb;

CREATE INDEX IF NOT EXISTS idx_document_chunks_embedding_null
  ON document_chunks ((embedding IS NULL))
  WHERE embedding IS NULL;

CREATE TABLE IF NOT EXISTS rag_feedback (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID REFERENCES users(id) ON DELETE SET NULL,
    question TEXT NOT NULL,
    answer TEXT,
    helpful BOOLEAN NOT NULL,
    confidence REAL,
    sources JSONB NOT NULL DEFAULT '[]',
    citations JSONB NOT NULL DEFAULT '[]',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_rag_feedback_created ON rag_feedback(created_at DESC);
CREATE INDEX IF NOT EXISTS idx_rag_feedback_helpful ON rag_feedback(helpful) WHERE helpful = false;

CREATE TABLE IF NOT EXISTS rag_learned_synonyms (
    term VARCHAR(128) PRIMARY KEY,
    synonyms TEXT[] NOT NULL DEFAULT '{}',
    hit_count INT NOT NULL DEFAULT 1,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
