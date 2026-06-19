package fts

import "testing"

func TestExtractKeywordsRussian(t *testing.T) {
	kw := ExtractKeywords("как локально деплоить?")
	if len(kw) < 2 {
		t.Fatalf("expected at least 2 keywords, got %v", kw)
	}
	if kw[0] != "локально" && kw[1] != "локально" {
		t.Fatalf("expected локально in keywords, got %v", kw)
	}
}

func TestExtractKeywordsStripsStopWords(t *testing.T) {
	kw := ExtractKeywords("как установить?")
	if len(kw) != 1 || kw[0] != "установить" {
		t.Fatalf("expected [установить], got %v", kw)
	}
}

func TestKeywordPrefix(t *testing.T) {
	if got := KeywordPrefix("деплоить"); got != "депло" {
		t.Fatalf("unexpected prefix: %q", got)
	}
}
