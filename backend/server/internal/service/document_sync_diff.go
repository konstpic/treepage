package service

import (
	"context"
	"strings"

	"github.com/konstpic/treepage/backend/pkg/models"
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
