package book

import (
	"fmt"
	"path/filepath"
	"sort"
	"strings"
)

const (
	StrategyGraph  = "graph"
	StrategyFolder = "folder"
)

var skippedRoots = map[string]bool{
	"orphans": true, "prompts": true, "@ignored": true,
}

type Chapter struct {
	ID      string `json:"id"`
	Title   string `json:"title"`
	Slug    string `json:"slug"`
	Path    string `json:"path"`
	Level   int    `json:"level"`
	Kind    string `json:"kind"`
	Section string `json:"section,omitempty"`
}

type Summary struct {
	ID           string `json:"id"`
	Title        string `json:"title"`
	Description  string `json:"description"`
	RootPath     string `json:"root_path"`
	DocCount     int    `json:"doc_count"`
	ChapterCount int    `json:"chapter_count"`
	Strategy     string `json:"strategy"`
}

type Preview struct {
	Summary
	Chapters      []Chapter `json:"chapters"`
	Markdown      string    `json:"markdown,omitempty"`
	Enhanced      bool      `json:"enhanced,omitempty"`
	Audience      string    `json:"audience,omitempty"`
	EnhanceNote   string    `json:"enhance_note,omitempty"`
	IntroMarkdown string    `json:"intro_markdown,omitempty"`
}

type Generator struct {
	minDocsPerBook int
}

func NewGenerator() *Generator {
	return &Generator{minDocsPerBook: 3}
}

func (g *Generator) GenerateAll(idx *Index) []Summary {
	roots := g.discoverRoots(idx)
	summaries := make([]Summary, 0, len(roots))
	for _, root := range roots {
		preview := g.GenerateBook(idx, root)
		if preview.DocCount >= g.minDocsPerBook {
			summaries = append(summaries, preview.Summary)
		}
	}
	sort.Slice(summaries, func(i, j int) bool {
		return summaries[i].Title < summaries[j].Title
	})
	return summaries
}

func (g *Generator) GenerateBook(idx *Index, root string) Preview {
	docs := filterByRoot(idx, root)
	if len(docs) == 0 {
		return Preview{Summary: Summary{ID: root, RootPath: root}}
	}

	overview := findOverview(docs, root)
	chapters := make([]Chapter, 0, len(docs))
	visited := map[string]bool{}
	strategy := StrategyFolder

	if overview != nil {
		chapters = append(chapters, toChapter(overview, 1, "Overview", ""))
		visited[overview.ID] = true
		graphChapters := g.walkGraph(idx, overview, root, visited, 2)
		if len(graphChapters) > 0 {
			strategy = StrategyGraph
			chapters = append(chapters, graphChapters...)
		}
	}

	remaining := make([]*DocRecord, 0)
	for _, d := range docs {
		if !visited[d.ID] {
			remaining = append(remaining, d)
		}
	}
	chapters = append(chapters, g.folderChapters(remaining, root)...)

	title := humanizeRoot(root)
	desc := fmt.Sprintf("Auto-generated book from %d documents under %s/", len(docs), root)
	if overview != nil && overview.Title != "" {
		title = overview.Title
		if overview.Meta.C4Level == 1 {
			desc = fmt.Sprintf("Architecture overview and %d related pages", len(docs))
		}
	}

	preview := Preview{
		Summary: Summary{
			ID:           root,
			Title:        title,
			Description:  desc,
			RootPath:     root,
			DocCount:     len(docs),
			ChapterCount: len(chapters),
			Strategy:     strategy,
		},
		Chapters: chapters,
	}
	preview.Markdown = CompileMarkdownWithIndex(preview, idx)
	return preview
}

func (g *Generator) discoverRoots(idx *Index) []string {
	counts := map[string]int{}
	for _, rec := range idx.All {
		root := pathRoot(rec.Path)
		if root == "" || skippedRoots[strings.ToLower(root)] {
			continue
		}
		counts[root]++
	}
	roots := make([]string, 0, len(counts))
	for root, n := range counts {
		if n >= g.minDocsPerBook {
			roots = append(roots, root)
		}
	}
	sort.Strings(roots)
	return roots
}

func filterByRoot(idx *Index, root string) []*DocRecord {
	out := make([]*DocRecord, 0)
	for _, rec := range idx.All {
		if underRoot(rec.Path, root) {
			out = append(out, rec)
		}
	}
	sort.Slice(out, func(i, j int) bool {
		return out[i].Path < out[j].Path
	})
	return out
}

