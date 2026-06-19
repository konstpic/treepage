# Дорожная карта и фазы развития

TreePage развивается итерациями — от production-ready базы до enterprise knowledge base с RAG. Этот документ фиксирует **реализованные фазы** и их возможности.

---

## Обзор фаз

| Фаза | Миграция | Статус | Фокус |
|------|----------|--------|-------|
| **1 — Production Hardening** | `012` | ✅ | Безопасность sync, конфликты Git↔UI, audit |
| **2 — Team KB** | `013` | ✅ | Избранное, уведомления, вложения, черновики |
| **3 — Enterprise KB** | `014` | ✅ | ACL страниц, комментарии, workflow, аналитика, RAG |
| **Search & RAG** | `015`, `016` | ✅ | Multilingual FTS, embeddings, цитаты, feedback |
| **Platform** | — | ✅ | Автомигратор, pre-built Docker-образы |

---

## Фаза 1 — Production Hardening

**Цель:** подготовить платформу к production без потери локальных правок при Git sync.

### Backend

| Возможность | Описание |
|-------------|----------|
| **INTERNAL_SERVICE_TOKEN** | Защита API `backend-sync` — только `backend-server` с общим токеном |
| **Pending changes** | Поля `has_pending_changes`, `synced_content_hash`, `last_synced_at` |
| **Conflict skip** | Sync **не перезаписывает** документы с локальными правками; счётчик `conflicts_skipped` |
| **Audit indexes** | Индексы на `audit_logs` для быстрого просмотра |
| **OIDC + Redis** | State OIDC в Redis (не in-memory) — несколько реплик `backend-auth` |

### Где читать

- [Устранение неполадок — Sync API 401](../operator/troubleshooting.md)
- [Устранение неполадок — Git sync не перезаписывает](../operator/troubleshooting.md)
- [Редактирование — конфликты Git](../user/editing-docs.md)

### Переменные окружения

```bash
INTERNAL_SERVICE_TOKEN=your-long-random-secret   # server + sync
REDIS_ADDR=redis:6379                            # auth (несколько реплик)
```

---

## Фаза 2 — Team KB

**Цель:** командная работа — избранное, уведомления, вложения, черновики.

### Backend

| Возможность | API / таблицы |
|-------------|---------------|
| **Избранное** | `/api/me/favorites`, `user_favorites` |
| **Недавние** | `/api/me/recent`, `user_recent_views` |
| **Уведомления** | `/api/notifications`, `notifications` |
| **Вложения** | `/api/documents/:id/attachments`, `document_attachments` |
| **Черновики** | `is_published=false`, publish-local |
| **GitLab MR** | Provider `gitlab` в настройках репозитория |
| **Orphan cleanup** | После sync удаляются документы, которых нет в Git (кроме pending changes) |
| **OpenSearch (opt-in)** | `SEARCH_BACKEND=opensearch` — заготовка, пока делегирует в PostgreSQL |

### Frontend

- `/me` — избранное и недавние
- Колокольчик уведомлений в навигации
- Звёздочка, черновики, вложения на странице документа

### Где читать

- [Редактирование документов](../user/editing-docs.md)

---

## Фаза 3 — Enterprise KB

**Цель:** права на уровне страницы, совместная работа, аналитика, базовый RAG.

### Backend

| Возможность | Описание |
|-------------|----------|
| **Page ACL** | Правила на путь внутри space (`page_acl_rules`), наследование от родительского path |
| **Комментарии** | Треды на документ, `@mentions` |
| **Workflow** | Состояния `draft` → `review` → `approved` → `published` |
| **Аналитика** | Просмотры, популярные документы, лог поисковых запросов |
| **RAG chunks** | Таблица `document_chunks` — индексация при sync и backfill при старте |

### Frontend

- Комментарии и workflow на странице документа
- `/admin/analytics` — дашборд аналитики
- `/search` — вкладка **Спросить документацию** (RAG)

### Где читать

- [Поиск — RAG](../user/search.md)
- [RBAC](../admin/rbac.md)

---

## Search & RAG (015, 016 и далее)

**Цель:** умный поиск на русском и английском, гибридный retrieval, ответы с цитатами.

### Multilingual FTS (`015`)

- PostgreSQL FTS: конфигурации `english`, `russian`, `simple`
- Обычный поиск и RAG retrieval учитывают кириллицу

### RAG enhancements (`016`)

| Компонент | Описание |
|-----------|----------|
| **Embeddings** | Колонка `document_chunks.embedding` (JSONB), Ollama или OpenAI |
| **Hybrid search** | FTS + cosine similarity (≈40% FTS + 60% vector) |
| **Multi-strategy retrieval** | phrase, websearch, keywords, ILIKE, LLM-expanded queries |
| **Цитаты** | Точные фрагменты из документа в ответе |
| **Confidence** | Оценка уверенности; при низкой — follow-up вопросы |
| **Feedback 👍/👎** | `/api/rag/feedback`, обучение синонимов (`rag_learned_synonyms`) |

### Архитектура RAG

```
Git sync / startup backfill
    → document_chunks (+ embedding)
    → POST /api/rag/ask
        → retrieval (FTS + vector + learned synonyms)
        → LLM (OpenAI-compatible)
        → answer + sources + citations + confidence
```

RAG живёт в **backend-server** (`backend/server/internal/rag/`), отдельного контейнера нет.

### Настройка LLM и embeddings

```yaml
# docker-compose / Helm extraEnv на backend-server
LLM_ENABLED: "true"
LLM_API_URL: http://host:11434/v1          # Ollama или OpenAI-compatible
LLM_MODEL: llama3.2:latest
LLM_API_KEY: ""                            # не нужен для локального Ollama

EMBEDDING_ENABLED: "true"
EMBEDDING_MODEL: nomic-embed-text          # ollama pull nomic-embed-text
```

Подробнее: [Конфигурация — LLM](../operator/configuration.md), [Секреты](../operator/secrets.md).

---

## Platform — инфраструктура

### Автомигратор

- Папка `migrations/*.up.sql` сканируется при старте `backend-server`
- Версии в `schema_migrations`; уже применённые **пропускаются**
- Новая миграция = файл `017_*.up.sql` + restart server
- См. [Миграции](../operator/migrations.md)

### Docker Compose (dev / demo)

| Было | Стало |
|------|-------|
| Air hot-reload, `go build` при каждом старте | Pre-built бинарник в alpine-образе |
| 14 volume mount'ов миграций в postgres | Одна строка `./migrations:/app/migrations` |

- **Frontend** — Vite hot reload (как раньше)
- **Backend** — после изменений Go: `docker compose up -d --build backend-server …`
- Активная разработка backend без Docker: `go run` / Air локально

См. [Docker Compose](../installation/docker-compose.md).

---

## Матрица миграций

| Файл | Фаза |
|------|------|
| `001`–`011` | Базовая платформа |
| `012_production_hardening` | Фаза 1 |
| `013_team_kb` | Фаза 2 |
| `014_enterprise_kb` | Фаза 3 |
| `015_multilingual_search` | Search & RAG |
| `016_rag_enhancements` | Search & RAG |

> `004` пропущена в нумерации.

---

## Связанные разделы

- [Архитектура](architecture.md)
- [Миграции](../operator/migrations.md)
- [Поиск (пользователь)](../user/search.md)
- [Устранение неполадок — LLM/RAG](../operator/troubleshooting.md)
