# First login

## Local administrator (first startup)

On first startup TreePage automatically creates an administrator account and a **welcome space** with bundled documentation.

### Credentials

| Field | Value |
|-------|--------|
| Email | `admin@local` |
| Password | `admin` |
| Role | `super_admin` |

> Available when `DEV_MODE=true` (Docker Compose default). **Change the password right after login:** Settings → Users.

### Step by step

1. Run `docker compose up --build`
2. Open http://localhost:5173
3. Click **Sign in** → `/auth`
4. Enter `admin@local` / `admin` (fields may be pre-filled)
5. After login open **Spaces** → **Welcome** (`/spaces/welcome`)

More about the welcome space: [Welcome space](welcome-space.md).

## Development mode (Docker Compose)

When running via `docker compose up`, dev mode is enabled (`DEV_MODE=true`). The auth service bootstrap creates `admin@local` on first start.

## Production: OIDC login

In production, SSO is used through an OIDC provider (Keycloak, Okta, Azure AD, etc.).

### How to sign in

1. Open your TreePage URL (e.g. `https://docs.example.com`)
2. Click **Sign in** → **Continue with OIDC**
3. You are redirected to the identity provider
4. After successful authentication, you return to TreePage with an active session

### What happens on first OIDC login

1. TreePage receives an authorization code from the OIDC provider
2. Exchanges the code for tokens
3. Creates or updates the user account in the database
4. Maps roles and groups from OIDC claims
5. Issues JWT access/refresh tokens
6. Redirects to `/auth/callback` → `/spaces`

## Session and tokens

| Parameter | Default value |
|-----------|---------------|
| Access token TTL | 15 minutes |
| Refresh token TTL | 7 days |

The frontend automatically refreshes the access token when it expires. On logout, the refresh token is revoked.

## Public access

Some spaces may be marked as **public** — their documentation is available without login. The **Welcome** space is public by default. The **Open spaces** button on the home page works without authorization.

## Login issues

| Symptom | Solution |
|---------|----------|
| "OIDC unavailable" | Check OIDC settings in Helm/config and provider availability |
| Redirect loop | Check `frontendUrl` and redirect URL in the OIDC client |
| "Login failed" (dev) | Ensure `DEV_MODE=true` and `ENV != prod` |
| 401 on API after login | Verify `JWT_SECRET` matches between auth and server |

Details: [Troubleshooting](../operator/troubleshooting.md).

## Next step

[Initial setup](initial-setup.md)
