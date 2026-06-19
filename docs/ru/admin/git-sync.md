# Синхронизация Git

Sync-worker (`backend-sync`) клонирует Git-репозитории и импортирует Markdown в PostgreSQL.

## Режимы синхронизации

| Режим | Описание | Когда использовать |
|-------|----------|-------------------|
| `manual` | Только по кнопке или API | Редко обновляемая документация |
| `scheduled` | Автоматически по интервалу | Основной режим (default: 300 сек) |
| `webhook` | По HTTP-запросу от Git | Мгновенное обновление при push |

## Процесс синхронизации

```
1. Создаётся запись sync_jobs (status: running)
2. git clone --depth 1 --branch <branch> <url>
3. Обход {clone}/{docs_path}/**/*.md
4. Для каждого файла: title, slug, tags, content
5. Upsert в documents по (space_id, slug)
6. Обновление last_sync_at, last_sync_status
```

## Триггеры

### 1. Scheduled (фоновый)

Sync-worker запускает ticker с интервалом из конфигурации:

- Глобальный default: `sync.git.syncInterval` (300s в Helm)
- Per-repo override: `sync_interval_seconds` в настройках репозитория

Синхронизируются только репозитории с `enabled: true`.

### 2. Manual (ручной)

**Из UI:**
- `/admin/repositories` → кнопка sync
- `/spaces/{slug}` → **Синхронизировать** в sidebar

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

Sync-worker слушает на порту 8083 (внутри кластера, не через Ingress).

#### Настройка webhook в GitHub

1. Repository → Settings → Webhooks → Add webhook
2. Payload URL: `http://backend-sync:8083/api/sync/webhook/{repo-id}` (internal)
   - Для external: настройте reverse proxy или используйте scheduled
3. Content type: `application/json`
4. Secret: значение `GIT_WEBHOOK_SECRET`
5. Events: Push events

#### Настройка webhook в GitLab

1. Project → Settings → Webhooks
2. URL: `http://backend-sync:8083/api/sync/webhook/{repo-id}`
3. Secret token: значение `GIT_WEBHOOK_SECRET`
4. Trigger: Push events

> В production webhook URL должен быть доступен из Git-платформы. Обычно sync не публикуется через Ingress — используйте internal network или scheduled sync.

## Хранение клонов

| Среда | Путь |
|-------|------|
| Docker Compose | `/tmp/treepage-repos` (volume `sync_repos`) |
| Kubernetes | `/data/repos` (PVC, default 2Gi) |

Клоны shallow (`--depth 1`) — пересоздаются при каждой синхронизации.

## Мониторинг sync

| Поле | Где смотреть |
|------|-------------|
| `last_sync_at` | Список репозиториев в admin |
| `last_sync_status` | success / failed / completed |
| `last_sync_error` | Текст ошибки при failed |
| `sync_jobs` | Таблица в PostgreSQL |

## Типичные ошибки

| Ошибка | Причина | Решение |
|--------|---------|---------|
| Authentication failed | Неверный или отсутствующий token | Проверьте `GIT_ACCESS_TOKEN` |
| Branch not found | Неверное имя ветки | Укажите правильную ветку |
| Path not found | Неверный `docs_path` | Проверьте структуру репозитория |
| Connection timeout | Git-сервер недоступен | Проверьте сеть из pod sync |

Подробнее: [Устранение неполадок](../operator/troubleshooting.md)

## Связанные разделы

- [Репозитории](repositories.md)
- [Конфигурация](../operator/configuration.md)
