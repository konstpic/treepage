package rag

import (
	"context"
	"encoding/json"
	"strings"

	"github.com/konstpic/treepage/backend/pkg/models"
	"gorm.io/gorm"
)

func (s *Service) loadLearnedSynonyms(ctx context.Context) map[string][]string {
	var rows []models.RAGLearnedSynonym
	if err := s.db.WithContext(ctx).Find(&rows).Error; err != nil {
		return nil
	}
	out := make(map[string][]string, len(rows))
	for _, r := range rows {
		if len(r.Synonyms) > 0 {
			out[r.Term] = r.Synonyms
		}
	}
	return out
}

func (s *Service) learnFromNegativeFeedback(ctx context.Context, question string, sources []Source) {
	if !s.llm.Available() {
		return
	}
	prompt := `Extract 1-3 Russian/English keyword stems from this failed documentation search.
Return JSON: {"terms":[{"term":"stem","synonyms":["word1","word2"]}]}
Only include terms that could improve future search.

Question: ` + question
	raw, err := s.llm.ChatJSON(ctx, "Output valid JSON only.", prompt)
	if err != nil {
		return
	}
	var parsed struct {
		Terms []struct {
			Term     string   `json:"term"`
			Synonyms []string `json:"synonyms"`
		} `json:"terms"`
	}
	if json.Unmarshal([]byte(raw), &parsed) != nil {
		return
	}
	for _, t := range parsed.Terms {
		if len([]rune(t.Term)) < 3 || len(t.Synonyms) == 0 {
			continue
		}
		_ = s.upsertLearnedSynonym(ctx, strings.ToLower(t.Term), t.Synonyms)
	}
}

func (s *Service) upsertLearnedSynonym(ctx context.Context, term string, synonyms []string) error {
	var row models.RAGLearnedSynonym
	err := s.db.WithContext(ctx).Where("term = ?", term).First(&row).Error
	if err == gorm.ErrRecordNotFound {
		return s.db.WithContext(ctx).Create(&models.RAGLearnedSynonym{
			Term: term, Synonyms: mergeSynonyms(nil, synonyms), HitCount: 1,
		}).Error
	}
	if err != nil {
		return err
	}
	merged := mergeSynonyms(row.Synonyms, synonyms)
	return s.db.WithContext(ctx).Model(&row).Updates(map[string]any{
		"synonyms":   merged,
		"hit_count":  gorm.Expr("hit_count + 1"),
		"updated_at": gorm.Expr("NOW()"),
	}).Error
}

func mergeSynonyms(existing, add []string) []string {
	seen := map[string]struct{}{}
	var out []string
	for _, s := range append(existing, add...) {
		s = strings.ToLower(strings.TrimSpace(s))
		if len([]rune(s)) < 2 {
			continue
		}
		if _, ok := seen[s]; ok {
			continue
		}
		seen[s] = struct{}{}
		out = append(out, s)
	}
	return out
}
