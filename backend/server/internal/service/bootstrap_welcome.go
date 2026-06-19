package service

import (
	"context"
	"errors"
	"os"

	"github.com/konstpic/treepage/backend/pkg/models"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

const (
	WelcomeSpaceSlug        = "welcome"
	WelcomeSpaceName        = "Welcome"
	DefaultWelcomeRepoURL   = "https://github.com/konstpic/treepage.git"
	DefaultWelcomeRepoName  = "TreePage Docs"
	DefaultWelcomeDocsPath  = "docs"
	DefaultWelcomeBranch    = "main"
	BootstrapAdminEmail     = "admin@local"
)

type WelcomeBootstrapConfig struct {
	Enabled  bool
	RepoURL  string
	Branch   string
	DocsPath string
}

func LoadWelcomeBootstrapConfig() WelcomeBootstrapConfig {
	enabled := os.Getenv("WELCOME_SPACE_ENABLED") != "false"
	repoURL := os.Getenv("WELCOME_REPO_URL")
	if repoURL == "" {
		repoURL = DefaultWelcomeRepoURL
	}
	branch := os.Getenv("WELCOME_REPO_BRANCH")
	if branch == "" {
		branch = DefaultWelcomeBranch
	}
	docsPath := os.Getenv("WELCOME_DOCS_PATH")
	if docsPath == "" {
		docsPath = DefaultWelcomeDocsPath
	}
	return WelcomeBootstrapConfig{
		Enabled:  enabled,
		RepoURL:  repoURL,
		Branch:   branch,
		DocsPath: docsPath,
	}
}

// BootstrapWelcomeSpace creates the public welcome space and docs repository on first startup.
// The second return value is true when a new repository was created (initial sync should run).
func BootstrapWelcomeSpace(ctx context.Context, db *gorm.DB, cfg WelcomeBootstrapConfig, log *zap.Logger) (*models.Repository, bool, error) {
	if !cfg.Enabled {
		return nil, false, nil
	}

	var existing models.Space
	err := db.WithContext(ctx).Where("slug = ?", WelcomeSpaceSlug).First(&existing).Error
	if err == nil {
		var repo models.Repository
		if err := db.WithContext(ctx).Where("space_id = ?", existing.ID).First(&repo).Error; err != nil {
			return nil, false, nil
		}
		return &repo, false, nil
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, false, err
	}

	var adminUser models.User
	var createdBy *string
	if err := db.WithContext(ctx).Where("email = ?", BootstrapAdminEmail).First(&adminUser).Error; err == nil {
		createdBy = &adminUser.ID
	}

	space := models.Space{
		Slug:        WelcomeSpaceSlug,
		Name:        WelcomeSpaceName,
		Description: "TreePage documentation — installation, user guide, and admin manual (EN/RU).",
		IsPublic:    true,
		CreatedBy:   createdBy,
	}
	if err := db.WithContext(ctx).Create(&space).Error; err != nil {
		return nil, false, err
	}

	if createdBy != nil {
		var adminRole models.Role
		if err := db.WithContext(ctx).Where("name = ?", "admin").First(&adminRole).Error; err == nil {
			_ = db.WithContext(ctx).Create(&models.SpaceMember{
				SpaceID: space.ID,
				UserID:  *createdBy,
				RoleID:  adminRole.ID,
			}).Error
		}
	}

	repo := models.Repository{
		SpaceID:             space.ID,
		Name:                DefaultWelcomeRepoName,
		URL:                 cfg.RepoURL,
		Branch:              cfg.Branch,
		Provider:            "github",
		DocsPath:            cfg.DocsPath,
		SyncMode:            "manual",
		SyncIntervalSeconds: 300,
		AccessTokenRef:      "GIT_ACCESS_TOKEN",
		Enabled:             true,
	}
	if err := db.WithContext(ctx).Create(&repo).Error; err != nil {
		return nil, false, err
	}

	if log != nil {
		log.Info("welcome space bootstrapped",
			zap.String("space_slug", WelcomeSpaceSlug),
			zap.String("repo_url", cfg.RepoURL),
			zap.String("docs_path", cfg.DocsPath),
		)
	}
	return &repo, true, nil
}
