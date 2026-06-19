package service

import (
	"context"
	"strings"

	"github.com/konstpic/treepage/backend/pkg/models"
)

type DocumentVersionRow struct {
	ID            string `json:"id"`
	VersionNumber int    `json:"version_number"`
	Title         string `json:"title"`
	AuthorName    string `json:"author_name,omitempty"`
	CreatedAt     string `json:"created_at"`
}

type DocumentVersionDetail struct {
	ID            string `json:"id"`
	VersionNumber int    `json:"version_number"`
	Title         string `json:"title"`
	Content       string `json:"content"`
	AuthorName    string `json:"author_name,omitempty"`
	CreatedAt     string `json:"created_at"`
}

type VersionDiffLine struct {
	Type    string `json:"type"` // add, remove, same
	Content string `json:"content"`
}

type VersionDiff struct {
	FromVersion int               `json:"from_version"`
	ToVersion   int               `json:"to_version"`
	Lines       []VersionDiffLine `json:"lines"`
}

func (d *DocumentService) ListVersions(ctx context.Context, docID string) ([]DocumentVersionRow, error) {
	var versions []models.DocumentVersion
	err := d.db.WithContext(ctx).
		Select("id", "version_number", "title", "author_name", "created_at").
		Where("document_id = ?", docID).
		Order("version_number DESC").
		Find(&versions).Error
	if err != nil {
		return nil, err
	}
	out := make([]DocumentVersionRow, 0, len(versions))
	for _, v := range versions {
		out = append(out, DocumentVersionRow{
			ID:            v.ID,
			VersionNumber: v.VersionNumber,
			Title:         v.Title,
			AuthorName:    v.AuthorName,
			CreatedAt:     v.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		})
	}
	return out, nil
}

func (d *DocumentService) GetVersion(ctx context.Context, docID string, versionNumber int) (*DocumentVersionDetail, error) {
	var v models.DocumentVersion
	err := d.db.WithContext(ctx).
		Where("document_id = ? AND version_number = ?", docID, versionNumber).
		First(&v).Error
	if err != nil {
		return nil, err
	}
	return &DocumentVersionDetail{
		ID:            v.ID,
		VersionNumber: v.VersionNumber,
		Title:         v.Title,
		Content:       v.Content,
		AuthorName:    v.AuthorName,
		CreatedAt:     v.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}, nil
}

func (d *DocumentService) DiffVersions(ctx context.Context, docID string, fromVersion, toVersion int) (*VersionDiff, error) {
	from, err := d.GetVersion(ctx, docID, fromVersion)
	if err != nil {
		return nil, err
	}
	to, err := d.GetVersion(ctx, docID, toVersion)
	if err != nil {
		return nil, err
	}
	fromLines := strings.Split(from.Content, "\n")
	toLines := strings.Split(to.Content, "\n")
	lines := lineDiff(fromLines, toLines)
	return &VersionDiff{
		FromVersion: fromVersion,
		ToVersion:   toVersion,
		Lines:       lines,
	}, nil
}

func lineDiff(a, b []string) []VersionDiffLine {
	m, n := len(a), len(b)
	dp := make([][]int, m+1)
	for i := range dp {
		dp[i] = make([]int, n+1)
	}
	for i := m - 1; i >= 0; i-- {
		for j := n - 1; j >= 0; j-- {
			if a[i] == b[j] {
				dp[i][j] = dp[i+1][j+1] + 1
			} else if dp[i+1][j] >= dp[i][j+1] {
				dp[i][j] = dp[i+1][j]
			} else {
				dp[i][j] = dp[i][j+1]
			}
		}
	}
	var out []VersionDiffLine
	i, j := 0, 0
	for i < m && j < n {
		if a[i] == b[j] {
			out = append(out, VersionDiffLine{Type: "same", Content: a[i]})
			i++
			j++
		} else if dp[i+1][j] >= dp[i][j+1] {
			out = append(out, VersionDiffLine{Type: "remove", Content: a[i]})
			i++
		} else {
			out = append(out, VersionDiffLine{Type: "add", Content: b[j]})
			j++
		}
	}
	for i < m {
		out = append(out, VersionDiffLine{Type: "remove", Content: a[i]})
		i++
	}
	for j < n {
		out = append(out, VersionDiffLine{Type: "add", Content: b[j]})
		j++
	}
	return out
}
