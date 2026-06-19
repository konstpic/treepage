# Local development (without Docker)

Run services directly on the developer machine.

## Prerequisites

- Go 1.22+
- Node.js 22+
- PostgreSQL 16+
- Git

## Step 1. Database

```bash
# Create user and database
createuser -P treepage   # password: treepage (or your own)
createdb -O treepage treepage

# Apply migrations
export PGPASSWORD=treepage
for f in migrations/*_up.sql; do
  psql -U treepage -d treepage -f "$f"
done
```

> Migration order: `001`, `002`, `003`, `005`, `006`, `007`, `008`, `009`, `010`, `011`.

## Step 2. Environment variables

```bash
export DB_PASSWORD=treepage
export JWT_SECRET=dev-secret
export DEV_MODE=true
export ENV=dev
```

## Step 3. Backend services

Start each service in a separate terminal:

```bash
# Terminal 1 — Auth
cd backend/auth
CONFIG_PATH=conf/config.yml go run ./cmd

# Terminal 2 — Server
cd backend/server
CONFIG_PATH=conf/config.yml go run ./cmd

# Terminal 3 — Sync
cd backend/sync
CONFIG_PATH=conf/config.yml go run ./cmd
```

Default ports: auth `8081`, server `8082`, sync `8083`.

## Step 4. Frontend

```bash
cd frontend
npm install
VITE_API_URL=http://localhost:8082 \
VITE_AUTH_URL=http://localhost:8081 \
npm run dev
```

Frontend is available at http://localhost:5173.

## Configuration

Non-secret settings are in YAML files:

| File | Service |
|------|---------|
| `backend/auth/conf/config.yml` | OIDC, JWT, CORS, frontend URL |
| `backend/server/conf/config.yml` | Search, audit log, CORS |
| `backend/sync/conf/config.yml` | Git sync interval, work_dir |

Secrets are passed via environment variables (see [Configuration](../operator/configuration.md)).

## LLM (optional)

For AI books and auto-translation, set variables for `backend-server`:

```bash
export LLM_ENABLED=true
export LLM_API_URL=https://api.openai.com/v1
export LLM_API_KEY=sk-...
export LLM_MODEL=gpt-4o-mini
```

## Next steps

- [First login](../getting-started/first-login.md)
- [Architecture](../reference/architecture.md)
