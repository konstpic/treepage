# Git repositories

**URL:** `/admin/repositories`

Repositories are sources of Markdown documentation for spaces.

## Creating a repository

| Field | Required | Default | Description |
|-------|:--------:|---------|-------------|
| Space | ✅ | — | Space the repository is linked to |
| Name | ✅ | — | Human-readable name |
| URL | ✅ | — | HTTPS Git repository URL |
| Branch | ✅ | `main` | Branch to clone |
| Docs path | ✅ | `docs` | Subdirectory with `.md` files |
| Sync mode | ✅ | `manual` | manual / scheduled / webhook |
| Interval (sec) | ❌ | `300` | For scheduled mode |
| Access token | ❌ | — | Env ref or literal token |
| Webhook secret | ❌ | — | Env ref or literal secret |
| Enabled | ❌ | ✅ | Participates in scheduled sync |

### Supported Git platforms

GitHub, GitLab, Gitea, Bitbucket, and any Git server with HTTPS.

### URL format

```
https://github.com/org/repo.git
https://gitlab.example.com/group/project.git
https://gitea.example.com/user/docs.git
```

## Editing

Click the pencil icon → change fields → save.

## Manual synchronization

Click the **refresh** icon (RefreshCw) on the repository card.

Status is shown in the sync column:

| Status | Meaning |
|--------|---------|
| success / completed | Synchronization successful |
| failed | Error (see last_sync_error) |
| — | Not synchronized yet |

## Deletion

Deleting a repository does not remove documents from Git — only the sync configuration in TreePage.

## Access tokens

### Via env ref (recommended)

Specify the environment variable name:

```
GIT_ACCESS_TOKEN
```

Value is set in Kubernetes Secret or `.env`.

### Literal token

You can specify the token directly (not recommended for production):

```
ghp_xxxxxxxxxxxxxxxxxxxx
```

### Creating a token

| Platform | Minimum permissions |
|----------|---------------------|
| GitHub | `repo` (read) for private repos |
| GitLab | `read_repository` |
| Gitea | `read` access to repository |

## docs_path structure

TreePage scans `{clone}/{docs_path}/**/*.md`:

```
repo/
└── docs/                    ← docs_path = "docs"
    ├── index.md             → slug: index
    ├── getting-started/
    │   └── install.md       → slug: getting-started/install
    └── api/
        └── reference.md     → slug: api/reference
```

### Metadata from Markdown

| Source | Field |
|--------|-------|
| First `# H1` | Title |
| Filename | Slug (if no H1) |
| `tags: tag1, tag2` (first line) | Tags for search |

## Linking to another space

Via **Spaces** → edit → unlink/link.

Or API: `POST /api/admin/spaces/{id}/bind-repo`

## Related sections

- [Git Sync](git-sync.md)
- [Secrets](../operator/secrets.md)
