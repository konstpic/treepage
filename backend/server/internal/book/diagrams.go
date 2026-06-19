package book

import (
	"context"
	"fmt"
	"regexp"
	"sort"
	"strings"
	"unicode"
)

const maxDiagramNodes = 20

var mermaidFenceRE = regexp.MustCompile("(?is)^```(?:mermaid)?\\s*([\\s\\S]*?)```$")

// BuildArchitectureDiagrams returns markdown blocks (mermaid fences) for architect books.
func BuildArchitectureDiagrams(preview Preview, idx *Index) string {
	var parts []string
	if block := buildHasPartFlowchart(preview, idx); block != "" {
		parts = append(parts, "### Карта компонентов\n\n"+block)
	}
	if block := buildSectionFlowchart(preview); block != "" && len(parts) == 0 {
		parts = append(parts, "### Структура документации\n\n"+block)
	}
	return strings.Join(parts, "\n\n")
}

func (r *ReadableGenerator) generateArchitectureFlowchart(
	_ context.Context,
	preview Preview,
	idx *Index,
	_ string,
) (string, error) {
	brief := buildDiagramBrief(preview, idx)
	block := buildIntegrationFlowchart(brief)
	if block == "" {
		return "", nil
	}
	return block, nil
}

type diagramBrief struct {
	Components []string
	Edges      []string
}

func (d diagramBrief) String() string {
	var b strings.Builder
	for _, c := range d.Components {
		b.WriteString("- ")
		b.WriteString(c)
		b.WriteString("\n")
	}
	if len(d.Edges) > 0 {
		b.WriteString("Relations:\n")
		for _, e := range d.Edges {
			b.WriteString("- ")
			b.WriteString(e)
			b.WriteString("\n")
		}
	}
	return b.String()
}

func buildDiagramBrief(preview Preview, idx *Index) diagramBrief {
	inBook := chapterIDSet(preview)
	seen := map[string]bool{}
	var components []string
	var edges []string

	addComp := func(title string) {
		title = strings.TrimSpace(title)
		if title == "" || seen[title] {
			return
		}
		seen[title] = true
		components = append(components, title)
	}

	for _, ch := range preview.Chapters {
		rec := idx.ByPath[normPath(ch.Path)]
		if rec == nil {
			continue
		}
		addComp(ch.Title)
		for _, ref := range rec.Meta.HasPart {
			target := idx.Resolve(ref)
			if target == nil || !inBook[target.ID] {
				continue
			}
			edges = append(edges, fmt.Sprintf("%s --> %s", ch.Title, target.Title))
		}
		if rec.Meta.PartOf != "" {
			parent := idx.Resolve(rec.Meta.PartOf)
			if parent != nil && inBook[parent.ID] {
				edges = append(edges, fmt.Sprintf("%s --> %s", parent.Title, ch.Title))
			}
		}
	}

	sort.Strings(components)
	sort.Strings(edges)
	if len(components) > maxDiagramNodes {
		components = components[:maxDiagramNodes]
	}
	if len(edges) > 30 {
		edges = edges[:30]
	}
	return diagramBrief{Components: components, Edges: edges}
}

func chapterIDSet(preview Preview) map[string]bool {
	m := make(map[string]bool, len(preview.Chapters))
	for _, ch := range preview.Chapters {
		m[ch.ID] = true
	}
	return m
}

func buildHasPartFlowchart(preview Preview, idx *Index) string {
	inBook := chapterIDSet(preview)
	root := findDiagramRoot(preview, idx)
	if root == nil {
		return ""
	}
	recs, edgePairs := collectBoundedGraph(root, preview, idx, inBook, maxDiagramNodes)
	if len(recs) < 2 || len(edgePairs) == 0 {
		return ""
	}
	return renderFlowchart(recs, edgePairs)
}

func findDiagramRoot(preview Preview, idx *Index) *DocRecord {
	var best *DocRecord
	for _, ch := range preview.Chapters {
		rec := idx.ByPath[normPath(ch.Path)]
		if rec == nil {
			continue
		}
		if rec.Meta.Kind == "overview" || rec.Meta.C4Level <= 1 {
			if best == nil || depth(rec.Path) < depth(best.Path) {
				best = rec
			}
		}
	}
	if best != nil {
		return best
	}
	if len(preview.Chapters) > 0 {
		return idx.ByPath[normPath(preview.Chapters[0].Path)]
	}
	return nil
}

