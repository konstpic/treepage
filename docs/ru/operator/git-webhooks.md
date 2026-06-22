# Git webhooks (GitHub / GitLab)

TreePage может синхронизировать документацию при push в Git-репозиторий. Webhook принимает **backend-sync** (порт 8083), не публичный UI.

---

## Endpoint

```
POST /api/sync/webhook/{repository_id}
Header: X-Hub-Signature-256: sha256=<hmac>   # GitHub
Header: X-Gitlab-Token: <secret>              # GitLab
```

`{repository_id}` — UUID из **Admin → Repositories** или `GET /api/admin/repositories`.

---

## Маршрутизация в production

`backend-sync` обычно **не** выставлен на публичный Ingress. Варианты:

| Подход | Когда |
|--------|--------|
| **Internal URL** | CI/runner в той же сети вызывает `http://backend-sync:8083/...` |
| **Ingress path** | Только `/api/sync/webhook/*` на sync + rate limit + IP allowlist |
| **Manual / cron** | Без webhook; `POST /api/sync/repositories/{id}` или расписание |

Пример Ingress (nginx):

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

## GitHub

1. Repository → **Settings → Webhooks → Add webhook**
2. **Payload URL:** `https://docs.example.com/api/sync/webhook/<repository-uuid>`
3. **Content type:** `application/json`
4. **Secret:** то же значение, что `GIT_WEBHOOK_SECRET` на backend-sync
5. События: **Push**

```bash
GIT_WEBHOOK_SECRET=your-long-random-secret
```

---

## GitLab

1. Project → **Settings → Webhooks**
2. **URL:** `https://docs.example.com/api/sync/webhook/<repository-uuid>`
3. **Secret token:** `GIT_WEBHOOK_SECRET`
4. Trigger: **Push events**

---

## Режимы sync

| `sync_mode` | Поведение |
|-------------|-----------|
| `scheduled` | Периодический pull |
| `manual` | Только ручной/API |
| `webhook` | Push → sync |

---

## Troubleshooting

- **401** — неверный `GIT_WEBHOOK_SECRET`
- **404** — неверный UUID в URL
- **Документы пропускаются** — локальные правки (`has_pending_changes`)

См. [Secrets](secrets.md), [Troubleshooting](troubleshooting.md).
