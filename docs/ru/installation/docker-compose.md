# Установка через Docker Compose

Самый быстрый способ запустить TreePage локально. Frontend — hot reload через Vite; backend — готовый бинарник в образе (быстрый старт контейнеров).

## Шаг 1. Клонирование репозитория

```bash
git clone https://github.com/konstpic/treepage.git
cd treepage
```

## Шаг 2. Запуск

```bash
docker compose up --build
```

Docker Compose поднимает:

| Сервис | URL | Описание |
|--------|-----|----------|
| frontend | http://localhost:5173 | Веб-интерфейс |
| backend-auth | http://localhost:8081 | Аутентификация |
| backend-server | http://localhost:8082 | API документации |
| backend-sync | http://localhost:8083 | Git sync worker |
| postgres | localhost:5432 | База данных |

Миграции БД применяются автоматически при старте **backend-server** (сканирует папку `migrations/`).

После изменений в Go-коде backend пересоберите образ:

```bash
docker compose up -d --build backend-server backend-auth backend-sync
```

Для активной разработки backend без Docker см. [Локальная разработка](local-development.md) (`go run` / Air).

## Шаг 3. Проверка

1. Откройте http://localhost:5173/auth
2. Войдите как **`admin@local`** / **`admin`** (локальный super_admin)
3. Откройте welcome-документацию: http://localhost:5173/spaces/welcome

Подробнее: [Первый вход](../getting-started/first-login.md), [Welcome space](../getting-started/welcome-space.md).

## Переменные окружения

Создайте файл `.env` в корне проекта для переопределения секретов:

```bash
# Опционально — для доступа к приватным Git-репозиториям
GIT_ACCESS_TOKEN=ghp_xxxxxxxxxxxx
GIT_WEBHOOK_SECRET=my-webhook-secret

# Опционально — OIDC в dev
OIDC_CLIENT_SECRET=your-oidc-secret
```

В Docker Compose по умолчанию:

| Переменная | Значение по умолчанию |
|------------|----------------------|
| `DB_PASSWORD` | `treepage` |
| `JWT_SECRET` | `dev-jwt-secret-change-in-production` |
| `DEV_MODE` | `true` (включает локальный вход) |

> **Важно:** значения по умолчанию предназначены только для разработки. В production используйте Kubernetes/Helm и надёжные секреты.

## Опциональный Redis

```bash
docker compose --profile cache up --build
```

## Остановка и очистка

```bash
# Остановить
docker compose down

# Остановить и удалить volumes (данные БД и клоны Git)
docker compose down -v
```

## Структура volumes

| Volume | Назначение |
|--------|------------|
| `postgres_data` | Данные PostgreSQL |
| `sync_repos` | Клоны Git-репозиториев (`/tmp/treepage-repos`) |

## Следующие шаги

- [Первый вход](../getting-started/first-login.md)
- [Первоначальная настройка](../getting-started/initial-setup.md)