type diagramEdge struct {
	from string
	to   string
}

func collectBoundedGraph(
	root *DocRecord,
	preview Preview,
	idx *Index,
	inBook map[string]bool,
	max int,
) ([]*DocRecord, []diagramEdge) {
	inPreview := map[string]*DocRecord{}
	for _, ch := range preview.Chapters {
		if rec := idx.ByPath[normPath(ch.Path)]; rec != nil {
			inPreview[rec.ID] = rec
		}
	}

	seen := map[string]bool{root.ID: true}
	queue := []*DocRecord{root}
	recs := []*DocRecord{root}
	var edges []diagramEdge

	for len(queue) > 0 && len(recs) < max {
		cur := queue[0]
		queue = queue[1:]

		neighbors := append([]string{}, cur.Meta.HasPart...)
		if cur.Meta.PartOf != "" {
			neighbors = append(neighbors, cur.Meta.PartOf)
		}
		for _, ch := range preview.Chapters {
			rec := inPreview[ch.ID]
			if rec == nil {
				continue
			}
			if rec.Meta.PartOf != "" {
				if p := idx.Resolve(rec.Meta.PartOf); p != nil && p.ID == cur.ID {
					neighbors = append(neighbors, rec.Title)
				}
			}
		}

		for _, ref := range neighbors {
			target := idx.Resolve(ref)
			if target == nil || !inBook[target.ID] {
				continue
			}
			if inPreview[target.ID] == nil {
				continue
			}
			edges = append(edges, diagramEdge{from: cur.ID, to: target.ID})
			if !seen[target.ID] && len(recs) < max {
				seen[target.ID] = true
				recs = append(recs, target)
				queue = append(queue, target)
			}
		}
	}

	if len(recs) < 2 {
		return nil, nil
	}
	return recs, edges
}

func renderFlowchart(recs []*DocRecord, edges []diagramEdge) string {
	idFor := map[string]string{}
	labelFor := map[string]string{}
	for i, rec := range recs {
		id := fmt.Sprintf("n%d", i)
		idFor[rec.ID] = id
		label := rec.Title
		if label == "" {
			label = humanizeSegment(pathBase(rec.Path))
		}
		labelFor[rec.ID] = mermaidEscape(label)
	}

	var b strings.Builder
	b.WriteString("```mermaid\nflowchart TB\n")
	for _, rec := range recs {
		b.WriteString(fmt.Sprintf("  %s(\"%s\")\n", idFor[rec.ID], labelFor[rec.ID]))
	}
	seen := map[string]bool{}
	for _, e := range edges {
		from, ok1 := idFor[e.from]
		to, ok2 := idFor[e.to]
		if !ok1 || !ok2 || from == to {
			continue
		}
		key := from + "->" + to
		if seen[key] {
			continue
		}
		seen[key] = true
		b.WriteString(fmt.Sprintf("  %s --> %s\n", from, to))
	}
	b.WriteString("```")
	return b.String()
}

func buildIntegrationFlowchart(brief diagramBrief) string {
	if len(brief.Components) < 2 {
		return ""
	}
	idByLabel := map[string]string{}
	for i, label := range brief.Components {
		id := fmt.Sprintf("c%d", i)
		idByLabel[strings.ToLower(label)] = id
		idByLabel[label] = id
	}

	var b strings.Builder
	b.WriteString("```mermaid\nflowchart LR\n")
	for i, label := range brief.Components {
		b.WriteString(fmt.Sprintf("  c%d(\"%s\")\n", i, mermaidEscape(label)))
	}

	seen := map[string]bool{}
	addEdge := func(from, to string) {
		if from == "" || to == "" || from == to {
			return
		}
		key := from + "->" + to
		if seen[key] {
			return
		}
		seen[key] = true
		b.WriteString(fmt.Sprintf("  %s --> %s\n", from, to))
	}

	for _, e := range brief.Edges {
		parts := strings.Split(e, "-->")
		if len(parts) != 2 {
			continue
		}
		from := resolveBriefID(idByLabel, strings.TrimSpace(parts[0]))
		to := resolveBriefID(idByLabel, strings.TrimSpace(parts[1]))
		addEdge(from, to)
	}
	if len(seen) == 0 {
		for i := 0; i+1 < len(brief.Components) && i < 12; i++ {
			addEdge(fmt.Sprintf("c%d", i), fmt.Sprintf("c%d", i+1))
		}
	}
	b.WriteString("```")
	return b.String()
}

