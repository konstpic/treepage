# Kubernetes / Helm installation

Production deployment of TreePage using open-source Helm charts.

## Prerequisites

- Kubernetes 1.24+
- Helm 3.10+
- Ingress controller
- PostgreSQL 16+ (external)
- Container registry with TreePage images

## Step 1. Build images

```bash
docker build -f deploy/docker/Dockerfile.auth -t treepage/backend-auth:latest .
docker build -f deploy/docker/Dockerfile.server -t treepage/backend-server:latest .
docker build -f deploy/docker/Dockerfile.sync -t treepage/backend-sync:latest .
docker build -f deploy/docker/Dockerfile.frontend -t treepage/frontend:latest .
```

Push images to your registry and set `global.imageRegistry` or full paths in values.

## Step 2. Database migrations

Before the first deploy, run SQL migrations from `migrations/` on PostgreSQL:

```bash
for f in migrations/*_up.sql; do
  psql -h <postgres-host> -U treepage -d treepage -f "$f"
done
```

Details: [Database migrations](../operator/migrations.md).

## Step 3. Choose deployment scheme

### Option A — Umbrella chart (single release)

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

### Option B — Separate releases (recommended for GitOps)

```bash
# Backend: routes /api, /api/auth
helm upgrade --install treepage-backend backend/ \
  -f backend/values.yaml \
  --set ingress.host=docs.example.com \
  --set postgresql.host=postgres.default.svc \
  --set secret.dbPassword='...' \
  --set secret.jwtSecret='...'

# Frontend: route /
helm upgrade --install treepage-frontend .helm/frontend/ \
  -f .helm/frontend/values.yaml \
  --set ingress.host=docs.example.com \
  --set backend.releaseName=treepage-backend
```

### Option C — Single-ingress (frontend proxies API)

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

## Step 4. Secrets

Do not store secrets in git. Options:

1. `--set secret.dbPassword=...` at install time
2. `secret.existingSecret: my-treepage-secrets` with keys: `db-password`, `jwt-secret`, `oidc-client-secret`, `csrf-secret`, `git-access-token`, `git-webhook-secret`
3. External Secrets Operator / Sealed Secrets

Details: [Secrets](../operator/secrets.md).

## Step 5. OIDC

Configure OIDC in `backend/values.yaml`:

```yaml
auth:
  oidc:
    enabled: true
    issuerUrl: https://keycloak.example.com/realms/treepage
    clientId: treepage
    scopes: openid profile email

global:
  frontendUrl: https://docs.example.com
```

Redirect URL in the OIDC provider: `https://docs.example.com/api/auth/callback`

## Step 6. Verify

```bash
# Lint charts
helm lint backend --strict
helm lint .helm/frontend --strict

# Health checks
kubectl get pods -l app.kubernetes.io/part-of=treepage
curl https://docs.example.com/api/auth/me  # 401 without token — expected
```

## Ingress routing

| Path | Service |
|------|---------|
| `/` | frontend |
| `/api/auth/*` | backend-auth |
| `/api/*` | backend-server |

The sync service is cluster-internal only (ClusterIP).

## LLM in Kubernetes

```yaml
# backend/values.yaml
server:
  extraEnv:
    - name: LLM_ENABLED
      value: "true"
    - name: LLM_API_URL
      value: https://api.openai.com/v1
    - name: LLM_API_KEY
      valueFrom:
        secretKeyRef:
          name: treepage-llm
          key: api-key
    - name: LLM_MODEL
      value: gpt-4o-mini
```

## See also

- [Detailed Helm guide](../reference/helm-deployment.md)
- [Monitoring](../operator/monitoring.md)
- [Initial setup](../getting-started/initial-setup.md)
