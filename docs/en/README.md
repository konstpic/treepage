# TreePage — Documentation

**Language:** English · [Русский](../ru/README.md)

**TreePage** is a self-hosted documentation platform with Git synchronization, Markdown and Mermaid rendering, full-text search, RBAC, and OIDC authentication.

**Repository:** [github.com/konstpic/treepage](https://github.com/konstpic/treepage)

This folder contains the complete guide: from installation to day-to-day use by users and administrators.

---

## Where to start

| Role | Recommended path |
|------|-------------------|
| **Operator / DevOps** | [Installation](installation/README.md) → [Initial setup](getting-started/initial-setup.md) |
| **Platform administrator** | [Quick start](getting-started/README.md) → [Administrator guide](admin/README.md) |
| **User / editor** | [First login](getting-started/first-login.md) → [User guide](user/README.md) |

---

## Sections

### [Installation](installation/README.md)

Deploying TreePage: Docker Compose, local development, Kubernetes/Helm.

- [Requirements](installation/requirements.md)
- [Docker Compose](installation/docker-compose.md)
- [Local development](installation/local-development.md)
- [Kubernetes / Helm](installation/kubernetes.md)

### [Quick start](getting-started/README.md)

First launch and basic configuration after installation.

- [First login](getting-started/first-login.md)
- [Welcome space](getting-started/welcome-space.md)
- [Initial setup](getting-started/initial-setup.md)

### [User guide](user/README.md)

Working with documentation: spaces, reading, search, editing, books.

- [Interface navigation](user/navigation.md)
- [Spaces](user/spaces.md)
- [Reading documents](user/reading-docs.md)
- [Search](user/search.md)
- [Editing documents](user/editing-docs.md)
- [Books (AI compilations)](user/books.md)

### [Administrator guide](admin/README.md)

Platform management: spaces, repositories, users, OIDC, settings.

- [Roles and permissions (RBAC)](admin/rbac.md)
- [Spaces](admin/spaces.md)
- [Git repositories](admin/repositories.md)
- [Git synchronization](admin/git-sync.md)
- [Users](admin/users.md)
- [Groups](admin/groups.md)
- [System settings](admin/settings.md)
- [OIDC providers](admin/oidc.md)

### [Operations](operator/README.md)

Configuration, secrets, migrations, monitoring, troubleshooting.

- [Configuration](operator/configuration.md)
- [Secrets and environment variables](operator/secrets.md)
- [Database migrations](operator/migrations.md)
- [Monitoring and health checks](operator/monitoring.md)
- [Troubleshooting](operator/troubleshooting.md)

### [Reference](reference/README.md)

Technical documentation for developers and integrators.

- [Architecture](reference/architecture.md)
- [Helm deployment (details)](reference/helm-deployment.md)
- [REST API](reference/api.md)

---

## Architecture (overview)

```
frontend (React)  →  backend-auth (OIDC/JWT)
                  →  backend-server (Docs API, Search, Admin)
                  →  backend-sync (Git sync worker)
                  →  PostgreSQL
```

Details: [Architecture](reference/architecture.md).
