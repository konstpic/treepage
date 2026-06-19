# Installation with Docker Compose

The fastest way to run TreePage locally. Frontend uses Vite hot reload; backend runs a pre-built binary in the image (fast container startup).

## Step 1. Clone the repository

```bash
git clone https://github.com/konstpic/treepage.git
cd treepage
```

## Step 2. Start

```bash
docker compose up --build
```

Docker Compose starts:

| Service | URL | Description |
|---------|-----|-------------|
| frontend | http://localhost:5173 | Web interface |
| backend-auth | http://localhost:8081 | Authentication |
| backend-server | http://localhost:8082 | Documentation API |
| backend-sync | http://localhost:8083 | Git sync worker |
| postgres | localhost:5432 | Database |

Database migrations are applied automatically when **backend-server** starts (scans the `migrations/` folder).

After Go backend changes, rebuild the image:

```bash
docker compose up -d --build backend-server backend-auth backend-sync
```

For active backend development without Docker, see [Local development](local-development.md) (`go run` / Air).

## Step 3. Verify

1. Open http://localhost:5173/auth
2. Sign in as **`admin@local`** / **`admin`** (local super_admin)
3. Open welcome documentation: http://localhost:5173/spaces/welcome

Details: [First login](../getting-started/first-login.md), [Welcome space](../getting-started/welcome-space.md).

## Environment variables

Create a `.env` file in the project root to override secrets:

```bash
# Optional — for access to private Git repositories
GIT_ACCESS_TOKEN=ghp_xxxxxxxxxxxx
GIT_WEBHOOK_SECRET=my-webhook-secret

# Optional — OIDC in dev
OIDC_CLIENT_SECRET=your-oidc-secret
```

Docker Compose defaults:

| Variable | Default value |
|----------|---------------|
| `DB_PASSWORD` | `treepage` |
| `JWT_SECRET` | `dev-jwt-secret-change-in-production` |
| `DEV_MODE` | `true` (enables local login) |

> **Important:** default values are for development only. In production, use Kubernetes/Helm and strong secrets.

## Optional Redis

```bash
docker compose --profile cache up --build
```

## Stop and cleanup

```bash
# Stop
docker compose down

# Stop and remove volumes (DB data and Git clones)
docker compose down -v
```

## Volume structure

| Volume | Purpose |
|--------|---------|
| `postgres_data` | PostgreSQL data |
| `sync_repos` | Git repository clones (`/tmp/treepage-repos`) |

## Next steps

- [First login](../getting-started/first-login.md)
- [Initial setup](../getting-started/initial-setup.md)
