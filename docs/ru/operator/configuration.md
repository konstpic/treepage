# Конфигурация

## Принцип загрузки

```
/opt/app/conf/config.yml  →  Environment variables  →  Validation
```

YAML содержит несекретные defaults. Секреты — только через env.

## Файлы конфигурации

| Файл | Сервис | Ключевые секции |
|------|--------|----------------|
| `backend/auth/conf/config.yml` | auth | server, postgres, oidc, jwt, frontend, security |
| `backend/server/conf/config.yml` | server | server, postgres, jwt, search, security |
| `backend/sync/conf/config.yml` | sync | server, postgres, git |

## backend-auth

```yaml
server:
  host: 0.0.0.0
  port: 8081

postgres:
  host: postgres
  port: 5432
  database: treepage
  user: treepage
  ssl_mode: disable

oidc:
  enabled: true
  issuer_url: https://keycloak.example.com/realms/treepage
  client_id: treepage
  redirect_url: https://docs.example.com/api/auth/callback
  scopes: openid profile email

jwt:
  access_ttl: 15m
  refresh_ttl: 168h
  issuer: treepage-auth
  audience: treepage

frontend:
  url: https://docs.example.com

security:
  rate_limit_rps: 100
  allowed_origins: https://docs.example.com
```

## backend-server

```yaml
server:
  host: 0.0.0.0
  port: 8082

jwt:
  issuer: treepage-auth
  audience: treepage

search:
  default_limit: 20
  max_limit: 100

security:
  rate_limit_rps: 200
  enable_audit_log: true
```

Env override для sync service URL:

```bash
SYNC_SERVICE_URL=http://treepage-backend-sync:8083
```

## backend-sync

```yaml
server:
  host: 0.0.0.0
  port: 8083

git:
  sync_interval: 300s
  work_dir: /data/repos
```

## Frontend

### Development (Vite)

```bash
VITE_USE_PROXY=true
VITE_PROXY_AUTH=http://backend-auth:8081
VITE_PROXY_API=http://backend-server:8082
# или
VITE_API_URL=http://localhost:8082
VITE_AUTH_URL=http://localhost:8081
```

### Production

Runtime config: `/config.json` (генерируется Helm ConfigMap):

```json
{
  "apiUrl": "/api",
  "authUrl": "/api/auth"
}
```

При `frontend.proxy.enabled: true` nginx проксирует API.

## Helm values

Основные файлы:

| Chart | Values |
|-------|--------|
| Backend | `backend/values.yaml` |
| Frontend | `.helm/frontend/values.yaml` |
| Umbrella | `deploy/helm/treepage/values.yaml` |

### Ключевые параметры backend

```yaml
global:
  frontendUrl: https://docs.example.com

ingress:
  host: docs.example.com

postgresql:
  host: postgres.default.svc
  port: 5432
  database: treepage
  user: treepage

auth:
  replicas: 2
  oidc:
    issuerUrl: ...
    clientId: ...

server:
  replicas: 2
  extraEnv: []  # LLM vars

sync:
  replicas: 1
  git:
    syncInterval: 300s
  persistence:
    enabled: true
    size: 2Gi
```

## Runtime settings (UI)

Настройки из **Настройки системы** хранятся в PostgreSQL (`system_settings`) и имеют приоритет над defaults для:

- UI theme / language
- Auth flags (oidc_enabled, local_auth_fallback)
- Git defaults
- Platform settings (search limits, cache, auto_translate)

## CONFIG_PATH

Override пути к config.yml:

```bash
CONFIG_PATH=/custom/path/config.yml
```

## Связанные разделы

- [Секреты](secrets.md)
- [Helm deployment](../reference/helm-deployment.md)
