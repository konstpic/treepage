# Резервное копирование и восстановление

TreePage хранит весь контент в **PostgreSQL**. Git-репозитории остаются upstream для синхронизируемых spaces; в БД — рабочая копия, версии, пользователи и настройки.

## Что бэкапить

| Данные | Где |
|--------|-----|
| Документы, версии, пользователи, RBAC | PostgreSQL, БД `treepage` |
| Кэш git clone (опционально) | Volume sync worker |
| Секреты | Kubernetes Secrets / `.env` — отдельно в vault |

## Бэкап PostgreSQL

```bash
# Docker Compose
docker compose exec postgres pg_dump -U treepage -Fc treepage > treepage-$(date +%Y%m%d).dump

# Plain SQL
docker compose exec postgres pg_dump -U treepage treepage > treepage-$(date +%Y%m%d).sql
```

## Восстановление

```bash
docker compose stop backend-auth backend-server backend-sync frontend

docker compose exec -T postgres pg_restore -U treepage -d treepage --clean --if-exists < treepage-20260101.dump

docker compose start backend-auth backend-server backend-sync frontend
```

## Проверка после restore

1. Откройте `/spaces` — spaces и документы на месте.
2. Войдите как admin → **Admin → Журнал аудита**.
3. Запустите **Sync**; при локальных правках в ответе будет `conflicts_skipped`.
4. Проверьте **История версий → Восстановить** на документе.

## Production checklist

- [ ] Ежедневный `pg_dump` в object storage
- [ ] Политика хранения (30 daily / 12 monthly)
- [ ] Ежеквартальный тест restore на staging
- [ ] Секреты (`INTERNAL_SERVICE_TOKEN`, `JWT_SECRET`, `DB_PASSWORD`) — в secrets manager
- [ ] Redis (OIDC state) эфемерен — после потери пользователи просто перелогинятся

## Связанные разделы

- [Миграции](migrations.md)
- [Troubleshooting](troubleshooting.md)
- [Конфигурация](configuration.md)
