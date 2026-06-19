# Локальная разработка (без Docker)

Запуск сервисов напрямую на машине разработчика.

## Предварительные требования

- Go 1.22+
- Node.js 22+
- PostgreSQL 16+
- Git

## Шаг 1. База данных

```bash
# Создать пользователя и базу
createuser -P treepage   # пароль: treepage (или свой)
createdb -O treepage treepage

# Применить миграции (legacy — предпочтительно автомигратор при старте server)
for f in migrations/*.up.sql; do
  psql -U treepage -d treepage -f "$f"
done
```

> Порядок миграций `001`–`016` (см. [Дорожная карта](ru/reference/roadmap.md)). Предпочтительно запускать `backend-server` — он применит pending-миграции через `schema_migrations`.

## Шаг 2. Переменные окружения

```bash
export DB_PASSWORD=treepage
export JWT_SECRET=dev-secret
export DEV_MODE=true
export ENV=dev
```

## Шаг 3. Backend-сервисы

Запустите каждый сервис в отдельном терминале:

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

Порты по умолчанию: auth `8081`, server `8082`, sync `8083`.

## Шаг 4. Frontend

```bash
cd frontend
npm install
VITE_API_URL=http://localhost:8082 \
VITE_AUTH_URL=http://localhost:8081 \
npm run dev
```

Frontend доступен на http://localhost:5173.

## Конфигурация

Несекретные настройки — в YAML-файлах:

| Файл | Сервис |
|------|--------|
| `backend/auth/conf/config.yml` | OIDC, JWT, CORS, frontend URL |
| `backend/server/conf/config.yml` | Search, audit log, CORS |
| `backend/sync/conf/config.yml` | Git sync interval, work_dir |

Секреты передаются через переменные окружения (см. [Конфигурация](../operator/configuration.md)).

## LLM (опционально)

Для AI-книг и автоперевода задайте переменные для `backend-server`:

```bash
export LLM_ENABLED=true
export LLM_API_URL=https://api.openai.com/v1
export LLM_API_KEY=sk-...
export LLM_MODEL=gpt-4o-mini
```

## Следующие шаги

- [Первый вход](../getting-started/first-login.md)
- [Архитектура](../reference/architecture.md)
