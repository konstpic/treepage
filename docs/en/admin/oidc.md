# OIDC providers

**URL:** `/admin/oidc`

> Available only to `super_admin`.

## Purpose

OIDC (OpenID Connect) provides Single Sign-On through a corporate Identity Provider:

- Keycloak
- Okta
- Azure AD / Entra ID
- Google Workspace
- Auth0
- Any OIDC-compatible provider

## Creating a provider

| Field | Required | Example |
|-------|:--------:|---------|
| Name | ✅ | `Keycloak Production` |
| Issuer URL | ✅ | `https://keycloak.example.com/realms/treepage` |
| Client ID | ✅ | `treepage` |
| Redirect URL | ✅ | `https://docs.example.com/api/auth/callback` |
| Scopes | ✅ | `openid profile email` |
| Role claim | ❌ | `roles` |
| Group claim | ❌ | `groups` |
| Sync groups | ❌ | ✅ |

Client secret is set via env (`OIDC_CLIENT_SECRET`) or Kubernetes Secret — not through UI.

## Provider-side configuration

### Authentik (local dev)

See [Authentik OIDC test](../getting-started/authentik-oidc-test.md) for a Docker Compose overlay with preconfigured OAuth2 client.

### Keycloak

1. **Clients** → Create client
   - Client ID: `treepage`
   - Client authentication: ON
   - Standard flow: ON
2. **Valid redirect URIs:** `https://docs.example.com/api/auth/callback`
3. **Web origins:** `https://docs.example.com`
4. **Credentials** → copy Client secret → `OIDC_CLIENT_SECRET`
5. **Client scopes** → add mapper for `roles` and `groups`

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

## Authentication flow

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

## Role mapping

TreePage reads roles from JWT claim (default: `roles`):

```json
{
  "email": "user@company.com",
  "roles": ["editor"],
  "groups": ["developers", "platform-team"]
}
```

Values must match TreePage system roles: `super_admin`, `admin`, `editor`, `viewer`.

### Keycloak: mapper for roles

1. Client scopes → `{client}-dedicated`
2. Add mapper → User Realm Role
3. Token Claim Name: `roles`
4. Add to access token: ON

## Group mapping

When `sync_groups: true`:

1. Groups claim is synchronized on each login
2. User is added to TreePage groups with matching names
3. Space permissions apply via group assignments

## Helm OIDC configuration

Alternative to UI — configure via Helm:

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

## Multiple providers

UI supports CRUD for multiple OIDC providers. The active provider is determined by auth service configuration.

## Troubleshooting

| Issue | Solution |
|-------|----------|
| Redirect URI mismatch | Check URL in OIDC client and `frontendUrl` |
| Invalid client | Check `OIDC_CLIENT_SECRET` |
| No roles assigned | Configure role claim mapper |
| Groups not synced | Enable `sync_groups`, check group claim |

Details: [Troubleshooting](../operator/troubleshooting.md)

## Related sections

- [First login](../getting-started/first-login.md)
- [RBAC](rbac.md)
- [Secrets](../operator/secrets.md)
