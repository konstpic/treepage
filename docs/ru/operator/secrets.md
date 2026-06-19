# Секреты и переменные окружения

## Обязательные секреты

| Переменная | Сервисы | Описание |
|------------|---------|----------|
| `DB_PASSWORD` | auth, server, sync | Пароль PostgreSQL |
| `JWT_SECRET` | auth, server | Подпись JWT (должен совпадать!) |

## OIDC

| Переменная | Сервис | Описание |
|------------|--------|----------|
| `OIDC_CLIENT_SECRET` | auth | Client secret OIDC-провайдера |
| `CSRF_SECRET` | auth | CSRF/state token (опционально) |

## Git Sync

| Переменная | Сервис | Описание |
|------------|--------|----------|
| `GIT_ACCESS_TOKEN` | sync | HTTPS token для Git |
| `GIT_WEBHOOK_SECRET` | sync | Secret для webhook validation |

## LLM (опционально)

| Переменная | Сервис | Описание |
|------------|--------|----------|
| `LLM_ENABLED` | server | `true` / `false` |
| `LLM_API_URL` | server | OpenAI-compatible API URL |
| `LLM_API_KEY` | server | API key |
| `LLM_MODEL` | server | Model name |

## Dev-режим

| Переменная | Сервис | Описание |
|------------|--------|----------|
| `DEV_MODE` | auth | `true` — локальный вход |
| `ENV` | auth | `prod` отключает dev login |

## Прочие

| Переменная | Сервис | Описание |
|------------|--------|----------|
| `CONFIG_PATH` | all | Путь к config.yml |
| `SYNC_SERVICE_URL` | server | URL sync worker |

## Kubernetes Secret

Helm создаёт Secret с ключами:

| Key | Env var |
|-----|---------|
| `db-password` | `DB_PASSWORD` |
| `jwt-secret` | `JWT_SECRET` |
| `oidc-client-secret` | `OIDC_CLIENT_SECRET` |
| `csrf-secret` | `CSRF_SECRET` |
| `git-access-token` | `GIT_ACCESS_TOKEN` |
| `git-webhook-secret` | `GIT_WEBHOOK_SECRET` |

### Существующий Secret

```yaml
secret:
  create: false
  existingSecret: my-treepage-secrets
```

### Установка через --set

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

## Генерация секретов

```bash
# JWT secret
openssl rand -base64 48

# CSRF secret
openssl rand -hex 32

# Webhook secret
openssl rand -hex 16
```

## Безопасность

- ❌ Не коммитьте секреты в git
- ❌ Не используйте defaults в production
- ✅ Используйте External Secrets Operator / Vault / Sealed Secrets
- ✅ Ротируйте `JWT_SECRET` с осторожностью (инвалидирует все сессии)
- ✅ Минимальные права для Git token (read-only)

## Per-repo token refs

В настройках репозитория поле **Токен доступа** может содержать:

1. **Env ref** — имя переменной: `GIT_ACCESS_TOKEN`
2. **Literal** — сам token (не рекомендуется)

Global defaults задаются в **Настройки системы** → **Git-интеграция**.

## Связанные разделы

- [Конфигурация](configuration.md)
- [Git Sync](../admin/git-sync.md)
