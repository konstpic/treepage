# Administrator guide

The admin panel is available to users with system roles `admin` or `super_admin`.

**URL:** `/admin`

## Panel sections

| Section | URL | Min. role |
|---------|-----|-----------|
| [Spaces](spaces.md) | `/admin/spaces` | admin |
| [Repositories](repositories.md) | `/admin/repositories` | admin |
| [Users](users.md) | `/admin/users` | admin (limited) |
| [Groups](groups.md) | `/admin/groups` | admin |
| [System settings](settings.md) | `/admin/settings` | admin (write: super_admin) |
| [OIDC providers](oidc.md) | `/admin/oidc` | super_admin |
| Analytics | `/admin/analytics` | admin |
| [RAG indexing](rag.md) | `/admin/rag` | admin |
| Audit | `/admin/audit` | super_admin |

## System roles

| Role | Capabilities |
|------|-------------|
| `super_admin` | Full access: OIDC, system settings, all users |
| `admin` | Spaces, repositories, groups, limited user management |
| `editor` | Edit documents, sync, books (no /admin access) |
| `viewer` | Read only |

Details: [RBAC](rbac.md)

## Typical administrator workflow

```
1. Configure OIDC and system parameters
2. Create spaces
3. Connect Git repositories
4. Run synchronization
5. Create groups and assign permissions
6. Invite users (or wait for OIDC provisioning)
```

## Restrictions for admin (not super_admin)

| Action | admin | super_admin |
|--------|:-----:|:-----------:|
| Save system settings | ❌ (read only) | ✅ |
| Manage OIDC | ❌ | ✅ |
| Create users | ❌ | ✅ |
| Assign super_admin/admin | ❌ | ✅ |
| Manage viewer/editor | ✅ | ✅ |
| Spaces and repositories | ✅ | ✅ |

## Audit

When audit log is enabled (`enableAuditLog: true` in Helm), administrator actions are recorded in the `audit_log` table.

## Related sections

- [Initial setup](../getting-started/initial-setup.md)
- [Operations](../operator/README.md)
