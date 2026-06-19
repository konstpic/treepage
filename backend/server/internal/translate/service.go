package translate

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strings"
	"unicode"

	"github.com/konstpic/treepage/backend/pkg/models"
	"github.com/konstpic/treepage/backend/server/internal/llm"
	"github.com/konstpic/treepage/backend/server/internal/service"
	"gorm.io/gorm"
)

type Service struct {
	db    *gorm.DB
	llm   *llm.Client
	admin *service.AdminService
}

func NewService(db *gorm.DB, llmClient *llm.Client, admin *service.AdminService) *Service {
	return &Service{db: db, llm: llmClient, admin: admin}
}

type DocumentView struct {
	models.Document
	Translated      bool   `json:"translated,omitempty"`
	SourceLanguage  string `json:"source_language,omitempty"`
	DisplayLanguage string `json:"display_language,omitempty"`
}

func (s *Service) LocalizeDocument(ctx context.Context, doc *models.Document, targetLang string, allowLLM bool) (*DocumentView, error) {
	targetLang = normalizeLang(targetLang)
	view := &DocumentView{Document: *doc}

	if targetLang == "" {
		return view, nil
	}
	view.DisplayLanguage = targetLang
	view.SourceLanguage = detectLanguage(doc.Title + "\n" + doc.Content)

	if view.SourceLanguage == targetLang {
		return view, nil
	}
	if !allowLLM || !s.autoTranslateEnabled(ctx) || !s.llm.Available() {
		return view, nil
	}

	hash := hashContent(doc.Content)
	var cached models.DocumentTranslation
	err := s.db.WithContext(ctx).
		Where("document_id = ? AND locale = ? AND source_hash = ?", doc.ID, targetLang, hash).
		First(&cached).Error
	if err == nil {
		view.Title = cached.Title
		view.Content = cached.Content
		view.Translated = true
		return view, nil
	}
	if err != nil && err != gorm.ErrRecordNotFound {
		return view, err
	}

	title, content, err := s.translateDocument(ctx, doc.Title, doc.Content, view.SourceLanguage, targetLang)
	if err != nil {
		return view, nil
	}

	row := models.DocumentTranslation{
		DocumentID: doc.ID,
		Locale:     targetLang,
		SourceHash: hash,
		Title:      title,
		Content:    content,
	}
	_ = s.db.WithContext(ctx).
		Where("document_id = ? AND locale = ?", doc.ID, targetLang).
		Assign(models.DocumentTranslation{SourceHash: hash, Title: title, Content: content}).
		FirstOrCreate(&row).Error

	view.Title = title
	view.Content = content
	view.Translated = true
	return view, nil
}

type BookLocalization struct {
	Title           string
	Description     string
	Content         string
	Translated      bool   `json:"translated,omitempty"`
	SourceLanguage  string `json:"source_language,omitempty"`
	DisplayLanguage string `json:"display_language,omitempty"`
}

func (s *Service) LocalizeBook(ctx context.Context, bookID, title, description, content, targetLang string, allowLLM bool) (*BookLocalization, error) {
	targetLang = normalizeLang(targetLang)
	out := &BookLocalization{
		Title:       title,
		Description: description,
		Content:     content,
	}

	if targetLang == "" || content == "" {
		return out, nil
	}
	out.DisplayLanguage = targetLang
	out.SourceLanguage = detectLanguage(title + "\n" + description + "\n" + content)

	if out.SourceLanguage == targetLang {
		return out, nil
	}
	if !allowLLM || !s.autoTranslateEnabled(ctx) || !s.llm.Available() {
		return out, nil
	}

	hash := hashContent(content)
	var cached models.BookTranslation
	err := s.db.WithContext(ctx).
		Where("book_id = ? AND locale = ? AND source_hash = ?", bookID, targetLang, hash).
		First(&cached).Error
	if err == nil {
		out.Title = cached.Title
		out.Description = cached.Description
		out.Content = cached.Content
		out.Translated = true
		return out, nil
	}
	if err != nil && err != gorm.ErrRecordNotFound {
		return out, err
	}

	t, desc, body, err := s.translateBook(ctx, title, description, content, out.SourceLanguage, targetLang)
	if err != nil {
		return out, nil
	}

	row := models.BookTranslation{
		BookID:      bookID,
		Locale:      targetLang,
		SourceHash:  hash,
		Title:       t,
		Description: desc,
		Content:     body,
	}
	_ = s.db.WithContext(ctx).
		Where("book_id = ? AND locale = ?", bookID, targetLang).
		Assign(models.BookTranslation{SourceHash: hash, Title: t, Description: desc, Content: body}).
		FirstOrCreate(&row).Error

	out.Title = t
	out.Description = desc
	out.Content = body
	out.Translated = true
	return out, nil
}

