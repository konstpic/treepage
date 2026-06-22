package handler

import (
	"context"
	"time"

	"github.com/konstpic/treepage/backend/pkg/models"
	"github.com/konstpic/treepage/backend/server/internal/search"
)

func (h *Handler) indexDocumentAsync(doc *models.Document) {
	if h.docIndexer == nil || doc == nil {
		return
	}
	go func(d models.Document) {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
		space, err := h.spaces.GetByID(ctx, d.SpaceID)
		if err != nil {
			return
		}
		_ = h.docIndexer.IndexDocument(ctx, &d, space.Slug)
	}(*doc)
}

func (h *Handler) deleteDocumentFromIndex(docID string) {
	if h.docIndexer == nil {
		return
	}
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
		defer cancel()
		_ = h.docIndexer.DeleteDocument(ctx, docID)
	}()
}

func pickDocumentIndexer(searcher search.Searcher) search.DocumentIndexer {
	if idx, ok := searcher.(search.DocumentIndexer); ok {
		return idx
	}
	return nil
}
