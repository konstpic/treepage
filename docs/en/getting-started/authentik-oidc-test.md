# OIDC testing with Authentik (dev)

Optional Authentik stack for local OpenID Connect testing alongside TreePage.

## Quick start

```bash
cp .env.authentik.example .env.authentik
# merge into .env or export vars manually
cat .env.authentik >> .env

chmod +x scripts/deploy-dev-authentik.sh
./scripts/deploy-dev-authentik.sh
```

| Service | URL |
|---------|-----|
| TreePage UI | http://localhost:8080 |
| Authentik admin | http://localhost:9000/if/admin/ |
| OIDC issuer (internal) | `http://authentik-server:9000/application/o/treepage/` |

## Default credentials

### TreePage local admin (always available)

| Field | Value |
|-------|--------|
| Email | `admin@local` |
| Password | `admin` |

Use for first setup and when Authentik/OIDC is down. Change the password after login.

### Authentik admin (IdP only)

| Field | Value |
|-------|--------|
| Email | `admin@authentik.local` |
| Password | `authentik` |

Create end-user accounts in Authentik → **Directory → Users** for OIDC login tests.

## Test users (blueprint)

Blueprint `deploy/authentik/blueprints/treepage-test-users.yaml` creates **10 users**. Password for all: **`Test123!`**

| Email | TreePage role | Authentik groups |
|-------|---------------|------------------|
| `alice.admin@test.local` | super_admin | platform-team, admins |
| `bob.admin@test.local` | admin | ops-team |
| `carol.editor@test.local` | editor | developers, docs-writers |
| `dave.editor@test.local` | editor | developers |
| `eve.viewer@test.local` | viewer | docs-readers |
| `frank.viewer@test.local` | viewer | *(none — role only)* |
| `grace.editor@test.local` | editor | *(none — role only)* |
| `henry.admin@test.local` | admin | *(none — role only)* |
| `iris.viewer@test.local` | viewer | guests, docs-readers |
| `jack.editor@test.local` | editor | contractors |

Roles come from user attribute `treepage_roles`; groups from Authentik group membership. TreePage syncs groups when `oidc.sync_groups: true` in auth config.

## Dev server (192.168.0.64)

```bash
cat .env.authentik.server.example >> .env
COMPOSE_FILE=docker-compose.dev.yml:docker-compose.authentik.yml ./scripts/deploy-dev.sh
```

| Service | URL |
|---------|-----|
| TreePage | http://192.168.0.64:8090 |
| Authentik admin | http://192.168.0.64:9000/if/admin/ |

## What gets configured automatically

Blueprints under `deploy/authentik/blueprints/` create:

- OAuth2/OIDC provider **treepage** with `roles` and `groups` claims
- Application slug **treepage**
- Redirect URIs for localhost and dev server
- Client ID `treepage`, secret `treepage-authentik-dev-secret` (dev only)
- 8 groups and 10 test users (see above)

`backend-auth` uses `config.authentik.yml` (local) or `config.authentik.server.yml` (dev server via `AUTH_CONFIG_PATH`).

## Test OIDC login

1. Open http://localhost:8080/auth
2. Click **Continue with OIDC**
3. Sign in with an Authentik user
4. You return to TreePage with a JWT session

## Test local fallback

1. Stop Authentik: `docker compose -f docker-compose.dev.yml -f docker-compose.authentik.yml stop authentik-server authentik-worker`
2. Open http://localhost:8080/auth
3. Sign in with `admin@local` / `admin`

## Custom ports

If `:8080` or `:9000` are taken, set in `.env`:

```bash
FRONTEND_PORT=8090
AUTHENTIK_PORT=9001
```

Then update redirect URIs in `deploy/authentik/blueprints/treepage.yaml` and `config.authentik.yml` to match.

## Related

- [First login](first-login.md)
- [OIDC providers](../admin/oidc.md)
