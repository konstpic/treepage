# Установка в Kubernetes / Helm

Production-развёртывание TreePage с помощью open-source Helm charts.

## Предварительные требования

- Kubernetes 1.24+
- Helm 3.10+
- Ingress controller
- PostgreSQL 16+ (внешний)
- Container registry с образами TreePage

## Шаг 1. Сборка образов

```bash
docker build -f deploy/docker/Dockerfile.auth -t treepage/backend-auth:latest .
docker build -f deploy/docker/Dockerfile.server -t treepage/backend-server:latest .
docker build -f deploy/docker/Dockerfile.sync -t treepage/backend-sync:latest .
docker build -f deploy/docker/Dockerfile.frontend -t treepage/frontend:latest .
```

Загрузите образы в ваш registry и укажите `global.imageRegistry` или полные пути в values.

## Шаг 2. Миграции БД

Перед первым деплоем выполните SQL-миграции из `migrations/` на PostgreSQL:

```bash
for f in migrations/*_up.sql; do
  psql -h <postgres-host> -U treepage -d treepage -f "$f"
done
```

Подробнее: [Миграции базы данных](../operator/migrations.md).

## Шаг 3. Выбор схемы деплоя

### Вариант A — Umbrella chart (один release)

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

### Вариант B — Раздельные releases (рекомендуется для GitOps)

```bash
# Backend: маршруты /api, /api/auth
helm upgrade --install treepage-backend backend/ \
  -f backend/values.yaml \
  --set ingress.host=docs.example.com \
  --set postgresql.host=postgres.default.svc \
  --set secret.dbPassword='...' \
  --set secret.jwtSecret='...'

# Frontend: маршрут /
helm upgrade --install treepage-frontend .helm/frontend/ \
  -f .helm/frontend/values.yaml \
  --set ingress.host=docs.example.com \
  --set backend.releaseName=treepage-backend
```

### Вариант C — Single-ingress (frontend проксирует API)

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

## Шаг 4. Секреты

Не храните секреты в git. Варианты:

1. `--set secret.dbPassword=...` при установке
2. `secret.existingSecret: my-treepage-secrets` с ключами: `db-password`, `jwt-secret`, `oidc-client-secret`, `csrf-secret`, `git-access-token`, `git-webhook-secret`
3. External Secrets Operator / Sealed Secrets

Подробнее: [Секреты](../operator/secrets.md).

## Шаг 5. OIDC

Настройте OIDC в `backend/values.yaml`:

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

Redirect URL в OIDC-провайдере: `https://docs.example.com/api/auth/callback`

## Шаг 6. Проверка

```bash
# Lint charts
helm lint backend --strict
helm lint .helm/frontend --strict

# Health checks
kubectl get pods -l app.kubernetes.io/part-of=treepage
curl https://docs.example.com/api/auth/me  # 401 без токена — норма
```

## Маршрутизация Ingress

| Путь | Сервис |
|------|--------|
| `/` | frontend |
| `/api/auth/*` | backend-auth |
| `/api/*` | backend-server |

Sync-сервис доступен только внутри кластера (ClusterIP).

## LLM в Kubernetes

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

## Дополнительно

- [Детальное руководство Helm](../reference/helm-deployment.md)
- [Мониторинг](../operator/monitoring.md)
- [Первоначальная настройка](../getting-started/initial-setup.md)
