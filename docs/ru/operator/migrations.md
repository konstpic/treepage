# Миграции базы данных

SQL-миграции находятся в `migrations/`.

## Список миграций

| Файл | Описание |
|------|----------|
| `001_initial_schema.up.sql` | Core: users, roles, spaces, repos, documents, sync_jobs, audit |
| `002_local_auth.up.sql` | password_hash для локального входа |
| `003_system_settings.up.sql` | Таблица system_settings |
| `005_ui_theme.up.sql` | UI theme setting |
| `006_books.up.sql` | AI books |
| `007_ui_language.up.sql` | UI language setting |
| `008_document_translations.up.sql` | Переводы документов |
| `009_book_translations.up.sql` | Переводы книг |
| `010_space_groups.up.sql` | Группы в пространствах |
| `011_oidc_sync_groups.up.sql` | sync_groups flag для OIDC |
| `012_production_hardening.up.sql` | Pending changes, audit, OIDC Redis |
| `013_team_kb.up.sql` | Избранное, недавние, уведомления, вложения |
| `014_enterprise_kb.up.sql` | ACL страниц, комментарии, workflow, аналитика, RAG |
| `015_multilingual_search.up.sql` | Multilingual FTS (RU + EN) |
| `016_rag_enhancements.up.sql` | Embeddings, feedback, learned synonyms |

> Миграция `004` отсутствует (пропущена в нумерации).

## Автоматическое применение

Все файлы `migrations/*_up.sql` применяются **автоматически** при старте `backend-server`:

1. Сканируется папка `migrations/` (env `MIGRATIONS_DIR`, по умолчанию `/app/migrations` в Docker)
2. Файлы сортируются по имени (`001_…`, `002_…`, …, `100_…`)
3. Неприменённые версии записываются в таблицу `schema_migrations`

**Новая миграция** = положить файл `017_*.up.sql` в `migrations/` и перезапустить server. Редактировать `docker-compose.yml` не нужно.

Docker Compose монтирует `./migrations:/app/migrations:ro` в `backend-server`.

## Docker Compose

PostgreSQL больше **не** монтирует миграции по одной. Схема поднимается через migrator в `backend-server` (после `postgres` healthy).

`backend-auth` и `backend-sync` стартуют после `backend-server`, чтобы миграции успели примениться.

## Ручное применение (без Docker)

```bash
export PGPASSWORD=<password>
export MIGRATIONS_DIR=migrations
# через backend-server при старте — предпочтительно

# или вручную одной командой (legacy):
for f in migrations/*_up.sql; do
  echo "Applying $f..."
  psql -h <host> -U treepage -d treepage -f "$f"
done
```

## Проверка

```sql
SELECT version, applied_at FROM schema_migrations ORDER BY version;
```

## Kubernetes

Chart не включает init job для миграций. Примените вручную перед первым деплоем:

```bash
# Port-forward или direct connection
kubectl run psql-client --rm -it --image=postgres:16-alpine -- \
  psql -h postgres.default.svc -U treepage -d treepage

# Или через Job (пример)
kubectl create configmap migrations --from-file=migrations/
```

## Откат

Для каждой миграции есть `.down.sql`:

```bash
psql -U treepage -d treepage -f migrations/011_oidc_sync_groups.down.sql
```

> Откатывайте в обратном порядке. В production откат — крайняя мера.

## Проверка

```sql
-- Проверить таблицы
\dt

-- Пользователи
SELECT email, display_name FROM users;

-- Системные настройки
SELECT * FROM system_settings;
```

## Backup перед миграцией

```bash
pg_dump -h <host> -U treepage -d treepage -F c -f treepage_backup_$(date +%Y%m%d).dump
```

## Связанные разделы

- [Установка](../installation/README.md)
- [Устранение неполадок](troubleshooting.md)
