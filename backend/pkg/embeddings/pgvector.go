package embeddings

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"gorm.io/gorm"
)

const defaultVectorDims = 768

// PgVectorAvailable reports whether the pgvector extension is installed.
func PgVectorAvailable(db *gorm.DB) bool {
	var ok bool
	err := db.Raw(`SELECT EXISTS(SELECT 1 FROM pg_extension WHERE extname = 'vector')`).Scan(&ok).Error
	return err == nil && ok
}

// VectorLiteral formats a vector for PostgreSQL pgvector casts.
func VectorLiteral(v Vector) string {
	if len(v) == 0 {
		return "[]"
	}
	parts := make([]string, len(v))
	for i, x := range v {
		parts[i] = fmt.Sprintf("%g", x)
	}
	return "[" + strings.Join(parts, ",") + "]"
}

// SetChunkVector stores embedding in the pgvector column when available.
func SetChunkVector(ctx context.Context, db *gorm.DB, chunkID string, v Vector) error {
	if len(v) == 0 || chunkID == "" || !PgVectorAvailable(db) {
		return nil
	}
	lit := VectorLiteral(v)
	return db.WithContext(ctx).Exec(
		`UPDATE document_chunks SET embedding_vector = $1::vector WHERE id = $2`,
		lit, chunkID,
	).Error
}

// BackfillVectorsFromJSONB copies JSONB embeddings into embedding_vector in batches.
func BackfillVectorsFromJSONB(ctx context.Context, db *gorm.DB, limit int) (int, error) {
	if !PgVectorAvailable(db) || limit <= 0 {
		return 0, nil
	}
	type row struct {
		ID        string
		Embedding []byte
	}
	var rows []row
	err := db.WithContext(ctx).Raw(`
		SELECT id, embedding
		FROM document_chunks
		WHERE embedding IS NOT NULL
		  AND embedding_vector IS NULL
		ORDER BY document_id, chunk_index
		LIMIT ?`, limit).Scan(&rows).Error
	if err != nil {
		return 0, err
	}
	n := 0
	for _, r := range rows {
		var v Vector
		if err := json.Unmarshal(r.Embedding, &v); err != nil || len(v) == 0 {
			continue
		}
		if err := SetChunkVector(ctx, db, r.ID, v); err != nil {
			return n, err
		}
		n++
	}
	return n, nil
}
