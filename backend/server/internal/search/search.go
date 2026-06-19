package search

import (
	"context"
	"fmt"
	"strings"

	"github.com/konstpic/treepage/backend/pkg/fts"
	"github.com/konstpic/treepage/backend/pkg/models"
	"gorm.io/gorm"
)

type Query struct {
	Text            string
	SpaceID         string
	AllowedSpaceIDs []string // nil = no filter (super admin)
	Repository      string
	Author          string
	Tags            []string
	Limit           int
	Offset          int
}

type Result struct {
	ID         string   `json:"id"`
	SpaceID    string   `json:"space_id"`
	SpaceSlug  string   `json:"space_slug"`
	Title      string   `json:"title"`
	Slug       string   `json:"slug"`
	Path       string   `json:"path"`
	Snippet    string   `json:"snippet"`
	Tags       []string `json:"tags"`
	AuthorName string   `json:"author_name"`
	Rank       float64  `json:"rank"`
}

// Searcher abstracts full-text search (PostgreSQL now, OpenSearch later).
type Searcher interface {
	Search(ctx context.Context, q Query) ([]Result, int64, error)
}

type PostgresSearcher struct {
	db *gorm.DB
}

func NewPostgresSearcher(db *gorm.DB) *PostgresSearcher {
	return &PostgresSearcher{db: db}
}

func (s *PostgresSearcher) Search(ctx context.Context, q Query) ([]Result, int64, error) {
	limit := q.Limit
	if limit <= 0 {
		limit = 20
	}
	if limit > 100 {
		limit = 100
	}

	if q.AllowedSpaceIDs != nil && len(q.AllowedSpaceIDs) == 0 {
		return []Result{}, 0, nil
	}

	base := s.db.WithContext(ctx).Model(&models.Document{}).
		Joins("JOIN spaces ON spaces.id = documents.space_id").
		Where("documents.is_published = ?", true)
	if q.SpaceID != "" {
		base = base.Where("documents.space_id = ?", q.SpaceID)
	}
	if q.AllowedSpaceIDs != nil {
		base = base.Where("documents.space_id IN ?", q.AllowedSpaceIDs)
	}
	if q.Author != "" {
		base = base.Where("documents.author_name ILIKE ?", "%"+q.Author+"%")
	}
	if len(q.Tags) > 0 {
		base = base.Where("documents.tags && ?", fmt.Sprintf("{%s}", strings.Join(q.Tags, ",")))
	}
	if q.Repository != "" {
		base = base.Joins("JOIN repositories ON repositories.id = documents.repository_id").
			Where("repositories.name ILIKE ?", "%"+q.Repository+"%")
	}

	tsQuery := strings.TrimSpace(q.Text)
	vectorMatchSQL := `(
  documents.search_vector @@ plainto_tsquery('english', ?)
  OR documents.search_vector @@ plainto_tsquery('russian', ?)
  OR documents.search_vector @@ plainto_tsquery('simple', ?)
)`
	vectorRankSQL := `GREATEST(
  ts_rank(documents.search_vector, plainto_tsquery('english', ?)),
  ts_rank(documents.search_vector, plainto_tsquery('russian', ?)),
  ts_rank(documents.search_vector, plainto_tsquery('simple', ?))
)`

	var total int64
	countQuery := base
	if tsQuery != "" {
		countQuery = countQuery.Where(vectorMatchSQL, fts.QueryArgs(tsQuery)...)
	}
	if err := countQuery.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	type row struct {
		models.Document
		SpaceSlug string  `gorm:"column:space_slug"`
		Rank      float64 `gorm:"column:rank"`
	}

	var rows []row
	query := base
	if tsQuery != "" {
		query = query.
			Select("documents.*, spaces.slug as space_slug, "+vectorRankSQL+" as rank", fts.RankArgs(tsQuery)...).
			Where(vectorMatchSQL, fts.QueryArgs(tsQuery)...).
			Order("rank DESC")
	} else {
		query = query.
			Select("documents.*, spaces.slug as space_slug, 0 as rank").
			Order("documents.updated_at DESC")
	}
	if err := query.Limit(limit).Offset(q.Offset).Find(&rows).Error; err != nil {
		return nil, 0, err
	}

	results := make([]Result, len(rows))
	for i, r := range rows {
		snippet := r.Content
		if len(snippet) > 200 {
			snippet = snippet[:200] + "..."
		}
		results[i] = Result{
			ID: r.ID, SpaceID: r.SpaceID, SpaceSlug: r.SpaceSlug,
			Title: r.Title, Slug: r.Slug, Path: r.Path, Snippet: snippet,
			Tags: r.Tags, AuthorName: r.AuthorName, Rank: r.Rank,
		}
	}
	return results, total, nil
}
