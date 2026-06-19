package service

import (
	"context"
	"time"

	"github.com/konstpic/treepage/backend/pkg/models"
	"gorm.io/gorm"
)

type AnalyticsService struct {
	db *gorm.DB
}

func NewAnalyticsService(db *gorm.DB) *AnalyticsService {
	return &AnalyticsService{db: db}
}

func (s *AnalyticsService) RecordView(ctx context.Context, documentID string) {
	now := time.Now()
	s.db.WithContext(ctx).Exec(`
		INSERT INTO document_view_stats (document_id, view_count, last_viewed_at)
		VALUES (?, 1, ?)
		ON CONFLICT (document_id) DO UPDATE SET
			view_count = document_view_stats.view_count + 1,
			last_viewed_at = EXCLUDED.last_viewed_at
	`, documentID, now)
}

func (s *AnalyticsService) LogSearch(ctx context.Context, userID, query string, resultCount int) {
	if len(query) == 0 {
		return
	}
	if len(query) > 512 {
		query = query[:512]
	}
	var uid *string
	if userID != "" {
		uid = &userID
	}
	_ = s.db.WithContext(ctx).Create(&models.SearchQueryLog{
		UserID: uid, QueryText: query, ResultCount: resultCount,
	}).Error
}

type TopDocumentRow struct {
	DocumentID string `json:"document_id"`
	SpaceID    string `json:"space_id"`
	SpaceSlug  string `json:"space_slug"`
	Title      string `json:"title"`
	DocSlug    string `json:"doc_slug"`
	ViewCount  int64  `json:"view_count"`
}

type StaleDocumentRow struct {
	DocumentID string    `json:"document_id"`
	SpaceSlug  string    `json:"space_slug"`
	Title      string    `json:"title"`
	DocSlug    string    `json:"doc_slug"`
	UpdatedAt  time.Time `json:"updated_at"`
}

type TopSearchRow struct {
	QueryText string `json:"query_text"`
	Count     int64  `json:"count"`
}

type AnalyticsOverview struct {
	TotalViews      int64              `json:"total_views"`
	TopDocuments    []TopDocumentRow   `json:"top_documents"`
	StaleDocuments  []StaleDocumentRow `json:"stale_documents"`
	TopSearches     []TopSearchRow     `json:"top_searches"`
}

func (s *AnalyticsService) Overview(ctx context.Context, staleDays int) (*AnalyticsOverview, error) {
	if staleDays <= 0 {
		staleDays = 90
	}
	out := &AnalyticsOverview{}
	s.db.WithContext(ctx).Table("document_view_stats").Select("COALESCE(SUM(view_count), 0)").Scan(&out.TotalViews)

	s.db.WithContext(ctx).Table("document_view_stats dvs").
		Select(`dvs.document_id, d.space_id, sp.slug AS space_slug, d.title, d.slug AS doc_slug, dvs.view_count`).
		Joins("JOIN documents d ON d.id = dvs.document_id").
		Joins("JOIN spaces sp ON sp.id = d.space_id").
		Order("dvs.view_count DESC").
		Limit(20).
		Scan(&out.TopDocuments)

	cutoff := time.Now().AddDate(0, 0, -staleDays)
	s.db.WithContext(ctx).Table("documents d").
		Select(`d.id AS document_id, sp.slug AS space_slug, d.title, d.slug AS doc_slug, d.updated_at`).
		Joins("JOIN spaces sp ON sp.id = d.space_id").
		Where("d.is_published = ? AND d.updated_at < ?", true, cutoff).
		Order("d.updated_at ASC").
		Limit(20).
		Scan(&out.StaleDocuments)

	s.db.WithContext(ctx).Table("search_query_log").
		Select("query_text, COUNT(*) AS count").
		Where("created_at > ?", time.Now().AddDate(0, 0, -30)).
		Group("query_text").
		Order("count DESC").
		Limit(20).
		Scan(&out.TopSearches)

	return out, nil
}
