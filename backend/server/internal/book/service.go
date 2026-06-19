package book

import (
	"context"
	"errors"
	"fmt"

	"github.com/konstpic/treepage/backend/server/internal/llm"
	"github.com/konstpic/treepage/backend/server/internal/service"
	"gorm.io/gorm"
)

type Service struct {
	docs      *service.DocumentService
	generator *Generator
	enhancer  *Enhancer
	readable  *ReadableGenerator
	store     *Store
}

func NewService(docs *service.DocumentService, db *gorm.DB, llmClient *llm.Client) *Service {
	return &Service{
		docs:      docs,
		generator: NewGenerator(),
		enhancer:  NewEnhancer(llmClient),
		readable:  NewReadableGenerator(llmClient),
		store:     NewStore(db),
	}
}

func (s *Service) LLMAvailable() bool {
	return s.readable.Available()
}

func (s *Service) ListCandidates(ctx context.Context, spaceID string) ([]Summary, error) {
	idx, err := s.loadIndex(ctx, spaceID)
	if err != nil {
		return nil, err
	}
	return s.generator.GenerateAll(idx), nil
}

type ListResponse struct {
	Saved       []StoredBookView `json:"saved"`
	Candidates  []Summary        `json:"candidates"`
	LLMAvailable bool            `json:"llm_available"`
}

func (s *Service) List(ctx context.Context, spaceID string) (*ListResponse, error) {
	saved, err := s.store.ListBySpace(ctx, spaceID)
	if err != nil {
		return nil, err
	}
	idx, err := s.loadIndex(ctx, spaceID)
	if err != nil {
		return nil, err
	}
	candidates := s.generator.GenerateAll(idx)

	views := make([]StoredBookView, 0, len(saved))
	for _, b := range saved {
		hash := HashSourceDocs(idx.All, b.RootPath)
		views = append(views, ToStoredView(b, hash))
	}
	return &ListResponse{
		Saved:        views,
		Candidates:   candidates,
		LLMAvailable: s.LLMAvailable(),
	}, nil
}

func (s *Service) GetSaved(ctx context.Context, spaceID, slug string, withContent bool) (*StoredBookView, error) {
	b, err := s.store.GetBySlug(ctx, spaceID, slug)
	if err != nil {
		return nil, err
	}
	idx, _ := s.loadIndex(ctx, spaceID)
	hash := ""
	if idx != nil {
		hash = HashSourceDocs(idx.All, b.RootPath)
	}
	view := ToStoredView(*b, hash)
	if !withContent {
		view.ContentMarkdown = ""
	}
	return &view, nil
}

func (s *Service) Create(ctx context.Context, spaceID string, userID *string, input CreateBookInput) (*StoredBookView, error) {
	idx, err := s.loadIndex(ctx, spaceID)
	if err != nil {
		return nil, err
	}
	root := input.RootPath
	preview := s.generator.GenerateBook(idx, root)
	if preview.DocCount == 0 {
		return nil, errors.New("no documents for root_path")
	}
	b, err := s.store.Create(ctx, spaceID, userID, input, preview)
	if err != nil {
		return nil, err
	}
	hash := HashSourceDocs(idx.All, b.RootPath)
	view := ToStoredView(*b, hash)
	return &view, nil
}

func (s *Service) Generate(ctx context.Context, spaceID, slug string, force bool) (*StoredBookView, error) {
	b, err := s.store.GetBySlug(ctx, spaceID, slug)
	if err != nil {
		return nil, err
	}
	if b.Status == "generating" {
		return nil, errors.New("generation already in progress")
	}

	idx, err := s.loadIndex(ctx, spaceID)
	if err != nil {
		return nil, err
	}
	sourceHash := HashSourceDocs(idx.All, b.RootPath)
	if !force && b.Status == "ready" && b.ContentMarkdown != "" && b.SourceHash == sourceHash {
		view := ToStoredView(*b, sourceHash)
		return &view, nil
	}

	if err := s.store.MarkGenerating(ctx, b); err != nil {
		return nil, err
	}

	preview := s.generator.GenerateBook(idx, b.RootPath)
	preview = ApplyAudience(preview, idx, b.Audience)
	if len(preview.Chapters) == 0 {
		err := fmt.Errorf("no chapters match audience %q", b.Audience)
		_ = s.store.MarkFailed(ctx, b, err.Error())
		return nil, err
	}

	enhancedOK := false
	if (b.Audience == AudienceDeveloper || b.Audience == AudienceArchitect) && s.enhancer.Available() {
		if ep, err := s.enhancer.Enhance(ctx, preview, idx, EnhanceOptions{
			Audience: b.Audience,
			Focus:    b.Focus,
		}); err == nil {
			preview = ApplyAudience(ep, idx, b.Audience)
			enhancedOK = true
		}
	}

	content, err := s.readable.Build(ctx, preview, idx, b.Audience)
	if err != nil {
		_ = s.store.MarkFailed(ctx, b, err.Error())
		return nil, err
	}

	if err := s.store.SaveGenerated(ctx, b, preview, content, sourceHash, enhancedOK); err != nil {
		return nil, err
	}
	view := ToStoredView(*b, sourceHash)
	return &view, nil
}

func (s *Service) Delete(ctx context.Context, spaceID, slug string) error {
	return s.store.Delete(ctx, spaceID, slug)
}

func (s *Service) loadIndex(ctx context.Context, spaceID string) (*Index, error) {
	docs, err := s.docs.ListFullBySpace(ctx, spaceID)
	if err != nil {
		return nil, err
	}
	return BuildIndex(docs), nil
}
