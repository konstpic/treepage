# Backup and restore

TreePage stores all content in **PostgreSQL**. Git repositories remain the upstream source for synced spaces; the database holds the working copy, versions, users, and settings.

## What to back up

| Data | Location |
|------|----------|
| Documents, versions, users, RBAC | PostgreSQL `treepage` database |
| Git clone cache (optional) | Sync worker volume (`/tmp/treepage-repos` or Helm PVC) |
| Secrets | Kubernetes Secrets / `.env` — store separately in a vault |

## PostgreSQL backup

### Manual dump (recommended baseline)

```bash
# Docker Compose
docker compose exec postgres pg_dump -U treepage -Fc treepage > treepage-$(date +%Y%m%d).dump

# Plain SQL (portable)
docker compose exec postgres pg_dump -U treepage treepage > treepage-$(date +%Y%m%d).sql
```

### Restore from custom format

```bash
# Stop writers (frontend + backends) or put platform in maintenance mode
docker compose stop backend-auth backend-server backend-sync frontend

docker compose exec -T postgres pg_restore -U treepage -d treepage --clean --if-exists < treepage-20260101.dump

docker compose start backend-auth backend-server backend-sync frontend
```

### Restore from SQL

```bash
docker compose exec -T postgres psql -U treepage -d treepage < treepage-20260101.sql
```

## Verify restore

1. Open http://localhost:5173/spaces — spaces and documents load.
2. Sign in as admin; check **Admin → Audit log** for recent entries.
3. Trigger **Sync** on a Git-backed space; confirm `conflicts_skipped` is reported when local edits exist.
4. Open a document **Version history** and test **Restore**.

## Production checklist

- [ ] Daily automated `pg_dump` to object storage (S3, GCS, MinIO)
- [ ] Retention policy (e.g. 30 daily, 12 monthly)
- [ ] Quarterly restore drill on a staging cluster
- [ ] `INTERNAL_SERVICE_TOKEN`, `JWT_SECRET`, `DB_PASSWORD` backed up in a secrets manager (not in the dump)
- [ ] Redis (OIDC state) is ephemeral — no backup required; users re-login after Redis loss

## Kubernetes / Helm

Use your platform's PostgreSQL operator backup (CloudNativePG, Zalando, RDS snapshots) or a CronJob:

```yaml
# Example CronJob fragment — adjust for your cluster
command: ["pg_dump", "-Fc", "-h", "postgres", "-U", "treepage", "treepage"]
```

Point the job output to durable storage and test restore on a non-production namespace before relying on it.

## Related

- [Migrations](migrations.md)
- [Troubleshooting](troubleshooting.md)
- [Configuration](configuration.md)
