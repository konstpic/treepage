# TreePage — документация

**Language:** Русский · [English](../en/README.md)

**TreePage** — self-hosted платформа документации с синхронизацией из Git, рендером Markdown и Mermaid, полнотекстовым и RAG-поиском, RBAC и OIDC-аутентификацией.

**Репозиторий:** [github.com/konstpic/treepage](https://github.com/konstpic/treepage)

Эта папка содержит полное руководство: от установки до ежедневной работы пользователей и администраторов.

---

## С чего начать

| Роль | Рекомендуемый путь |
|------|-------------------|
| **Оператор / DevOps** | [Установка](installation/README.md) → [Первоначальная настройка](getting-started/initial-setup.md) |
| **Администратор платформы** | [Быстрый старт](getting-started/README.md) → [Руководство администратора](admin/README.md) |
| **Пользователь / редактор** | [Первый вход](getting-started/first-login.md) → [Руководство пользователя](user/README.md) |

---

## Разделы

### [Установка](installation/README.md)

Развёртывание TreePage: Docker Compose, локальная разработка, Kubernetes/Helm.

- [Требования](installation/requirements.md)
- [Docker Compose](installation/docker-compose.md)
- [Локальная разработка](installation/local-development.md)
- [Kubernetes / Helm](installation/kubernetes.md)

### [Быстрый старт](getting-started/README.md)

Первый запуск и базовая конфигурация после установки.

- [Первый вход](getting-started/first-login.md)
- [Welcome space](getting-started/welcome-space.md)
- [Первоначальная настройка](getting-started/initial-setup.md)

### [Руководство пользователя](user/README.md)

Работа с документацией: пространства, чтение, поиск, редактирование, книги.

- [Навигация по интерфейсу](user/navigation.md)
- [Пространства](user/spaces.md)
- [Чтение документов](user/reading-docs.md)
- [Поиск](user/search.md)
- [Редактирование документов](user/editing-docs.md)
- [Книги (AI-сборки)](user/books.md)

### [Руководство администратора](admin/README.md)

Управление платформой: пространства, репозитории, пользователи, OIDC, настройки.

- [Роли и права (RBAC)](admin/rbac.md)
- [Пространства](admin/spaces.md)
- [Репозитории Git](admin/repositories.md)
- [Синхронизация Git](admin/git-sync.md)
- [Пользователи](admin/users.md)
- [Группы](admin/groups.md)
- [Системные настройки](admin/settings.md)
- [OIDC-провайдеры](admin/oidc.md)

### [Эксплуатация](operator/README.md)

Конфигурация, секреты, миграции, мониторинг, устранение неполадок.

- [Конфигурация](operator/configuration.md)
- [Секреты и переменные окружения](operator/secrets.md)
- [Миграции базы данных](operator/migrations.md)
- [Мониторинг и health checks](operator/monitoring.md)
- [Устранение неполадок](operator/troubleshooting.md)

### [Справочник](reference/README.md)

Техническая документация для разработчиков и интеграторов.

- [Архитектура](reference/architecture.md)
- [Дорожная карта (фазы 1–3, RAG)](reference/roadmap.md)
- [Развёртывание Helm (детали)](reference/helm-deployment.md)
- [REST API](reference/api.md)

---

## Архитектура (кратко)

```
frontend (React)  →  backend-auth (OIDC/JWT)
                  →  backend-server (Docs API, Search, RAG, Admin)
                  →  backend-sync (Git sync worker)
                  →  PostgreSQL
```

Подробнее: [Архитектура](reference/architecture.md).
