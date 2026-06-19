package rag

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"strings"

	"github.com/konstpic/treepage/backend/pkg/models"
	"github.com/konstpic/treepage/backend/server/internal/llm"
	"github.com/konstpic/treepage/backend/server/internal/service"
	"gorm.io/gorm"
)

type Service struct {
	db     *gorm.DB
	llm    *llm.Client
	spaces *service.SpaceService
	pageACL *service.PageACLService
}

func New(db *gorm.DB, llmClient *llm.Client, spaces *service.SpaceService, pageACL *service.PageACLService) *Service {
	return &Service{db: db, llm: llmClient, spaces: spaces, pageACL: pageACL}
}

type Source struct {
	DocumentID string  `json:"document_id"`
	SpaceSlug  string  `json:"space_slug"`
	DocSlug    string  `json:"doc_slug"`
	Title      string  `json:"title"`
	Snippet    string  `json:"snippet"`
	Score      float64 `json:"score"`
}

type Answer struct {
	Answer  string   `json:"answer"`
	Sources []Source `json:"sources"`
}

func (s *Service) ReindexDocument(ctx context.Context, doc *models.Document) error {
	hash := contentHash(doc.Content)
	s.db.WithContext(ctx).Where("document_id = ?", doc.ID).Delete(&models.DocumentChunk{})
	chunks := splitChunks(doc.Content, 1200)
	for i, chunk := range chunks {
		if strings.TrimSpace(chunk) == "" {
			continue
		}
		if err := s.db.WithContext(ctx).Create(&models.DocumentChunk{
			DocumentID: doc.ID, ChunkIndex: i, Content: chunk, ContentHash: hash,
		}).Error; err != nil {
			return err
		}
	}
	return nil
}

func (s *Service) Ask(ctx context.Context, question, userID string, globalRoles []string, limit int) (*Answer, error) {
	question = strings.TrimSpace(question)
	if question == "" {
		return nil, errors.New("question is required")
	}
	if !s.llm.Available() {
		return nil, errors.New("LLM is not configured (set LLM_ENABLED=true and LLM_API_URL)")
	}
	if limit <= 0 || limit > 10 {
		limit = 5
	}
	isSuperAdmin := service.HasRole(globalRoles, "super_admin")
	allowed, err := s.spaces.AccessibleSpaceIDs(ctx, userID, isSuperAdmin)
	if err != nil {
		return nil, err
	}
	if allowed != nil && len(allowed) == 0 {
		return &Answer{Answer: "No accessible documentation found for your account.", Sources: nil}, nil
	}

	type chunkRow struct {
		ChunkID    string
		DocumentID string
		Content    string
		Title      string
		Slug       string
		SpaceSlug  string
		SpaceID    string
		Path       string
		Rank       float64
	}
	var rows []chunkRow
	q := s.db.WithContext(ctx).Table("document_chunks dc").
		Select(`dc.id AS chunk_id, dc.document_id, dc.content, d.title, d.slug, sp.slug AS space_slug,
			d.space_id, d.path,
			ts_rank(to_tsvector('english', dc.content), plainto_tsquery('english', ?)) AS rank`, question).
		Joins("JOIN documents d ON d.id = dc.document_id").
		Joins("JOIN spaces sp ON sp.id = d.space_id").
		Where("d.is_published = ?", true).
		Where("to_tsvector('english', dc.content) @@ plainto_tsquery('english', ?)", question)
	if allowed != nil {
		q = q.Where("d.space_id IN ?", allowed)
	}
	if err := q.Order("rank DESC").Limit(limit * 3).Scan(&rows).Error; err != nil {
		return nil, err
	}

	sources := make([]Source, 0, limit)
	var contextParts []string
	seen := map[string]struct{}{}
	for _, r := range rows {
		spaceRole, _ := s.spaces.EffectiveRole(ctx, r.SpaceID, userID, globalRoles)
		if _, ok := s.pageACL.EffectiveDocumentRole(ctx, r.SpaceID, r.Path, userID, spaceRole, isSuperAdmin); !ok {
			continue
		}
		if _, dup := seen[r.DocumentID]; dup {
			continue
		}
		seen[r.DocumentID] = struct{}{}
		snippet := r.Content
		if len(snippet) > 300 {
			snippet = snippet[:300] + "..."
		}
		sources = append(sources, Source{
			DocumentID: r.DocumentID, SpaceSlug: r.SpaceSlug, DocSlug: r.Slug,
			Title: r.Title, Snippet: snippet, Score: r.Rank,
		})
		contextParts = append(contextParts, fmt.Sprintf("## %s (%s)\n%s", r.Title, r.Path, r.Content))
		if len(sources) >= limit {
			break
		}
	}
	if len(sources) == 0 {
		return &Answer{Answer: "I could not find relevant documentation for your question.", Sources: []Source{}}, nil
	}

	prompt := fmt.Sprintf(`Answer the question using ONLY the documentation excerpts below.
If the answer is not in the excerpts, say you don't know.
Cite document titles when relevant.

Question: %s

Documentation:
%s`, question, strings.Join(contextParts, "\n\n---\n\n"))

	answer, err := s.llm.Chat(ctx,
		"You are a helpful documentation assistant. Be concise and accurate. Use only the provided excerpts.",
		prompt,
	)
	if err != nil {
		return nil, err
	}
	return &Answer{Answer: strings.TrimSpace(answer), Sources: sources}, nil
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
