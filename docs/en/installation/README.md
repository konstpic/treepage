# Installing TreePage

TreePage can be deployed in three ways — choose the one that fits your environment.

| Method | When to use |
|--------|-------------|
| [Docker Compose](docker-compose.md) | Quick start, local development, demo |
| [Local development](local-development.md) | Development without Docker |
| [Kubernetes / Helm](kubernetes.md) | Production, GitOps, scaling |

## Common requirements

Before installation, ensure [system requirements](requirements.md) are met.

## After installation

1. Open the web interface (default `http://localhost:5173` in dev).
2. Complete [first login](../getting-started/first-login.md).
3. Follow [initial setup](../getting-started/initial-setup.md).

## Secrets

Regardless of installation method, set strong values for:

| Variable | Purpose |
|----------|---------|
| `DB_PASSWORD` | PostgreSQL password |
| `JWT_SECRET` | JWT token signing |
| `OIDC_CLIENT_SECRET` | OIDC client secret (production) |
| `GIT_ACCESS_TOKEN` | Git repository access token |
| `GIT_WEBHOOK_SECRET` | Webhook validation secret |

Details: [Secrets and environment variables](../operator/secrets.md).
