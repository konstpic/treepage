package ragindex

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"strings"

	"github.com/konstpic/treepage/backend/pkg/embeddings"
	"github.com/konstpic/treepage/backend/pkg/models"
	"gorm.io/gorm"
)

type Embedder interface {
	Available() bool
	Embed(ctx context.Context, text string) (embeddings.Vector, error)
}

func IndexDocument(ctx context.Context, db *gorm.DB, doc *models.Document) error {
	return IndexDocumentWithEmbedder(ctx, db, doc, nil)
}

func IndexDocumentWithEmbedder(ctx context.Context, db *gorm.DB, doc *models.Document, embedder Embedder) error {
	hash := contentHash(doc.Content)
	if err := db.WithContext(ctx).Where("document_id = ?", doc.ID).Delete(&models.DocumentChunk{}).Error; err != nil {
		return err
	}
	chunks := splitChunks(doc.Content, 1200)
	for i, chunk := range chunks {
		if strings.TrimSpace(chunk) == "" {
			continue
		}
		row := models.DocumentChunk{
			DocumentID: doc.ID, ChunkIndex: i, Content: chunk, ContentHash: hash,
		}
		if embedder != nil && embedder.Available() {
			if emb, err := embedder.Embed(ctx, chunk); err == nil && len(emb) > 0 {
				row.Embedding = emb
			}
		}
		if err := db.WithContext(ctx).Create(&row).Error; err != nil {
			return err
		}
		if len(row.Embedding) > 0 {
			_ = embeddings.SetChunkVector(ctx, db, row.ID, row.Embedding)
		}
	}
	return nil
}

func splitChunks(content string, maxLen int) []string {
	paras := strings.Split(content, "\n\n")
	var chunks []string
	var buf strings.Builder
	for _, p := range paras {
		p = strings.TrimSpace(p)
		if p == "" {
			continue
		}
		if buf.Len()+len(p)+2 > maxLen && buf.Len() > 0 {
			chunks = append(chunks, buf.String())
			buf.Reset()
		}
		if buf.Len() > 0 {
			buf.WriteString("\n\n")
		}
		buf.WriteString(p)
	}
	if buf.Len() > 0 {
		chunks = append(chunks, buf.String())
	}
	if len(chunks) == 0 && strings.TrimSpace(content) != "" {
		return []string{content}
	}
	return chunks
}

func contentHash(content string) string {
	sum := sha256.Sum256([]byte(content))
	return hex.EncodeToString(sum[:])
}
