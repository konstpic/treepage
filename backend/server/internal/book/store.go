package book

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/konstpic/treepage/backend/pkg/models"
	"gorm.io/gorm"
)

type CreateBookInput struct {
	RootPath  string `json:"root_path" binding:"required"`
	Title     string `json:"title"`
	Audience  string `json:"audience"`
	Focus     string `json:"focus"`
}

type StoredBookView struct {
	ID              string     `json:"id"`
	Slug            string     `json:"slug"`
	Title           string     `json:"title"`
	Description     string     `json:"description"`
	RootPath        string     `json:"root_path"`
	Audience        string     `json:"audience"`
	Focus           string     `json:"focus"`
	Status          string     `json:"status"`
	SourceHash      string     `json:"source_hash,omitempty"`
	CurrentHash     string     `json:"current_source_hash,omitempty"`
	SourcesStale    bool       `json:"sources_stale"`
	ChapterCount    int        `json:"chapter_count"`
	Enhanced        bool       `json:"enhanced"`
	ErrorMessage    string     `json:"error_message,omitempty"`
	ContentMarkdown string     `json:"content_markdown,omitempty"`
	CreatedAt       time.Time  `json:"created_at"`
	UpdatedAt       time.Time  `json:"updated_at"`
	GeneratedAt     *time.Time `json:"generated_at,omitempty"`
}

type Store struct {
	db *gorm.DB
}

func NewStore(db *gorm.DB) *Store {
	return &Store{db: db}
}

func (s *Store) ListBySpace(ctx context.Context, spaceID string) ([]models.Book, error) {
	var books []models.Book
	err := s.db.WithContext(ctx).
		Where("space_id = ?", spaceID).
		Order("updated_at DESC").
		Find(&books).Error
	return books, err
}

func (s *Store) GetBySlug(ctx context.Context, spaceID, slug string) (*models.Book, error) {
	var b models.Book
	err := s.db.WithContext(ctx).Where("space_id = ? AND slug = ?", spaceID, slug).First(&b).Error
	if err != nil {
		return nil, err
	}
	return &b, nil
}

func (s *Store) Create(ctx context.Context, spaceID string, userID *string, input CreateBookInput, preview Preview) (*models.Book, error) {
	root := strings.Trim(strings.TrimSpace(input.RootPath), "/")
	if root == "" {
		return nil, errors.New("root_path is required")
	}
	title := strings.TrimSpace(input.Title)
	if title == "" {
		title = preview.Title
	}
	audience := strings.TrimSpace(input.Audience)
	if audience == "" {
		audience = "developer"
	}
	slug := slugify(root)
	outline, _ := json.Marshal(preview.Chapters)

	book := models.Book{
		SpaceID:     spaceID,
		Slug:        slug,
		Title:       title,
		Description: preview.Description,
		RootPath:    root,
		Audience:    audience,
		Focus:       strings.TrimSpace(input.Focus),
		Status:      "draft",
		OutlineJSON: string(outline),
		CreatedBy:   userID,
	}
	err := s.db.WithContext(ctx).Create(&book).Error
	if err != nil && strings.Contains(err.Error(), "duplicate") {
		return nil, fmt.Errorf("book already exists for %s", root)
	}
	return &book, err
}

func (s *Store) SaveGenerated(ctx context.Context, book *models.Book, preview Preview, content, sourceHash string, enhanced bool) error {
	outline, _ := json.Marshal(preview.Chapters)
	now := time.Now()
	book.Title = preview.Title
	book.Description = preview.Description
	book.OutlineJSON = string(outline)
	book.ContentMarkdown = content
	book.SourceHash = sourceHash
	book.Enhanced = enhanced
	book.Status = "ready"
	book.ErrorMessage = ""
	book.GeneratedAt = &now
	book.UpdatedAt = now
	return s.db.WithContext(ctx).Save(book).Error
}

func (s *Store) MarkGenerating(ctx context.Context, book *models.Book) error {
	book.Status = "generating"
	book.ErrorMessage = ""
	book.UpdatedAt = time.Now()
	return s.db.WithContext(ctx).Save(book).Error
}

func (s *Store) MarkFailed(ctx context.Context, book *models.Book, msg string) error {
	book.Status = "failed"
	book.ErrorMessage = msg
	book.UpdatedAt = time.Now()
	return s.db.WithContext(ctx).Save(book).Error
}

func (s *Store) Delete(ctx context.Context, spaceID, slug string) error {
	res := s.db.WithContext(ctx).Where("space_id = ? AND slug = ?", spaceID, slug).Delete(&models.Book{})
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

func HashSourceDocs(docs []*DocRecord, root string) string {
	type row struct {
		Path      string
		UpdatedAt string
	}
	rows := make([]row, 0)
	for _, d := range docs {
		if underRoot(d.Path, root) {
			rows = append(rows, row{Path: d.Path, UpdatedAt: d.UpdatedAt.Format(time.RFC3339Nano)})
		}
	}
	sort.Slice(rows, func(i, j int) bool { return rows[i].Path < rows[j].Path })
	h := sha256.New()
	for _, r := range rows {
		h.Write([]byte(r.Path + "|" + r.UpdatedAt + "\n"))
	}
	return hex.EncodeToString(h.Sum(nil))
}

func ToStoredView(b models.Book, currentHash string) StoredBookView {
	var chapters []Chapter
	_ = json.Unmarshal([]byte(b.OutlineJSON), &chapters)
	stale := b.SourceHash != "" && currentHash != "" && b.SourceHash != currentHash
	return StoredBookView{
		ID:              b.ID,
		Slug:            b.Slug,
		Title:           b.Title,
		Description:     b.Description,
		RootPath:        b.RootPath,
		Audience:        b.Audience,
		Focus:           b.Focus,
		Status:          b.Status,
		SourceHash:      b.SourceHash,
		CurrentHash:     currentHash,
		SourcesStale:    stale,
		ChapterCount:    len(chapters),
		Enhanced:        b.Enhanced,
		ErrorMessage:    b.ErrorMessage,
		ContentMarkdown: b.ContentMarkdown,
		CreatedAt:       b.CreatedAt,
		UpdatedAt:       b.UpdatedAt,
		GeneratedAt:     b.GeneratedAt,
	}
}
