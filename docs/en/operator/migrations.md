# Database migrations

SQL migrations are in `migrations/`.

## Migration list

| File | Description |
|------|-------------|
| `001_initial_schema.up.sql` | Core: users, roles, spaces, repos, documents, sync_jobs, audit |
| `002_local_auth.up.sql` | password_hash for local login |
| `003_system_settings.up.sql` | system_settings table |
| `005_ui_theme.up.sql` | UI theme setting |
| `006_books.up.sql` | AI books |
| `007_ui_language.up.sql` | UI language setting |
| `008_document_translations.up.sql` | Document translations |
| `009_book_translations.up.sql` | Book translations |
| `010_space_groups.up.sql` | Groups in spaces |
| `011_oidc_sync_groups.up.sql` | sync_groups flag for OIDC |
| `012_production_hardening.up.sql` | Pending changes, audit, OIDC Redis |
| `013_team_kb.up.sql` | Favorites, recent views, notifications, attachments |

> Migration `004` is missing (skipped in numbering).

## Docker Compose

Migrations are applied automatically on first PostgreSQL startup via `docker-entrypoint-initdb.d/`.

## Manual application

```bash
export PGPASSWORD=<password>

for f in migrations/00*_up.sql migrations/01*_up.sql; do
  echo "Applying $f..."
  psql -h <host> -U treepage -d treepage -f "$f"
done
```

## Kubernetes

The chart does not include an init job for migrations. Apply manually before first deploy:

```bash
# Port-forward or direct connection
kubectl run psql-client --rm -it --image=postgres:16-alpine -- \
  psql -h postgres.default.svc -U treepage -d treepage

# Or via Job (example)
kubectl create configmap migrations --from-file=migrations/
```

## Rollback

Each migration has a `.down.sql`:

```bash
psql -U treepage -d treepage -f migrations/011_oidc_sync_groups.down.sql
```

> Roll back in reverse order. In production, rollback is a last resort.

## Verification

```sql
-- Check tables
\dt

-- Users
SELECT email, display_name FROM users;

-- System settings
SELECT * FROM system_settings;
```

## Backup before migration

```bash
pg_dump -h <host> -U treepage -d treepage -F c -f treepage_backup_$(date +%Y%m%d).dump
```

## Related sections

- [Installation](../installation/README.md)
- [Troubleshooting](troubleshooting.md)
