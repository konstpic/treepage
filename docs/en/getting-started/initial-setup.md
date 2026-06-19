# Initial setup

Step-by-step guide after first login as local administrator `admin@local` / `admin`.

> The welcome space (`/spaces/welcome`) is created automatically — see [Welcome space](welcome-space.md).

## Step 0. Welcome space

Already available after first startup:

- **Space:** Welcome (`/spaces/welcome`)
- **Source:** [github.com/konstpic/treepage](https://github.com/konstpic/treepage) → `docs/` folder
- **Access:** public (readable without login)

If documents are missing — click **Sync now** in the sidebar (requires a GitHub push with the `docs/` folder).

---

## Step 1. System settings

Go to: **Settings** → **System settings** (`/admin/settings`)

### Interface language

Choose **English** or **Русский**. Applies globally for all users.

### Theme

| Theme | Description |
|-------|-------------|
| Fox White | Default light theme |
| Coral Night | Dark with accents |
| Light | Classic light |
| Dark | Classic dark |

### Authentication

| Parameter | Recommendation (production) |
|-----------|----------------------------|
| Enable OIDC | ✅ Yes |
| Allow local login | ❌ No (dev only) |

### Git integration (global defaults)

| Parameter | Default | Description |
|-----------|---------|-------------|
| Global token reference | `GIT_ACCESS_TOKEN` | Env variable name for Git token |
| Webhook secret ref | `GIT_WEBHOOK_SECRET` | Env variable name for webhook secret |
| Sync interval | `300` sec | Default for new repositories |
| Sync mode | `scheduled` | manual / scheduled / webhook |

### Platform

| Parameter | Default |
|-----------|---------|
| Search limit | 20 |
| Max search limit | 100 |
| Cache | disabled |
| Log level | info |
| Document auto-translation | disabled (requires LLM) |

Click **Save settings**.

---

## Step 2. OIDC configuration (production)

Go to: **Settings** → **OIDC providers** (`/admin/oidc`)

> Available only to `super_admin`.

### Creating a provider

| Field | Example |
|-------|---------|
| Name | `Keycloak` |
| Issuer URL | `https://keycloak.example.com/realms/treepage` |
| Client ID | `treepage` |
| Redirect URL | `https://docs.example.com/api/auth/callback` |
| Scopes | `openid profile email` |
| Role claim | `roles` |
| Group claim | `groups` |
| Sync groups | ✅ |

### Provider-side configuration

1. Create a **confidential** or **public** client (with PKCE)
2. Set redirect URI: `https://<your-domain>/api/auth/callback`
3. Add scopes: `openid`, `profile`, `email`
4. Store client secret in Kubernetes Secret (`oidc-client-secret`)

Details: [OIDC providers](../admin/oidc.md)

---

## Step 3. Create a space

Go to: **Settings** → **Spaces** (`/admin/spaces`)

1. Fill in the form:
   - **Slug** — URL identifier (e.g. `engineering`)
   - **Name** — display name (e.g. `Engineering Docs`)
   - **Description** — brief description (optional)
   - **Public** — enable if documentation should be accessible without login
2. Click **Create**

Details: [Spaces (admin)](../admin/spaces.md)

---

## Step 4. Connect a Git repository

Go to: **Settings** → **Repositories** (`/admin/repositories`)

1. Click **Add repository**
2. Fill in:

| Field | Example |
|-------|---------|
| Space | `Engineering Docs` |
| Name | `Main Docs Repo` |
| URL | `https://github.com/org/docs.git` |
| Branch | `main` |
| Docs path | `docs` |
| Sync mode | `scheduled` |
| Interval (sec) | `300` |
| Access token | `GIT_ACCESS_TOKEN` (or literal token) |
| Enabled | ✅ |

3. Save

### Repository structure

TreePage scans `{docs_path}/**/*.md`. Example:

```
docs/
├── README.md           → slug: readme
├── guides/
│   ├── installation.md → slug: guides/installation
│   └── first-steps.md  → slug: guides/first-steps
└── api/
    └── reference.md    → slug: api/reference
```

Page title is taken from the first `# H1` or from the filename.

Optional tags via frontmatter:

```markdown
tags: kubernetes, helm, deployment

# Document title
```

Details: [Repositories](../admin/repositories.md), [Git Sync](../admin/git-sync.md)

---

## Step 5. First synchronization

### From the admin panel

1. **Settings** → **Repositories**
2. Find the repository → click **Synchronize** (refresh icon)

### From the space interface

1. Open `/spaces/<slug>`
2. In the sidebar, click **Synchronize**

### Verification

- Sync status: `success` / `completed`
- Documents appear in the tree on the left
- On error — check `last_sync_error` in the repository list

---

## Step 6. Configure users and access

### Creating users

**Settings** → **Users** (`/admin/users`)

Create accounts or wait for automatic creation via OIDC.

### Groups

**Settings** → **Groups** (`/admin/groups`)

Create groups (e.g. `developers`, `ops-team`) for bulk permission assignment.

### Space permissions

**Settings** → **Spaces** → edit space:

- **Members** — assign users with role `viewer`, `editor`, or `admin`
- **Groups** — assign groups with a role

Details: [RBAC](../admin/rbac.md), [Users](../admin/users.md), [Groups](../admin/groups.md)

---

## Step 7. Verification

1. Open `/spaces/<slug>` — documents are displayed
2. Try **Search** (`/search?q=...`)
3. Sign in as a user with `viewer` role — verify access
4. Sign in as `editor` — verify editing

## Done!

TreePage is configured. Share the link to `/spaces` and this guide with users.

## See also

- [LLM for AI books](../user/books.md) — configure `LLM_ENABLED` on server
- [Webhook sync](../admin/git-sync.md) — instant sync on push
- [Monitoring](../operator/monitoring.md) — health checks and metrics
