# Эксплуатация TreePage

Документация для DevOps и операторов платформы.

## Разделы

| Тема | Документ |
|------|----------|
| Конфигурация сервисов | [Конфигурация](configuration.md) |
| Секреты и env vars | [Секреты](secrets.md) |
| Миграции БД | [Миграции](migrations.md) |
| Backup / restore | [Резервное копирование](backup-restore.md) |
| Health checks и метрики | [Мониторинг](monitoring.md) |
| Диагностика проблем | [Устранение неполадок](troubleshooting.md) |
| Дорожная карта (фазы 1–3) | [Roadmap](../reference/roadmap.md) |

## Компоненты

| Сервис | Порт | Обязательный |
|--------|------|:------------:|
| frontend | 5173 (dev) / 80 (prod) | ✅ |
| backend-auth | 8081 | ✅ |
| backend-server | 8082 | ✅ |
| backend-sync | 8083 | ✅ |
| PostgreSQL | 5432 | ✅ |
| Redis | 6379 | ❌ |

## Load order конфигурации

```
1. YAML: /opt/app/conf/config.yml
2. Environment variables (override)
3. Validation → fail fast
```

## Production checklist

- [ ] Надёжные секреты (не defaults)
- [ ] `DEV_MODE=false`, `ENV=prod`
- [ ] OIDC настроен
- [ ] TLS на Ingress
- [ ] Миграции применены
- [ ] Backup PostgreSQL настроен
- [ ] Мониторинг health endpoints
- [ ] PVC для sync (≥ 2Gi)
- [ ] Resource limits в Helm values

## Связанные разделы

- [Kubernetes / Helm](../installation/kubernetes.md)
- [Архитектура](../reference/architecture.md)
