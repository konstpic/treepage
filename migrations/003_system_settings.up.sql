CREATE TABLE IF NOT EXISTS system_settings (
    key VARCHAR(128) PRIMARY KEY,
    value JSONB NOT NULL DEFAULT '{}',
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_by UUID REFERENCES users(id) ON DELETE SET NULL
);

INSERT INTO system_settings (key, value) VALUES
    ('platform', '{"search_default_limit":20,"search_max_limit":100,"cache_enabled":false,"logging_level":"info"}'),
    ('auth', '{"oidc_enabled":true,"local_auth_fallback":false}'),
    ('git', '{"access_token_ref":"GIT_ACCESS_TOKEN","webhook_secret_ref":"GIT_WEBHOOK_SECRET","default_sync_interval_seconds":300,"default_sync_mode":"scheduled"}')
ON CONFLICT (key) DO NOTHING;
