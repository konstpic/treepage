# Search

TreePage supports full-text search across documentation.

## How to search

1. Click **Search** in the navigation or go to `/search`
2. Enter a query
3. Click **Search** or press Enter

### Direct link with query

```
/search?q=kubernetes&space=engineering&tags=helm,deployment
```

## Search fields

| Field | Description |
|-------|-------------|
| Title | H1 and document title |
| Content | Markdown text |
| Tags | Frontmatter `tags:` |
| Repository | Git repository name |
| Author | Last change author |

## Filters

| Filter | Description |
|--------|-------------|
| Space | Limit search to one space |
| Author | Filter by author email |
| Tags | Comma-separated (e.g. `kubernetes, helm`) |

## Results

Each result contains:

- Document title (link)
- Space
- Text snippet with highlighted matches
- Document slug

## Limits

| Parameter | Default value |
|-----------|---------------|
| Result limit | 20 |
| Maximum limit | 100 |

Configured by the administrator in **System settings** → **Platform**.

## Access

- Search in **public** spaces is available without login
- Search in **private** spaces — only after authorization and subject to RBAC

## Tips

- Use tags in Markdown for better discoverability:

  ```markdown
  tags: api, rest, authentication

  # API Reference
  ```

- Short queries (1–2 words) return more results
- Combine a space filter with a text query for precision

## Related sections

- [Spaces](spaces.md)
- [System settings (admin)](../admin/settings.md)
