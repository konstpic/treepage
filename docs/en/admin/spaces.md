# Managing spaces

**URL:** `/admin/spaces`

## Creating a space

| Field | Required | Description |
|-------|:--------:|-------------|
| Slug | ✅ | URL identifier (Latin letters, hyphens). Example: `engineering` |
| Name | ✅ | Display name |
| Description | ❌ | Brief description |
| Public | ❌ | Access without authorization |

After creation, the space is available at `/spaces/{slug}`.

## Editing

Click the pencil icon on the space card.

### Basic fields

- Name, description, public flag

### Linked repositories

In edit mode, Git repositories linked to the space are shown.

- **Unlink** — remove association (repository remains in the system)
- To link a new repository — create it in **Repositories** with this space selected

### Members

| Field | Description |
|-------|-------------|
| User | Select from list |
| Role | `viewer`, `editor`, `admin` |

To add: select user → role → **Add**.

### Groups

| Field | Description |
|-------|-------------|
| Group | Select from group list |
| Role | `viewer`, `editor`, `admin` |

All group members receive the specified role in the space.

## Deletion

Deleting a space removes associated documents from TreePage (not from Git). This operation is irreversible.

## Slug rules

- Latin letters, digits, hyphens only
- Unique in the system
- Used in URL: `/spaces/{slug}`
- Slug cannot be changed after creation — create a new space

## Configuration examples

### Public product documentation

```
Slug: product-docs
Public: ✅
Members: not required for reading
```

### Internal team documentation

```
Slug: platform-team
Public: ❌
Groups: platform-engineers → editor
        platform-leads → admin
```

## Related sections

- [Repositories](repositories.md)
- [RBAC](rbac.md)
- [Spaces (user)](../user/spaces.md)
