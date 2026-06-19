package service

import (
	"context"
	"errors"

	"github.com/konstpic/treepage/backend/pkg/models"
	"gorm.io/gorm"
)

type RepositoryWithSpace struct {
	models.Repository
	SpaceSlug string `json:"space_slug"`
	SpaceName string `json:"space_name"`
}

type RepositoryDetail struct {
	RepositoryWithSpace
	LatestJob *models.SyncJob `json:"latest_job,omitempty"`
}

func (r *RepositoryService) ListAll(ctx context.Context) ([]RepositoryWithSpace, error) {
	var rows []RepositoryWithSpace
	err := r.db.WithContext(ctx).
		Table("repositories").
		Select(`repositories.*, spaces.slug as space_slug, spaces.name as space_name`).
		Joins("JOIN spaces ON spaces.id = repositories.space_id").
		Order("repositories.name ASC").
		Scan(&rows).Error
	return rows, err
}

func (r *RepositoryService) GetByID(ctx context.Context, id string) (*RepositoryDetail, error) {
	var row RepositoryWithSpace
	err := r.db.WithContext(ctx).
		Table("repositories").
		Select(`repositories.*, spaces.slug as space_slug, spaces.name as space_name`).
		Joins("JOIN spaces ON spaces.id = repositories.space_id").
		Where("repositories.id = ?", id).
		Scan(&row).Error
	if err != nil {
		return nil, err
	}
	if row.ID == "" {
		return nil, gorm.ErrRecordNotFound
	}
	detail := &RepositoryDetail{RepositoryWithSpace: row}
	var job models.SyncJob
	if err := r.db.WithContext(ctx).
		Where("repository_id = ?", id).
		Order("created_at DESC").
		First(&job).Error; err == nil {
		detail.LatestJob = &job
	}
	return detail, nil
}

type AdminCreateRepositoryInput struct {
	SpaceID              string `json:"space_id" binding:"required"`
	Name                 string `json:"name" binding:"required"`
	URL                  string `json:"url" binding:"required"`
	Branch               string `json:"branch"`
	Provider             string `json:"provider"`
	DocsPath             string `json:"docs_path"`
	SyncMode             string `json:"sync_mode"`
	SyncIntervalSeconds  int    `json:"sync_interval_seconds"`
	AccessTokenRef       string `json:"access_token_ref"`
	WebhookSecretRef     string `json:"webhook_secret_ref"`
	Enabled              *bool  `json:"enabled"`
}

func (r *RepositoryService) AdminCreate(ctx context.Context, input AdminCreateRepositoryInput) (*models.Repository, error) {
	var space models.Space
	if err := r.db.WithContext(ctx).First(&space, "id = ?", input.SpaceID).Error; err != nil {
		return nil, err
	}
	ci := CreateRepositoryInput{
		Name: input.Name, URL: input.URL, Branch: input.Branch,
		Provider: input.Provider, DocsPath: input.DocsPath,
	}
	repo, err := r.Create(ctx, space.ID, ci)
	if err != nil {
		return nil, err
	}
	if input.SyncMode != "" {
		repo.SyncMode = input.SyncMode
	}
	if input.SyncIntervalSeconds > 0 {
		repo.SyncIntervalSeconds = input.SyncIntervalSeconds
	}
	if input.AccessTokenRef != "" {
		repo.AccessTokenRef = input.AccessTokenRef
	}
	if input.WebhookSecretRef != "" {
		repo.WebhookSecretRef = input.WebhookSecretRef
	}
	if input.Enabled != nil {
		repo.Enabled = *input.Enabled
	}
	if err := r.db.WithContext(ctx).Save(repo).Error; err != nil {
		return nil, err
	}
	return repo, nil
}

type UpdateRepositoryInput struct {
	Name                string `json:"name"`
	URL                 string `json:"url"`
	Branch              string `json:"branch"`
	Provider            string `json:"provider"`
	DocsPath            string `json:"docs_path"`
	SyncMode            string `json:"sync_mode"`
	SyncIntervalSeconds int    `json:"sync_interval_seconds"`
	AccessTokenRef      string `json:"access_token_ref"`
	WebhookSecretRef    string `json:"webhook_secret_ref"`
	Enabled             *bool  `json:"enabled"`
}

func (r *RepositoryService) Update(ctx context.Context, id string, input UpdateRepositoryInput) (*models.Repository, error) {
	var repo models.Repository
	if err := r.db.WithContext(ctx).First(&repo, "id = ?", id).Error; err != nil {
		return nil, err
	}
	if input.Name != "" {
		repo.Name = input.Name
	}
	if input.URL != "" {
		repo.URL = input.URL
	}
	if input.Branch != "" {
		repo.Branch = input.Branch
	}
	if input.Provider != "" {
		repo.Provider = input.Provider
	}
	if input.DocsPath != "" {
		repo.DocsPath = input.DocsPath
	}
	if input.SyncMode != "" {
		repo.SyncMode = input.SyncMode
	}
	if input.SyncIntervalSeconds > 0 {
		repo.SyncIntervalSeconds = input.SyncIntervalSeconds
	}
	if input.AccessTokenRef != "" {
		repo.AccessTokenRef = input.AccessTokenRef
	}
	if input.WebhookSecretRef != "" {
		repo.WebhookSecretRef = input.WebhookSecretRef
	}
	if input.Enabled != nil {
		repo.Enabled = *input.Enabled
	}
	if err := r.db.WithContext(ctx).Save(&repo).Error; err != nil {
		return nil, err
	}
	return &repo, nil
}

func (r *RepositoryService) BindToSpace(ctx context.Context, spaceID, repoID string) (*models.Repository, error) {
	var space models.Space
	if err := r.db.WithContext(ctx).First(&space, "id = ?", spaceID).Error; err != nil {
		return nil, err
	}
	var repo models.Repository
	if err := r.db.WithContext(ctx).First(&repo, "id = ?", repoID).Error; err != nil {
		return nil, err
	}
	repo.SpaceID = space.ID
	if err := r.db.WithContext(ctx).Save(&repo).Error; err != nil {
		return nil, err
	}
	return &repo, nil
}

func (r *RepositoryService) Delete(ctx context.Context, id string) error {
	res := r.db.WithContext(ctx).Delete(&models.Repository{}, "id = ?", id)
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

// UnbindFromSpace removes a repository from a space and deletes its synced documents.
func (r *RepositoryService) UnbindFromSpace(ctx context.Context, spaceID, repoID string) error {
	var repo models.Repository
	if err := r.db.WithContext(ctx).First(&repo, "id = ? AND space_id = ?", repoID, spaceID).Error; err != nil {
		return err
	}
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("repository_id = ? AND space_id = ?", repoID, spaceID).
			Delete(&models.Document{}).Error; err != nil {
			return err
		}
		if err := tx.Where("repository_id = ?", repoID).Delete(&models.SyncJob{}).Error; err != nil {
			return err
		}
		return tx.Delete(&repo).Error
	})
}

var ErrSpaceNotFound = errors.New("space not found")

func (s *SpaceService) GetByID(ctx context.Context, id string) (*models.Space, error) {
	var space models.Space
	if err := s.db.WithContext(ctx).First(&space, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return &space, nil
}
