package book

import (
	"fmt"
	"strings"
)

func CompileMarkdownEnhanced(book Preview, idx *Index, intro string) string {
	chapterDocs := make(map[string]*DocRecord, len(book.Chapters))
	for _, ch := range book.Chapters {
		if rec := idx.ByPath[normPath(ch.Path)]; rec != nil {
			chapterDocs[ch.ID] = rec
		}
	}

	var b strings.Builder
	b.WriteString(fmt.Sprintf("# %s\n\n", book.Title))
	b.WriteString(fmt.Sprintf("> %s\n\n", book.Description))
	if book.Enhanced {
		b.WriteString(fmt.Sprintf("*AI-enhanced · audience: %s · %d chapters*\n\n", book.Audience, book.ChapterCount))
		if book.EnhanceNote != "" {
			b.WriteString(fmt.Sprintf("*Editorial note: %s*\n\n", book.EnhanceNote))
		}
	} else {
		b.WriteString(fmt.Sprintf("*Test build — %d chapters from `%s/`*\n\n", book.ChapterCount, book.RootPath))
	}

	if strings.TrimSpace(intro) != "" {
		b.WriteString("## Introduction\n\n")
		b.WriteString(strings.TrimSpace(intro))
		b.WriteString("\n\n---\n\n")
	}

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
