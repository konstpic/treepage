package book

import (
	"regexp"
	"strings"
)

var (
	wikiLinkInlineRE = regexp.MustCompile(`\[\[([^\]|]+)(?:\|[^\]]+)?\]\]`)
	icepanelLineRE   = regexp.MustCompile(`(?im)^icepanel\.\w+:\s*.*$\n?`)
	statusLineRE     = regexp.MustCompile(`(?im)^status:\s*.*$\n?`)
	c4LineRE         = regexp.MustCompile(`(?im)^c4\.level:\s*.*$\n?`)
)

// cleanMarkdown normalizes synced docs for book output without LLM.
func cleanMarkdown(content string) string {
	body := stripFrontmatter(content)
	body = wikiLinkInlineRE.ReplaceAllString(body, "$1")
	body = icepanelLineRE.ReplaceAllString(body, "")
	body = statusLineRE.ReplaceAllString(body, "")
	body = c4LineRE.ReplaceAllString(body, "")
	body = strings.ReplaceAll(body, "\r\n", "\n")
	lines := strings.Split(body, "\n")
	out := make([]string, 0, len(lines))
	prevBlank := false
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" {
			if !prevBlank {
				out = append(out, "")
				prevBlank = true
			}
			continue
		}
		prevBlank = false
		if strings.HasPrefix(trimmed, "tags:") {
			continue
		}
		out = append(out, line)
	}
	return strings.TrimSpace(strings.Join(out, "\n"))
}

func isDevelopPath(path string) bool {
	base := strings.ToLower(strings.TrimSuffix(pathBase(path), ".md"))
	return base == "develop" || base == "development"
}

func isDeployPath(path string) bool {
	return strings.ToLower(strings.TrimSuffix(pathBase(path), ".md")) == "deploy"
}

func isMaintainPath(path string) bool {
	base := strings.ToLower(strings.TrimSuffix(pathBase(path), ".md"))
	return base == "maintain" || base == "monitoring"
}

func primaryLanguage(chapters []Chapter, idx *Index) string {
	cyrillic := 0
	latin := 0
	for _, ch := range chapters {
		rec := idx.ByPath[normPath(ch.Path)]
		if rec == nil {
			continue
		}
		for _, r := range rec.Title + rec.Content {
			if r >= 'а' && r <= 'я' || r >= 'А' && r <= 'Я' {
				cyrillic++
			}
			if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') {
				latin++
			}
		}
	}
	if cyrillic > latin {
		return "ru"
	}
	return "en"
}
