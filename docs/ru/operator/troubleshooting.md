# Устранение неполадок

## Аутентификация

### «OIDC недоступен»

**Причины:**
- OIDC disabled в config
- Issuer URL недоступен из pod auth
- Неверный client_id

**Диагностика:**
```bash
kubectl logs -l app.kubernetes.io/component=auth --tail=50
curl -s https://keycloak.example.com/realms/treepage/.well-known/openid-configuration
```

### Redirect loop после OIDC login

**Причины:**
- `frontendUrl` не совпадает с реальным URL
- Redirect URI в OIDC client не совпадает

**Решение:**
- Helm: `global.frontendUrl: https://docs.example.com`
- OIDC client redirect: `https://docs.example.com/api/auth/callback`

### 401 на все API-запросы

**Причины:**
- `JWT_SECRET` различается между auth и server
- Expired token без refresh

**Решение:**
```bash
# Проверить одинаковость secret
kubectl get secret treepage-backend -o jsonpath='{.data.jwt-secret}' | base64 -d
```

### Dev login не работает

- Проверьте `DEV_MODE=true` на backend-auth
- Проверьте `ENV != prod`
- Учётные данные: `admin@local` / `admin`

---

## Git Sync

### Sync failed: authentication failed

```bash
# Проверить token
kubectl get secret treepage-backend -o jsonpath='{.data.git-access-token}' | base64 -d

# Проверить из sync pod
kubectl exec -it deploy/treepage-backend-sync -- sh
git ls-remote https://token@github.com/org/repo.git
```

### Sync failed: branch not found

- Проверьте имя ветки в настройках репозитория
- Default: `main` (не `master`)

### Документы не появляются после sync

- Проверьте `docs_path` (default: `docs`)
- Убедитесь, что файлы имеют расширение `.md`
- Проверьте `last_sync_status` в admin UI

### Webhook не срабатывает

- Sync service не exposed через Ingress — webhook URL должен быть internal
- Проверьте header `X-Webhook-Secret`
- Альтернатива: используйте `scheduled` sync mode

---

## Frontend

### Blank page / 502

```bash
kubectl get pods -l app.kubernetes.io/part-of=treepage
kubectl logs -l app.kubernetes.io/component=frontend
```

- Проверьте Ingress routes
- Проверьте `/config.json` доступен

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

- Проверьте `postgresql.host` в Helm values
- Проверьте `DB_PASSWORD`

### Migration errors

- Примените миграции в правильном порядке
- См. [Миграции](migrations.md)

---

## LLM / Books

### «LLM не настроен»

```bash
kubectl set env deploy/treepage-backend-server \
  LLM_ENABLED=true \
  LLM_API_URL=https://api.openai.com/v1 \
  LLM_MODEL=gpt-4o-mini
# + secret для LLM_API_KEY
```

### Book generation failed

- Проверьте доступность LLM API из server pod
- Проверьте лимиты API
- Проверьте наличие документов в пространстве

---

## Полезные команды

```bash
# Статус всех pods
kubectl get pods -l app.kubernetes.io/part-of=treepage

# Логи auth
kubectl logs -f deploy/treepage-backend-auth

# Логи sync
kubectl logs -f deploy/treepage-backend-sync

# Helm values
helm get values treepage-backend

# Lint charts
helm lint backend --strict
```

## Docker Compose

```bash
# Логи всех сервисов
docker compose logs -f

# Перезапуск одного сервиса
docker compose restart backend-sync

# Проверка postgres
docker compose exec postgres psql -U treepage -d treepage -c '\dt'
```

## Связанные разделы

- [Мониторинг](monitoring.md)
- [Секреты](secrets.md)
- [Git Sync](../admin/git-sync.md)
