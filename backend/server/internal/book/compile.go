package book

import (
	"fmt"
	"strings"
)

func CompileMarkdownWithIndex(book Preview, idx *Index) string {
	chapterDocs := make(map[string]*DocRecord, len(book.Chapters))
	for _, ch := range book.Chapters {
		if rec := idx.ByPath[normPath(ch.Path)]; rec != nil {
			chapterDocs[ch.ID] = rec
		}
	}

	var b strings.Builder
	b.WriteString(fmt.Sprintf("# %s\n\n", book.Title))
	b.WriteString(fmt.Sprintf("> %s\n\n", book.Description))
	b.WriteString(fmt.Sprintf("*Test build — %d chapters from `%s/` · strategy: %s*\n\n", book.ChapterCount, book.RootPath, book.Strategy))

	b.WriteString("## Table of Contents\n\n")
	currentSection := ""
	for i, ch := range book.Chapters {
		if ch.Section != "" && ch.Section != currentSection {
			currentSection = ch.Section
			b.WriteString(fmt.Sprintf("\n### %s\n\n", currentSection))
		}
		indent := strings.Repeat("  ", ch.Level-1)
		b.WriteString(fmt.Sprintf("%s%d. %s\n", indent, i+1, ch.Title))
	}
	b.WriteString("\n---\n\n")

	currentSection = ""
	for i, ch := range book.Chapters {
		rec := chapterDocs[ch.ID]
		if rec == nil {
			continue
		}
		if ch.Section != "" && ch.Section != currentSection {
			currentSection = ch.Section
			b.WriteString(fmt.Sprintf("\n---\n\n# %s\n\n", currentSection))
		}
		b.WriteString(fmt.Sprintf("## Chapter %d. %s\n\n", i+1, ch.Title))
		body := stripFrontmatter(rec.Content)
		body = demoteHeadings(body, 1)
		b.WriteString(body)
		b.WriteString("\n\n")
	}
	return strings.TrimSpace(b.String()) + "\n"
}

func demoteHeadings(content string, levels int) string {
	if levels <= 0 {
		return content
	}
	prefix := strings.Repeat("#", levels)
	lines := strings.Split(content, "\n")
	inFence := false
	for i, line := range lines {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "```") {
			inFence = !inFence
			continue
		}
		if inFence {
			continue
		}
		if strings.HasPrefix(line, "#") {
			j := 0
			for j < len(line) && line[j] == '#' {
				j++
			}
			if j < len(line) && line[j] == ' ' {
				lines[i] = prefix + line
			}
		}
	}
	return strings.Join(lines, "\n")
}
