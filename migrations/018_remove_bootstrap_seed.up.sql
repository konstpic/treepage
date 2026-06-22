-- P0: seed data previously applied only via main.go bootstrap SQL.

UPDATE system_settings
SET value = value || '{"auto_translate_docs": true}'::jsonb
WHERE key = 'platform' AND NOT (value ? 'auto_translate_docs');