func (s *Service) translateBook(ctx context.Context, title, description, content, from, to string) (string, string, string, error) {
	targetName := langName(to)
	sourceName := langName(from)
	system := fmt.Sprintf(`You translate internal technical documentation books from %s to %s.
Rules:
- Preserve Markdown structure (headings, lists, tables, links)
- Do NOT translate content inside fenced code blocks or mermaid diagrams
- Keep [[wiki links]] and URLs unchanged
- Do not add or remove sections
- Output format exactly:
TITLE: <translated title>
DESCRIPTION: <translated description>
---
<translated markdown body>`, sourceName, targetName)

	const maxChunk = 6000
	if len(content) <= maxChunk {
		user := fmt.Sprintf("TITLE: %s\nDESCRIPTION: %s\n\nBODY:\n%s", title, description, content)
		raw, err := s.llm.Chat(ctx, system, user)
		if err != nil {
			return "", "", "", err
		}
		t, desc, body, err := parseBookTranslationResponse(raw, title, description, content)
		if err != nil {
			return "", "", "", err
		}
		return t, desc, body, nil
	}

	translatedTitle := title
	translatedDesc := description
	var parts []string
	for i, chunk := range splitBySections(content, maxChunk) {
		user := fmt.Sprintf("TITLE: %s\nDESCRIPTION: %s\n\nBODY:\n%s", translatedTitle, translatedDesc, chunk)
		raw, err := s.llm.Chat(ctx, system, user)
		if err != nil {
			return "", "", "", err
		}
		t, desc, body, err := parseBookTranslationResponse(raw, translatedTitle, translatedDesc, chunk)
		if err != nil {
			return "", "", "", err
		}
		if i == 0 {
			if t != "" {
				translatedTitle = t
			}
			if desc != "" {
				translatedDesc = desc
			}
		}
		parts = append(parts, body)
	}
	return translatedTitle, translatedDesc, strings.Join(parts, "\n\n"), nil
}

func parseBookTranslationResponse(raw, fallbackTitle, fallbackDesc, fallbackBody string) (string, string, string, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return fallbackTitle, fallbackDesc, fallbackBody, fmt.Errorf("empty translation")
	}
	if i := strings.Index(raw, "\n---\n"); i >= 0 {
		head := strings.TrimSpace(raw[:i])
		body := strings.TrimSpace(raw[i+5:])
		title := fallbackTitle
		desc := fallbackDesc
		for _, line := range strings.Split(head, "\n") {
			line = strings.TrimSpace(line)
			upper := strings.ToUpper(line)
			if strings.HasPrefix(upper, "TITLE:") {
				title = strings.TrimSpace(line[6:])
			}
			if strings.HasPrefix(upper, "DESCRIPTION:") {
				desc = strings.TrimSpace(line[12:])
			}
		}
		if body == "" {
			body = fallbackBody
		}
		return title, desc, body, nil
	}
	return fallbackTitle, fallbackDesc, raw, nil
}

func (s *Service) autoTranslateEnabled(ctx context.Context) bool {
	settings, err := s.admin.GetSystemSettings(ctx)
	if err != nil {
		return false
	}
	return platformBool(settings.Platform, "auto_translate_docs", false)
}

func platformBool(platform map[string]any, key string, defaultVal bool) bool {
	v, ok := platform[key]
	if !ok {
		return defaultVal
	}
	switch b := v.(type) {
	case bool:
		return b
	case string:
		return strings.EqualFold(strings.TrimSpace(b), "true") || b == "1"
	case float64:
		return b != 0
	default:
		return defaultVal
	}
}

