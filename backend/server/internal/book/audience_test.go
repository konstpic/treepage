package book

import (
	"testing"

	"github.com/konstpic/treepage/backend/pkg/models"
)

func TestApplyAudience_opsFiltersDevelopAndEntity(t *testing.T) {
	docs := []models.Document{
		{ID: "1", Path: "api/dataapi/api.dataapi.md", Title: "DataAPI", Content: "# DataAPI\n\nСистема.", Slug: "a"},
		{ID: "2", Path: "api/dataapi/dispatcher/deploy.md", Title: "Deploy Dispatcher", Content: "# Deploy\n\nreplicas: 1", Slug: "b"},
		{ID: "3", Path: "api/dataapi/jsonrpc/develop.md", Title: "Develop JSON-RPC", Content: "# Dev", Slug: "c"},
		{ID: "4", Path: "api/dataapi/rabbitmq/api.dataapi.rabbitmq.md", Title: "RabbitMQ", Content: "# RabbitMQ\n\nБрокер.", Slug: "d"},
		{ID: "5", Path: "api/dataapi/dispatcher/maintain.md", Title: "Maintain Dispatcher", Content: "# Maintain", Slug: "e"},
	}
	idx := BuildIndex(docs)
	base := Preview{
		Summary: Summary{RootPath: "api"},
		Chapters: []Chapter{
			{ID: "1", Title: "DataAPI", Path: docs[0].Path, Kind: "overview"},
			{ID: "2", Title: "Deploy Dispatcher", Path: docs[1].Path, Kind: "operations"},
			{ID: "3", Title: "Develop JSON-RPC", Path: docs[2].Path, Kind: "operations"},
			{ID: "4", Title: "RabbitMQ", Path: docs[3].Path, Kind: "entity"},
			{ID: "5", Title: "Maintain Dispatcher", Path: docs[4].Path, Kind: "operations"},
		},
	}
	out := ApplyAudience(base, idx, AudienceOps)
	if len(out.Chapters) != 3 {
		t.Fatalf("expected 3 ops chapters, got %d", len(out.Chapters))
	}
	for _, ch := range out.Chapters {
		if isDevelopPath(ch.Path) {
			t.Fatalf("develop should be excluded: %s", ch.Path)
		}
		if ch.Path == docs[3].Path {
			t.Fatalf("entity rabbitmq should be excluded for ops")
		}
	}
	if out.Chapters[0].Section != "Контекст" {
		t.Fatalf("first section want context, got %s", out.Chapters[0].Section)
	}
}

func TestCleanMarkdown_stripsWiki(t *testing.T) {
	raw := "---\nstatus: live\n---\n\n# T\n\nSee [[api.foo]] for more."
	out := cleanMarkdown(raw)
	if contains(out, "[[") {
		t.Fatalf("wiki links should be stripped: %q", out)
	}
	if !contains(out, "api.foo") {
		t.Fatalf("link label should remain: %q", out)
	}
}

func contains(s, sub string) bool {
	return len(s) >= len(sub) && (s == sub || len(sub) == 0 || indexOfStr(s, sub) >= 0)
}

func indexOfStr(s, sub string) int {
	for i := 0; i+len(sub) <= len(s); i++ {
		if s[i:i+len(sub)] == sub {
			return i
		}
	}
	return -1
}
