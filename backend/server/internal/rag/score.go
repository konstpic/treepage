package rag

import (
	"strings"

	"github.com/konstpic/treepage/backend/pkg/fts"
)

var baseTermSynonyms = map[string][]string{
	"страниц":  {"страницы", "страница", "pages", "page", "документ"},
	"раздел":   {"раздел", "вкладка", "tab", "section", "панель", "sidebar"},
	"навигац":  {"навигация", "navigation", "меню", "интерфейс"},
	"установ":  {"установка", "install", "развернуть", "деплой", "deploy"},
	"депло":    {"деплой", "deploy", "развернуть", "установка"},
	"локаль":   {"локально", "local", "development", "разработка"},
	"поиск":    {"search", "найти"},
	"простран": {"пространство", "space", "spaces"},
}

func expandSearchTerms(keywords []string, learned map[string][]string) []string {
	seen := map[string]struct{}{}
	var out []string
	add := func(s string) {
		s = strings.ToLower(strings.TrimSpace(s))
		if len([]rune(s)) < 2 {
			return
		}
		if _, ok := seen[s]; ok {
			return
		}
		seen[s] = struct{}{}
		out = append(out, s)
	}
	allSynonyms := mergeSynonymMaps(baseTermSynonyms, learned)
	for _, kw := range keywords {
		add(kw)
		add(fts.KeywordPrefix(kw))
		for stem, syns := range allSynonyms {
			if strings.HasPrefix(kw, stem) || strings.HasPrefix(stem, kw) {
				for _, syn := range syns {
					add(syn)
				}
			}
		}
		for term, syns := range learned {
			if strings.Contains(kw, term) || strings.Contains(term, kw) {
				add(term)
				for _, syn := range syns {
					add(syn)
				}
			}
		}
	}
	return out
}

func mergeSynonymMaps(a, b map[string][]string) map[string][]string {
	out := make(map[string][]string, len(a)+len(b))
	for k, v := range a {
		out[k] = v
	}
	for k, v := range b {
		out[k] = mergeSynonyms(out[k], v)
	}
	return out
}

func scoreChunkWithLearned(r chunkRow, keywords []string, learned map[string][]string) float64 {
	score := r.Rank
	if score <= 0 {
		score = 0.01
	}
	if r.VectorSim > 0 {
		score += r.VectorSim * 5
	}
	terms := expandSearchTerms(keywords, learned)
	title := strings.ToLower(r.Title)
	path := strings.ToLower(r.Path)
	content := strings.ToLower(r.Content)

	titleHits, pathHits, contentHits := 0, 0, 0
	for _, term := range terms {
		if strings.Contains(title, term) {
			titleHits++
			score += 10
		}
		if strings.Contains(path, term) {
			pathHits++
			score += 5
		}
		n := strings.Count(content, term)
		if n > 0 {
			contentHits += n
			score += float64(n) * 1.5
		}
	}
	if titleHits > 0 && pathHits > 0 {
		score += 5
	}
	if contentHits > 0 && titleHits == 0 && pathHits == 0 && len([]rune(content)) > 2000 {
		score *= 0.7
	}
	if strings.Contains(path, "navigation") || strings.Contains(path, "user/navigation") {
		for _, term := range terms {
			if term == "страниц" || term == "страницы" || term == "страница" || term == "раздел" || term == "вкладка" || term == "pages" || term == "page" {
				score += 12
				break
			}
		}
	}
	return score
}

func rankChunksWithLearned(rows []chunkRow, keywords []string, learned map[string][]string) []chunkRow {
	if len(rows) == 0 {
		return rows
	}
	scored := make([]chunkRow, len(rows))
	copy(scored, rows)
	for i := range scored {
		scored[i].Rank = scoreChunkWithLearned(scored[i], keywords, learned)
	}
	for i := 1; i < len(scored); i++ {
		j := i
		for j > 0 && scored[j].Rank > scored[j-1].Rank {
			scored[j], scored[j-1] = scored[j-1], scored[j]
			j--
		}
	}
	return scored
}

func rankChunks(rows []chunkRow, keywords []string) []chunkRow {
	return rankChunksWithLearned(rows, keywords, nil)
}

func scoreChunk(r chunkRow, keywords []string) float64 {
	return scoreChunkWithLearned(r, keywords, nil)
}

// bestChunkPerDocument keeps one chunk per document (rows must be score-sorted desc).
func bestChunkPerDocument(rows []chunkRow) []chunkRow {
	seen := map[string]struct{}{}
	out := make([]chunkRow, 0, len(rows))
	for _, r := range rows {
		if _, dup := seen[r.DocumentID]; dup {
			continue
		}
		seen[r.DocumentID] = struct{}{}
		out = append(out, r)
	}
	return out
}
