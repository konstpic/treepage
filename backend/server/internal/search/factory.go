package search

import (
	"os"

	"go.uber.org/zap"
	"gorm.io/gorm"
)

// NewFromEnv selects the search backend from SEARCH_BACKEND (postgres|opensearch).
func NewFromEnv(db *gorm.DB, log *zap.Logger) Searcher {
	backend := os.Getenv("SEARCH_BACKEND")
	if backend == "opensearch" {
		url := os.Getenv("OPENSEARCH_URL")
		if url != "" {
			if log != nil {
				log.Info("search backend: opensearch", zap.String("url", url))
			}
			return NewOpenSearchSearcher(url, db, log)
		}
		if log != nil {
			log.Warn("SEARCH_BACKEND=opensearch but OPENSEARCH_URL is empty; using postgres")
		}
	}
	return NewPostgresSearcher(db)
}
