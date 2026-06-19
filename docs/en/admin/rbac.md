# Roles and permissions (RBAC)

TreePage uses a two-level access model: **system roles** and **space roles**.

## System roles

Assigned globally via **Users** (`/admin/users`).

| Role | Description |
|------|-------------|
| `super_admin` | Full platform control |
| `admin` | Manage spaces, repositories, groups |
| `editor` | Edit documents, sync, books |
| `viewer` | Read only |

### System permissions matrix

| Action | super_admin | admin | editor | viewer |
|--------|:-----------:|:-----:|:------:|:------:|
| System settings (write) | ✅ | ❌ | ❌ | ❌ |
| OIDC providers | ✅ | ❌ | ❌ | ❌ |
| User CRUD | ✅ | partial | ❌ | ❌ |
| Group CRUD | ✅ | ✅ | ❌ | ❌ |
| Space CRUD | ✅ | ✅ | ❌ | ❌ |
| Repository CRUD | ✅ | ✅ | ❌ | ❌ |
| Edit documents | ✅ | ✅ | ✅ | ❌ |
| Manual sync | ✅ | ✅ | ✅ | ❌ |
| AI books | ✅ | ✅ | ✅ | ❌ |
| Read documents | ✅ | ✅ | ✅ | ✅ |
| /admin access | ✅ | ✅ | ❌ | ❌ |

## Space roles

Assigned via **Spaces** → edit → **Members** / **Groups**.

| Role | Description |
|------|-------------|
| `admin` | Manage space members |
| `editor` | Create and edit documents |
| `viewer` | Read only |

### Permission assignment

Two methods:

1. **Direct membership** — user added to space with a role
2. **Via group** — user belongs to a group assigned to the space

Effective role is the maximum of all assignments (direct and via groups).

## Public spaces

A space with the **Public** flag (`is_public: true`):

- Documents are accessible without authorization
- Search on public documents — without login
- Editing — only for authorized users with editor+ role

## OIDC: role mapping

On OIDC login, roles and groups are synchronized from JWT claims:

| Claim | Default | Purpose |
|-------|---------|---------|
| Role claim | `roles` | System roles |
| Group claim | `groups` | User groups |

Configured in **OIDC providers** → `role_claim`, `group_claim`, `sync_groups`.

## admin restrictions when managing users

A user with `admin` role (not super_admin):

- ✅ Can create/edit users with `viewer`, `editor` roles
- ❌ Cannot manage users with `admin`, `super_admin` roles
- ❌ Cannot assign `admin`, `super_admin` roles
- ❌ Cannot create new users (super_admin only)

## Recommendations

1. Minimize the number of `super_admin` users (1–2 people)
2. Use groups for bulk permission assignment
3. Public spaces — only for documentation without secrets
4. Sync groups from OIDC (`sync_groups: true`)

## Related sections

- [Users](users.md)
- [Groups](groups.md)
- [Spaces](spaces.md)
- [OIDC providers](oidc.md)
