# Системные требования

## Минимальные ресурсы (production)

| Компонент | CPU | RAM | Диск |
|-----------|-----|-----|------|
| frontend | 50m | 64 Mi | — |
| backend-auth | 50m | 64 Mi | — |
| backend-server | 100m | 128 Mi | — |
| backend-sync | 100m | 128 Mi | 2 Gi (клоны репозиториев) |
| PostgreSQL | 500m | 512 Mi | 10 Gi+ |

Рекомендуется запускать backend-сервисы с минимум 2 репликами (кроме sync).

## Программное обеспечение

### Docker Compose

- Docker 24+
- Docker Compose v2

### Локальная разработка

- Go 1.22+
- Node.js 22+
- PostgreSQL 16+
- Git 2.x

### Kubernetes / Helm

- Kubernetes 1.24+
- Helm 3.10+
- Ingress controller (nginx рекомендуется)
- PostgreSQL 16+ (внешний или managed)

## Сеть

| Порт (dev) | Сервис |
|------------|--------|
| 5173 | frontend (Vite dev) |
| 8081 | backend-auth |
| 8082 | backend-server |
| 8083 | backend-sync |
| 5432 | PostgreSQL |

В production все сервисы доступны через Ingress на портах 80/443. Сервис sync не публикуется наружу — только внутри кластера.

## Внешние зависимости

| Сервис | Обязательность | Назначение |
|--------|---------------|------------|
| PostgreSQL | Обязательно | Основное хранилище |
| OIDC-провайдер | Production | SSO (Keycloak, Okta, Azure AD и др.) |
| Git-репозиторий | Для Git Sync | Источник документации |
| LLM API | Опционально | AI-книги, автоперевод документов |
| Redis | Опционально | Кэш (profile `cache` в Docker Compose) |

## Браузеры

Поддерживаются современные браузеры с ES2020+: Chrome, Firefox, Safari, Edge (последние 2 версии).
