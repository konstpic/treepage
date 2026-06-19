package book

import (
	"testing"

	"github.com/konstpic/treepage/backend/pkg/models"
)

func TestGenerateAnalyticsBook(t *testing.T) {
	docs := []models.Document{
		{ID: "1", Slug: "analytics-analytics", Title: "analytics", Path: "analytics/analytics.md", Content: `---
status: current
c4.level: 1
hasPart:
  - "[[analytics.analytics]]"
  - "[[analytics.cpm]]"
---
# Analytics domain`},
		{ID: "2", Slug: "analytics-analytics-analytics-analytics", Title: "Микросервис server", Path: "analytics/analytics/analytics.analytics.md", Content: "# Микросервис server\n\nBody"},
		{ID: "3", Slug: "analytics-analytics-server-deploy", Title: "deploy", Path: "analytics/analytics/server/deploy.md", Content: "## Deploy\n\nSteps"},
		{ID: "4", Slug: "analytics-cpm-analytics-cpm", Title: "CPM", Path: "analytics/cpm/analytics.cpm.md", Content: "# CPM\n\nSystem"},
	}

	idx := BuildIndex(docs)
	gen := NewGenerator()
	gen.minDocsPerBook = 3
	preview := gen.GenerateBook(idx, "analytics")

	if preview.DocCount != 4 {
		t.Fatalf("expected 4 docs, got %d", preview.DocCount)
	}
	if preview.Strategy != StrategyGraph {
		t.Fatalf("expected graph strategy, got %s", preview.Strategy)
	}
	if preview.ChapterCount < 3 {
		t.Fatalf("expected at least 3 chapters, got %d", preview.ChapterCount)
	}
	if preview.Markdown == "" {
		t.Fatal("expected compiled markdown")
	}
}

func TestSkipsOrphansRoot(t *testing.T) {
	docs := []models.Document{
		{ID: "1", Path: "ORPHANS/foo.md", Title: "Orphan", Content: "# Orphan"},
		{ID: "2", Path: "ORPHANS/bar.md", Title: "Orphan 2", Content: "# Orphan 2"},
	}
	idx := BuildIndex(docs)
	gen := NewGenerator()
	books := gen.GenerateAll(idx)
	if len(books) != 0 {
		t.Fatalf("expected no books from ORPHANS, got %d", len(books))
	}
}
