package service

import (
	"context"
	"sort"
	"strings"
	"time"

	"github.com/konstpic/treepage/backend/pkg/models"
	"github.com/konstpic/treepage/backend/server/internal/syncclient"
)

type DocumentVersionRow struct {
	ID            string `json:"id"`
	Source        string `json:"source"` // local, git
	VersionNumber int    `json:"version_number,omitempty"`
	CommitSHA     string `json:"commit_sha,omitempty"`
	ShortSHA      string `json:"short_sha,omitempty"`
	Title         string `json:"title,omitempty"`
	AuthorName    string `json:"author_name,omitempty"`
	Message       string `json:"message,omitempty"`
	CreatedAt     string `json:"created_at"`
}

type DocumentVersionDetail struct {
	ID            string `json:"id"`
	Source        string `json:"source"`
	VersionNumber int    `json:"version_number,omitempty"`
	CommitSHA     string `json:"commit_sha,omitempty"`
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
	FromVersion int               `json:"from_version,omitempty"`
	ToVersion   int               `json:"to_version,omitempty"`
	FromSHA     string            `json:"from_sha,omitempty"`
	ToSHA       string            `json:"to_sha,omitempty"`
	Lines       []VersionDiffLine `json:"lines"`
}

func shortSHA(sha string) string {
	sha = strings.TrimSpace(sha)
	if len(sha) <= 8 {
		return sha
	}
	return sha[:8]
}

func (d *DocumentService) ListVersions(ctx context.Context, docID string) ([]DocumentVersionRow, error) {
	var versions []models.DocumentVersion
	err := d.db.WithContext(ctx).
		Select("id", "version_number", "title", "author_name", "commit_sha", "created_at").
		Where("document_id = ?", docID).
		Order("version_number DESC").
		Find(&versions).Error
	if err != nil {
		return nil, err
	}
	out := make([]DocumentVersionRow, 0, len(versions))
	for _, v := range versions {
		row := DocumentVersionRow{
			ID:            v.ID,
			Source:        "local",
			VersionNumber: v.VersionNumber,
			Title:         v.Title,
			AuthorName:    v.AuthorName,
			CommitSHA:     v.CommitSHA,
			CreatedAt:     v.CreatedAt.Format(time.RFC3339),
		}
		if row.CommitSHA != "" {
			row.ShortSHA = shortSHA(row.CommitSHA)
		}
		out = append(out, row)
	}
	return out, nil
}

func MergeVersionHistory(local []DocumentVersionRow, git []syncclient.GitFileRevision) []DocumentVersionRow {
	seenSHA := map[string]struct{}{}
	for _, row := range local {
		if row.CommitSHA != "" {
			seenSHA[strings.ToLower(row.CommitSHA)] = struct{}{}
		}
	}
	merged := append([]DocumentVersionRow{}, local...)
	for _, g := range git {
		if g.CommitSHA != "" {
			if _, ok := seenSHA[strings.ToLower(g.CommitSHA)]; ok {
				continue
			}
			seenSHA[strings.ToLower(g.CommitSHA)] = struct{}{}
		}
		merged = append(merged, DocumentVersionRow{
			ID:         g.CommitSHA,
			Source:     "git",
			CommitSHA:  g.CommitSHA,
			ShortSHA:   shortSHA(g.CommitSHA),
			AuthorName: g.AuthorName,
			Message:    g.Message,
			CreatedAt:  g.CreatedAt,
		})
	}
	sort.Slice(merged, func(i, j int) bool {
		ti, _ := time.Parse(time.RFC3339, merged[i].CreatedAt)
		tj, _ := time.Parse(time.RFC3339, merged[j].CreatedAt)
		return ti.After(tj)
	})
	return merged
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
		Source:        "local",
		VersionNumber: v.VersionNumber,
		CommitSHA:     v.CommitSHA,
		Title:         v.Title,
		Content:       v.Content,
		AuthorName:    v.AuthorName,
		CreatedAt:     v.CreatedAt.Format(time.RFC3339),
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
	return diffContent(from.Content, to.Content, fromVersion, toVersion, "", ""), nil
}

func (d *DocumentService) DiffGitVersions(ctx context.Context, sync *syncclient.Client, doc *models.Document, fromSHA, toSHA string) (*VersionDiff, error) {
	if doc.RepositoryID == nil || sync == nil {
		return nil, ErrDocumentNotFound
	}
	fromContent, err := sync.FileContentAt(ctx, *doc.RepositoryID, doc.Path, fromSHA)
	if err != nil {
		return nil, err
	}
	toContent, err := sync.FileContentAt(ctx, *doc.RepositoryID, doc.Path, toSHA)
	if err != nil {
		return nil, err
	}
	return diffContent(fromContent, toContent, 0, 0, fromSHA, toSHA), nil
}

func (d *DocumentService) DiffMixedVersions(
	ctx context.Context,
	sync *syncclient.Client,
	doc *models.Document,
	fromVersion int,
	toSHA string,
) (*VersionDiff, error) {
	from, err := d.GetVersion(ctx, doc.ID, fromVersion)
	if err != nil {
		return nil, err
	}
	if doc.RepositoryID == nil || sync == nil {
		return nil, ErrDocumentNotFound
	}
	toContent, err := sync.FileContentAt(ctx, *doc.RepositoryID, doc.Path, toSHA)
	if err != nil {
		return nil, err
	}
	return diffContent(from.Content, toContent, fromVersion, 0, "", toSHA), nil
}

func (d *DocumentService) DiffMixedVersionsReverse(
	ctx context.Context,
	sync *syncclient.Client,
	doc *models.Document,
	fromSHA string,
	toVersion int,
) (*VersionDiff, error) {
	if doc.RepositoryID == nil || sync == nil {
		return nil, ErrDocumentNotFound
	}
	fromContent, err := sync.FileContentAt(ctx, *doc.RepositoryID, doc.Path, fromSHA)
	if err != nil {
		return nil, err
	}
	to, err := d.GetVersion(ctx, doc.ID, toVersion)
	if err != nil {
		return nil, err
	}
	return diffContent(fromContent, to.Content, 0, toVersion, fromSHA, ""), nil
}

func diffContent(fromContent, toContent string, fromVersion, toVersion int, fromSHA, toSHA string) *VersionDiff {
	fromLines := strings.Split(fromContent, "\n")
	toLines := strings.Split(toContent, "\n")
	return &VersionDiff{
		FromVersion: fromVersion,
		ToVersion:   toVersion,
		FromSHA:     fromSHA,
		ToSHA:       toSHA,
		Lines:       lineDiff(fromLines, toLines),
	}
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
