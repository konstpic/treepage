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

var mentionRe = regexp.MustCompile(`@([a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+)`)

type CommentService struct {
	db *gorm.DB
}

func NewCommentService(db *gorm.DB) *CommentService {
	return &CommentService{db: db}
}

type MentionLabel struct {
	Email       string `json:"email"`
	DisplayName string `json:"display_name"`
}

type CommentView struct {
	models.DocumentComment
	MentionLabels []MentionLabel `json:"mention_labels,omitempty"`
	Replies       []CommentView  `json:"replies,omitempty"`
}

func (s *CommentService) ListThread(ctx context.Context, documentID string) ([]CommentView, error) {
	var rows []models.DocumentComment
	err := s.db.WithContext(ctx).Table("document_comments c").
		Select("c.*, COALESCE(NULLIF(TRIM(u.display_name), ''), u.email, '') AS author_name").
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
	for i := range roots {
		s.enrichCommentView(ctx, &roots[i])
	}
	return roots, nil
}

func (s *CommentService) enrichCommentView(ctx context.Context, cv *CommentView) {
	if cv.AuthorID != nil && *cv.AuthorID != "" && cv.AuthorName == "" {
		cv.AuthorName, _ = s.authorDisplayName(ctx, *cv.AuthorID)
	}
	cv.MentionLabels = s.mentionLabelsForComment(ctx, cv.Body, []string(cv.Mentions))
	for i := range cv.Replies {
		s.enrichCommentView(ctx, &cv.Replies[i])
	}
}

func (s *CommentService) authorDisplayName(ctx context.Context, userID string) (string, error) {
	var user models.User
	if err := s.db.WithContext(ctx).Select("email", "display_name").First(&user, "id = ?", userID).Error; err != nil {
		return "", err
	}
	return displayNameOrEmail(user.DisplayName, user.Email), nil
}

func displayNameOrEmail(displayName, email string) string {
	if strings.TrimSpace(displayName) != "" {
		return strings.TrimSpace(displayName)
	}
	return email
}

func (s *CommentService) mentionLabels(ctx context.Context, userIDs []string) []MentionLabel {
	if len(userIDs) == 0 {
		return nil
	}
	var users []models.User
	if err := s.db.WithContext(ctx).Select("email", "display_name").Where("id IN ?", userIDs).Find(&users).Error; err != nil {
		return nil
	}
	labels := make([]MentionLabel, 0, len(users))
	for _, u := range users {
		labels = append(labels, MentionLabel{
			Email:       u.Email,
			DisplayName: displayNameOrEmail(u.DisplayName, u.Email),
		})
	}
	return labels
}

func (s *CommentService) mentionLabelsForComment(ctx context.Context, body string, mentionIDs []string) []MentionLabel {
	labels := s.mentionLabels(ctx, mentionIDs)
	seen := map[string]struct{}{}
	for _, l := range labels {
		seen[strings.ToLower(l.Email)] = struct{}{}
	}
	for _, m := range mentionRe.FindAllStringSubmatch(body, -1) {
		if len(m) < 2 {
			continue
		}
		email := strings.ToLower(m[1])
		if _, ok := seen[email]; ok {
			continue
		}
		var user models.User
		if err := s.db.WithContext(ctx).Where("LOWER(email) = ?", email).First(&user).Error; err != nil {
			continue
		}
		labels = append(labels, MentionLabel{
			Email:       user.Email,
			DisplayName: displayNameOrEmail(user.DisplayName, user.Email),
		})
		seen[email] = struct{}{}
	}
	return labels
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
	c.AuthorName, _ = s.authorDisplayName(ctx, authorID)
	return &c, mentions, nil
}

func (s *CommentService) GetByID(ctx context.Context, commentID string) (*models.DocumentComment, error) {
	var c models.DocumentComment
	if err := s.db.WithContext(ctx).First(&c, "id = ?", commentID).Error; err != nil {
		return nil, err
	}
	return &c, nil
}

// IsMentionedOnDocument reports whether the user was @mentioned on any comment in the document.
func (s *CommentService) IsMentionedOnDocument(ctx context.Context, documentID, userID string) (bool, error) {
	if userID == "" {
		return false, nil
	}
	var count int64
	err := s.db.WithContext(ctx).Model(&models.DocumentComment{}).
		Where("document_id = ? AND ? = ANY(mentions)", documentID, userID).
		Count(&count).Error
	return count > 0, err
}

// IsMentionedInSpace reports whether the user was @mentioned on any document in the space.
func (s *CommentService) IsMentionedInSpace(ctx context.Context, spaceID, userID string) (bool, error) {
	if userID == "" {
		return false, nil
	}
	var count int64
	err := s.db.WithContext(ctx).Table("document_comments c").
		Joins("JOIN documents d ON d.id = c.document_id").
		Where("d.space_id = ? AND ? = ANY(c.mentions)", spaceID, userID).
		Count(&count).Error
	return count > 0, err
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

func (s *CommentService) SuggestMentionUsers(ctx context.Context, q string, limit int) ([]models.User, error) {
	if limit <= 0 || limit > 20 {
		limit = 10
	}
	q = strings.TrimSpace(q)
	var users []models.User
	db := s.db.WithContext(ctx).Where("is_active = ?", true)
	if q != "" {
		like := "%" + q + "%"
		db = db.Where("email ILIKE ? OR display_name ILIKE ?", like, like)
	}
	err := db.Order("display_name ASC, email ASC").Limit(limit).Find(&users).Error
	return users, err
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
