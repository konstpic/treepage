# Troubleshooting

## Authentication

### "OIDC unavailable"

**Causes:**
- OIDC disabled in config
- Issuer URL unreachable from auth pod
- Invalid client_id

**Diagnosis:**
```bash
kubectl logs -l app.kubernetes.io/component=auth --tail=50
curl -s https://keycloak.example.com/realms/treepage/.well-known/openid-configuration
```

### Redirect loop after OIDC login

**Causes:**
- `frontendUrl` does not match actual URL
- Redirect URI in OIDC client does not match

**Solution:**
- Helm: `global.frontendUrl: https://docs.example.com`
- OIDC client redirect: `https://docs.example.com/api/auth/callback`

### 401 on all API requests

**Causes:**
- `JWT_SECRET` differs between auth and server
- Expired token without refresh

**Solution:**
```bash
# Verify secret matches
kubectl get secret treepage-backend -o jsonpath='{.data.jwt-secret}' | base64 -d
```

### Dev login not working

- Check `DEV_MODE=true` on backend-auth
- Check `ENV != prod`
- Credentials: `admin@local` / `admin`

---

## Git Sync

### Sync failed: authentication failed

```bash
# Check token
kubectl get secret treepage-backend -o jsonpath='{.data.git-access-token}' | base64 -d

# Test from sync pod
kubectl exec -it deploy/treepage-backend-sync -- sh
git ls-remote https://token@github.com/org/repo.git
```

### Sync failed: branch not found

- Check branch name in repository settings
- Default: `main` (not `master`)

### Documents do not appear after sync

- Check `docs_path` (default: `docs`)
- Ensure files have `.md` extension
- Check `last_sync_status` in admin UI

### Webhook not firing

- Sync service is not exposed via Ingress — webhook URL must be internal
- Check `X-Webhook-Secret` header
- Alternative: use `scheduled` sync mode

---

## Frontend

### Blank page / 502

```bash
kubectl get pods -l app.kubernetes.io/part-of=treepage
kubectl logs -l app.kubernetes.io/component=frontend
```

- Check Ingress routes
- Check `/config.json` is accessible

### API proxy errors (dev)

```bash
# Vite proxy config
VITE_USE_PROXY=true
VITE_PROXY_AUTH=http://backend-auth:8081
VITE_PROXY_API=http://backend-server:8082
```

---

## Database

### Connection refused

```bash
kubectl exec -it deploy/treepage-backend-auth -- sh
nc -zv postgres 5432
```

- Check `postgresql.host` in Helm values
- Check `DB_PASSWORD`

### Migration errors

- Apply migrations in correct order
- See [Migrations](migrations.md)

---

## LLM / Books

### "LLM not configured"

```bash
kubectl set env deploy/treepage-backend-server \
  LLM_ENABLED=true \
  LLM_API_URL=https://api.openai.com/v1 \
  LLM_MODEL=gpt-4o-mini
# + secret for LLM_API_KEY
```

### Book generation failed

- Check LLM API availability from server pod
- Check API limits
- Check documents exist in the space

---

## Useful commands

```bash
# Status of all pods
kubectl get pods -l app.kubernetes.io/part-of=treepage

# Auth logs
kubectl logs -f deploy/treepage-backend-auth

# Sync logs
kubectl logs -f deploy/treepage-backend-sync

# Helm values
helm get values treepage-backend

# Lint charts
helm lint backend --strict
```

## Docker Compose

### `listing workers for Build: EOF`

BuildKit/Bake error in Docker Compose v5 on Docker Desktop — not a TreePage bug.

**Fix:**

1. Restart Docker Desktop
2. Create `.env` in the project root (see `.env.example`):

```bash
echo 'COMPOSE_BAKE=false' > .env
```

3. Run:

```bash
docker compose up -d --build
```

```bash
# Logs for all services
docker compose logs -f

# Restart one service
docker compose restart backend-sync

# Check postgres
docker compose exec postgres psql -U treepage -d treepage -c '\dt'
```

## Related sections

- [Monitoring](monitoring.md)
- [Secrets](secrets.md)
- [Git Sync](../admin/git-sync.md)
