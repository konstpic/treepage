package rag

import (
	"context"
)

// IndexStats reflects persisted RAG index state in PostgreSQL.
type IndexStats struct {
	PublishedDocuments  int   `json:"published_documents"`
	DocumentsWithChunks int   `json:"documents_with_chunks"`
	ChunksTotal         int64 `json:"chunks_total"`
	ChunksEmbedded      int64 `json:"chunks_embedded"`
	ChunksPending       int64 `json:"chunks_pending"`
	EmbeddingsEnabled   bool  `json:"embeddings_enabled"`
}

func (s *Service) IndexStats(ctx context.Context) (IndexStats, error) {
	var stats IndexStats
	stats.EmbeddingsEnabled = s.embed != nil && s.embed.Available()

	var published int64
	if err := s.db.WithContext(ctx).Table("documents").Where("is_published = ?", true).Count(&published).Error; err != nil {
		return stats, err
	}
	stats.PublishedDocuments = int(published)

	var withChunks int64
	_ = s.db.WithContext(ctx).Raw(`SELECT COUNT(DISTINCT document_id) FROM document_chunks`).Scan(&withChunks).Error
	stats.DocumentsWithChunks = int(withChunks)

	_ = s.db.WithContext(ctx).Table("document_chunks").Count(&stats.ChunksTotal).Error
	_ = s.db.WithContext(ctx).Table("document_chunks").Where("embedding IS NOT NULL").Count(&stats.ChunksEmbedded).Error
	_ = s.db.WithContext(ctx).Table("document_chunks").Where("embedding IS NULL").Count(&stats.ChunksPending).Error
	return stats, nil
}
