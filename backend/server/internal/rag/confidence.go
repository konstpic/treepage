package rag

import (
	"strings"

	"github.com/konstpic/treepage/backend/pkg/fts"
)

const lowConfidenceThreshold = 0.42

type confidenceResult struct {
	Score          float64
	LowConfidence  bool
	FollowUp       []string
}

func computeConfidence(rows []chunkRow, answer string, citations []Citation) confidenceResult {
	if len(rows) == 0 {
		return confidenceResult{Score: 0, LowConfidence: true}
	}
	top := rows[0].Rank
	second := 0.0
	if len(rows) > 1 {
		second = rows[1].Rank
	}
	gap := top - second
	if gap < 0 {
		gap = 0
	}

	score := 0.0
	if top > 0 {
		score += minFloat(top/20, 0.45)
	}
	if rows[0].VectorSim > 0 {
		score += rows[0].VectorSim * 0.35
	}
	if gap > 2 {
		score += 0.1
	} else if gap < 0.5 {
		score -= 0.08
	}
	if len(citations) > 0 && citations[0].Score > 2 {
		score += 0.12
	}
	ans := strings.ToLower(strings.TrimSpace(answer))
	if strings.Contains(ans, "не знаю") || strings.Contains(ans, "don't know") || strings.Contains(ans, "i don't know") {
		score -= 0.25
	}
	if score < 0 {
		score = 0
	}
	if score > 1 {
		score = 1
	}
	return confidenceResult{
		Score:         score,
		LowConfidence: score < lowConfidenceThreshold,
	}
}

func (s *Service) buildFollowUpQuestions(_ string, rows []chunkRow, cyrillic bool) []string {
	if len(rows) == 0 {
		if cyrillic {
			return []string{
				"О каком разделе интерфейса или документации вы спрашиваете?",
				"В каком пространстве (space) вы работаете?",
			}
		}
		return []string{
			"Which part of the UI or documentation do you mean?",
			"Which space or feature are you asking about?",
		}
	}
	return defaultFollowUps(cyrillic)
}

func defaultFollowUps(cyrillic bool) []string {
	if cyrillic {
		return []string{
			"Уточните, о какой функции или странице интерфейса идёт речь?",
			"Вы спрашиваете про установку, навигацию или редактирование документов?",
		}
	}
	return []string{
		"Which feature or page are you asking about?",
		"Are you asking about installation, navigation, or editing docs?",
	}
}

func minFloat(a, b float64) float64 {
	if a < b {
		return a
	}
	return b
}

func isCyrillicQuestion(q string) bool {
	return fts.HasCyrillic(q)
}
