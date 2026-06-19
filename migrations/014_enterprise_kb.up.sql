-- Phase 3: Enterprise KB — page ACL, comments, workflow, analytics, RAG chunks

CREATE TABLE IF NOT EXISTS page_acl_rules (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    space_id UUID NOT NULL REFERENCES spaces(id) ON DELETE CASCADE,
    path_prefix VARCHAR(512) NOT NULL DEFAULT '',
    subject_type VARCHAR(16) NOT NULL CHECK (subject_type IN ('user', 'group')),
    subject_id UUID NOT NULL,
    role VARCHAR(32) NOT NULL DEFAULT 'viewer' CHECK (role IN ('viewer', 'editor', 'admin', 'none')),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (space_id, path_prefix, subject_type, subject_id)
);

CREATE INDEX IF NOT EXISTS idx_page_acl_space_prefix ON page_acl_rules(space_id, path_prefix);

CREATE TABLE IF NOT EXISTS document_comments (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    document_id UUID NOT NULL REFERENCES documents(id) ON DELETE CASCADE,
    parent_id UUID REFERENCES document_comments(id) ON DELETE CASCADE,
    author_id UUID REFERENCES users(id) ON DELETE SET NULL,
    body TEXT NOT NULL,
    mentions UUID[] NOT NULL DEFAULT '{}',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    resolved_at TIMESTAMPTZ
);

CREATE INDEX IF NOT EXISTS idx_document_comments_doc ON document_comments(document_id, created_at ASC);

ALTER TABLE documents ADD COLUMN IF NOT EXISTS workflow_state VARCHAR(32) NOT NULL DEFAULT 'published';
UPDATE documents SET workflow_state = 'draft' WHERE is_published = false AND workflow_state = 'published';

CREATE TABLE IF NOT EXISTS search_query_log (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID REFERENCES users(id) ON DELETE SET NULL,
    query_text VARCHAR(512) NOT NULL,
    result_count INT NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_search_query_log_created ON search_query_log(created_at DESC);

CREATE TABLE IF NOT EXISTS document_view_stats (
    document_id UUID PRIMARY KEY REFERENCES documents(id) ON DELETE CASCADE,
    view_count BIGINT NOT NULL DEFAULT 0,
    last_viewed_at TIMESTAMPTZ
);

CREATE TABLE IF NOT EXISTS document_chunks (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    document_id UUID NOT NULL REFERENCES documents(id) ON DELETE CASCADE,
    chunk_index INT NOT NULL,
    content TEXT NOT NULL,
    content_hash VARCHAR(64) NOT NULL,
    UNIQUE (document_id, chunk_index)
);

CREATE INDEX IF NOT EXISTS idx_document_chunks_doc ON document_chunks(document_id);
