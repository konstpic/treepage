# Git webhooks (GitHub / GitLab)

TreePage can sync documentation when a Git repository receives a push. The webhook endpoint lives on **backend-sync** (port 8083), not on the public UI.

---

## Endpoint

```
POST /api/sync/webhook/{repository_id}
Header: X-Hub-Signature-256: sha256=<hmac>   # GitHub
Header: X-Gitlab-Token: <secret>              # GitLab
```

The `{repository_id}` is the UUID from **Admin → Repositories** or `GET /api/admin/repositories`.

---

## Production routing

`backend-sync` is typically **not** exposed on the public Ingress. Choose one:

| Approach | When to use |
|----------|-------------|
| **Internal webhook URL** | GitLab/GitHub runner or CI in the same VPC calls `http://backend-sync:8083/...` |
| **Ingress path** | Expose only `/api/sync/webhook/*` to sync service with rate limit + IP allowlist |
| **Manual / scheduled sync** | No webhook; use cron or `POST /api/sync/repositories/{id}` from backend-server |

Example Ingress snippet (nginx):

```yaml
- path: /api/sync/webhook
  pathType: Prefix
  backend:
    service:
      name: backend-sync
      port:
        number: 8083
```

---

## GitHub setup

1. Repository → **Settings → Webhooks → Add webhook**
2. **Payload URL:** `https://docs.example.com/api/sync/webhook/<repository-uuid>`
3. **Content type:** `application/json`
4. **Secret:** same value as `GIT_WEBHOOK_SECRET` on backend-sync
5. Events: **Just the push event**

Set in backend-sync environment:

```bash
GIT_WEBHOOK_SECRET=your-long-random-secret
```

---

## GitLab setup

1. Project → **Settings → Webhooks**
2. **URL:** `https://docs.example.com/api/sync/webhook/<repository-uuid>`
3. **Secret token:** same as `GIT_WEBHOOK_SECRET`
4. Trigger: **Push events**

---

## Sync modes

In repository settings (`sync_mode`):

| Mode | Behavior |
|------|----------|
| `scheduled` | Periodic pull (default) |
| `manual` | Only admin trigger or API |
| `webhook` | Push triggers sync (use with webhook URL above) |

---

## Troubleshooting

- **401 / invalid signature** — `GIT_WEBHOOK_SECRET` mismatch between Git provider and sync service
- **404 repository** — wrong UUID in webhook URL
- **Sync skips documents** — local edits with `has_pending_changes`; see [Editing docs — Git conflicts](../user/editing-docs.md)

See also [Secrets](secrets.md) and [Troubleshooting](troubleshooting.md).
