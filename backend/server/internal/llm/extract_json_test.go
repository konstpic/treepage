package llm

import (
	"strings"
	"testing"
)

func TestExtractJSON_fenced(t *testing.T) {
	raw := "Here is the plan:\n\n```json\n{\"title\":\"Test\",\"parts\":[]}\n```\n"
	got := ExtractJSON(raw)
	if got != `{"title":"Test","parts":[]}` {
		t.Fatalf("got %q", got)
	}
}

func TestExtractJSON_preamble(t *testing.T) {
	raw := `Sure! {"title":"X","description":"Y","parts":[{"title":"A","chapter_ids":[]}]}`
	got := ExtractJSON(raw)
	if !strings.Contains(got, `"title":"X"`) {
		t.Fatalf("got %q", got)
	}
}

func TestExtractJSON_backticksOnly(t *testing.T) {
	raw := "```\n{\"notes\":\"ok\",\"parts\":[{\"title\":\"P\",\"chapter_ids\":[\"a\"]}]}\n```"
	got := ExtractJSON(raw)
	if got[0] != '{' {
		t.Fatalf("expected object, got %q", got)
	}
}

func TestExtractJSON_thinkBlocks(t *testing.T) {
	raw := "\x3cthink\x3ereasoning\x3c/think\x3e\n```json\n{\"parts\":[{\"title\":\"A\",\"chapter_ids\":[]}]}\n```"
	got := ExtractJSON(raw)
	if !strings.Contains(got, `"parts"`) {
		t.Fatalf("got %q", got)
	}
}
