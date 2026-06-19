package service

import (
	"context"
	"errors"
	"regexp"
	"strings"
	"time"

	"github.com/konstpic/treepage/backend/pkg/models"
	"github.com/lib/pq"
	"gorm.io/gorm"
)

const (
	WorkflowDraft     = "draft"
	WorkflowInReview  = "in_review"
	WorkflowApproved  = "approved"
	WorkflowPublished = "published"
)

var mentionRe = regexp.MustCompile(`@([a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,})`)

type CommentService struct {
	db *gorm.DB
}

func NewCommentService(db *gorm.DB) *CommentService {
	return &CommentService{db: db}
}

type CommentView struct {
	models.DocumentComment
	Replies []CommentView `json:"replies,omitempty"`
}

func (s *CommentService) ListThread(ctx context.Context, documentID string) ([]CommentView, error) {
	var rows []models.DocumentComment
	err := s.db.WithContext(ctx).Table("document_comments c").
		Select("c.*, COALESCE(u.display_name, u.email, '') AS author_name").
		Joins("LEFT JOIN users u ON u.id = c.author_id").
		Where("c.document_id = ?", documentID).
		Order("c.created_at ASC").
		Scan(&rows).Error
	if err != nil {
		return nil, err
	}
	byID := map[string]*CommentView{}
	var roots []CommentView
	for i := range rows {
		cv := CommentView{DocumentComment: rows[i]}
		byID[cv.ID] = &cv
	}
	for i := range rows {
		cv := byID[rows[i].ID]
		if rows[i].ParentID != nil && *rows[i].ParentID != "" {
			if parent, ok := byID[*rows[i].ParentID]; ok {
				parent.Replies = append(parent.Replies, *cv)
			}
		} else {
			roots = append(roots, *cv)
		}
	}
	return roots, nil
}

func (s *CommentService) Create(ctx context.Context, documentID, authorID, body string, parentID *string) (*models.DocumentComment, []string, error) {
	body = strings.TrimSpace(body)
	if body == "" {
		return nil, nil, errors.New("comment body is required")
	}
	mentions := s.parseMentions(ctx, body)
	now := time.Now()
	c := models.DocumentComment{
		DocumentID: documentID, ParentID: parentID, AuthorID: &authorID,
		Body: body, Mentions: pq.StringArray(mentions), CreatedAt: now, UpdatedAt: now,
	}
	if err := s.db.WithContext(ctx).Create(&c).Error; err != nil {
		return nil, nil, err
	}
	return &c, mentions, nil
}

func (s *CommentService) GetByID(ctx context.Context, commentID string) (*models.DocumentComment, error) {
	var c models.DocumentComment
	if err := s.db.WithContext(ctx).First(&c, "id = ?", commentID).Error; err != nil {
		return nil, err
	}
	return &c, nil
}

func (s *CommentService) Delete(ctx context.Context, commentID, userID string, isAdmin bool) error {
	var c models.DocumentComment
	if err := s.db.WithContext(ctx).First(&c, "id = ?", commentID).Error; err != nil {
		return err
	}
	if !isAdmin && (c.AuthorID == nil || *c.AuthorID != userID) {
		return errors.New("forbidden")
	}
	return s.db.WithContext(ctx).Delete(&c).Error
}

func (s *CommentService) parseMentions(ctx context.Context, body string) []string {
	seen := map[string]struct{}{}
	var ids []string
	for _, m := range mentionRe.FindAllStringSubmatch(body, -1) {
		if len(m) < 2 {
			continue
		}
		email := strings.ToLower(m[1])
		if _, ok := seen[email]; ok {
			continue
		}
		seen[email] = struct{}{}
		var user models.User
		if err := s.db.WithContext(ctx).Where("LOWER(email) = ?", email).First(&user).Error; err == nil {
			ids = append(ids, user.ID)
		}
	}
	return ids
}

func (d *DocumentService) SetWorkflowState(ctx context.Context, docID, state string) error {
	return d.db.WithContext(ctx).Model(&models.Document{}).
		Where("id = ?", docID).
		Update("workflow_state", state).Error
}

func (d *DocumentService) SubmitReview(ctx context.Context, docID string) (*models.Document, error) {
	var doc models.Document
	if err := d.db.WithContext(ctx).First(&doc, "id = ?", docID).Error; err != nil {
		return nil, err
	}
	doc.WorkflowState = WorkflowInReview
	doc.IsPublished = false
	if err := d.db.WithContext(ctx).Save(&doc).Error; err != nil {
		return nil, err
	}
	return &doc, nil
}

func (d *DocumentService) ApproveReview(ctx context.Context, docID string) (*models.Document, error) {
	var doc models.Document
	if err := d.db.WithContext(ctx).First(&doc, "id = ?", docID).Error; err != nil {
		return nil, err
	}
	doc.WorkflowState = WorkflowApproved
	if err := d.db.WithContext(ctx).Save(&doc).Error; err != nil {
		return nil, err
	}
	return &doc, nil
}

func (d *DocumentService) RejectReview(ctx context.Context, docID string) (*models.Document, error) {
	var doc models.Document
	if err := d.db.WithContext(ctx).First(&doc, "id = ?", docID).Error; err != nil {
		return nil, err
	}
	doc.WorkflowState = WorkflowDraft
	doc.IsPublished = false
	if err := d.db.WithContext(ctx).Save(&doc).Error; err != nil {
		return nil, err
	}
	return &doc, nil
}

func (d *DocumentService) PublishWorkflow(ctx context.Context, docID string) (*models.Document, error) {
	var doc models.Document
	if err := d.db.WithContext(ctx).First(&doc, "id = ?", docID).Error; err != nil {
		return nil, err
	}
	doc.WorkflowState = WorkflowPublished
	doc.IsPublished = true
	if err := d.db.WithContext(ctx).Save(&doc).Error; err != nil {
		return nil, err
	}
	return &doc, nil
}

func DocumentVisibleToViewer(doc *models.Document) bool {
	if !doc.IsPublished {
		return false
	}
	switch doc.WorkflowState {
	case WorkflowPublished, WorkflowApproved, "":
		return true
	default:
		return false
	}
}

func syncWorkflowOnPublish(isPublished bool, current string) string {
	if isPublished {
		return WorkflowPublished
	}
	if current == WorkflowPublished || current == "" {
		return WorkflowDraft
	}
	return current
}
