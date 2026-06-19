# Spaces

A **Space** is a logical container for a documentation collection. Each space can be linked to one or more Git repositories.

## Viewing spaces

Go to **Spaces** (`/spaces`).

### Without authorization

Only **public** spaces are shown.

### After login

All spaces you have access to are shown (direct or via group).

## Space card

| Field | Description |
|-------|-------------|
| Name | Human-readable name |
| Slug | URL identifier |
| Description | Brief description of contents |
| Access | Public / Private |

## Opening a space

Click a space → opens `/spaces/{slug}`.

### What you see

- Document tree in the sidebar
- List of linked repositories
- Synchronize button (if you have editor+ permissions)
- **Books** tab

## Searching spaces

The `/spaces` page has a search box — filters by name and description.

## Access permissions

| Space type | Who can see |
|------------|-------------|
| Public | Everyone (including unauthenticated users) |
| Private | Members and group members only |

### Space roles

| Role | Capabilities |
|------|-------------|
| `viewer` | Read documents and books |
| `editor` | Read + edit + sync + create books |
| `admin` | All of the above + manage space members |

## If there are no spaces

- **Without login:** "No public spaces yet"
- **After login:** "No spaces yet. Create one in the admin panel"

Contact an administrator to create a space.

## Related sections

- [Reading documents](reading-docs.md)
- [Spaces (admin)](../admin/spaces.md)
