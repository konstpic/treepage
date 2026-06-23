# RAG indexing

**URL:** `/admin/rag`

> Requires `admin` or `super_admin`.

## Purpose

RAG (Retrieval-Augmented Generation) indexes document text into chunks and embeddings so signed-in users can **Ask documentation** on the Search page (`/search` → AI tab).

## Dashboard

| Metric | Description |
|--------|-------------|
| Documents with chunks | Pages split for indexing |
| Text chunks | Total chunks in `document_chunks` |
| Chunks with embeddings | Ready for semantic search |
| Chunks without embeddings | Pending embedding worker |

Use **Reindex all** to rebuild the index after bulk imports or LLM configuration changes.

## Requirements

- `LLM_ENABLED=true` on `backend-server` for embeddings and `/api/rag/ask`
- Documents must be visible to the user asking the question (same ACL as search)

## Related sections

- [Search (user)](../user/search.md)
- [System settings](settings.md)
- [REST API — RAG](../reference/api.md)
