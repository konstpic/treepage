package rag

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/konstpic/treepage/backend/pkg/embeddings"
	"github.com/konstpic/treepage/backend/pkg/fts"
	"github.com/konstpic/treepage/backend/pkg/models"
	"github.com/konstpic/treepage/backend/pkg/ragindex"
	"github.com/konstpic/treepage/backend/server/internal/llm"
	"github.com/konstpic/treepage/backend/server/internal/service"
	"gorm.io/gorm"
)

type Service struct {
	db      *gorm.DB
	llm     *llm.Client
	embed   *embeddings.Client
	spaces  *service.SpaceService
	pageACL *service.PageACLService
}

func New(db *gorm.DB, llmClient *llm.Client, embedClient *embeddings.Client, spaces *service.SpaceService, pageACL *service.PageACLService) *Service {
	return &Service{db: db, llm: llmClient, embed: embedClient, spaces: spaces, pageACL: pageACL}
}

type Source struct {
	DocumentID string  `json:"document_id"`
	SpaceSlug  string  `json:"space_slug"`
	DocSlug    string  `json:"doc_slug"`
	Title      string  `json:"title"`
	Snippet    string  `json:"snippet"`
	Score      float64 `json:"score"`
}

type Answer struct {
	Answer            string     `json:"answer"`
	Sources           []Source   `json:"sources"`
	Citations         []Citation `json:"citations"`
	Confidence        float64    `json:"confidence"`
	LowConfidence     bool       `json:"low_confidence"`
	FollowUpQuestions []string   `json:"follow_up_questions,omitempty"`
}

func (s *Service) ReindexDocument(ctx context.Context, doc *models.Document) error {
	return ragindex.IndexDocumentWithEmbedder(ctx, s.db, doc, s.embed)
}

func (s *Service) ReindexAllPublished(ctx context.Context) (int, error) {
	var docs []models.Document
	if err := s.db.WithContext(ctx).Where("is_published = ?", true).Find(&docs).Error; err != nil {
		return 0, err
	}
	for i := range docs {
		if err := ragindex.IndexDocumentWithEmbedder(ctx, s.db, &docs[i], s.embed); err != nil {
			return i, err
		}
	}
	return len(docs), nil
}

func (s *Service) retrieveAndRank(ctx context.Context, question string, keywords []string, allowed []string, fetchLimit int, expanded []string) ([]chunkRow, error) {
	passes := s.buildRetrievalPasses(question, expanded)
	ftsRows, err := s.retrieveChunks(ctx, passes, allowed, fetchLimit)
	if err != nil {
		return nil, err
	}
	vecRows, err := s.vectorSearchChunks(ctx, question, allowed, fetchLimit)
	if err != nil {
		return nil, err
	}
	merged := mergeChunkRows(ftsRows, vecRows)
	merged = s.rerankWithEmbeddings(ctx, merged, question, keywords)
	learned := s.loadLearnedSynonyms(ctx)
	merged = rankChunksWithLearned(merged, keywords, learned)
	return bestChunkPerDocument(merged), nil
}

