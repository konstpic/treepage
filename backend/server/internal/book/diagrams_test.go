package book

import (
	"strings"
	"testing"

	"github.com/konstpic/treepage/backend/pkg/models"
)

func TestBuildHasPartFlowchart(t *testing.T) {
	parent := models.Document{
		ID: "p", Path: "api/api.md", Title: "API Gateway", Slug: "api",
		Content: "---\nhasPart:\n  - [[api.dataapi]]\n---\n\n# API",
	}
	child := models.Document{
		ID: "c", Path: "api/dataapi/api.dataapi.md", Title: "DataAPI", Slug: "dataapi",
		Content: "# DataAPI\n\nCore service.",
	}
	idx := BuildIndex([]models.Document{parent, child})
	preview := Preview{
		Summary: Summary{Title: "API"},
		Chapters: []Chapter{
			{ID: "p", Title: "API Gateway", Path: parent.Path},
			{ID: "c", Title: "DataAPI", Path: child.Path},
		},
	}
	out := buildHasPartFlowchart(preview, idx)
	if !strings.Contains(out, "```mermaid") {
		t.Fatalf("expected mermaid block: %q", out)
	}
	if !strings.Contains(out, "API Gateway") || !strings.Contains(out, "DataAPI") {
		t.Fatalf("expected node labels: %q", out)
	}
	if !strings.Contains(out, "-->") {
		t.Fatalf("expected edge: %q", out)
	}
}

func TestMermaidEscape_apostrophe(t *testing.T) {
	out := mermaidEscape("Система 'Софтфон'")
	if strings.Contains(out, "'") {
		t.Fatalf("apostrophe should be removed: %q", out)
	}
}

func TestBuildIntegrationFlowchart(t *testing.T) {
	brief := diagramBrief{
		Components: []string{"Omni Chats", "Omni Core"},
		Edges:      []string{"Omni Chats --> Omni Core"},
	}
	out := buildIntegrationFlowchart(brief)
	if !strings.Contains(out, "flowchart LR") {
		t.Fatalf("expected LR flowchart: %q", out)
	}
	if !strings.Contains(out, "c0 --> c1") {
		t.Fatalf("expected edge: %q", out)
	}
}

func TestBuildSectionFlowchart_fallback(t *testing.T) {
	preview := Preview{
		Summary: Summary{Title: "Analytics"},
		Chapters: []Chapter{
			{ID: "1", Title: "Overview", Section: "Обзор"},
			{ID: "2", Title: "Events", Section: "Компоненты"},
			{ID: "3", Title: "Reports", Section: "Компоненты"},
		},
	}
	out := buildSectionFlowchart(preview)
	if !strings.Contains(out, "flowchart") {
		t.Fatalf("expected flowchart: %q", out)
	}
	if !strings.Contains(out, "root --> sec") {
		t.Fatalf("expected link to subgraph child: %q", out)
	}
}

func TestSanitizeMermaid_stripsFence(t *testing.T) {
	raw := "```mermaid\nflowchart LR\n  A --> B\n```"
	out := sanitizeMermaid(raw)
	if !strings.HasPrefix(out, "flowchart LR") {
		t.Fatalf("got %q", out)
	}
}
