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

> Миграция `004` отсутствует (пропущена в нумерации).

## Docker Compose

Миграции применяются автоматически при первом запуске PostgreSQL через `docker-entrypoint-initdb.d/`.

## Ручное применение

```bash
export PGPASSWORD=<password>

for f in migrations/00*_up.sql migrations/01*_up.sql; do
  echo "Applying $f..."
  psql -h <host> -U treepage -d treepage -f "$f"
done
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
