package rag

import (
	"context"
	"fmt"
	"strings"

	"github.com/konstpic/treepage/backend/pkg/fts"
)

type chunkRow struct {
	ChunkID    string
	DocumentID string
	Content    string
	Title      string
	Slug       string
	SpaceSlug  string
	SpaceID    string
	Path       string
	Rank       float64
	VectorSim  float64
}

type retrievalPass struct {
	name     string
	matchSQL string
	rankSQL  string
	args     []any
	rankArgs []any
}

func (s *Service) buildRetrievalPasses(question string, expanded []string) []retrievalPass {
	question = strings.TrimSpace(question)
	keywords := fts.ExtractKeywords(question)
	passes := make([]retrievalPass, 0, 8)

	addFTS := func(name, matchSQL, rankSQL string, query string) {
		if strings.TrimSpace(query) == "" {
			return
		}
		passes = append(passes, retrievalPass{
			name:     name,
			matchSQL: matchSQL,
			rankSQL:  rankSQL,
			args:     fts.QueryArgs(query),
			rankArgs: fts.RankArgs(query),
		})
	}

	addFTS("phrase", fts.MatchSQL("dc.content"), fts.RankSQL("dc.content"), question)
	addFTS("websearch", fts.WebsearchMatchSQL("dc.content"), fts.WebsearchRankSQL("dc.content"), question)

	if len(keywords) > 0 {
		kwMatch := fts.KeywordORMatchSQL("dc.content", len(keywords))
		kwRank := fts.KeywordORRankSQL("dc.content", len(keywords))
		passes = append(passes, retrievalPass{
			name:     "keywords",
			matchSQL: kwMatch,
			rankSQL:  kwRank,
			args:     fts.KeywordORArgs(keywords),
			rankArgs: fts.KeywordORArgs(keywords),
		})

		ilikeMatch := fts.ILIKEKeywordMatchSQL("dc.content", "d.title", len(keywords))
		passes = append(passes, retrievalPass{
			name:     "ilike",
			matchSQL: ilikeMatch,
			rankSQL:  "1.0",
			args:     fts.ILIKEKeywordArgs(keywords),
			rankArgs: nil,
		})
	}

	seen := map[string]struct{}{question: {}}
	for _, q := range expanded {
		q = strings.TrimSpace(q)
		if q == "" {
			continue
		}
		if _, dup := seen[strings.ToLower(q)]; dup {
			continue
		}
		seen[strings.ToLower(q)] = struct{}{}
		addFTS("expanded:"+q, fts.WebsearchMatchSQL("dc.content"), fts.WebsearchRankSQL("dc.content"), q)
	}

	return passes
}

func (s *Service) retrieveChunks(ctx context.Context, passes []retrievalPass, allowed []string, fetchLimit int) ([]chunkRow, error) {
	merged := map[string]chunkRow{}
	order := make([]string, 0, fetchLimit*len(passes))

	for _, pass := range passes {
		var rows []chunkRow
		q := s.db.WithContext(ctx).Table("document_chunks dc").
			Select(`dc.id AS chunk_id, dc.document_id, dc.content, d.title, d.slug, sp.slug AS space_slug,
				d.space_id, d.path, `+pass.rankSQL+` AS rank`, pass.rankArgs...).
			Joins("JOIN documents d ON d.id = dc.document_id").
			Joins("JOIN spaces sp ON sp.id = d.space_id").
			Where("d.is_published = ?", true).
			Where(pass.matchSQL, pass.args...)
		if allowed != nil {
			q = q.Where("d.space_id IN ?", allowed)
		}
		if err := q.Order("rank DESC").Limit(fetchLimit).Scan(&rows).Error; err != nil {
			return nil, fmt.Errorf("rag retrieval %s: %w", pass.name, err)
		}
		for _, r := range rows {
			if r.ChunkID == "" {
				continue
			}
			if prev, ok := merged[r.ChunkID]; ok {
				if r.Rank > prev.Rank {
					merged[r.ChunkID] = r
				}
				continue
			}
			merged[r.ChunkID] = r
			order = append(order, r.ChunkID)
		}
	}

	out := make([]chunkRow, 0, len(order))
	for _, id := range order {
		out = append(out, merged[id])
	}
	return out, nil
}

func (s *Service) expandQueryWithLLM(ctx context.Context, question string) []string {
	prompt := fmt.Sprintf(`You help search internal documentation. The user asked:
"%s"

Return 4 short search phrases (one per line) that would find relevant docs.
Mix Russian and English. Include synonyms (deploy/деплой/развернуть, install/установка, local/локально, docker).
No numbering, no explanations, only phrases.`, question)

	raw, err := s.llm.Chat(ctx,
		"You output only search keyword phrases, one per line.",
		prompt,
	)
	if err != nil || strings.TrimSpace(raw) == "" {
		return nil
	}
	var out []string
	for _, line := range strings.Split(raw, "\n") {
		line = strings.TrimSpace(strings.TrimLeft(line, "0123456789.-) "))
		if line != "" {
			out = append(out, line)
		}
	}
	if len(out) > 6 {
		out = out[:6]
	}
	return out
}
