# Репозитории Git

**URL:** `/admin/repositories`

Репозитории — источники Markdown-документации для пространств.

## Создание репозитория

| Поле | Обязательно | По умолчанию | Описание |
|------|:-----------:|-------------|----------|
| Пространство | ✅ | — | Пространство, к которому привязан репозиторий |
| Название | ✅ | — | Человекочитаемое имя |
| URL | ✅ | — | HTTPS URL Git-репозитория |
| Ветка | ✅ | `main` | Ветка для clone |
| Путь к документам | ✅ | `docs` | Поддиректория с `.md` файлами |
| Режим синхронизации | ✅ | `manual` | manual / scheduled / webhook |
| Интервал (сек) | ❌ | `300` | Для режима scheduled |
| Токен доступа | ❌ | — | Env ref или literal token |
| Webhook secret | ❌ | — | Env ref или literal secret |
| Включён | ❌ | ✅ | Участвует в scheduled sync |

### Поддерживаемые Git-платформы

GitHub, GitLab, Gitea, Bitbucket и любой Git-сервер с HTTPS.

### URL формат

```
https://github.com/org/repo.git
https://gitlab.example.com/group/project.git
https://gitea.example.com/user/docs.git
```

## Редактирование

Нажмите иконку карандаша → измените поля → сохраните.

## Ручная синхронизация

Нажмите иконку **обновления** (RefreshCw) на карточке репозитория.

Статус отображается в колонке sync:

| Статус | Значение |
|--------|----------|
| success / completed | Синхронизация успешна |
| failed | Ошибка (см. last_sync_error) |
| — | Ещё не синхронизировался |

## Удаление

Удаление репозитория не удаляет документы из Git — только конфигурацию sync в TreePage.

## Токены доступа

### Через env ref (рекомендуется)

Укажите имя переменной окружения:

```
GIT_ACCESS_TOKEN
```

Значение задаётся в Kubernetes Secret или `.env`.

### Literal token

Можно указать токен напрямую (не рекомендуется для production):

```
ghp_xxxxxxxxxxxxxxxxxxxx
```

### Создание токена

| Платформа | Минимальные права |
|-----------|------------------|
| GitHub | `repo` (read) для private repos |
| GitLab | `read_repository` |
| Gitea | `read` access to repository |

## Структура docs_path

TreePage сканирует `{clone}/{docs_path}/**/*.md`:

```
repo/
└── docs/                    ← docs_path = "docs"
    ├── index.md             → slug: index
    ├── getting-started/
    │   └── install.md       → slug: getting-started/install
    └── api/
        └── reference.md     → slug: api/reference
```

### Метаданные из Markdown

| Источник | Поле |
|----------|------|
| Первый `# H1` | Заголовок |
| Имя файла | Slug (если нет H1) |
| `tags: tag1, tag2` (первая строка) | Теги для поиска |

## Привязка к другому пространству

Через **Пространства** → редактирование → отвязать/привязать.

Или API: `POST /api/admin/spaces/{id}/bind-repo`

## Связанные разделы

- [Git Sync](git-sync.md)
- [Секреты](../operator/secrets.md)
