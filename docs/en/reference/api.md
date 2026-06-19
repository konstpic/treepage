# REST API

TreePage provides a REST API through two backend services. Full specification: [`openapi/openapi.yaml`](../../openapi/openapi.yaml).

## Base URLs

| Service | Dev | Production (via Ingress) |
|---------|-----|--------------------------|
| Auth | `http://localhost:8081` | `https://docs.example.com/api/auth` |
| Server | `http://localhost:8082` | `https://docs.example.com/api` |
| Sync | `http://localhost:8083` | Internal only |

## Authentication

Most endpoints require a JWT Bearer token:

```
Authorization: Bearer <access_token>
```

Obtain token: OIDC flow or `POST /api/auth/login` (dev only).

## Auth API (`backend-auth`)

| Method | Path | Auth | Description |
|--------|------|:----:|-------------|
| GET | `/api/auth/login` | ❌ | Start OIDC flow |
| POST | `/api/auth/login` | ❌ | Local login (dev) |
| GET | `/api/auth/callback` | ❌ | OIDC callback |
| POST | `/api/auth/refresh` | ❌ | Refresh access token |
| POST | `/api/auth/logout` | ✅ | Logout |
| GET | `/api/auth/me` | ✅ | Current user profile |

## Public API (`backend-server`)

| Method | Path | Auth | Description |
|--------|------|:----:|-------------|
| GET | `/api/public/branding` | ❌ | Platform branding |
| GET | `/api/public/ui-theme` | ❌ | Current UI theme |
| GET | `/api/public/spaces` | ❌ | Public spaces |

## Spaces & Documents

| Method | Path | Auth | Description |
|--------|------|:----:|-------------|
| GET | `/api/spaces` | ✅ | List spaces |
| GET | `/api/spaces/{slug}` | ✅/❌ | Space metadata |
| GET | `/api/spaces/{slug}/documents` | ✅/❌ | Document tree |
| GET | `/api/spaces/{slug}/documents/{docSlug}` | ✅/❌ | Document content |
| PUT | `/api/spaces/{slug}/documents/{docSlug}` | ✅ | Update document |
| POST | `/api/spaces/{slug}/documents` | ✅ | Create document |

## Search

| Method | Path | Auth | Description |
|--------|------|:----:|-------------|
| GET | `/api/search` | ✅/❌ | Full-text search |

Query params: `q`, `space`, `author`, `tags`, `limit`, `offset`

## Admin API (`backend-server`)

> Requires `admin` or `super_admin` role.

| Method | Path | Min role | Description |
|--------|------|----------|-------------|
| GET/PUT | `/api/admin/system-settings` | admin/super | System settings |
| PUT | `/api/admin/system-settings/ui-theme` | super | UI theme |
| PUT | `/api/admin/system-settings/ui-language` | super | UI language |
| GET/POST | `/api/admin/repositories` | admin | Repository CRUD |
| GET/PUT/DELETE | `/api/admin/repositories/{id}` | admin | Repository management |
| GET/POST | `/api/admin/spaces` | admin | Space CRUD |
| PATCH | `/api/admin/spaces/{id}` | admin | Update space |
| POST | `/api/admin/spaces/{id}/bind-repo` | admin | Bind repository |
| POST | `/api/admin/sync/{repoId}` | admin | Trigger sync |
| GET/POST/PUT/DELETE | `/api/admin/oidc-providers` | super | OIDC CRUD |
| GET/POST/PUT/DELETE | `/api/admin/users` | admin/super | Users CRUD |
| GET/POST/PUT/DELETE | `/api/admin/groups` | admin | Groups CRUD |

## Sync API (`backend-sync`)

| Method | Path | Auth | Description |
|--------|------|:----:|-------------|
| POST | `/api/sync/repositories/{id}` | Internal | Trigger sync |
| POST | `/api/sync/webhook/{id}` | Secret header | Webhook trigger |

## Health

| Method | Path | Description |
|--------|------|-------------|
| GET | `/liveness` | Process alive |
| GET | `/readiness` | Dependencies ready |
| GET | `/metrics` | Prometheus metrics |

## Examples

### Search

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

## Related sections

- [RBAC](../admin/rbac.md)
- [Architecture](architecture.md)
