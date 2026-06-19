CREATE TABLE IF NOT EXISTS book_translations (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    book_id UUID NOT NULL REFERENCES books(id) ON DELETE CASCADE,
    locale VARCHAR(8) NOT NULL,
    source_hash VARCHAR(64) NOT NULL,
    title VARCHAR(512) NOT NULL DEFAULT '',
    description TEXT NOT NULL DEFAULT '',
    content TEXT NOT NULL DEFAULT '',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (book_id, locale)
);

CREATE INDEX IF NOT EXISTS idx_book_translations_book ON book_translations(book_id, locale);
