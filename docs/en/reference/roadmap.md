# Roadmap and development phases

TreePage evolves in iterations — from a production-ready base to an enterprise knowledge base with RAG. This document describes **implemented phases** and their capabilities.

---

## Phase overview

| Phase | Migration | Status | Focus |
|-------|-----------|--------|-------|
| **1 — Production Hardening** | `012` | ✅ | Sync security, Git↔UI conflicts, audit |
| **2 — Team KB** | `013` | ✅ | Favorites, notifications, attachments, drafts |
| **3 — Enterprise KB** | `014` | ✅ | Page ACL, comments, workflow, analytics, RAG |
| **Search & RAG** | `015`, `016` | ✅ | Multilingual FTS, embeddings, citations, feedback |
| **Platform** | — | ✅ | Auto migrator, pre-built Docker images |

---

## Phase 1 — Production Hardening

**Goal:** production readiness without losing local edits during Git sync.

### Backend

| Capability | Description |
|------------|-------------|
| **INTERNAL_SERVICE_TOKEN** | Protects `backend-sync` API — only `backend-server` with shared token |
| **Pending changes** | Fields `has_pending_changes`, `synced_content_hash`, `last_synced_at` |
| **Conflict skip** | Sync **does not overwrite** documents with local edits; `conflicts_skipped` counter |
| **Audit indexes** | Indexes on `audit_logs` for faster queries |
| **OIDC + Redis** | OIDC state in Redis (not in-memory) — multiple `backend-auth` replicas |

### Where to read more

- [Troubleshooting — Sync API 401](../operator/troubleshooting.md)
- [Troubleshooting — Git sync skips documents](../operator/troubleshooting.md)
- [Editing — Git conflicts](../user/editing-docs.md)

### Environment variables

```bash
INTERNAL_SERVICE_TOKEN=your-long-random-secret   # server + sync
REDIS_ADDR=redis:6379                            # auth (multiple replicas)
```

---

## Phase 2 — Team KB

**Goal:** team collaboration — favorites, notifications, attachments, drafts.

### Backend

| Capability | API / tables |
|------------|--------------|
| **Favorites** | `/api/me/favorites`, `user_favorites` |
| **Recent views** | `/api/me/recent`, `user_recent_views` |
| **Notifications** | `/api/notifications`, `notifications` |
| **Attachments** | `/api/documents/:id/attachments`, `document_attachments` |
| **Drafts** | `is_published=false`, publish-local |
| **GitLab MR** | Provider `gitlab` in repository settings |
| **Orphan cleanup** | After sync, removes docs missing from Git (except pending changes) |
| **OpenSearch (opt-in)** | `SEARCH_BACKEND=opensearch` — foundation, currently delegates to PostgreSQL |

### Frontend

- `/me` — favorites and recent
- Notification bell in navigation
- Star, drafts, attachments on document page

### Where to read more

- [Editing documents](../user/editing-docs.md)

---

## Phase 3 — Enterprise KB

**Goal:** page-level permissions, collaboration, analytics, basic RAG.

### Backend

| Capability | Description |
|------------|-------------|
| **Page ACL** | Rules on path within a space (`page_acl_rules`), parent path inheritance |
| **Comments** | Document threads, `@mentions` |
| **Workflow** | States `draft` → `review` → `approved` → `published` |
| **Analytics** | Views, popular documents, search query log |
| **RAG chunks** | `document_chunks` table — indexed on sync and backfilled on startup |

### Frontend

- Comments and workflow on document page
- `/admin/analytics` — analytics dashboard
- `/search` — **Ask documentation** tab (RAG)

### Where to read more

- [Search — RAG](../user/search.md)
- [RBAC](../admin/rbac.md)

---

## Search & RAG (015, 016 and beyond)

**Goal:** smart search in Russian and English, hybrid retrieval, answers with citations.

### Multilingual FTS (`015`)

- PostgreSQL FTS: `english`, `russian`, `simple` configurations
- Standard search and RAG retrieval support Cyrillic

### RAG enhancements (`016`)

| Component | Description |
|-----------|-------------|
| **Embeddings** | `document_chunks.embedding` (JSONB), Ollama or OpenAI |
| **Hybrid search** | FTS + cosine similarity (~40% FTS + 60% vector) |
| **Multi-strategy retrieval** | phrase, websearch, keywords, ILIKE, LLM-expanded queries |
| **Citations** | Exact document quotes in the answer |
| **Confidence** | Confidence score; low confidence → follow-up questions |
| **Feedback 👍/👎** | `/api/rag/feedback`, synonym learning (`rag_learned_synonyms`) |

### RAG architecture

