# REST API

TreePage предоставляет REST API через два backend-сервиса. Полная спецификация: [`openapi/openapi.yaml`](../../openapi/openapi.yaml).

## Base URLs

| Сервис | Dev | Production (via Ingress) |
|--------|-----|--------------------------|
| Auth | `http://localhost:8081` | `https://docs.example.com/api/auth` |
| Server | `http://localhost:8082` | `https://docs.example.com/api` |
| Sync | `http://localhost:8083` | Internal only |

## Аутентификация

Большинство endpoints требуют JWT Bearer token:

```
Authorization: Bearer <access_token>
```

Получение токена: OIDC flow или `POST /api/auth/login` (dev only).

## Auth API (`backend-auth`)

| Method | Path | Auth | Описание |
|--------|------|:----:|----------|
| GET | `/api/auth/login` | ❌ | Начать OIDC flow |
| POST | `/api/auth/login` | ❌ | Локальный вход (dev) |
| GET | `/api/auth/callback` | ❌ | OIDC callback |
| POST | `/api/auth/refresh` | ❌ | Обновить access token |
| POST | `/api/auth/logout` | ✅ | Выход |
| GET | `/api/auth/me` | ✅ | Профиль текущего пользователя |

## Public API (`backend-server`)

| Method | Path | Auth | Описание |
|--------|------|:----:|----------|
| GET | `/api/public/branding` | ❌ | Брендинг платформы |
| GET | `/api/public/ui-theme` | ❌ | Текущая UI тема |
| GET | `/api/public/spaces` | ❌ | Публичные пространства |

## Spaces & Documents

| Method | Path | Auth | Описание |
|--------|------|:----:|----------|
| GET | `/api/spaces` | ✅ | Список пространств |
| GET | `/api/spaces/{slug}` | ✅/❌ | Метаданные пространства |
| GET | `/api/spaces/{slug}/documents` | ✅/❌ | Дерево документов |
| GET | `/api/spaces/{slug}/documents/{docSlug}` | ✅/❌ | Содержимое документа |
| PUT | `/api/spaces/{slug}/documents/{docSlug}` | ✅ | Обновить документ |
| POST | `/api/spaces/{slug}/documents` | ✅ | Создать документ |

## Search

| Method | Path | Auth | Описание |
|--------|------|:----:|----------|
| GET | `/api/search` | ✅/❌ | Полнотекстовый поиск |

Query params: `q`, `space`, `author`, `tags`, `limit`, `offset`

## Admin API (`backend-server`)

> Требуется роль `admin` или `super_admin`.

| Method | Path | Min role | Описание |
|--------|------|----------|----------|
| GET/PUT | `/api/admin/system-settings` | admin/super | Системные настройки |
| PUT | `/api/admin/system-settings/ui-theme` | super | UI тема |
| PUT | `/api/admin/system-settings/ui-language` | super | UI язык |
| GET/POST | `/api/admin/repositories` | admin | CRUD репозиториев |
| GET/PUT/DELETE | `/api/admin/repositories/{id}` | admin | Управление репозиторием |
| GET/POST | `/api/admin/spaces` | admin | CRUD пространств |
| PATCH | `/api/admin/spaces/{id}` | admin | Обновить пространство |
| POST | `/api/admin/spaces/{id}/bind-repo` | admin | Привязать репозиторий |
| POST | `/api/admin/sync/{repoId}` | admin | Trigger sync |
| GET/POST/PUT/DELETE | `/api/admin/oidc-providers` | super | OIDC CRUD |
| GET/POST/PUT/DELETE | `/api/admin/users` | admin/super | Users CRUD |
| GET/POST/PUT/DELETE | `/api/admin/groups` | admin | Groups CRUD |

## Sync API (`backend-sync`)

| Method | Path | Auth | Описание |
|--------|------|:----:|----------|
| POST | `/api/sync/repositories/{id}` | Internal | Trigger sync |
| POST | `/api/sync/webhook/{id}` | Secret header | Webhook trigger |

## Health

| Method | Path | Описание |
|--------|------|----------|
| GET | `/liveness` | Process alive |
| GET | `/readiness` | Dependencies ready |
| GET | `/metrics` | Prometheus metrics |

## Примеры

### Поиск

```bash
curl "http://localhost:8082/api/search?q=kubernetes&limit=10" \
  -H "Authorization: Bearer $TOKEN"
```

### Trigger sync (admin)

```bash
curl -X POST "http://localhost:8082/api/admin/sync/{repoId}" \
  -H "Authorization: Bearer $TOKEN"
```

### Refresh token

```bash
curl -X POST "http://localhost:8081/api/auth/refresh" \
  -H "Content-Type: application/json" \
  -d '{"refresh_token": "..."}'
```

## Связанные разделы

- [RBAC](../admin/rbac.md)
- [Архитектура](architecture.md)
