package fts

import "fmt"

// WebsearchMatchSQL uses websearch_to_tsquery (more forgiving than plainto_tsquery).
func WebsearchMatchSQL(column string) string {
	return fmt.Sprintf(`(
  to_tsvector('english', %s) @@ websearch_to_tsquery('english', ?)
  OR to_tsvector('russian', %s) @@ websearch_to_tsquery('russian', ?)
  OR to_tsvector('simple', %s) @@ websearch_to_tsquery('simple', ?)
)`, column, column, column)
}

// WebsearchRankSQL ranks using websearch_to_tsquery.
func WebsearchRankSQL(column string) string {
	return fmt.Sprintf(`GREATEST(
  ts_rank(to_tsvector('english', %s), websearch_to_tsquery('english', ?)),
  ts_rank(to_tsvector('russian', %s), websearch_to_tsquery('russian', ?)),
  ts_rank(to_tsvector('simple', %s), websearch_to_tsquery('simple', ?))
)`, column, column, column)
}

// KeywordORMatchSQL matches if ANY keyword hits in ANY language config.
func KeywordORMatchSQL(column string, keywordCount int) string {
	if keywordCount <= 0 {
		return "FALSE"
	}
	var parts []string
	for range keywordCount {
		parts = append(parts, fmt.Sprintf(`(
  to_tsvector('english', %s) @@ plainto_tsquery('english', ?)
  OR to_tsvector('russian', %s) @@ plainto_tsquery('russian', ?)
  OR to_tsvector('simple', %s) @@ plainto_tsquery('simple', ?)
)`, column, column, column))
	}
	return "(" + joinOR(parts) + ")"
}

// KeywordORRankSQL sums ranks across keywords (any hit boosts score).
func KeywordORRankSQL(column string, keywordCount int) string {
	if keywordCount <= 0 {
		return "0"
	}
	var parts []string
	for range keywordCount {
		parts = append(parts, fmt.Sprintf(`GREATEST(
  ts_rank(to_tsvector('english', %s), plainto_tsquery('english', ?)),
  ts_rank(to_tsvector('russian', %s), plainto_tsquery('russian', ?)),
  ts_rank(to_tsvector('simple', %s), plainto_tsquery('simple', ?))
)`, column, column, column))
	}
	if len(parts) == 1 {
		return parts[0]
	}
	return "(" + joinPlus(parts) + ")"
}

// ILIKEKeywordMatchSQL fuzzy-matches keyword prefixes in content and title.
func ILIKEKeywordMatchSQL(contentCol, titleCol string, keywordCount int) string {
	if keywordCount <= 0 {
		return "FALSE"
	}
	var parts []string
	for range keywordCount {
		parts = append(parts, fmt.Sprintf("(%s ILIKE ? OR %s ILIKE ?)", contentCol, titleCol))
	}
	return "(" + joinOR(parts) + ")"
}

func ILIKEKeywordArgs(keywords []string) []any {
	args := make([]any, 0, len(keywords)*2)
	for _, kw := range keywords {
		prefix := KeywordPrefix(kw)
		pattern := "%" + prefix + "%"
		args = append(args, pattern, pattern)
	}
	return args
}

func KeywordORArgs(keywords []string) []any {
	args := make([]any, 0, len(keywords)*3)
	for _, kw := range keywords {
		args = append(args, kw, kw, kw)
	}
	return args
}

func joinOR(parts []string) string {
	out := parts[0]
	for i := 1; i < len(parts); i++ {
		out += " OR " + parts[i]
	}
	return out
}

func joinPlus(parts []string) string {
	out := parts[0]
	for i := 1; i < len(parts); i++ {
		out += " + " + parts[i]
	}
	return out
}
