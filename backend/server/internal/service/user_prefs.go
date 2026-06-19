package service

import (
	"context"
	"time"

	"github.com/konstpic/treepage/backend/pkg/models"
	"gorm.io/gorm"
)

type UserPrefsService struct {
	db *gorm.DB
}

func NewUserPrefsService(db *gorm.DB) *UserPrefsService {
	return &UserPrefsService{db: db}
}

type FavoriteItem struct {
	DocumentID string    `json:"document_id"`
	SpaceID    string    `json:"space_id"`
	SpaceSlug  string    `json:"space_slug"`
	SpaceName  string    `json:"space_name"`
	DocSlug    string    `json:"doc_slug"`
	Title      string    `json:"title"`
	CreatedAt  time.Time `json:"created_at"`
}

func (s *UserPrefsService) ListFavorites(ctx context.Context, userID string) ([]FavoriteItem, error) {
	var rows []struct {
		DocumentID string
		SpaceID    string
		SpaceSlug  string
		SpaceName  string
		DocSlug    string
		Title      string
		CreatedAt  time.Time
	}
	err := s.db.WithContext(ctx).Table("user_favorites uf").
		Select(`uf.document_id, d.space_id, sp.slug AS space_slug, sp.name AS space_name,
			d.slug AS doc_slug, d.title, uf.created_at`).
		Joins("JOIN documents d ON d.id = uf.document_id").
		Joins("JOIN spaces sp ON sp.id = d.space_id").
		Where("uf.user_id = ? AND d.is_published = ?", userID, true).
		Order("uf.created_at DESC").
		Limit(100).
		Scan(&rows).Error
	if err != nil {
		return nil, err
	}
	out := make([]FavoriteItem, len(rows))
	for i, r := range rows {
		out[i] = FavoriteItem(r)
	}
	return out, nil
}

func (s *UserPrefsService) AddFavorite(ctx context.Context, userID, documentID string) error {
	return s.db.WithContext(ctx).Create(&models.UserFavorite{
		UserID: userID, DocumentID: documentID,
	}).Error
}

func (s *UserPrefsService) RemoveFavorite(ctx context.Context, userID, documentID string) error {
	return s.db.WithContext(ctx).
		Where("user_id = ? AND document_id = ?", userID, documentID).
		Delete(&models.UserFavorite{}).Error
}

func (s *UserPrefsService) IsFavorite(ctx context.Context, userID, documentID string) (bool, error) {
	var count int64
	err := s.db.WithContext(ctx).Model(&models.UserFavorite{}).
		Where("user_id = ? AND document_id = ?", userID, documentID).
		Count(&count).Error
	return count > 0, err
}

type RecentItem struct {
	DocumentID string    `json:"document_id"`
	SpaceID    string    `json:"space_id"`
	SpaceSlug  string    `json:"space_slug"`
	SpaceName  string    `json:"space_name"`
	DocSlug    string    `json:"doc_slug"`
	Title      string    `json:"title"`
	ViewedAt   time.Time `json:"viewed_at"`
}

func (s *UserPrefsService) ListRecent(ctx context.Context, userID string, limit int) ([]RecentItem, error) {
	if limit <= 0 || limit > 50 {
		limit = 20
	}
	var rows []struct {
		DocumentID string
		SpaceID    string
		SpaceSlug  string
		SpaceName  string
		DocSlug    string
		Title      string
		ViewedAt   time.Time
	}
	err := s.db.WithContext(ctx).Table("user_recent_views urv").
		Select(`urv.document_id, d.space_id, sp.slug AS space_slug, sp.name AS space_name,
			d.slug AS doc_slug, d.title, urv.viewed_at`).
		Joins("JOIN documents d ON d.id = urv.document_id").
		Joins("JOIN spaces sp ON sp.id = d.space_id").
		Where("urv.user_id = ? AND d.is_published = ?", userID, true).
		Order("urv.viewed_at DESC").
		Limit(limit).
		Scan(&rows).Error
	if err != nil {
		return nil, err
	}
	out := make([]RecentItem, len(rows))
	for i, r := range rows {
		out[i] = RecentItem(r)
	}
	return out, nil
}

func (s *UserPrefsService) RecordView(ctx context.Context, userID, documentID, spaceID string) error {
	now := time.Now()
	return s.db.WithContext(ctx).Exec(`
		INSERT INTO user_recent_views (user_id, document_id, space_id, viewed_at)
		VALUES (?, ?, ?, ?)
		ON CONFLICT (user_id, document_id) DO UPDATE SET viewed_at = EXCLUDED.viewed_at, space_id = EXCLUDED.space_id
	`, userID, documentID, spaceID, now).Error
}
