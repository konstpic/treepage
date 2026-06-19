package rag

import (
	"strings"
)

type Citation struct {
	DocumentID string  `json:"document_id"`
	SpaceSlug  string  `json:"space_slug"`
	DocSlug    string  `json:"doc_slug"`
	Title      string  `json:"title"`
	Path       string  `json:"path"`
	Quote      string  `json:"quote"`
	Score      float64 `json:"score"`
}

func extractCitations(rows []chunkRow, keywords []string, learned map[string][]string, limit int) []Citation {
	if limit <= 0 {
		limit = 3
	}
	terms := expandSearchTerms(keywords, learned)
	var out []Citation
	seen := map[string]struct{}{}

	for _, r := range rows {
		if len(out) >= limit {
			break
		}
		quote, score := bestQuoteLine(r.Content, terms)
		if quote == "" {
			continue
		}
		key := r.DocumentID + "|" + quote
		if _, dup := seen[key]; dup {
			continue
		}
		seen[key] = struct{}{}
		out = append(out, Citation{
			DocumentID: r.DocumentID, SpaceSlug: r.SpaceSlug, DocSlug: r.Slug,
			Title: r.Title, Path: r.Path, Quote: quote, Score: score + r.Rank,
		})
	}
	for i := 1; i < len(out); i++ {
		j := i
		for j > 0 && out[j].Score > out[j-1].Score {
			out[j], out[j-1] = out[j-1], out[j]
			j--
		}
	}
	return out
}

func bestQuoteLine(content string, terms []string) (string, float64) {
	lines := strings.Split(content, "\n")
	var best string
	var bestScore float64
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if len([]rune(line)) < 10 {
			continue
		}
		if strings.HasPrefix(line, "#") {
			continue
		}
		lower := strings.ToLower(line)
		score := 0.0
		for _, term := range terms {
			if strings.Contains(lower, term) {
				score += 2
			}
		}
		if len(terms) == 0 {
			score = 0.5
		}
		if score > bestScore {
			bestScore = score
			best = trimQuote(line, 280)
		}
	}
	if best == "" && len(lines) > 0 {
		for _, line := range lines {
			line = strings.TrimSpace(line)
			if len([]rune(line)) >= 20 && !strings.HasPrefix(line, "#") {
				return trimQuote(line, 280), 0.1
			}
		}
	}
	return best, bestScore
}

func trimQuote(s string, maxRunes int) string {
	runes := []rune(s)
	if len(runes) <= maxRunes {
		return s
	}
	return string(runes[:maxRunes]) + "…"
}
