package book

import (
	"regexp"
	"strings"
)

var (
	frontmatterRE = regexp.MustCompile(`(?s)^---\r?\n([\s\S]*?)\r?\n---\r?\n?`)
	wikiLinkRE    = regexp.MustCompile(`\[\[([^\]|]+)(?:\|[^\]]+)?\]\]`)
)

type DocMeta struct {
	Status   string
	C4Level  int
	HasPart  []string
	PartOf   string
	Skipped  bool
	Kind     string // overview, entity, operations, general
}

var opsBasenames = map[string]bool{
	"deploy": true, "maintain": true, "develop": true,
	"development": true, "monitoring": true, "operations": true,
}

func ParseMeta(content, path string) DocMeta {
	meta := DocMeta{Kind: "general"}
	if strings.Contains(strings.ToLower(path), "@ignored") {
		meta.Skipped = true
		return meta
	}

	body := content
	if fm := frontmatterRE.FindStringSubmatch(content); len(fm) == 2 {
		parseFrontmatterBlock(fm[1], &meta)
		body = content[len(fm[0]):]
	}
	parseLeadingMeta(body, &meta)

	base := strings.ToLower(strings.TrimSuffix(pathBase(path), ".md"))
	if opsBasenames[base] {
		meta.Kind = "operations"
	}
	if meta.C4Level == 1 || isOverviewPath(path) {
		meta.Kind = "overview"
	}
	if strings.Contains(pathBase(path), ".") && meta.Kind == "general" {
		meta.Kind = "entity"
	}
	if meta.Status == "ignored" || meta.Status == "draft" {
		meta.Skipped = true
	}
	return meta
}

func parseFrontmatterBlock(block string, meta *DocMeta) {
	inHasPart := false
	for _, line := range strings.Split(block, "\n") {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" {
			inHasPart = false
			continue
		}
		if wikiLinkRE.MatchString(trimmed) || (inHasPart && strings.HasPrefix(trimmed, "-")) {
			inHasPart = true
			for _, m := range wikiLinkRE.FindAllStringSubmatch(trimmed, -1) {
				meta.HasPart = append(meta.HasPart, strings.TrimSpace(m[1]))
			}
			continue
		}
		lower := strings.ToLower(trimmed)
		if strings.HasPrefix(lower, "haspart:") {
			inHasPart = true
			continue
		}
		inHasPart = false
		if strings.HasPrefix(lower, "status:") {
			meta.Status = strings.TrimSpace(trimmed[len("status:"):])
			continue
		}
		if strings.HasPrefix(lower, "c4.level:") {
			meta.C4Level = atoi(strings.TrimSpace(trimmed[len("c4.level:"):]))
			continue
		}
		if strings.HasPrefix(lower, "partof:") {
			val := strings.TrimSpace(trimmed[strings.Index(trimmed, ":")+1:])
			if m := wikiLinkRE.FindStringSubmatch(val); len(m) == 2 {
				meta.PartOf = strings.TrimSpace(m[1])
			} else {
				meta.PartOf = strings.Trim(val, `"`)
			}
		}
	}
}

func parseLeadingMeta(body string, meta *DocMeta) {
	lines := strings.Split(body, "\n")
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" {
			continue
		}
		if strings.HasPrefix(trimmed, "#") {
			break
		}
		lower := strings.ToLower(trimmed)
		if strings.HasPrefix(lower, "status:") && meta.Status == "" {
			meta.Status = strings.TrimSpace(trimmed[len("status:"):])
		}
		if strings.HasPrefix(lower, "c4.level:") && meta.C4Level == 0 {
			meta.C4Level = atoi(strings.TrimSpace(trimmed[len("c4.level:"):]))
		}
	}
}

func isOverviewPath(path string) bool {
	parts := strings.Split(strings.ReplaceAll(path, "\\", "/"), "/")
	if len(parts) != 2 {
		return false
	}
	name := strings.TrimSuffix(parts[1], ".md")
	return strings.EqualFold(name, parts[0])
}

func pathBase(path string) string {
	p := strings.ReplaceAll(path, "\\", "/")
	if i := strings.LastIndex(p, "/"); i >= 0 {
		return p[i+1:]
	}
	return p
}

func atoi(s string) int {
	n := 0
	for _, c := range s {
		if c < '0' || c > '9' {
			break
		}
		n = n*10 + int(c-'0')
	}
	return n
}

func stripFrontmatter(content string) string {
	if fm := frontmatterRE.FindString(content); fm != "" {
		return strings.TrimLeft(content[len(fm):], "\r\n")
	}
	return content
}
