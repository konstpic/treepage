INSERT INTO system_settings (key, value)
VALUES ('ui_language', '"en"')
ON CONFLICT (key) DO NOTHING;
