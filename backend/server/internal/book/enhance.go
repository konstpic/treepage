package book

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/konstpic/treepage/backend/server/internal/llm"
)

const StrategyAI = "ai"

type EnhanceOptions struct {
	Audience string `json:"audience"` // architect | developer | ops | onboarding
	Focus    string `json:"focus"`    // optional free-text goal
}

type aiPlan struct {
	Title             string   `json:"title"`
	Description       string   `json:"description"`
	IntroMarkdown     string   `json:"intro_markdown"`
	Parts             []aiPart `json:"parts"`
	ExcludeChapterIDs []string `json:"exclude_chapter_ids"`
	Notes             string   `json:"notes"`
}

type aiPart struct {
	Title      string   `json:"title"`
	ChapterIDs []string `json:"chapter_ids"`
}

type Enhancer struct {
	llm *llm.Client
}

func NewEnhancer(client *llm.Client) *Enhancer {
	return &Enhancer{llm: client}
}

func (e *Enhancer) Available() bool {
	return e.llm != nil && e.llm.Available()
}

func (e *Enhancer) Enhance(ctx context.Context, preview Preview, idx *Index, opts EnhanceOptions) (Preview, error) {
	if !e.Available() {
		return preview, fmt.Errorf("LLM is not configured")
	}
	if opts.Audience == "" {
		opts.Audience = "developer"
	}

	plan, err := e.requestPlan(ctx, preview, idx, opts)
	if err != nil {
		return preview, err
	}
	return applyAIPlan(preview, idx, plan, opts.Audience), nil
}

func (e *Enhancer) requestPlan(ctx context.Context, preview Preview, idx *Index, opts EnhanceOptions) (*aiPlan, error) {
	maxBriefs := 0
	if opts.Audience == AudienceArchitect {
		maxBriefs = 50
	}
	chapters := buildChapterBriefs(preview, idx, maxBriefs)

	userPayload, _ := json.Marshal(map[string]any{
		"book_id":     preview.ID,
		"root_path":   preview.RootPath,
		"audience":    opts.Audience,
		"focus":       opts.Focus,
		"base_title":  preview.Title,
		"doc_count":   preview.DocCount,
		"chapters":    chapters,
		"constraints": chapterConstraints(opts.Audience),
	})

	systemPrompt := `You are a technical documentation editor. Given a draft book outline from a docs repository, produce a clearer book structure for a specific audience.

Return ONLY valid JSON with this schema:
{
  "title": "string — compelling book title",
  "description": "string — 1-2 sentences",
  "intro_markdown": "string — markdown introduction (2-4 paragraphs) explaining scope and how to read the book",
  "parts": [
    {"title": "Part name", "chapter_ids": ["uuid-from-input", "..."]}
  ],
  "exclude_chapter_ids": ["uuid", "..."],
  "notes": "string — brief note on editorial choices"
}

Rules:
- Use ONLY chapter ids from the input; never invent ids.
- Each included chapter appears at most once across all parts.
- Order chapters logically for the audience (overview → components → details).
- Group into 3-8 parts with human-readable titles (not raw folder names).
- exclude_chapter_ids: drop irrelevant pages (wrong audience, duplicates, empty stubs).
- For audience "ops": ONLY include deploy, maintain/monitoring, and overview pages. EXCLUDE develop/development and all architecture entity pages.
- For audience "architect": prioritize overview/entity pages; exclude most deploy/maintain unless critical.
- For audience "onboarding": short path, max ~25 chapters, skip appendix noise.
- Prefer titles from input; do not rename chapter titles in output (only regroup).
- Write intro in the same language as the majority of chapter titles.
- Output raw JSON only. No markdown fences, no backticks, no commentary before or after.`

	raw, err := e.llm.ChatJSON(ctx, systemPrompt, string(userPayload))
	if err != nil {
		return nil, err
	}
	raw = llm.ExtractJSON(raw)

	var plan aiPlan
	if err := json.Unmarshal([]byte(raw), &plan); err != nil {
		return nil, fmt.Errorf("parse LLM plan: %w (response starts with: %.80q)", err, raw)
	}
	if len(plan.Parts) == 0 {
		return nil, fmt.Errorf("LLM returned empty parts")
	}
	return &plan, nil
}

type chapterBrief struct {
	ID      string `json:"id"`
	Title   string `json:"title"`
	Path    string `json:"path"`
	Kind    string `json:"kind"`
	Snippet string `json:"snippet"`
}

func buildChapterBriefs(preview Preview, idx *Index, maxChapters int) []chapterBrief {
	out := make([]chapterBrief, 0, len(preview.Chapters))
	for _, ch := range preview.Chapters {
		if maxChapters > 0 && len(out) >= maxChapters {
			break
		}
		snippet := ""
		if rec := idx.ByPath[normPath(ch.Path)]; rec != nil {
			body := stripFrontmatter(rec.Content)
			body = strings.Join(strings.Fields(body), " ")
			if len(body) > 160 {
				body = body[:160] + "…"
			}
			snippet = body
		}
		out = append(out, chapterBrief{
			ID: ch.ID, Title: ch.Title, Path: ch.Path, Kind: ch.Kind, Snippet: snippet,
		})
	}
	return out
}

func chapterConstraints(audience string) string {
	switch audience {
	case "ops":
		return "Operations runbook: deployment, monitoring, maintenance, troubleshooting."
	case "architect":
		return "Architecture reference: system context, components, integrations, data flows."
	case "onboarding":
		return "New team member guide: essentials only, readable in under 2 hours."
	default:
		return "Developer handbook: architecture plus practical implementation details."
	}
}

func applyAIPlan(preview Preview, idx *Index, plan *aiPlan, audience string) Preview {
	byID := map[string]Chapter{}
	for _, ch := range preview.Chapters {
		byID[ch.ID] = ch
	}
	excluded := map[string]bool{}
	for _, id := range plan.ExcludeChapterIDs {
		excluded[id] = true
	}

	newChapters := make([]Chapter, 0, len(preview.Chapters))
	used := map[string]bool{}
	for _, part := range plan.Parts {
		for _, id := range part.ChapterIDs {
			if excluded[id] || used[id] {
				continue
			}
			ch, ok := byID[id]
			if !ok {
				continue
			}
			used[id] = true
			ch.Section = part.Title
			ch.Level = 2
			newChapters = append(newChapters, ch)
		}
	}
	// Append any chapters the model missed (stable tail).
	for _, ch := range preview.Chapters {
		if excluded[ch.ID] || used[ch.ID] {
			continue
		}
		ch.Section = "Additional"
		ch.Level = 2
		newChapters = append(newChapters, ch)
	}

	out := preview
	out.Title = plan.Title
	if out.Title == "" {
		out.Title = preview.Title
	}
	out.Description = plan.Description
	if out.Description == "" {
		out.Description = preview.Description
	}
	out.Chapters = newChapters
	out.ChapterCount = len(newChapters)
	out.Strategy = StrategyAI
	out.Enhanced = true
	out.Audience = audience
	out.EnhanceNote = plan.Notes
	out.IntroMarkdown = plan.IntroMarkdown
	out.Markdown = CompileMarkdownEnhanced(out, idx, plan.IntroMarkdown)
	return out
}
