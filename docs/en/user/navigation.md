# Interface navigation

## Home page (`/`)

Landing page describing TreePage features. The **Open spaces** button leads to the space list.

## Top navigation

| Item | URL | Description |
|------|-----|-------------|
| Spaces | `/spaces` | List of accessible spaces |
| Search | `/search` | Full-text search |
| Settings | `/admin/*` | Admin panel (admin/super_admin only) |
| Sign in / Account | `/auth` | Login and logout |

## Space page (`/spaces/:slug`)

```
┌─────────────────────────────────────────────────────┐
│  ← Back to spaces    Space name                     │
├──────────────┬──────────────────────────────────────┤
│  Pages       │  Document content                    │
│  Books       │  (Markdown + Mermaid)                │
│              │                                      │
│  📄 doc-1    │  Breadcrumbs: Space > Folder > Page  │
│  📁 folder/  │                                      │
│    📄 doc-2  │                                      │
│              │                                      │
│  [Sync]      │                                      │
└──────────────┴──────────────────────────────────────┘
```

### Sidebar

- **Pages** — document tree (default tab)
- **Books** — saved AI compilations
- **Synchronize** — manual sync trigger (editors+)

### Space display modes

On the `/spaces` page, three modes are available:

| Mode | Description |
|------|-------------|
| Tile | Space cards |
| Table | Tabular view with slug and access |
| List | Compact list |

## Document page (`/spaces/:slug/docs/:docSlug`)

- **Breadcrumbs** — navigation trail
- **Edit** — switch to editor mode (editors+)
- **Version history** — view previous versions and diff

## Search (`/search`)

Query string + filters by space, author, tags.

## Admin panel (`/admin`)

Available to users with `admin` or `super_admin` roles:

| Section | URL |
|---------|-----|
| Spaces | `/admin/spaces` |
| Repositories | `/admin/repositories` |
| Users | `/admin/users` |
| Groups | `/admin/groups` |
| System settings | `/admin/settings` |
| OIDC providers | `/admin/oidc` |

## Themes

The administrator sets the global theme. Available options: Fox White, Coral Night, Light, Dark.

## Quick links

- Document: `/spaces/{slug}/docs/{doc-slug}`
- Book: `/spaces/{slug}/books/{book-slug}`
- Search with query: `/search?q=kubernetes`
