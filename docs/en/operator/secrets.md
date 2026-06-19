# Secrets and environment variables

## Required secrets

| Variable | Services | Description |
|----------|----------|-------------|
| `DB_PASSWORD` | auth, server, sync | PostgreSQL password |
| `JWT_SECRET` | auth, server | JWT signing (must match!) |

## OIDC

| Variable | Service | Description |
|----------|---------|-------------|
| `OIDC_CLIENT_SECRET` | auth | OIDC provider client secret |
| `CSRF_SECRET` | auth | CSRF/state token (optional) |

## Git Sync

| Variable | Service | Description |
|----------|---------|-------------|
| `GIT_ACCESS_TOKEN` | sync | HTTPS token for Git |
| `GIT_WEBHOOK_SECRET` | sync | Secret for webhook validation |

## LLM (optional)

| Variable | Service | Description |
|----------|---------|-------------|
| `LLM_ENABLED` | server | `true` / `false` |
| `LLM_API_URL` | server | OpenAI-compatible API URL |
| `LLM_API_KEY` | server | API key |
| `LLM_MODEL` | server | Model name |

## Dev mode

| Variable | Service | Description |
|----------|---------|-------------|
| `DEV_MODE` | auth | `true` — local login |
| `ENV` | auth | `prod` disables dev login |

## Other

| Variable | Service | Description |
|----------|---------|-------------|
| `CONFIG_PATH` | all | Path to config.yml |
| `SYNC_SERVICE_URL` | server | Sync worker URL |

## Kubernetes Secret

Helm creates a Secret with keys:

| Key | Env var |
|-----|---------|
| `db-password` | `DB_PASSWORD` |
| `jwt-secret` | `JWT_SECRET` |
| `oidc-client-secret` | `OIDC_CLIENT_SECRET` |
| `csrf-secret` | `CSRF_SECRET` |
| `git-access-token` | `GIT_ACCESS_TOKEN` |
| `git-webhook-secret` | `GIT_WEBHOOK_SECRET` |

### Existing Secret

```yaml
secret:
  create: false
  existingSecret: my-treepage-secrets
```

### Install via --set

```bash
helm upgrade --install treepage-backend backend/ \
  --set secret.dbPassword='strong-password' \
  --set secret.jwtSecret='long-random-string-min-32-chars' \
  --set secret.oidcClientSecret='oidc-secret' \
  --set secret.gitAccessToken='ghp_xxxx'
```

## Docker Compose (.env)

```bash
DB_PASSWORD=treepage
JWT_SECRET=dev-jwt-secret-change-in-production
GIT_ACCESS_TOKEN=ghp_xxxxxxxxxxxx
GIT_WEBHOOK_SECRET=my-webhook-secret
OIDC_CLIENT_SECRET=your-oidc-secret
```

## Generating secrets

```bash
# JWT secret
openssl rand -base64 48

# CSRF secret
openssl rand -hex 32

# Webhook secret
openssl rand -hex 16
```

## Security

- ❌ Do not commit secrets to git
- ❌ Do not use defaults in production
- ✅ Use External Secrets Operator / Vault / Sealed Secrets
- ✅ Rotate `JWT_SECRET` carefully (invalidates all sessions)
- ✅ Minimum permissions for Git token (read-only)

## Per-repo token refs

In repository settings, the **Access token** field can contain:

1. **Env ref** — variable name: `GIT_ACCESS_TOKEN`
2. **Literal** — the token itself (not recommended)

Global defaults are set in **System settings** → **Git integration**.

## Related sections

- [Configuration](configuration.md)
- [Git Sync](../admin/git-sync.md)
