# Git synchronization

The sync worker (`backend-sync`) clones Git repositories and imports Markdown into PostgreSQL.

## Sync modes

| Mode | Description | When to use |
|------|-------------|-------------|
| `manual` | Button or API only | Rarely updated documentation |
| `scheduled` | Automatically on interval | Primary mode (default: 300 sec) |
| `webhook` | HTTP request from Git | Instant update on push |

## Synchronization process

```
1. sync_jobs record created (status: running)
2. git clone --depth 1 --branch <branch> <url>
3. Walk {clone}/{docs_path}/**/*.md
4. For each file: title, slug, tags, content
5. Upsert into documents by (space_id, slug)
6. Update last_sync_at, last_sync_status
```

## Triggers

### 1. Scheduled (background)

Sync worker runs a ticker with interval from configuration:

- Global default: `sync.git.syncInterval` (300s in Helm)
- Per-repo override: `sync_interval_seconds` in repository settings

Only repositories with `enabled: true` are synchronized.

### 2. Manual

**From UI:**
- `/admin/repositories` → sync button
- `/spaces/{slug}` → **Synchronize** in sidebar

**API:**
```
POST /api/admin/sync/{repoId}        # admin
POST /api/spaces/{slug}/repositories/{repoId}/sync  # editor+
```

### 3. Webhook

**Endpoint:**
```
POST /api/sync/webhook/{repositoryId}
Header: X-Webhook-Secret: <secret>
```

Sync worker listens on port 8083 (inside cluster, not via Ingress).

#### GitHub webhook setup

1. Repository → Settings → Webhooks → Add webhook
2. Payload URL: `http://backend-sync:8083/api/sync/webhook/{repo-id}` (internal)
   - For external: configure reverse proxy or use scheduled sync
3. Content type: `application/json`
4. Secret: value of `GIT_WEBHOOK_SECRET`
5. Events: Push events

#### GitLab webhook setup

1. Project → Settings → Webhooks
2. URL: `http://backend-sync:8083/api/sync/webhook/{repo-id}`
3. Secret token: value of `GIT_WEBHOOK_SECRET`
4. Trigger: Push events

> In production, the webhook URL must be reachable from the Git platform. Sync is usually not exposed via Ingress — use internal network or scheduled sync.

## Clone storage

| Environment | Path |
|-------------|------|
| Docker Compose | `/tmp/treepage-repos` (volume `sync_repos`) |
| Kubernetes | `/data/repos` (PVC, default 2Gi) |

Clones are shallow (`--depth 1`) — recreated on each sync.

## Sync monitoring

| Field | Where to look |
|-------|---------------|
| `last_sync_at` | Repository list in admin |
| `last_sync_status` | success / failed / completed |
| `last_sync_error` | Error text when failed |
| `sync_jobs` | Table in PostgreSQL |

## Common errors

| Error | Cause | Solution |
|-------|-------|----------|
| Authentication failed | Invalid or missing token | Check `GIT_ACCESS_TOKEN` |
| Branch not found | Wrong branch name | Set correct branch |
| Path not found | Wrong `docs_path` | Check repository structure |
| Connection timeout | Git server unreachable | Check network from sync pod |

Details: [Troubleshooting](../operator/troubleshooting.md)

## Related sections

- [Repositories](repositories.md)
- [Configuration](../operator/configuration.md)