func (s *Service) Ask(ctx context.Context, question, userID string, globalRoles []string, limit int) (*Answer, error) {
	question = strings.TrimSpace(question)
	if question == "" {
		return nil, errors.New("question is required")
	}
	if !s.llm.Available() {
		return nil, errors.New("LLM is not configured (set LLM_ENABLED=true and LLM_API_URL)")
	}
	if limit <= 0 || limit > 10 {
		limit = 5
	}
	isSuperAdmin := service.HasRole(globalRoles, "super_admin")
	allowed, err := s.spaces.AccessibleSpaceIDs(ctx, userID, isSuperAdmin)
	if err != nil {
		return nil, err
	}
	if allowed != nil && len(allowed) == 0 {
		return &Answer{Answer: "No accessible documentation found for your account.", Sources: nil}, nil
	}

	fetchLimit := limit * 5
	keywords := fts.ExtractKeywords(question)
	learned := s.loadLearnedSynonyms(ctx)

	rows, err := s.retrieveAndRank(ctx, question, keywords, allowed, fetchLimit, nil)
	if err != nil {
		return nil, err
	}

	sources, contextParts, rankedRows := s.filterAndBuildSources(ctx, rows, userID, globalRoles, isSuperAdmin, limit)
	if len(sources) == 0 {
		expanded := s.expandQueryWithLLM(ctx, question)
		if len(expanded) > 0 {
			rows, err = s.retrieveAndRank(ctx, question, keywords, allowed, fetchLimit, expanded)
			if err != nil {
				return nil, err
			}
			sources, contextParts, rankedRows = s.filterAndBuildSources(ctx, rows, userID, globalRoles, isSuperAdmin, limit)
		}
	}

	if len(sources) == 0 {
		conf := computeConfidence(nil, "", nil)
		followUp := s.buildFollowUpQuestions(ctx, question, nil, isCyrillicQuestion(question))
		return &Answer{
			Answer:            notFoundMessage(question),
			Sources:           []Source{},
			Citations:         []Citation{},
			Confidence:        conf.Score,
			LowConfidence:     true,
			FollowUpQuestions: followUp,
		}, nil
	}

	citations := extractCitations(rankedRows, keywords, learned, 3)

	systemPrompt := "You are a helpful documentation assistant. Be concise and accurate. Use only the provided excerpts."
	userPromptIntro := `Excerpts are ordered by relevance (most relevant first).
Read ALL excerpts before answering. If the answer appears in any excerpt, give a clear direct answer and cite the document title.
Only say you don't know if the excerpts truly do not contain the information.

`
	if fts.HasCyrillic(question) {
		systemPrompt = "Ты помощник по документации. Отвечай кратко и точно на языке вопроса. Используй только приведённые фрагменты."
		userPromptIntro = `Фрагменты отсортированы по релевантности (самый подходящий — первый).
Прочитай ВСЕ фрагменты. Если ответ есть в тексте — дай конкретный ответ и укажи название документа.
Не говори «не знаю», если информация есть во фрагментах.

`
	}

	prompt := userPromptIntro + fmt.Sprintf(`Question: %s

Documentation:
%s`, question, strings.Join(contextParts, "\n\n---\n\n"))

	answer, err := s.llm.Chat(ctx, systemPrompt, prompt)
	if err != nil {
		return nil, err
	}
	answer = strings.TrimSpace(answer)

	conf := computeConfidence(rankedRows, answer, citations)
	followUp := []string(nil)
	if conf.LowConfidence {
		followUp = s.buildFollowUpQuestions(ctx, question, rankedRows, isCyrillicQuestion(question))
	}

	return &Answer{
		Answer:            answer,
		Sources:           sources,
		Citations:         citations,
		Confidence:        conf.Score,
		LowConfidence:     conf.LowConfidence,
		FollowUpQuestions: followUp,
	}, nil
}

func notFoundMessage(question string) string {
	if fts.HasCyrillic(question) {
		return "Не удалось найти подходящую документацию по вашему вопросу."
	}
	return "I could not find relevant documentation for your question."
}

func (s *Service) filterAndBuildSources(ctx context.Context, rows []chunkRow, userID string, globalRoles []string, isSuperAdmin bool, limit int) ([]Source, []string, []chunkRow) {
	sources := make([]Source, 0, limit)
	var contextParts []string
	var ranked []chunkRow
	seen := map[string]struct{}{}
	for _, r := range rows {
		spaceRole, _ := s.spaces.EffectiveRole(ctx, r.SpaceID, userID, globalRoles)
		if _, ok := s.pageACL.EffectiveDocumentRole(ctx, r.SpaceID, r.Path, userID, spaceRole, isSuperAdmin); !ok {
			continue
		}
		ranked = append(ranked, r)
		if _, dup := seen[r.DocumentID]; dup {
			continue
		}
		seen[r.DocumentID] = struct{}{}
		snippet := r.Content
		if len(snippet) > 300 {
			snippet = snippet[:300] + "..."
		}
		sources = append(sources, Source{
			DocumentID: r.DocumentID, SpaceSlug: r.SpaceSlug, DocSlug: r.Slug,
			Title: r.Title, Snippet: snippet, Score: r.Rank,
		})
		contextParts = append(contextParts, fmt.Sprintf("## %s (%s)\n%s", r.Title, r.Path, r.Content))
		if len(sources) >= limit {
			break
		}
	}
	return sources, contextParts, ranked
}
