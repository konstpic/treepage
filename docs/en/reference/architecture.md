# TreePage architecture

## System overview

```mermaid
flowchart TB
    subgraph Client
        FE[frontend<br/>React + Vite]
    end

    subgraph Auth
        AUTH[backend-auth<br/>OIDC + JWT]
    end

    subgraph Core
        SRV[backend-server<br/>Docs API + Search]
        SYNC[backend-sync<br/>Git Sync Worker]
    end

    subgraph Data
        PG[(PostgreSQL)]
        RD[(Redis<br/>optional cache)]
    end

    subgraph External
        OIDC[OIDC Provider<br/>Keycloak / Okta / Azure AD]
        GIT[Git Repositories<br/>GitHub / GitLab / Gitea]
    end

    FE -->|Bearer JWT| AUTH
    FE -->|Bearer JWT| SRV
    AUTH --> OIDC
    AUTH --> PG
    SRV --> PG
    SRV --> RD
    SYNC --> PG
    SYNC --> GIT
    GIT -->|webhook| SYNC
```

## Microservices

| Service | Port (dev) | Responsibility |
|---------|------------|----------------|
| frontend | 5173 | UI, markdown/mermaid, admin panel |
| backend-auth | 8081 | OIDC, JWT issue/refresh, user sync |
| backend-server | 8082 | Spaces, documents, search, RBAC, admin API |
| backend-sync | 8083 | Git clone, parse, index, webhooks |
| postgres | 5432 | Primary datastore |
| redis | 6379 | Cache (optional) |

## Authentication Flow

```mermaid
sequenceDiagram
    participant U as User
    participant FE as Frontend
    participant AUTH as backend-auth
    participant OIDC as OIDC Provider
    participant API as backend-server

    U->>FE: Open app
    FE->>AUTH: GET /auth/login
    AUTH->>OIDC: Redirect authorization
    OIDC->>U: Login
    OIDC->>AUTH: Authorization code
    AUTH->>OIDC: Exchange tokens
    AUTH->>AUTH: Map roles/groups, upsert user
    AUTH->>FE: Redirect with tokens
    FE->>API: API calls with Bearer JWT
    API->>API: Validate JWT, check RBAC
```

## RBAC Model

```mermaid
erDiagram
    users ||--o{ user_roles : has
    roles ||--o{ user_roles : assigned
    roles ||--o{ role_permissions : grants
    permissions ||--o{ role_permissions : includes
    users ||--o{ group_members : belongs
    groups ||--o{ group_members : contains
    spaces ||--o{ space_members : has
    users ||--o{ space_members : member
    spaces ||--o{ space_groups : has
    groups ||--o{ space_groups : assigned
    spaces ||--o{ documents : contains
    repositories ||--o{ spaces : feeds
    documents ||--o{ document_versions : versioned
    spaces ||--o{ books : contains
```

### Roles

| Role | Scope | Capabilities |
|------|-------|--------------|
| super_admin | System | All settings, OIDC, users, repos |
| admin | System/Space | Manage spaces, repos, members |
| editor | Space | Create/edit docs, trigger sync |
| viewer | Space | Read docs |

Details: [RBAC](../admin/rbac.md)

## Search Architecture

- **Phase 1:** PostgreSQL `tsvector` full-text search on title, content, tags.
- **Phase 2:** OpenSearch adapter (interface in `backend/server/internal/search`).

Search fields: title, content, tags, repository, author.

## Git Sync Architecture

```
Git Repo → backend-sync (clone + parse) → PostgreSQL (documents)
                ↑
    scheduled / manual / webhook triggers
```

Server proxies sync requests to sync worker via `SYNC_SERVICE_URL`.

## Configuration

```
/opt/app/conf/config.yml   ← non-secret defaults
Environment variables      ← secrets + overrides
```

Load order: YAML → ENV override → validation → fail fast.

## Kubernetes Probes

All Go services expose:

| Endpoint | Purpose |
|----------|---------|
| `/liveness` | Process alive |
| `/readiness` | DB connected |
| `/metrics` | Prometheus metrics |

## Security

- JWT validation on all protected routes
- CSRF token for OIDC state
- Rate limiting (in-memory or Redis)
- Audit log for admin actions
- Secure headers (HSTS, X-Frame-Options, CSP)
- Secrets only via ENV

## Deployment

| Environment | Tool |
|-------------|------|
| Local dev | Docker Compose + hot reload |
| Production | Open-source Helm charts (`backend/`, `.helm/frontend/`) |

See [Kubernetes / Helm](../installation/kubernetes.md), [Helm deployment](helm-deployment.md).

## LLM Integration

Optional OpenAI-compatible API for:

- AI book generation (structure + introduction)
- Document auto-translation

Configured via `LLM_*` env vars on backend-server.
