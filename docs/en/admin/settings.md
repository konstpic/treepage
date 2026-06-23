# System settings

**URL:** `/admin/settings`

> Saving settings — `super_admin` only. Users with `admin` role see settings in read-only mode.

## Interface language

Global UI language for all users.

| Language | ID |
|----------|-----|
| English | `en` |
| Русский | `ru` |

Applied immediately on selection (auto-save). All **admin panel** screens (spaces, repositories, users, RAG, OIDC, etc.) use the same locale.

## Contextual help

On **System settings**, **RAG indexing**, and other admin cards, the **?** icon next to the section title shows a short hint about that block.

## Appearance (theme)

| Theme | ID | Description |
|-------|-----|-------------|
| Fox White | `fox_white` | Default light theme |
| Coral Night | `coral_night` | Dark with accents |
| Light | `light` | Classic light |
| Dark | `dark` | Classic dark |

## Document auto-translation

| Parameter | Description |
|-----------|-------------|
| Auto-translate documents to interface language (LLM) | Translates documents and books to UI language |

**Requirements:** LLM must be configured (`LLM_ENABLED=true` on backend-server).

When disabled, all users see original text.

## Authentication

| Parameter | Production | Dev |
|-----------|-----------|-----|
| Enable OIDC | ✅ | ✅ |
| Allow local login | ❌ | ✅ |

Local login works only when `DEV_MODE=true` on backend-auth.

## Git integration

| Parameter | Default | Description |
|-----------|---------|-------------|
| Global token reference | `GIT_ACCESS_TOKEN` | Env var for Git token |
| Webhook secret ref | `GIT_WEBHOOK_SECRET` | Env var for webhook secret |
| Sync interval (sec) | `300` | Default for new repositories |
| Sync mode | `scheduled` | Default sync mode |

## Platform

| Parameter | Default | Description |
|-----------|---------|-------------|
| Default search limit | `20` | Results per page |
| Max search limit | `100` | Upper bound |
| Enable cache | `false` | Redis/in-memory cache |
| Log level | `info` | debug / info / warn / error |

## Saving

Click **Save settings** for Authentication, Git, and Platform sections.

Language, theme, and auto-translation save automatically on change.

## Helm vs UI

Some settings are duplicated in Helm values and UI:

| Setting | Helm | UI |
|---------|------|-----|
| OIDC issuer/client | `auth.oidc.*` | OIDC Providers |
| JWT TTL | `auth.jwt.*` | — |
| Search limits | `server.search.*` | Platform settings |
| Audit log | `server.security.enableAuditLog` | — |
| LLM | `server.extraEnv` | — |

Helm settings apply at deploy time. UI settings — at runtime via PostgreSQL (`system_settings`).

## Related sections

- [OIDC providers](oidc.md)
- [Configuration](../operator/configuration.md)
- [Books (LLM)](../user/books.md)
