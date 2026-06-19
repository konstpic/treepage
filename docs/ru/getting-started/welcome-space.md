# Welcome space

При первом запуске TreePage автоматически создаёт **публичное welcome-пространство** с документацией из этого репозитория.

## Что создаётся

| Сущность | Значение |
|----------|----------|
| Slug пространства | `welcome` |
| Название | Welcome |
| Публичный доступ | Да (чтение без входа) |
| Git-репозиторий | [github.com/konstpic/treepage](https://github.com/konstpic/treepage) |
| Путь к документам | `docs/` (руководства EN + RU) |
| Первая синхронизация | Запускается один раз после bootstrap |

Открыть: **http://localhost:5173/spaces/welcome** (dev) или `/spaces/welcome` на вашем домене.

## Первый вход (локальный администратор)

Перед настройкой платформы войдите под bootstrap-администратором:

| Поле | Значение |
|------|----------|
| Email | `admin@local` |
| Пароль | `admin` |
| Роль | `super_admin` |

> Доступно при `DEV_MODE=true` (по умолчанию в Docker Compose). Смените пароль после первого входа.

Шаги:

1. Запустите TreePage (`docker compose up --build`)
2. Откройте `/auth` и войдите как `admin@local` / `admin`
3. Перейдите в **Пространства** → **Welcome** (или `/spaces/welcome`)
4. Если документов нет — нажмите **Синхронизировать** (нужен push в GitHub с папкой `docs/`)

## Переменные окружения

| Переменная | По умолчанию | Описание |
|------------|-------------|----------|
| `WELCOME_SPACE_ENABLED` | `true` | `false` — не создавать welcome space |
| `WELCOME_REPO_URL` | `https://github.com/konstpic/treepage.git` | URL репозитория |
| `WELCOME_REPO_BRANCH` | `main` | Ветка для clone |
| `WELCOME_DOCS_PATH` | `docs` | Поддиректория с Markdown |

## Production

- `DEV_MODE=false` + OIDC для доступа администраторов
- Welcome space остаётся **публичным** — подходит для документации продукта
- `WELCOME_REPO_URL` — ваш fork или internal mirror
- `WELCOME_SPACE_ENABLED=false` — если пространства создаются вручную

## См. также

- [Первый вход](first-login.md)
- [Первоначальная настройка](initial-setup.md)
- [Пространства (admin)](../admin/spaces.md)
