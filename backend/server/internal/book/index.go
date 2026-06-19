package book

import (
	"path/filepath"
	"regexp"
	"strings"

	"github.com/konstpic/treepage/backend/pkg/models"
)

var slugRe = regexp.MustCompile(`[^a-z0-9]+`)

type DocRecord struct {
	models.Document
	Meta DocMeta
}

type Index struct {
	ByPath map[string]*DocRecord
	ByRef  map[string]*DocRecord
	BySlug map[string]*DocRecord
	All    []*DocRecord
}

func BuildIndex(docs []models.Document) *Index {
	idx := &Index{
		ByPath: make(map[string]*DocRecord),
		ByRef:  make(map[string]*DocRecord),
		BySlug: make(map[string]*DocRecord),
	}
	for i := range docs {
		d := docs[i]
		rec := &DocRecord{Document: d, Meta: ParseMeta(d.Content, d.Path)}
		if rec.Meta.Skipped {
			continue
		}
		idx.All = append(idx.All, rec)
		pathKey := normPath(d.Path)
		idx.ByPath[pathKey] = rec
		idx.BySlug[d.Slug] = rec
		registerRefs(pathKey, d.Slug, rec, idx.ByRef)
	}
	return idx
}

func registerRefs(pathKey, slug string, rec *DocRecord, byRef map[string]*DocRecord) {
	refs := []string{
		pathKey,
		strings.TrimSuffix(pathKey, ".md"),
		strings.ReplaceAll(strings.TrimSuffix(pathKey, ".md"), "/", "."),
		strings.ReplaceAll(strings.TrimSuffix(pathKey, ".md"), "/", "-"),
		slug,
		filepath.Base(pathKey),
		strings.TrimSuffix(filepath.Base(pathKey), ".md"),
	}
	for _, ref := range refs {
		if ref != "" {
			byRef[strings.ToLower(ref)] = rec
		}
	}
}

func (idx *Index) Resolve(ref string) *DocRecord {
	key := strings.ToLower(strings.TrimSpace(ref))
	key = strings.TrimPrefix(key, "/")
	key = strings.TrimSuffix(key, ".md")
	candidates := []string{
		key,
		strings.ReplaceAll(key, ".", "/"),
		strings.ReplaceAll(key, "/", "."),
		slugify(key),
		slugify(strings.ReplaceAll(key, ".", "/")),
	}
	for _, c := range candidates {
		if rec := idx.ByRef[c]; rec != nil {
			return rec
		}
		if rec := idx.BySlug[c]; rec != nil {
			return rec
		}
	}
	return nil
}

func normPath(path string) string {
	return strings.ToLower(strings.ReplaceAll(path, "\\", "/"))
}

func slugify(s string) string {
	s = strings.ToLower(strings.TrimSpace(s))
	s = strings.ReplaceAll(s, "/", "-")
	return strings.Trim(slugRe.ReplaceAllString(s, "-"), "-")
}

func pathRoot(path string) string {
	parts := strings.Split(normPath(path), "/")
	if len(parts) == 0 {
		return ""
	}
	return parts[0]
}

func underRoot(path, root string) bool {
	p := normPath(path)
	r := strings.ToLower(root)
	return p == r+".md" || strings.HasPrefix(p, r+"/")
}
