package search

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/konstpic/treepage/backend/pkg/models"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

const openSearchIndex = "treepage-documents"

type openSearchDocument struct {
	Title      string   `json:"title"`
	Content    string   `json:"content"`
	SpaceID    string   `json:"space_id"`
	SpaceSlug  string   `json:"space_slug"`
	Slug       string   `json:"slug"`
	Path       string   `json:"path"`
	Tags       []string `json:"tags"`
	AuthorName string   `json:"author_name"`
	IsPublished bool    `json:"is_published"`
}

// OpenSearchSearcher indexes and queries documents via OpenSearch HTTP API.
type OpenSearchSearcher struct {
	baseURL  string
	client   *http.Client
	fallback *PostgresSearcher
	logger   *zap.Logger
	db       *gorm.DB
}

func NewOpenSearchSearcher(baseURL string, db *gorm.DB, logger *zap.Logger) *OpenSearchSearcher {
	baseURL = strings.TrimRight(baseURL, "/")
	return &OpenSearchSearcher{
		baseURL:  baseURL,
		client:   &http.Client{Timeout: 15 * time.Second},
		fallback: NewPostgresSearcher(db),
		logger:   logger,
		db:       db,
	}
}

func (s *OpenSearchSearcher) Search(ctx context.Context, q Query) ([]Result, int64, error) {
	results, total, err := s.searchOpenSearch(ctx, q)
	if err != nil {
		if s.logger != nil {
			s.logger.Warn("opensearch query failed, falling back to postgres", zap.Error(err))
		}
		return s.fallback.Search(ctx, q)
	}
	return results, total, nil
}

func (s *OpenSearchSearcher) searchOpenSearch(ctx context.Context, q Query) ([]Result, int64, error) {
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

	filters := []map[string]any{
		{"term": map[string]any{"is_published": true}},
	}
	if q.SpaceID != "" {
		filters = append(filters, map[string]any{"term": map[string]any{"space_id": q.SpaceID}})
	}
	if q.AllowedSpaceIDs != nil {
		filters = append(filters, map[string]any{"terms": map[string]any{"space_id": q.AllowedSpaceIDs}})
	}
	if q.Author != "" {
		filters = append(filters, map[string]any{"wildcard": map[string]any{"author_name": "*" + q.Author + "*"}})
	}
	if len(q.Tags) > 0 {
		filters = append(filters, map[string]any{"terms": map[string]any{"tags": q.Tags}})
	}

	var query map[string]any
	tsQuery := strings.TrimSpace(q.Text)
	if tsQuery != "" {
		query = map[string]any{
			"bool": map[string]any{
				"must": []map[string]any{
					{"multi_match": map[string]any{
						"query":  tsQuery,
						"fields": []string{"title^3", "content", "path", "tags"},
					}},
				},
				"filter": filters,
			},
		}
	} else {
		query = map[string]any{
			"bool": map[string]any{
				"must":   []map[string]any{{"match_all": map[string]any{}}},
				"filter": filters,
			},
		}
	}

	body := map[string]any{
		"query": query,
		"from":  q.Offset,
		"size":  limit,
		"sort":  []map[string]any{{"_score": map[string]string{"order": "desc"}}},
	}
	if tsQuery == "" {
		body["sort"] = []map[string]any{{"title.keyword": map[string]string{"order": "asc"}}}
	}

	var resp struct {
		Hits struct {
			Total struct {
				Value int64 `json:"value"`
			} `json:"total"`
			Hits []struct {
				ID     string  `json:"_id"`
				Score  float64 `json:"_score"`
				Source struct {
					Title      string   `json:"title"`
					Content    string   `json:"content"`
					SpaceID    string   `json:"space_id"`
					SpaceSlug  string   `json:"space_slug"`
					Slug       string   `json:"slug"`
					Path       string   `json:"path"`
					Tags       []string `json:"tags"`
					AuthorName string   `json:"author_name"`
				} `json:"_source"`
			} `json:"hits"`
		} `json:"hits"`
	}
	if err := s.requestJSON(ctx, http.MethodPost, "/"+openSearchIndex+"/_search", body, &resp); err != nil {
		return nil, 0, err
	}

	results := make([]Result, 0, len(resp.Hits.Hits))
	for _, hit := range resp.Hits.Hits {
		snippet := hit.Source.Content
		if len(snippet) > 200 {
			snippet = snippet[:200] + "..."
		}
		results = append(results, Result{
			ID: hit.ID, SpaceID: hit.Source.SpaceID, SpaceSlug: hit.Source.SpaceSlug,
			Title: hit.Source.Title, Slug: hit.Source.Slug, Path: hit.Source.Path,
			Snippet: snippet, Tags: hit.Source.Tags, AuthorName: hit.Source.AuthorName,
			Rank: hit.Score,
		})
	}
	return results, resp.Hits.Total.Value, nil
}

