# Books (AI compilations)

**Books** are structured compilations of documentation from a space's page tree. Chapter text comes from real documents (no hallucinations). When LLM is connected, AI improves structure and writes the introduction.

> Creating and building books requires **editor** role or higher.

## Where to find them

- In a space: **Books** tab in the sidebar
- URL: `/spaces/{slug}/books`

## Audience types

| Type | Description |
|------|-------------|
| Developer reference | Developer reference with AI structure |
| Architecture overview | Flowchart + Mermaid diagrams |
| Operations runbook | Ops runbook |
| Onboarding (brief) | Short introduction for new team members |

## Creating a book

1. Open `/spaces/{slug}/books`
2. Fill in the **Create book from domain** form:
   - **Focus** (optional) — topic or document subtree
   - **Audience** — book type
3. Click **Create book**
4. Click **Build book** to generate chapters

## Build process

```
1. TreePage selects relevant documents from the space
2. Forms chapter structure
3. (If LLM connected) AI improves structure and writes introduction
4. Chapter text — verbatim from documentation
5. Book saved to DB with "ready" status
```

### Statuses

| Status | Description |
|--------|-------------|
| draft | Created, not built |
| building | Generation in progress |
| ready | Can read and download |
| error | Build failed |

## Reading a book

URL: `/spaces/{slug}/books/{book-slug}`

- Chapter navigation
- Download as Markdown
- "Sources changed" indicator — if documents were updated after build

## Rebuild

- **Rebuild** — refresh chapters from current documents
- **Regenerate** — full regeneration with LLM

## LLM

| State | Behavior |
|-------|----------|
| LLM connected | AI structure + introduction for developer/architect books |
| LLM not configured | Build works; introduction from book description |

### LLM configuration (administrator)

Environment variables for `backend-server`:

```bash
LLM_ENABLED=true
LLM_API_URL=https://api.openai.com/v1   # or Ollama, vLLM, etc.
LLM_API_KEY=sk-...
LLM_MODEL=gpt-4o-mini
```

## Book auto-translation

When **Document auto-translation** is enabled (System settings), books are also translated to the interface language.

## Permissions

| Role | Create | Build | Read | Delete |
|------|:------:|:-----:|:----:|:------:|
| viewer | ❌ | ❌ | ✅ | ❌ |
| editor | ✅ | ✅ | ✅ | ✅ |
| admin | ✅ | ✅ | ✅ | ✅ |

## Related sections

- [Spaces](spaces.md)
- [System settings](../admin/settings.md)
