# First login

## Local administrator (first startup)

On first startup TreePage automatically creates an administrator account and a **welcome space** with bundled documentation.

### Credentials

| Field | Value |
|-------|--------|
| Email | `admin@local` |
| Password | `admin` |
| Role | `super_admin` |

> **Always available** — for first setup and when the OIDC provider is unavailable. **Change the password right after login:** Settings → Users.

### Step by step

1. Run `docker compose up --build` (or `./scripts/deploy-dev.sh`)
2. Open http://localhost:8080
3. Click **Sign in** → `/auth`
4. Enter `admin@local` / `admin` (fields may be pre-filled)
5. After login open **Spaces** → **Welcome** (`/spaces/welcome`)

More about the welcome space: [Welcome space](welcome-space.md).

## Local admin vs OIDC

The login page always offers **local sign-in** (for accounts with a password) and **OIDC** when configured. Use local admin for initial setup; OIDC for day-to-day SSO.

To test OIDC locally with Authentik: [Authentik OIDC test](authentik-oidc-test.md).

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
| "Login failed" (local) | Check email/password; account must have a local password set |
| 401 on API after login | Verify `JWT_SECRET` matches between auth and server |

Details: [Troubleshooting](../operator/troubleshooting.md).

## Next step

[Initial setup](initial-setup.md)
