# Search

TreePage supports **full-text search** and **RAG** — LLM-powered answers over your documentation.

Page: `/search`

---

## Standard search

### How to search

1. Open **Search** in the navigation
2. Enter a query on the search tab
3. Click **Search** or press Enter

### Direct link

```
/search?q=kubernetes&space=engineering&tags=helm,deployment
```

### Search fields

| Field | Description |
|-------|-------------|
| Title | H1 and document title |
| Content | Markdown text |
| Tags | Frontmatter `tags:` |
| Repository | Git repository name |
| Author | Last change author |

### Multilingual FTS (Search & RAG phase)

Search works in **Russian and English** (PostgreSQL `russian`, `english`, `simple` configs).

### Filters

| Filter | Description |
|--------|-------------|
| Space | Limit to one space |
| Author | Author email |
| Tags | Comma-separated |

### Limits

| Parameter | Default |
|-----------|---------|
| Result limit | 20 |
| Maximum | 100 |

Configured in **System settings** → **Platform**.

---

## RAG — Ask documentation

> Requires LLM on `backend-server` (`LLM_ENABLED=true`). For hybrid search — embeddings (`EMBEDDING_ENABLED=true`).

### How to use

1. Go to `/search`
2. Enter a natural-language question (RU or EN)
3. Click **Ask documentation**

### What you get

| Element | Description |
|---------|-------------|
| **Answer** | Concise text in the question language, based on retrieved excerpts only |
| **Citations** | Exact quotes from documents with links |
| **Sources** | Document list; the first is marked as best match |
| **Confidence** | Low confidence → suggested follow-up questions |
| **👍 / 👎** | Feedback improves search (learned synonyms) |

### How it works (brief)

```
Question → FTS + keywords + vector (if embeddings)
         → RBAC + Page ACL filter
         → LLM generates answer
         → sources + citations
```

Details: [Roadmap — Search & RAG](../reference/roadmap.md).

### API

```http
POST /api/rag/ask
Content-Type: application/json

{ "question": "How do I deploy locally?" }
```

```http
POST /api/rag/feedback
{ "question": "...", "helpful": true, "answer": "...", "sources": [...] }
```

---

## Access

- **Public** spaces — search without login
- **Private** — after authorization, subject to RBAC and Page ACL (phase 3)
- RAG uses the same access rules as standard search

---

## Tips

- Add `tags:` in frontmatter for better discoverability
- For RAG, ask specific questions: “How do I configure OIDC?” rather than single words
- After `git pull` and sync, wait for chunk indexing (automatic on sync)

---

## Related sections

- [Roadmap](../reference/roadmap.md)
- [LLM configuration](../operator/configuration.md)
- [Troubleshooting — RAG](../operator/troubleshooting.md)
- [Spaces](spaces.md)
