ALTER TABLE oidc_providers
  ADD COLUMN IF NOT EXISTS sync_groups BOOLEAN NOT NULL DEFAULT false;

UPDATE oidc_providers
SET role_claim = COALESCE(NULLIF(role_claim, ''), 'roles'),
    group_claim = COALESCE(NULLIF(group_claim, ''), 'groups')
WHERE role_claim IS NULL OR role_claim = '' OR group_claim IS NULL OR group_claim = '';
