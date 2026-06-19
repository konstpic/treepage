# TreePage — Kubernetes / Helm deployment

Open-source Helm charts with **no external chart dependencies**. Standard Kubernetes resources only (Deployment, Service, Ingress, ConfigMap, Secret).

## Charts

| Chart | Path | Components |
|-------|------|------------|
| Backend | [`backend/`](../../backend/) | auth, server, sync |
| Frontend | [`.helm/frontend/`](../../.helm/frontend/) | React SPA + nginx |
| Umbrella (optional) | [`deploy/helm/treepage/`](../../deploy/helm/treepage/) | backend + frontend in one release |

## Prerequisites

- Kubernetes 1.24+
- Helm 3.10+
- Ingress controller (nginx recommended)
- PostgreSQL 16+ (external or managed)
- Container images built from this repo (see below)

## Build images

From repository root:

```bash
docker build -f deploy/docker/Dockerfile.auth -t treepage/backend-auth:latest .
docker build -f deploy/docker/Dockerfile.server -t treepage/backend-server:latest .
docker build -f deploy/docker/Dockerfile.sync -t treepage/backend-sync:latest .
docker build -f deploy/docker/Dockerfile.frontend -t treepage/frontend:latest .
```

Push to your registry and set `global.imageRegistry` or full `*.image.repository` in values.

## Quick install (umbrella)

```bash
helm dependency update deploy/helm/treepage

helm upgrade --install treepage deploy/helm/treepage \
  -f deploy/helm/treepage/values.yaml \
  --set backend.secret.dbPassword='...' \
  --set backend.secret.jwtSecret='...' \
  --set backend.postgresql.host='your-postgres-host' \
  --set backend.ingress.host='docs.example.com' \
  --set frontend.ingress.host='docs.example.com'
```

## Split releases (recommended for GitOps)

Two releases on the same host — ingress merges paths:

```bash
# Backend: /api, /api/auth
helm upgrade --install treepage-backend backend/ \
  -f backend/values.yaml \
  --set ingress.host=docs.example.com \
  --set postgresql.host=postgres.default.svc \
  --set secret.dbPassword='...' \
  --set secret.jwtSecret='...'

# Frontend: /
helm upgrade --install treepage-frontend .helm/frontend/ \
  -f .helm/frontend/values.yaml \
  --set ingress.host=docs.example.com \
  --set backend.releaseName=treepage-backend
```

## Single-ingress mode (frontend proxies API)

Disable backend ingress and enable nginx proxy in frontend:

```yaml
# backend values
ingress:
  enabled: false

# frontend values
frontend:
  proxy:
    enabled: true
backend:
  releaseName: treepage-backend
```

## Migration from corporate uc / react-spa charts

| Legacy (uc / react-spa) | New chart |
|-------------------------|-----------|
| `global.registry_url` | `global.imageRegistry` |
| `global.settings.relations.frontend_url` | `global.frontendUrl` or `ingress.host` |
| `global.settings.db.*` | `postgresql.*` |
| `secret.vault` | `secret.*` or `secret.existingSecret` |
| `auth.configs.config.yml` | auto-generated ConfigMap |
| `auth / server / sync` aliases | same keys under `auth`, `server`, `sync` |
| `frontend.configs.config.json` | auto-generated ConfigMap |

Service DNS names follow `{release}-auth`, `{release}-server`, `{release}-sync`.

## Secrets (production)

Do not commit real secrets. Options:

1. `--set secret.dbPassword=...` on install
2. `secret.existingSecret: my-treepage-secrets` with keys: `db-password`, `jwt-secret`, `oidc-client-secret`, `csrf-secret`, `git-access-token`, `git-webhook-secret`
3. External Secrets Operator / Sealed Secrets

Подробнее: [Секреты](../operator/secrets.md)

## Database migrations

Run SQL migrations from [`migrations/`](../../migrations/) against PostgreSQL before first deploy.

Подробнее: [Миграции](../operator/migrations.md)

## Lint

```bash
helm lint backend --strict
helm lint .helm/frontend --strict
helm lint deploy/helm/treepage --strict
```

## Связанные разделы

- [Kubernetes / Helm (руководство)](../installation/kubernetes.md)
- [Конфигурация](../operator/configuration.md)
- [Мониторинг](../operator/monitoring.md)
