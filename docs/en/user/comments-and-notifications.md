# Comments and notifications

> Requires sign-in. Comments are available on every document page for authenticated users.

## Comments sidebar

On the document page (`/spaces/{slug}/docs/{doc-slug}`), a **Comments** column appears on the right (on wide screens). On smaller screens it stacks below the document body.

| Action | How |
|--------|-----|
| Read thread | Scroll the comments panel |
| Post a comment | Type in the field at the bottom → **Post comment** |
| Reply | (flat thread today — use `@mention` to address someone) |

Comments are stored in TreePage (PostgreSQL), not in Git. They do not affect repository sync.

## @mentions

To notify a colleague:

1. In the comment field, type **`@`**
2. A dropdown lists active users (filtered as you type)
3. Select a user or press Enter — their **email** is inserted (`@user@example.com`)
4. Post the comment

### What the mentioned user gets

- An in-app **notification** in the bell icon (🔔)
- A **deep link** to the document and the specific comment
- **Read access** to that document even if they are not a member of a private space (mention acts as a one-off grant for that page)

Supported email formats include corporate domains (`user@company.com`) and local dev accounts (`admin@local`).

## Notifications

The **bell** in the top navigation shows recent events.

| Type | Description |
|------|-------------|
| `comment.mention` | Someone mentioned you in a comment |
| Space editor alerts | New documents, publishes, workflow changes (editors in the space) |

### Actions

| Action | Description |
|--------|-------------|
| Open notification | Click the row — for mentions, opens the document and scrolls to the comment (highlighted briefly) |
| Mark read | ✓ on an unread item |
| Mark all read | Link in the panel header |

Favorites and full recent history: **My pages** (`/me`).

## Direct links to comments

Share or bookmark a comment anchor:

```
/spaces/welcome/docs/getting-started#comment-{uuid}
```

The UI scrolls to the comment and highlights it when the hash is present.

## Related sections

- [Reading documents](reading-docs.md)
- [Navigation](navigation.md)
- [Editing documents](editing-docs.md)
