# OIDC-провайдеры

**URL:** `/admin/oidc`

> Доступно только для `super_admin`.

## Назначение

OIDC (OpenID Connect) обеспечивает Single Sign-On через корпоративный Identity Provider:

- Keycloak
- Okta
- Azure AD / Entra ID
- Google Workspace
- Auth0
- Любой OIDC-совместимый провайдер

## Создание провайдера

| Поле | Обязательно | Пример |
|------|:-----------:|--------|
| Name | ✅ | `Keycloak Production` |
| Issuer URL | ✅ | `https://keycloak.example.com/realms/treepage` |
| Client ID | ✅ | `treepage` |
| Redirect URL | ✅ | `https://docs.example.com/api/auth/callback` |
| Scopes | ✅ | `openid profile email` |
| Role claim | ❌ | `roles` |
| Group claim | ❌ | `groups` |
| Sync groups | ❌ | ✅ |

Client secret задаётся через env (`OIDC_CLIENT_SECRET`) или Kubernetes Secret — не через UI.

## Настройка на стороне провайдера

### Authentik (локальный dev)

См. [Тест OIDC с Authentik](../getting-started/authentik-oidc-test.md) — overlay Docker Compose с готовым OAuth2-клиентом.

### Keycloak

1. **Clients** → Create client
   - Client ID: `treepage`
   - Client authentication: ON
   - Standard flow: ON
2. **Valid redirect URIs:** `https://docs.example.com/api/auth/callback`
3. **Web origins:** `https://docs.example.com`
4. **Credentials** → скопируйте Client secret → `OIDC_CLIENT_SECRET`
5. **Client scopes** → добавьте mapper для `roles` и `groups`

### Azure AD / Entra ID

1. App registrations → New registration
2. Redirect URI: `https://docs.example.com/api/auth/callback`
3. Certificates & secrets → New client secret
4. Token configuration → Add optional claims: `groups`, custom roles

### Okta

1. Applications → Create App Integration → OIDC → Web
2. Sign-in redirect URI: `https://docs.example.com/api/auth/callback`
3. Assign groups
4. Configure group claims in Authorization Server

## Flow аутентификации

```
User → /auth → GET /api/auth/login
     → Redirect to OIDC provider
     → User authenticates
     → OIDC → GET /api/auth/callback?code=...
     → Exchange code for tokens
     → Upsert user, sync roles/groups
     → Redirect to /auth/callback?access_token=...&refresh_token=...
     → Frontend stores JWT
     → Redirect to /spaces
```

## Маппинг ролей

TreePage читает roles из JWT claim (default: `roles`):

```json
{
  "email": "user@company.com",
  "roles": ["editor"],
  "groups": ["developers", "platform-team"]
}
```

Значения должны совпадать с системными ролями TreePage: `super_admin`, `admin`, `editor`, `viewer`.

### Keycloak: mapper для roles

1. Client scopes → `{client}-dedicated`
2. Add mapper → User Realm Role
3. Token Claim Name: `roles`
4. Add to access token: ON

## Маппинг групп

При `sync_groups: true`:

1. Groups claim синхронизируется при каждом входе
2. Пользователь добавляется в группы TreePage с matching names
3. Права в пространствах применяются через group assignments

## Helm-конфигурация OIDC

Альтернатива UI — настройка через Helm:

```yaml
auth:
  oidc:
    enabled: true
    issuerUrl: https://keycloak.example.com/realms/treepage
    clientId: treepage
    scopes: openid profile email

secret:
  oidcClientSecret: "<client-secret>"

global:
  frontendUrl: https://docs.example.com
```

## Несколько провайдеров

UI поддерживает CRUD нескольких OIDC-провайдеров. Активный провайдер определяется конфигурацией auth-сервиса.

## Troubleshooting

| Проблема | Решение |
|----------|---------|
| Redirect URI mismatch | Проверьте URL в OIDC client и `frontendUrl` |
| Invalid client | Проверьте `OIDC_CLIENT_SECRET` |
| No roles assigned | Настройте role claim mapper |
| Groups not synced | Включите `sync_groups`, проверьте group claim |

Подробнее: [Устранение неполадок](../operator/troubleshooting.md)

## Связанные разделы

- [Первый вход](../getting-started/first-login.md)
- [RBAC](rbac.md)
- [Секреты](../operator/secrets.md)
