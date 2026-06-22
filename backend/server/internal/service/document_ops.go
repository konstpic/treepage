package service

import (
	"context"
	"errors"
	"time"

	"github.com/konstpic/treepage/backend/pkg/contenthash"
	"github.com/konstpic/treepage/backend/pkg/models"
)

var ErrDocumentNotFound = errors.New("document not found")

func (d *DocumentService) Delete(ctx context.Context, docID string) error {
	res := d.db.WithContext(ctx).Delete(&models.Document{}, "id = ?", docID)
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return ErrDocumentNotFound
	}
	return nil
}

func (d *DocumentService) RevertToVersion(ctx context.Context, docID, userID string, versionNum int) (*models.Document, error) {
	version, err := d.GetVersion(ctx, docID, versionNum)
	if err != nil {
		return nil, err
	}
	var doc models.Document
	if err := d.db.WithContext(ctx).First(&doc, "id = ?", docID).Error; err != nil {
		return nil, err
	}
	doc.Title = version.Title
	doc.Content = version.Content
	if doc.RepositoryID != nil {
		currentHash := contenthash.SHA256(doc.Content)
		doc.HasPendingChanges = doc.SyncedContentHash != "" && currentHash != doc.SyncedContentHash
	}
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

func (d *DocumentService) MarkPublished(ctx context.Context, docID, commitSHA, content string) error {
	hash := contenthash.SHA256(content)
	now := time.Now()
	return d.db.WithContext(ctx).Model(&models.Document{}).Where("id = ?", docID).Updates(map[string]interface{}{
		"commit_sha":            commitSHA,
		"synced_content_hash":   hash,
		"sync_snapshot_content": content,
		"has_pending_changes":   false,
		"last_synced_at":        now,
	}).Error
}
