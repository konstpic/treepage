package search

import (
	"context"

	"go.uber.org/zap"
	"gorm.io/gorm"
)

// OpenSearchSearcher is an opt-in backend. Until document indexing is wired,
// queries delegate to PostgreSQL full-text search.
type OpenSearchSearcher struct {
	baseURL  string
	fallback *PostgresSearcher
	logger   *zap.Logger
}

func NewOpenSearchSearcher(baseURL string, db *gorm.DB, logger *zap.Logger) *OpenSearchSearcher {
	return &OpenSearchSearcher{
		baseURL:  baseURL,
		fallback: NewPostgresSearcher(db),
		logger:   logger,
	}
}

func (s *OpenSearchSearcher) Search(ctx context.Context, q Query) ([]Result, int64, error) {
	if s.logger != nil {
		s.logger.Debug("opensearch search delegated to postgres (index sync not yet implemented)")
	}
	return s.fallback.Search(ctx, q)
}