func (s *Service) translateDocument(ctx context.Context, title, content, from, to string) (string, string, error) {
	targetName := langName(to)
	sourceName := langName(from)
	system := fmt.Sprintf(`You translate internal technical documentation from %s to %s.
Rules:
- Preserve Markdown structure (headings, lists, tables, links)
- Do NOT translate content inside fenced code blocks or mermaid diagrams
- Keep [[wiki links]] and URLs unchanged
- Do not add or remove sections
- Output format exactly:
TITLE: <translated title>
---
<translated markdown body>`, sourceName, targetName)

	const maxChunk = 6000
	if len(content) <= maxChunk {
		user := fmt.Sprintf("TITLE: %s\n\nBODY:\n%s", title, content)
		raw, err := s.llm.Chat(ctx, system, user)
		if err != nil {
			return "", "", err
		}
		return parseTranslationResponse(raw, title, content)
	}

	translatedTitle := title
	var parts []string
	for i, chunk := range splitBySections(content, maxChunk) {
		user := fmt.Sprintf("TITLE: %s\n\nBODY:\n%s", translatedTitle, chunk)
		raw, err := s.llm.Chat(ctx, system, user)
		if err != nil {
			return "", "", err
		}
		t, body, err := parseTranslationResponse(raw, translatedTitle, chunk)
		if err != nil {
			return "", "", err
		}
		if i == 0 && t != "" {
			translatedTitle = t
		}
		parts = append(parts, body)
	}
	return translatedTitle, strings.Join(parts, "\n\n"), nil
}

func parseTranslationResponse(raw, fallbackTitle, fallbackBody string) (string, string, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return fallbackTitle, fallbackBody, fmt.Errorf("empty translation")
	}
	if i := strings.Index(raw, "\n---\n"); i >= 0 {
		head := strings.TrimSpace(raw[:i])
		body := strings.TrimSpace(raw[i+5:])
		title := fallbackTitle
		if strings.HasPrefix(strings.ToUpper(head), "TITLE:") {
			title = strings.TrimSpace(head[6:])
		}
		if body == "" {
			body = fallbackBody
		}
		return title, body, nil
	}
	if strings.HasPrefix(strings.ToUpper(raw), "TITLE:") {
		lines := strings.SplitN(raw, "\n", 2)
		title := strings.TrimSpace(strings.TrimPrefix(lines[0], "TITLE:"))
		body := fallbackBody
		if len(lines) > 1 {
			body = strings.TrimSpace(lines[1])
		}
		return title, body, nil
	}
	return fallbackTitle, raw, nil
}

func splitBySections(content string, maxLen int) []string {
	if len(content) <= maxLen {
		return []string{content}
	}
	lines := strings.Split(content, "\n")
	var chunks []string
	var cur strings.Builder
	for _, line := range lines {
		if cur.Len() > 0 && strings.HasPrefix(line, "## ") && cur.Len() >= maxLen/2 {
			chunks = append(chunks, strings.TrimSpace(cur.String()))
			cur.Reset()
		}
		cur.WriteString(line)
		cur.WriteString("\n")
		if cur.Len() >= maxLen {
			chunks = append(chunks, strings.TrimSpace(cur.String()))
			cur.Reset()
		}
	}
	if cur.Len() > 0 {
		chunks = append(chunks, strings.TrimSpace(cur.String()))
	}
	if len(chunks) == 0 {
		return []string{content}
	}
	return chunks
}

func normalizeLang(lang string) string {
	lang = strings.ToLower(strings.TrimSpace(lang))
	if i := strings.IndexAny(lang, ",;"); i >= 0 {
		lang = lang[:i]
	}
	if len(lang) > 2 && strings.Contains(lang, "-") {
		lang = lang[:strings.Index(lang, "-")]
	}
	switch lang {
	case "en", "ru":
		return lang
	default:
		return ""
	}
}

func detectLanguage(text string) string {
	cyrillic, latin := 0, 0
	for _, r := range text {
		if r >= 'а' && r <= 'я' || r >= 'А' && r <= 'Я' || r == 'ё' || r == 'Ё' {
			cyrillic++
		}
		if unicode.IsLetter(r) && r < 128 {
			latin++
		}
	}
	if cyrillic == 0 && latin == 0 {
		return "en"
	}
	if cyrillic > latin {
		return "ru"
	}
	return "en"
}

func langName(code string) string {
	if code == "ru" {
		return "Russian"
	}
	return "English"
}

func hashContent(content string) string {
	sum := sha256.Sum256([]byte(content))
	return hex.EncodeToString(sum[:])
}
