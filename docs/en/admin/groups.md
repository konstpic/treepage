# Group management

**URL:** `/admin/groups`

Groups simplify bulk permission assignment in spaces.

## Why groups are useful

```
Group "platform-engineers"
  ├── user1@company.com
  ├── user2@company.com
  └── user3@company.com

Space "Platform Docs"
  └── Group "platform-engineers" → editor role
```

All three users get editor role in the space without individual assignment.

## Creating a group

| Field | Required | Description |
|-------|:--------:|-------------|
| Name | ✅ | Unique group name |
| Description | ❌ | Brief description |

## Managing members

1. Open the group
2. **Add member** — select a user
3. **Remove** — remove user from group

## Assigning a group to a space

**Spaces** → edit → **Groups**:

1. Select group
2. Specify role (`viewer`, `editor`, `admin`)
3. **Add**

## Sync from OIDC

When `sync_groups` is enabled in the OIDC provider:

1. On each login, user groups are synchronized from JWT claim
2. User is automatically added to matching TreePage groups
3. Space permissions apply via assigned groups

### Configuration

**OIDC providers** → `group_claim: groups`, `sync_groups: true`

On the OIDC provider side (Keycloak example):

1. Create groups: `developers`, `ops-team`
2. Add mapper for `groups` claim in access token
3. In TreePage, create groups with the same names
4. Assign groups in spaces

## Deleting a group

Deleting a group does not delete users — only group links to spaces and membership.

## Examples

### Development team

```
Group: backend-team
Space: api-docs → editor
Space: architecture → viewer
```

### DevOps

```
Group: ops-team
Space: runbooks → editor
Space: all-docs → viewer
```

## Related sections

- [RBAC](rbac.md)
- [Spaces](spaces.md)
- [OIDC providers](oidc.md)