```
Git sync / startup backfill
    → document_chunks (+ embedding)
    → POST /api/rag/ask
        → retrieval (FTS + vector + learned synonyms)
        → LLM (OpenAI-compatible)
        → answer + sources + citations + confidence
```

RAG runs inside **backend-server** (`backend/server/internal/rag/`), not a separate container.

### LLM and embeddings setup

```yaml
# docker-compose / Helm extraEnv on backend-server
LLM_ENABLED: "true"
LLM_API_URL: http://host:11434/v1          # Ollama or OpenAI-compatible
LLM_MODEL: llama3.2:latest
LLM_API_KEY: ""                            # not required for local Ollama

EMBEDDING_ENABLED: "true"
EMBEDDING_MODEL: nomic-embed-text          # ollama pull nomic-embed-text
```

Details: [Configuration — LLM](../operator/configuration.md), [Secrets](../operator/secrets.md).

---

## Platform — infrastructure

### Auto migrator

- Scans `migrations/*.up.sql` on `backend-server` startup
- Versions tracked in `schema_migrations`; applied migrations are **skipped**
- New migration = add `017_*.up.sql` + restart server
- See [Migrations](../operator/migrations.md)

### Docker Compose (dev / demo)

| Before | After |
|--------|-------|
| Air hot-reload, `go build` on every start | Pre-built binary in alpine image |
| 14 migration volume mounts on postgres | Single `./migrations:/app/migrations` mount |

- **Frontend** — Vite hot reload (unchanged)
- **Backend** — after Go changes: `docker compose up -d --build backend-server …`
- Active backend dev without Docker: local `go run` / Air

See [Docker Compose](../installation/docker-compose.md).

---

## Migration matrix

| File | Phase |
|------|-------|
| `001`–`011` | Core platform |
| `012_production_hardening` | Phase 1 |
| `013_team_kb` | Phase 2 |
| `014_enterprise_kb` | Phase 3 |
| `015_multilingual_search` | Search & RAG |
| `016_rag_enhancements` | Search & RAG |
| `017_p1_production_features` | P1 — pgvector, sync diff, scale |

---

## P1 — Production scale (017)

| Component | Description |
|-----------|-------------|
| **RAG worker** | Background reindex + embeddings; `GET /api/admin/rag/status` |
| **pgvector** | `embedding_vector` column + HNSW index (Postgres image `pgvector/pgvector:pg16`) |
| **OpenSearch** | Real HTTP index/search when `SEARCH_BACKEND=opensearch` |
| **Attachments S3** | `ATTACHMENTS_STORAGE=s3` + `S3_*` env vars |
| **Notification webhook** | `NOTIFY_WEBHOOK_URL` on in-app events |
| **Git conflict diff** | `GET /api/documents/:id/sync-diff` (local vs Git snapshot) |
| **Audit** | Admin settings, OIDC, users, repositories, RAG feedback |

Env examples:

```bash
ATTACHMENTS_STORAGE=s3
S3_ENDPOINT=minio:9000
S3_BUCKET=treepage-attachments
S3_ACCESS_KEY=...
S3_SECRET_KEY=...

NOTIFY_WEBHOOK_URL=https://hooks.example.com/treepage
NOTIFY_WEBHOOK_SECRET=optional-shared-secret

SEARCH_BACKEND=opensearch
OPENSEARCH_URL=http://opensearch:9200
```

See [Git webhooks](../operator/git-webhooks.md).

---

## P0 — Production operations (018+)

| Component | Description |
|-----------|-------------|
| **No bootstrap SQL** | Schema only via `migrations/` (018 removes main.go duplicates) |
| **Static frontend** | nginx + `deploy/docker/Dockerfile.frontend` in dev and prod compose |
| **docker-compose.prod.yml** | `DEV_MODE=false`, secrets via `.env.prod` |
| **Redis rate limit** | `REDIS_ADDR` → distributed limit across replicas |
| **App metrics** | `treepage_*` Prometheus counters/histograms on `/metrics` |
| **Helm ServiceMonitor** | `monitoring.serviceMonitor.enabled` in backend chart |
| **Helm backup CronJob** | `backup.enabled` — pg_dump schedule |
| **Sync → OpenSearch** | sync calls `POST /api/internal/documents/:id/reindex` after Git pull |

Deploy:

```bash
cp .env.prod.example .env.prod   # fill secrets
./scripts/deploy-prod.sh
cp docker-compose.dev.yml.example .env   # dev
./scripts/deploy-dev.sh
```

---

## Related sections

- [Architecture](architecture.md)
- [Migrations](../operator/migrations.md)
- [Search (user)](../user/search.md)
- [Troubleshooting — LLM/RAG](../operator/troubleshooting.md)
