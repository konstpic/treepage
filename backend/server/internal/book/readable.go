package book

import (
	"context"
	"fmt"
	"strings"

	"github.com/konstpic/treepage/backend/server/internal/llm"
)

type ReadableGenerator struct {
	llm *llm.Client
}

func NewReadableGenerator(client *llm.Client) *ReadableGenerator {
	return &ReadableGenerator{llm: client}
}

func (r *ReadableGenerator) Available() bool {
	return r.llm != nil && r.llm.Available()
}

// Build produces the final book markdown. Body text is always compiled deterministically
// from source docs (no LLM rewrite) to avoid hallucinations. LLM is used only for a short intro.
func (r *ReadableGenerator) Build(ctx context.Context, preview Preview, idx *Index, audience string) (string, error) {
	lang := primaryLanguage(preview.Chapters, idx)

	var out strings.Builder
	out.WriteString(fmt.Sprintf("# %s\n\n", preview.Title))
	if strings.TrimSpace(preview.Description) != "" {
		out.WriteString(fmt.Sprintf("> %s\n\n", preview.Description))
	}

	intro := strings.TrimSpace(preview.IntroMarkdown)
	if intro == "" && r.Available() && (audience == AudienceDeveloper || audience == AudienceArchitect) {
		generated, err := r.generateIntro(ctx, preview, idx, audience, lang)
		if err == nil {
			intro = generated
		}
	}
	if intro != "" {
		out.WriteString("## Введение\n\n")
		out.WriteString(intro)
		out.WriteString("\n\n---\n\n")
	}

	if audience == AudienceArchitect {
		diagrams := BuildArchitectureDiagrams(preview, idx)
		if diagrams != "" || r.Available() {
			out.WriteString("## Архитектурные диаграммы\n\n")
			if diagrams != "" {
				out.WriteString(diagrams)
				out.WriteString("\n\n")
			}
			if r.Available() {
				if flow, err := r.generateArchitectureFlowchart(ctx, preview, idx, lang); err == nil && flow != "" {
					out.WriteString("### Потоки и интеграции\n\n")
					out.WriteString(flow)
					out.WriteString("\n\n")
				}
			}
			out.WriteString("---\n\n")
		}
	}

	sections := groupChaptersBySection(preview.Chapters)
	for _, sec := range sections {
		body := compileChaptersClean(sec.chapters, idx)
		if strings.TrimSpace(body) == "" {
			continue
		}
		out.WriteString(fmt.Sprintf("## %s\n\n", sec.title))
		out.WriteString(body)
		out.WriteString("\n\n---\n\n")
	}
	return strings.TrimSpace(out.String()) + "\n", nil
}

func (r *ReadableGenerator) generateIntro(ctx context.Context, preview Preview, idx *Index, audience, lang string) (string, error) {
	langRule := "Write in Russian."
	if lang == "en" {
		langRule = "Write in English."
	}
	system := fmt.Sprintf(`You write a short introduction (2-3 paragraphs) for an internal technical book.
%s
Do not invent services, APIs, or endpoints. Only summarize what is listed in the chapter titles.
Output markdown paragraphs only — no headings, no JSON.`, langRule)

	var titles []string
	for _, ch := range preview.Chapters {
		titles = append(titles, ch.Title)
	}
	user := fmt.Sprintf("Book: %s\nAudience: %s\nDescription: %s\nChapters:\n- %s",
		preview.Title, audience, preview.Description, strings.Join(titles, "\n- "))
	return r.llm.Chat(ctx, system, user)
}

func compileChaptersClean(chapters []Chapter, idx *Index) string {
	var b strings.Builder
	for _, ch := range chapters {
		rec := idx.ByPath[normPath(ch.Path)]
		if rec == nil {
			continue
		}
		body := cleanMarkdown(rec.Content)
		body = demoteHeadings(body, 1)
		if strings.TrimSpace(body) == "" {
			continue
		}
		b.WriteString(fmt.Sprintf("### %s\n\n", ch.Title))
		b.WriteString(body)
		b.WriteString("\n\n")
	}
	return strings.TrimSpace(b.String())
}

type sectionGroup struct {
	title    string
	chapters []Chapter
}

func groupChaptersBySection(chapters []Chapter) []sectionGroup {
	order := make([]string, 0)
	groups := map[string][]Chapter{}
	for _, ch := range chapters {
		title := ch.Section
		if title == "" {
			title = "Документация"
		}
		if _, ok := groups[title]; !ok {
			order = append(order, title)
		}
		groups[title] = append(groups[title], ch)
	}
	out := make([]sectionGroup, 0, len(order))
	for _, title := range order {
		out = append(out, sectionGroup{title: title, chapters: groups[title]})
	}
	return out
}
