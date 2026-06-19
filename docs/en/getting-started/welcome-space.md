# Welcome space

On first startup TreePage automatically creates a **public welcome space** with documentation from this repository.

## What is created

| Entity | Value |
|--------|--------|
| Space slug | `welcome` |
| Space name | Welcome |
| Public access | Yes (readable without login) |
| Git repository | [github.com/konstpic/treepage](https://github.com/konstpic/treepage) |
| Docs path | `docs/` (EN + RU guides) |
| Initial sync | Triggered once after first bootstrap |

Open: **http://localhost:5173/spaces/welcome** (dev) or `/spaces/welcome` on your domain.

## First login (local admin)

Before configuring anything else, sign in with the bootstrap administrator:

| Field | Value |
|-------|--------|
| Email | `admin@local` |
| Password | `admin` |
| Role | `super_admin` |

> Available when `DEV_MODE=true` (Docker Compose default). Change the password after first login.

Steps:

1. Start TreePage (`docker compose up --build`)
2. Open `/auth` and sign in with `admin@local` / `admin`
3. Open **Spaces** → **Welcome** (or go to `/spaces/welcome`)
4. If docs are empty — click **Sync now** (requires the GitHub repo to contain the `docs/` folder)

## Environment variables

| Variable | Default | Description |
|----------|---------|-------------|
| `WELCOME_SPACE_ENABLED` | `true` | Set to `false` to skip bootstrap |
| `WELCOME_REPO_URL` | `https://github.com/konstpic/treepage.git` | Source repository |
| `WELCOME_REPO_BRANCH` | `main` | Branch to clone |
| `WELCOME_DOCS_PATH` | `docs` | Subdirectory with Markdown |

## Production

- Set `DEV_MODE=false` and configure OIDC for admin access
- Welcome space remains **public** — suitable for product documentation
- Point `WELCOME_REPO_URL` to your fork or internal mirror if needed
- Disable with `WELCOME_SPACE_ENABLED=false` if you manage spaces manually

## Related

- [First login](first-login.md)
- [Initial setup](initial-setup.md)
- [Spaces (admin)](../admin/spaces.md)
