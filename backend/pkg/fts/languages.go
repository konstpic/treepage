package fts

import "fmt"

// MatchSQL returns a WHERE fragment matching column against english, russian, and simple configs.
func MatchSQL(column string) string {
	return fmt.Sprintf(`(
  to_tsvector('english', %s) @@ plainto_tsquery('english', ?)
  OR to_tsvector('russian', %s) @@ plainto_tsquery('russian', ?)
  OR to_tsvector('simple', %s) @@ plainto_tsquery('simple', ?)
)`, column, column, column)
}

// RankSQL returns a GREATEST(ts_rank(...)) expression for the same configs.
func RankSQL(column string) string {
	return fmt.Sprintf(`GREATEST(
  ts_rank(to_tsvector('english', %s), plainto_tsquery('english', ?)),
  ts_rank(to_tsvector('russian', %s), plainto_tsquery('russian', ?)),
  ts_rank(to_tsvector('simple', %s), plainto_tsquery('simple', ?))
)`, column, column, column)
}

// QueryArgs repeats the same query text for each config placeholder in MatchSQL/RankSQL.
func QueryArgs(query string) []any {
	return []any{query, query, query}
}

// RankArgs repeats query for rank expression placeholders.
func RankArgs(query string) []any {
	return []any{query, query, query}
}
