# System requirements

## Minimum resources (production)

| Component | CPU | RAM | Disk |
|-----------|-----|-----|------|
| frontend | 50m | 64 Mi | — |
| backend-auth | 50m | 64 Mi | — |
| backend-server | 100m | 128 Mi | — |
| backend-sync | 100m | 128 Mi | 2 Gi (repository clones) |
| PostgreSQL | 500m | 512 Mi | 10 Gi+ |

It is recommended to run backend services with at least 2 replicas (except sync).

## Software

### Docker Compose

- Docker 24+
- Docker Compose v2

### Local development

- Go 1.22+
- Node.js 22+
- PostgreSQL 16+
- Git 2.x

### Kubernetes / Helm

- Kubernetes 1.24+
- Helm 3.10+
- Ingress controller (nginx recommended)
- PostgreSQL 16+ (external or managed)

## Network

| Port (dev) | Service |
|------------|---------|
| 5173 | frontend (Vite dev) |
| 8081 | backend-auth |
| 8082 | backend-server |
| 8083 | backend-sync |
| 5432 | PostgreSQL |

In production, all services are exposed through Ingress on ports 80/443. The sync service is not published externally — cluster-internal only.

## External dependencies

| Service | Required | Purpose |
|---------|----------|---------|
| PostgreSQL | Required | Primary storage |
| OIDC provider | Production | SSO (Keycloak, Okta, Azure AD, etc.) |
| Git repository | For Git Sync | Documentation source |
| LLM API | Optional | AI books, document auto-translation |
| Redis | Optional | Cache (Docker Compose `cache` profile) |

## Browsers

Modern browsers with ES2020+ support: Chrome, Firefox, Safari, Edge (last 2 versions).
