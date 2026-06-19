# Reference

**Language:** English · [Русский](../../ru/reference/README.md)

Technical documentation for developers, integrators, and operators.

## Documents

| Document | Description |
|----------|-------------|
| [Architecture](architecture.md) | Microservices, auth flow, RBAC, search, RAG, security |
| [Roadmap (phases)](roadmap.md) | Phases 1–3, Search & RAG, platform |
| [Helm deployment (details)](helm-deployment.md) | Charts, install modes, migration from legacy |
| [REST API](api.md) | OpenAPI endpoints overview |
| [Frontend analysis](frontend-analysis.md) | Design system reference (for UI developers) |

## OpenAPI spec

Full specification: [`openapi/openapi.yaml`](../../openapi/openapi.yaml)

## Source code

| Component | Path |
|-----------|------|
| Frontend | `frontend/` |
| Auth service | `backend/auth/` |
| Server API | `backend/server/` |
| Sync worker | `backend/sync/` |
| Shared libs | `backend/pkg/` |
| Migrations | `migrations/` |
| Helm backend | `backend/` |
| Helm frontend | `.helm/frontend/` |
| Umbrella chart | `deploy/helm/treepage/` |

## Related sections

- [Operations](../operator/README.md)
- [Installation](../installation/README.md)
