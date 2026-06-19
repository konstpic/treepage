package book

import (
	"fmt"
	"sort"
	"strings"
)

const (
	AudienceOps         = "ops"
	AudienceArchitect   = "architect"
	AudienceDeveloper   = "developer"
	AudienceOnboarding  = "onboarding"
)

// ApplyAudience filters and structures chapters for the target reader.
func ApplyAudience(preview Preview, idx *Index, audience string) Preview {
	if audience == "" {
		audience = AudienceDeveloper
	}
	chapters := filterChaptersForAudience(preview.Chapters, idx, audience, preview.RootPath)
	chapters = supplementChapters(preview, idx, audience, chapters)
	chapters = structureChapters(chapters, idx, audience)

	out := preview
	out.Chapters = chapters
	out.ChapterCount = len(chapters)
	out.Strategy = "audience"

	switch audience {
	case AudienceOps:
		out.Title = fmt.Sprintf("Рунбук: %s", humanizeRoot(preview.RootPath))
		out.Description = fmt.Sprintf(
			"Эксплуатация и развёртывание компонентов %s: deploy, maintain, диагностика.",
			humanizeRoot(preview.RootPath),
		)
	case AudienceArchitect:
		out.Title = fmt.Sprintf("Архитектура: %s", humanizeRoot(preview.RootPath))
		out.Description = fmt.Sprintf("Обзор системы и компонентов %s.", humanizeRoot(preview.RootPath))
	case AudienceOnboarding:
		out.Title = fmt.Sprintf("Введение: %s", humanizeRoot(preview.RootPath))
		out.Description = fmt.Sprintf("Краткий обзор %s для новых участников команды.", humanizeRoot(preview.RootPath))
	}
	return out
}

func filterChaptersForAudience(chapters []Chapter, idx *Index, audience, rootPath string) []Chapter {
	out := make([]Chapter, 0, len(chapters))
	for _, ch := range chapters {
		rec := idx.ByPath[normPath(ch.Path)]
		if rec == nil {
			continue
		}
		if rec.Meta.Skipped {
			continue
		}
		if includeChapter(rec, idx, audience, rootPath) {
			out = append(out, ch)
		}
	}
	return out
}

func includeChapter(rec *DocRecord, idx *Index, audience, rootPath string) bool {
	switch audience {
	case AudienceOps:
		if isDevelopPath(rec.Path) {
			return false
		}
		if isDeployPath(rec.Path) || isMaintainPath(rec.Path) {
			return true
		}
		if rec.Meta.Kind == "overview" {
			return true
		}
		if isOpsContextOverview(rec.Path) && componentHasOpsDocs(idx, rec.Path, rootPath) {
			return true
		}
		return false
	case AudienceArchitect:
		if isDeployPath(rec.Path) || isMaintainPath(rec.Path) || isDevelopPath(rec.Path) {
			return false
		}
		return rec.Meta.Kind == "overview" || rec.Meta.Kind == "entity" || rec.Meta.Kind == "general"
	case AudienceOnboarding:
		if isDeployPath(rec.Path) || isMaintainPath(rec.Path) || isDevelopPath(rec.Path) {
			return false
		}
		return rec.Meta.Kind == "overview" || rec.Meta.C4Level <= 2 || depth(rec.Path) <= 3
	default:
		return true
	}
}

// supplementChapters adds deploy/maintain pages under root that the graph walk may have missed.
func supplementChapters(preview Preview, idx *Index, audience string, chapters []Chapter) []Chapter {
	if audience != AudienceOps {
		return chapters
	}
	have := make(map[string]bool, len(chapters))
	for _, ch := range chapters {
		have[normPath(ch.Path)] = true
	}
	for _, rec := range idx.All {
		if !underRoot(rec.Path, preview.RootPath) || rec.Meta.Skipped {
			continue
		}
		if !includeChapter(rec, idx, audience, preview.RootPath) {
			continue
		}
		p := normPath(rec.Path)
		if have[p] {
			continue
		}
		chapters = append(chapters, toChapter(rec, 2, "", rec.Meta.Kind))
		have[p] = true
	}
	return chapters
}

func structureChapters(chapters []Chapter, idx *Index, audience string) []Chapter {
	if audience != AudienceOps {
		return assignGenericSections(chapters, audience)
	}
	type bucket struct {
		title string
		order int
		chs   []Chapter
	}
	buckets := map[string]*bucket{
		"context":    {title: "Контекст", order: 1},
		"deploy":     {title: "Развёртывание", order: 2},
		"maintain":   {title: "Эксплуатация и мониторинг", order: 3},
	}

	for _, ch := range chapters {
		rec := idx.ByPath[normPath(ch.Path)]
		if rec == nil {
			continue
		}
		ch.Level = 2
		switch {
		case isDeployPath(rec.Path):
			ch.Section = buckets["deploy"].title
			buckets["deploy"].chs = append(buckets["deploy"].chs, ch)
		case isMaintainPath(rec.Path):
			ch.Section = buckets["maintain"].title
			buckets["maintain"].chs = append(buckets["maintain"].chs, ch)
		default:
			ch.Section = buckets["context"].title
			buckets["context"].chs = append(buckets["context"].chs, ch)
		}
	}

	keys := []string{"context", "deploy", "maintain"}
	sort.Slice(keys, func(i, j int) bool { return buckets[keys[i]].order < buckets[keys[j]].order })

	out := make([]Chapter, 0, len(chapters))
	for _, k := range keys {
		chs := buckets[k].chs
		sort.Slice(chs, func(i, j int) bool { return chs[i].Path < chs[j].Path })
		out = append(out, chs...)
	}
	return out
}

func assignGenericSections(chapters []Chapter, audience string) []Chapter {
	out := make([]Chapter, 0, len(chapters))
	limit := len(chapters)
	if audience == AudienceOnboarding && limit > 22 {
		limit = 22
	}
	for i, ch := range chapters {
		if i >= limit {
			break
		}
		switch {
		case isDeployPath(ch.Path):
			ch.Section = "Развёртывание"
		case isMaintainPath(ch.Path):
			ch.Section = "Эксплуатация"
		case ch.Kind == "overview" || strings.Contains(ch.Path, ".md") && depth(ch.Path) <= 2:
			ch.Section = "Обзор"
		default:
			ch.Section = "Компоненты"
		}
		ch.Level = 2
		out = append(out, ch)
	}
	return out
}

// isOpsContextOverview includes short component intros (e.g. api/dataapi/api.dataapi.md)
// but not deep entity pages (e.g. api/.../rabbitmq/api.dataapi.rabbitmq.md).
func isOpsContextOverview(path string) bool {
	if isDeployPath(path) || isMaintainPath(path) || isDevelopPath(path) {
		return false
	}
	base := strings.TrimSuffix(pathBase(path), ".md")
	if strings.Count(base, ".") > 1 {
		return false
	}
	d := depth(path)
	return d >= 2 && d <= 3
}

func componentHasOpsDocs(idx *Index, path, root string) bool {
	if depth(path) == 2 {
		return true
	}
	comp := componentFromPath(path)
	if comp == "" {
		return false
	}
	prefix := normPath(root) + "/" + comp + "/"
	for _, r := range idx.All {
		if !strings.HasPrefix(normPath(r.Path), prefix) {
			continue
		}
		if isDeployPath(r.Path) || isMaintainPath(r.Path) {
			return true
		}
	}
	return false
}

func componentFromPath(path string) string {
	parts := strings.Split(normPath(path), "/")
	if len(parts) >= 2 {
		return parts[len(parts)-2]
	}
	return ""
}
