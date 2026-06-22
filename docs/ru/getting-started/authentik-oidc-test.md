# Тест OIDC с Authentik (dev)

Опциональный стек Authentik для локальной проверки OpenID Connect вместе с TreePage.

## Быстрый старт

```bash
cp .env.authentik.example .env.authentik
cat .env.authentik >> .env

chmod +x scripts/deploy-dev-authentik.sh
./scripts/deploy-dev-authentik.sh
```

| Сервис | URL |
|--------|-----|
| TreePage UI | http://localhost:8080 |
| Authentik admin | http://localhost:9000/if/admin/ |
| OIDC issuer (внутри Docker) | `http://authentik-server:9000/application/o/treepage/` |

## Учётные данные по умолчанию

### Локальный админ TreePage (всегда доступен)

| Поле | Значение |
|------|----------|
| Email | `admin@local` |
| Пароль | `admin` |

Для первичной настройки и когда Authentik/OIDC недоступен. Смените пароль после входа.

### Админ Authentik (только IdP)

| Поле | Значение |
|------|----------|
| Email | `admin@authentik.local` |
| Пароль | `authentik` |

Blueprint `deploy/authentik/blueprints/treepage-test-users.yaml` создаёт **10 пользователей**. Пароль у всех: **`Test123!`**

| Email | Роль TreePage | Группы Authentik |
|-------|---------------|------------------|
| `alice.admin@test.local` | super_admin | platform-team, admins |
| `bob.admin@test.local` | admin | ops-team |
| `carol.editor@test.local` | editor | developers, docs-writers |
| `dave.editor@test.local` | editor | developers |
| `eve.viewer@test.local` | viewer | docs-readers |
| `frank.viewer@test.local` | viewer | *(нет — только роль)* |
| `grace.editor@test.local` | editor | *(нет — только роль)* |
| `henry.admin@test.local` | admin | *(нет — только роль)* |
| `iris.viewer@test.local` | viewer | guests, docs-readers |
| `jack.editor@test.local` | editor | contractors |

Роли — атрибут `treepage_roles`, группы — членство в Authentik. TreePage синхронизирует группы при `oidc.sync_groups: true`.

## Dev-сервер (192.168.0.64)

```bash
cat .env.authentik.server.example >> .env
COMPOSE_FILE=docker-compose.dev.yml:docker-compose.authentik.yml ./scripts/deploy-dev.sh
```

| Сервис | URL |
|--------|-----|
| TreePage | http://192.168.0.64:8090 |
| Authentik admin | http://192.168.0.64:9000/if/admin/ |

## Что настраивается автоматически

Blueprint'ы в `deploy/authentik/blueprints/`:

- OAuth2/OIDC provider **treepage** с claims `roles` и `groups`
- Application **treepage**, redirect URI для localhost и dev-сервера
- Client ID `treepage`, secret `treepage-authentik-dev-secret`
- 8 групп и 10 тестовых пользователей

`backend-auth`: `config.authentik.yml` (локально) или `config.authentik.server.yml` (`AUTH_CONFIG_PATH` на сервере).

## Проверка OIDC

1. Откройте http://localhost:8080/auth
2. **Продолжить с OIDC**
3. Войдите пользователем Authentik
4. Вернётесь в TreePage с JWT-сессией

## Проверка локального fallback

1. Остановите Authentik: `docker compose -f docker-compose.dev.yml -f docker-compose.authentik.yml stop authentik-server authentik-worker`
2. Откройте http://localhost:8080/auth
3. Войдите как `admin@local` / `admin`

## Связанные разделы

- [Первый вход](first-login.md)
- [OIDC-провайдеры](../admin/oidc.md)
