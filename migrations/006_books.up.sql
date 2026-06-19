-- Saved books (compiled + LLM-readable versions)

CREATE TABLE books (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    space_id UUID NOT NULL REFERENCES spaces(id) ON DELETE CASCADE,
    slug VARCHAR(256) NOT NULL,
    title VARCHAR(512) NOT NULL,
    description TEXT NOT NULL DEFAULT '',
    root_path VARCHAR(512) NOT NULL,
    audience VARCHAR(64) NOT NULL DEFAULT 'developer',
    focus TEXT NOT NULL DEFAULT '',
    status VARCHAR(32) NOT NULL DEFAULT 'draft',
    source_hash VARCHAR(64),
    outline_json JSONB NOT NULL DEFAULT '[]',
    content_markdown TEXT NOT NULL DEFAULT '',
    error_message TEXT NOT NULL DEFAULT '',
    enhanced BOOLEAN NOT NULL DEFAULT false,
    created_by UUID REFERENCES users(id) ON DELETE SET NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    generated_at TIMESTAMPTZ,
    UNIQUE (space_id, slug)
);

CREATE INDEX idx_books_space ON books(space_id, updated_at DESC);