func resolveBriefID(idByLabel map[string]string, label string) string {
	if id, ok := idByLabel[label]; ok {
		return id
	}
	if id, ok := idByLabel[strings.ToLower(label)]; ok {
		return id
	}
	return ""
}

func buildSectionFlowchart(preview Preview) string {
	type section struct {
		title string
		chs   []string
	}
	sections := map[string]*section{}
	order := make([]string, 0)

	for _, ch := range preview.Chapters {
		title := ch.Section
		if title == "" {
			title = "Компоненты"
		}
		if sections[title] == nil {
			sections[title] = &section{title: title}
			order = append(order, title)
		}
		if len(sections[title].chs) < 4 {
			sections[title].chs = append(sections[title].chs, ch.Title)
		}
	}
	if len(order) < 2 {
		return ""
	}

	var b strings.Builder
	b.WriteString("```mermaid\nflowchart TB\n")
	rootID := "root"
	b.WriteString(fmt.Sprintf("  %s(\"%s\")\n", rootID, mermaidEscape(preview.Title)))
	for i, key := range order {
		secID := fmt.Sprintf("sec%d", i)
		firstChild := ""
		b.WriteString(fmt.Sprintf("  subgraph %s [\"%s\"]\n", secID, mermaidEscape(key)))
		for j, chTitle := range sections[key].chs {
			nodeID := fmt.Sprintf("%s_n%d", secID, j)
			if j == 0 {
				firstChild = nodeID
			}
			b.WriteString(fmt.Sprintf("    %s(\"%s\")\n", nodeID, mermaidEscape(chTitle)))
		}
		b.WriteString("  end\n")
		if firstChild != "" {
			b.WriteString(fmt.Sprintf("  %s --> %s\n", rootID, firstChild))
		}
	}
	b.WriteString("```")
	return b.String()
}

func mermaidNodeID(seed string) string {
	var b strings.Builder
	for _, r := range seed {
		if unicode.IsLetter(r) || unicode.IsDigit(r) {
			b.WriteRune(unicode.ToLower(r))
		} else if r == '-' || r == '_' {
			b.WriteRune(r)
		}
		if b.Len() >= 12 {
			break
		}
	}
	if b.Len() == 0 {
		return "n"
	}
	return "n_" + b.String()
}

func mermaidEscape(s string) string {
	replacer := strings.NewReplacer(
		`"`, "'",
		"'", " ",
		"\n", " ",
		"\r", " ",
		"#", "",
		";", ",",
		"(", " ",
		")", " ",
		"[", " ",
		"]", " ",
		"{", " ",
		"}", " ",
		"|", " ",
		"<", " ",
		">", " ",
		"&", " and ",
		"`", "",
		"\\", "",
	)
	s = replacer.Replace(s)
	s = strings.Join(strings.Fields(s), " ")
	if len([]rune(s)) > 48 {
		runes := []rune(s)
		s = string(runes[:45]) + "..."
	}
	return strings.TrimSpace(s)
}

func wrapMermaid(code string) string {
	code = strings.TrimSpace(code)
	if strings.HasPrefix(code, "```") {
		return code
	}
	return "```mermaid\n" + code + "\n```"
}

func sanitizeMermaid(raw string) string {
	s := strings.TrimSpace(raw)
	if m := mermaidFenceRE.FindStringSubmatch(s); len(m) == 2 {
		s = strings.TrimSpace(m[1])
	} else if i := strings.Index(s, "```"); i >= 0 {
		rest := s[i+3:]
		rest = strings.TrimLeft(rest, " \t\r\n")
		if strings.HasPrefix(strings.ToLower(rest), "mermaid") {
			rest = rest[7:]
			rest = strings.TrimLeft(rest, " \t\r\n")
		}
		if j := strings.Index(rest, "```"); j >= 0 {
			s = strings.TrimSpace(rest[:j])
		}
	}

	lines := strings.Split(s, "\n")
	start := -1
	for i, line := range lines {
		trimmed := strings.TrimSpace(strings.ToLower(line))
		if strings.HasPrefix(trimmed, "flowchart") || strings.HasPrefix(trimmed, "graph ") {
			start = i
			break
		}
	}
	if start < 0 {
		return ""
	}
	return strings.TrimSpace(strings.Join(lines[start:], "\n"))
}