func findOverview(docs []*DocRecord, root string) *DocRecord {
	target := strings.ToLower(root + "/" + root + ".md")
	for _, d := range docs {
		if normPath(d.Path) == target {
			return d
		}
	}
	var best *DocRecord
	for _, d := range docs {
		if d.Meta.Kind == "overview" || d.Meta.C4Level == 1 {
			if best == nil || depth(d.Path) < depth(best.Path) {
				best = d
			}
		}
	}
	if best != nil {
		return best
	}
	for _, d := range docs {
		if depth(d.Path) == 1 {
			return d
		}
	}
	if len(docs) > 0 {
		return docs[0]
	}
	return nil
}

func (g *Generator) walkGraph(idx *Index, start *DocRecord, root string, visited map[string]bool, level int) []Chapter {
	out := make([]Chapter, 0)
	queue := append([]string{}, start.Meta.HasPart...)
	seen := map[string]bool{}

	for len(queue) > 0 {
		ref := queue[0]
		queue = queue[1:]
		if seen[ref] {
			continue
		}
		seen[ref] = true

		rec := idx.Resolve(ref)
		if rec == nil || visited[rec.ID] {
			continue
		}
		if !underRoot(rec.Path, root) {
			continue
		}
		visited[rec.ID] = true
		section := componentSection(rec.Path, root)
		out = append(out, toChapter(rec, level, section, rec.Meta.Kind))
		for _, child := range rec.Meta.HasPart {
			queue = append(queue, child)
		}
	}
	return out
}

func (g *Generator) folderChapters(docs []*DocRecord, root string) []Chapter {
	type group struct {
		key   string
		docs  []*DocRecord
	}
	groups := map[string]*group{}
	order := make([]string, 0)

	for _, d := range docs {
		key := componentKey(d.Path, root)
		if groups[key] == nil {
			groups[key] = &group{key: key}
			order = append(order, key)
		}
		groups[key].docs = append(groups[key].docs, d)
	}
	sort.Strings(order)

	out := make([]Chapter, 0, len(docs))
	for _, key := range order {
		gd := groups[key]
		sort.Slice(gd.docs, func(i, j int) bool {
			return docSortKey(gd.docs[i]) < docSortKey(gd.docs[j])
		})
		section := humanizeSegment(key)
		if key == "" {
			section = "General"
		}
		for _, d := range gd.docs {
			lvl := 2
			if d.Meta.Kind == "operations" {
				lvl = 3
				section = "Appendix: Operations"
			}
			out = append(out, toChapter(d, lvl, section, d.Meta.Kind))
		}
	}
	return out
}

func toChapter(d *DocRecord, level int, section, kind string) Chapter {
	if kind == "" {
		kind = d.Meta.Kind
	}
	return Chapter{
		ID:      d.ID,
		Title:   d.Title,
		Slug:    d.Slug,
		Path:    d.Path,
		Level:   level,
		Kind:    kind,
		Section: section,
	}
}

func componentKey(path, root string) string {
	parts := strings.Split(normPath(path), "/")
	if len(parts) <= 1 {
		return ""
	}
	if len(parts) >= 3 {
		return parts[1]
	}
	return parts[len(parts)-1]
}

func componentSection(path, root string) string {
	key := componentKey(path, root)
	if key == "" {
		return "Components"
	}
	return humanizeSegment(key)
}

func docSortKey(d *DocRecord) string {
	priority := map[string]int{
		"overview": 0, "entity": 1, "general": 2, "operations": 3,
	}
	p := priority[d.Meta.Kind]
	if p == 0 && d.Meta.Kind == "" {
		p = 2
	}
	base := strings.ToLower(filepath.Base(d.Path))
	return fmt.Sprintf("%02d/%s/%s", p, base, d.Path)
}

func depth(path string) int {
	return len(strings.Split(normPath(path), "/"))
}

func humanizeRoot(root string) string {
	return humanizeSegment(root)
}

func humanizeSegment(s string) string {
	s = strings.ReplaceAll(s, "-", " ")
	s = strings.ReplaceAll(s, "_", " ")
	if s == "" {
		return s
	}
	return strings.ToUpper(s[:1]) + s[1:]
}
