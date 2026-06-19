DROP TABLE IF EXISTS rag_learned_synonyms;
DROP TABLE IF EXISTS rag_feedback;
ALTER TABLE document_chunks DROP COLUMN IF EXISTS embedding;
