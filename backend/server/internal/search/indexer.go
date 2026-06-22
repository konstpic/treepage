package search

import (
	"context"

	"github.com/konstpic/treepage/backend/pkg/models"
)

// DocumentIndexer syncs documents to an external search backend (OpenSearch).
type DocumentIndexer interface {
	IndexDocument(ctx context.Context, doc *models.Document, spaceSlug string) error
	DeleteDocument(ctx context.Context, docID string) error
}