func (s *OpenSearchSearcher) EnsureIndex(ctx context.Context) error {
	var exists struct {
		Exists bool `json:"exists"`
	}
	if err := s.requestJSON(ctx, http.MethodHead, "/"+openSearchIndex, nil, nil); err == nil {
		return nil
	}
	_ = exists
	mapping := map[string]any{
		"settings": map[string]any{
			"number_of_shards":   1,
			"number_of_replicas": 0,
		},
		"mappings": map[string]any{
			"properties": map[string]any{
				"title":        map[string]any{"type": "text"},
				"content":      map[string]any{"type": "text"},
				"space_id":     map[string]any{"type": "keyword"},
				"space_slug":   map[string]any{"type": "keyword"},
				"slug":         map[string]any{"type": "keyword"},
				"path":         map[string]any{"type": "text"},
				"tags":         map[string]any{"type": "keyword"},
				"author_name":  map[string]any{"type": "text"},
				"is_published": map[string]any{"type": "boolean"},
			},
		},
	}
	return s.requestJSON(ctx, http.MethodPut, "/"+openSearchIndex, mapping, nil)
}

func (s *OpenSearchSearcher) IndexDocument(ctx context.Context, doc *models.Document, spaceSlug string) error {
	if err := s.EnsureIndex(ctx); err != nil {
		return err
	}
	payload := openSearchDocument{
		Title: doc.Title, Content: doc.Content, SpaceID: doc.SpaceID, SpaceSlug: spaceSlug,
		Slug: doc.Slug, Path: doc.Path, Tags: []string(doc.Tags), AuthorName: doc.AuthorName,
		IsPublished: doc.IsPublished,
	}
	return s.requestJSON(ctx, http.MethodPut, "/"+openSearchIndex+"/_doc/"+doc.ID, payload, nil)
}

func (s *OpenSearchSearcher) DeleteDocument(ctx context.Context, docID string) error {
	return s.requestJSON(ctx, http.MethodDelete, "/"+openSearchIndex+"/_doc/"+docID, nil, nil)
}

func (s *OpenSearchSearcher) BackfillFromDB(ctx context.Context) (int, error) {
	if err := s.EnsureIndex(ctx); err != nil {
		return 0, err
	}
	type row struct {
		models.Document
		SpaceSlug string `gorm:"column:space_slug"`
	}
	var docs []row
	if err := s.db.WithContext(ctx).
		Table("documents").
		Select("documents.*, spaces.slug AS space_slug").
		Joins("JOIN spaces ON spaces.id = documents.space_id").
		Where("documents.is_published = ?", true).
		Find(&docs).Error; err != nil {
		return 0, err
	}
	n := 0
	for i := range docs {
		if err := s.IndexDocument(ctx, &docs[i].Document, docs[i].SpaceSlug); err != nil {
			return n, err
		}
		n++
	}
	return n, nil
}

func (s *OpenSearchSearcher) requestJSON(ctx context.Context, method, path string, body any, out any) error {
	var r io.Reader
	if body != nil {
		b, err := json.Marshal(body)
		if err != nil {
			return err
		}
		r = bytes.NewReader(b)
	}
	req, err := http.NewRequestWithContext(ctx, method, s.baseURL+path, r)
	if err != nil {
		return err
	}
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	resp, err := s.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 300 && method != http.MethodHead {
		raw, _ := io.ReadAll(io.LimitReader(resp.Body, 4096))
		return fmt.Errorf("opensearch %s %s: %s", method, path, strings.TrimSpace(string(raw)))
	}
	if out != nil && resp.StatusCode != http.StatusNotFound {
		return json.NewDecoder(resp.Body).Decode(out)
	}
	return nil
}
