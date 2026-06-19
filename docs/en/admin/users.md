# User management

**URL:** `/admin/users`

> User creation — `super_admin` only. Editing viewer/editor — `admin` and `super_admin`.

## User list

Shows all users with:

- Email
- Display name
- Roles
- Status (active / inactive)

## Creating a user (super_admin)

| Field | Required | Description |
|-------|:--------:|-------------|
| Email | ✅ | Unique email |
| Password | ✅ | For local login |
| Display name | ❌ | Name in UI |
| Roles | ✅ | One or more |
| Active account | ❌ | ✅ by default |

### Available roles

| Role | Description |
|------|-------------|
| super_admin | Full control |
| admin | Space management |
| editor | Editing |
| viewer | Read only |

## Editing

Click **Edit** on the user card.

| Field | Description |
|-------|-------------|
| Email | Change email |
| New password | Leave empty to keep unchanged |
| Display name | Name in UI |
| Roles | Role assignment |
| Active account | Deactivation blocks login |

## Restrictions for admin

A user with `admin` role (not super_admin):

| Action | Allowed |
|--------|---------|
| View all users | ✅ |
| Create users | ❌ |
| Edit viewer/editor | ✅ |
| Edit admin/super_admin | ❌ |
| Assign admin/super_admin | ❌ |
| Delete admin/super_admin | ❌ |

## Deletion

- Cannot delete your own account
- Cannot delete the last super_admin
- Admin cannot delete admin/super_admin

## OIDC provisioning

On OIDC login, users are created automatically:

1. Email from `email` claim
2. Display name from `name` or `preferred_username`
3. Roles from `roles` claim (configurable)
4. Groups from `groups` claim (if `sync_groups: true`)

Manual creation is needed only for local login (dev) or service accounts.

## Default dev account

When `DEV_MODE=true`, bootstrap creates:

| Email | Password | Role |
|-------|----------|------|
| admin@local | admin | super_admin |

> Change the password after first login.

## Related sections

- [RBAC](rbac.md)
- [OIDC providers](oidc.md)
- [Groups](groups.md)
