# Справочник

**Language:** Русский · [English](../../en/reference/README.md)

Техническая документация для разработчиков, интеграторов и операторов.

## Документы

| Документ | Описание |
|----------|----------|
| [Архитектура](architecture.md) | Микросервисы, auth flow, RBAC, search, security |
| [Helm deployment (детали)](helm-deployment.md) | Charts, install modes, migration from legacy |
| [REST API](api.md) | OpenAPI endpoints overview |
| [Frontend analysis](frontend-analysis.md) | Design system reference (для разработчиков UI) |

## OpenAPI spec

Полная спецификация: [`openapi/openapi.yaml`](../../openapi/openapi.yaml)

## Исходный код

| Компонент | Путь |
|-----------|------|
| Frontend | `frontend/` |
| Auth service | `backend/auth/` |
| Server API | `backend/server/` |
| Sync worker | `backend/sync/` |
| Shared libs | `backend/pkg/` |
| Migrations | `migrations/` |
| Helm backend | `backend/` |
| Helm frontend | `.helm/frontend/` |
| Umbrella chart | `deploy/helm/treepage/` |

## Связанные разделы

- [Эксплуатация](../operator/README.md)
- [Установка](../installation/README.md)
