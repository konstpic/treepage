INSERT INTO system_settings (key, value) VALUES
    ('ui_theme', '"fox_white"')
ON CONFLICT (key) DO NOTHING;
