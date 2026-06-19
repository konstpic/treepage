INSERT INTO system_settings (key, value)
VALUES ('platform', '{"search_default_limit":20,"search_max_limit":100,"cache_enabled":false,"logging_level":"info","auto_translate_docs":true}')
ON CONFLICT (key) DO UPDATE
SET value = system_settings.value || '{"auto_translate_docs": true}'::jsonb
WHERE NOT (system_settings.value ? 'auto_translate_docs');

CREATE TABLE IF NOT EXISTS document_translations (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    document_id UUID NOT NULL REFERENCES documents(id) ON DELETE CASCADE,
    locale VARCHAR(8) NOT NULL,
    source_hash VARCHAR(64) NOT NULL,
    title VARCHAR(512) NOT NULL DEFAULT '',
    content TEXT NOT NULL DEFAULT '',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (document_id, locale)
);

CREATE INDEX IF NOT EXISTS idx_document_translations_doc ON document_translations(document_id, locale);
