package rag

import (
	"context"
	"encoding/json"

	"github.com/konstpic/treepage/backend/pkg/models"
)

type FeedbackInput struct {
	Question   string     `json:"question" binding:"required"`
	Answer     string     `json:"answer"`
	Helpful    bool       `json:"helpful"`
	Confidence float64    `json:"confidence"`
	Sources    []Source   `json:"sources"`
	Citations  []Citation `json:"citations"`
}

func (s *Service) SubmitFeedback(ctx context.Context, userID string, in FeedbackInput) error {
	srcJSON, _ := json.Marshal(in.Sources)
	citJSON, _ := json.Marshal(in.Citations)
	row := models.RAGFeedback{
		Question:   in.Question,
		Answer:     in.Answer,
		Helpful:    in.Helpful,
		Confidence: float32(in.Confidence),
		Sources:    srcJSON,
		Citations:  citJSON,
	}
	if userID != "" {
		row.UserID = &userID
	}
	if err := s.db.WithContext(ctx).Create(&row).Error; err != nil {
		return err
	}
	if !in.Helpful {
		go s.learnFromNegativeFeedback(context.Background(), in.Question, in.Sources)
	}
	return nil
}
