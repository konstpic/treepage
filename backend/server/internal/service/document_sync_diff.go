package service

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/konstpic/treepage/backend/pkg/contenthash"
	"github.com/konstpic/treepage/backend/pkg/models"
)

var (
	ErrInvalidSyncStrategy = errors.New("strategy must be accept_git or keep_local")
	ErrNoSyncConflict      = errors.New("document has no git sync conflict")
)

type SyncConflictDiff struct {
	DocumentID string            `json:"document_id"`
	GitContent string            `json:"git_content"`
	LocalContent string          `json:"local_content"`
	Lines      []VersionDiffLine `json:"lines"`
}

func (d *DocumentService) SyncConflictDiff(ctx context.Context, docID string) (*SyncConflictDiff, error) {
	var doc models.Document
	if err := d.db.WithContext(ctx).First(&doc, "id = ?", docID).Error; err != nil {
		return nil, err
	}
	if doc.RepositoryID == nil {
		return nil, ErrDocumentNotFound
	}
	gitContent := doc.SyncSnapshotContent
	if gitContent == "" {
		gitContent = doc.Content
	}
	localContent := doc.Content
	lines := lineDiff(strings.Split(gitContent, "\n"), strings.Split(localContent, "\n"))
	return &SyncConflictDiff{
		DocumentID:   doc.ID,
		GitContent:   gitContent,
		LocalContent: localContent,
		Lines:        lines,
	}, nil
}

func (d *DocumentService) ResolveSyncConflict(ctx context.Context, docID, userID, strategy string) (*models.Document, error) {
	strategy = strings.TrimSpace(strings.ToLower(strategy))
	if strategy != "accept_git" && strategy != "keep_local" {
		return nil, ErrInvalidSyncStrategy
	}
	var doc models.Document
	if err := d.db.WithContext(ctx).First(&doc, "id = ?", docID).Error; err != nil {
		return nil, err
	}
	if doc.RepositoryID == nil || !doc.HasPendingChanges {
		return nil, ErrNoSyncConflict
	}
	if strategy == "keep_local" {
		return &doc, nil
	}
	gitContent := doc.SyncSnapshotContent
	if gitContent == "" {
		return nil, errors.New("no git snapshot available")
	}
	doc.Content = gitContent
	doc.Title = extractTitleFromContent(gitContent, doc.Title)
	hash := contenthash.SHA256(gitContent)
	doc.SyncedContentHash = hash
	doc.HasPendingChanges = false
	now := time.Now()
	doc.LastSyncedAt = &now
	if err := d.db.WithContext(ctx).Save(&doc).Error; err != nil {
		return nil, err
	}
	var maxVersion int
	d.db.WithContext(ctx).Model(&models.DocumentVersion{}).
		Where("document_id = ?", docID).
		Select("COALESCE(MAX(version_number), 0)").Scan(&maxVersion)
	d.db.WithContext(ctx).Create(&models.DocumentVersion{
		DocumentID: doc.ID, VersionNumber: maxVersion + 1,
		Title: doc.Title, Content: doc.Content, AuthorID: &userID,
	})
	return &doc, nil
}

func extractTitleFromContent(content, fallback string) string {
	for _, line := range strings.Split(content, "\n") {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "# ") {
			return strings.TrimSpace(strings.TrimPrefix(line, "# "))
		}
	}
	return fallback
}
