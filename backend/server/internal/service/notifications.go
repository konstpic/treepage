package service

import (
	"context"
	"fmt"
	"time"

	"github.com/konstpic/treepage/backend/pkg/models"
	"github.com/konstpic/treepage/backend/pkg/notify"
	"gorm.io/gorm"
)

type NotificationView struct {
	models.Notification
	Link string `json:"link,omitempty"`
}

type NotificationService struct {
	db      *gorm.DB
	webhook *notify.Webhook
	slack   *notify.Slack
	email   *notify.Email
}

func NewNotificationService(db *gorm.DB, webhook *notify.Webhook, slack *notify.Slack, email *notify.Email) *NotificationService {
	return &NotificationService{db: db, webhook: webhook, slack: slack, email: email}
}

func (s *NotificationService) List(ctx context.Context, userID string, limit int) ([]models.Notification, error) {
	if limit <= 0 || limit > 100 {
		limit = 50
	}
	var items []models.Notification
	err := s.db.WithContext(ctx).
		Where("user_id = ?", userID).
		Order("created_at DESC").
		Limit(limit).
		Find(&items).Error
	return items, err
}

func (s *NotificationService) ListViews(ctx context.Context, userID string, limit int) ([]NotificationView, error) {
	items, err := s.List(ctx, userID, limit)
	if err != nil {
		return nil, err
	}
	views := make([]NotificationView, len(items))
	for i, n := range items {
		views[i] = NotificationView{Notification: n, Link: s.resolveLink(ctx, &n)}
	}
	return views, nil
}

func (s *NotificationService) resolveLink(ctx context.Context, n *models.Notification) string {
	if n.ResourceType == nil || n.ResourceID == nil {
		return ""
	}
	switch *n.ResourceType {
	case "comment":
		return s.commentDeepLink(ctx, *n.ResourceID)
	case "document":
		return s.documentDeepLink(ctx, *n.ResourceID)
	default:
		return ""
	}
}

func (s *NotificationService) commentDeepLink(ctx context.Context, commentID string) string {
	var row struct {
		SpaceSlug string
		DocSlug   string
	}
	err := s.db.WithContext(ctx).Table("document_comments c").
		Select("s.slug AS space_slug, d.slug AS doc_slug").
		Joins("JOIN documents d ON d.id = c.document_id").
		Joins("JOIN spaces s ON s.id = d.space_id").
		Where("c.id = ?", commentID).
		Scan(&row).Error
	if err != nil || row.SpaceSlug == "" || row.DocSlug == "" {
		return ""
	}
	return fmt.Sprintf("/spaces/%s/docs/%s#comment-%s", row.SpaceSlug, row.DocSlug, commentID)
}

func (s *NotificationService) documentDeepLink(ctx context.Context, documentID string) string {
	var row struct {
		SpaceSlug string
		DocSlug   string
	}
	err := s.db.WithContext(ctx).Table("documents d").
		Select("s.slug AS space_slug, d.slug AS doc_slug").
		Joins("JOIN spaces s ON s.id = d.space_id").
		Where("d.id = ?", documentID).
		Scan(&row).Error
	if err != nil || row.SpaceSlug == "" || row.DocSlug == "" {
		return ""
	}
	return fmt.Sprintf("/spaces/%s/docs/%s", row.SpaceSlug, row.DocSlug)
}

func (s *NotificationService) UnreadCount(ctx context.Context, userID string) (int64, error) {
	var count int64
	err := s.db.WithContext(ctx).Model(&models.Notification{}).
		Where("user_id = ? AND read_at IS NULL", userID).
		Count(&count).Error
	return count, err
}

func (s *NotificationService) MarkRead(ctx context.Context, userID, notificationID string) error {
	now := time.Now()
	return s.db.WithContext(ctx).Model(&models.Notification{}).
		Where("id = ? AND user_id = ?", notificationID, userID).
		Update("read_at", now).Error
}

func (s *NotificationService) MarkAllRead(ctx context.Context, userID string) error {
	now := time.Now()
	return s.db.WithContext(ctx).Model(&models.Notification{}).
		Where("user_id = ? AND read_at IS NULL", userID).
		Update("read_at", now).Error
}

func (s *NotificationService) Create(ctx context.Context, userID, nType, title, body string, resourceType, resourceID *string) (*models.Notification, error) {
	n := models.Notification{
		UserID: userID, Type: nType, Title: title, Body: body,
		ResourceType: resourceType, ResourceID: resourceID,
	}
	if err := s.db.WithContext(ctx).Create(&n).Error; err != nil {
		return nil, err
	}
	s.dispatchExternal(userID, nType, title, body, resourceType, resourceID)
	return &n, nil
}

func (s *NotificationService) dispatchExternal(userID, nType, title, body string, resourceType, resourceID *string) {
	if s.webhook == nil && s.slack == nil && s.email == nil {
		return
	}
	payload := notify.Payload{
		Type: nType, Title: title, Body: body, UserID: userID,
		ResourceType: resourceType, ResourceID: resourceID,
	}
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		if s.webhook != nil {
			s.webhook.Notify(ctx, payload)
		}
		if s.slack != nil {
			s.slack.Notify(ctx, payload)
		}
		if s.email != nil {
			var userEmail string
			s.db.WithContext(ctx).Model(&models.User{}).Where("id = ?", userID).Select("email").Scan(&userEmail)
			s.email.Notify(ctx, userEmail, payload)
		}
	}()
}

// NotifySpaceEditors sends a notification to all editors+ in a space (excluding actor).
func (s *NotificationService) NotifySpaceEditors(ctx context.Context, spaceID, actorID, nType, title, body, resourceType, resourceID string) {
	var memberIDs []string
	s.db.WithContext(ctx).Table("space_members sm").
		Select("sm.user_id").
		Joins("JOIN roles r ON r.id = sm.role_id").
		Where("sm.space_id = ? AND r.name IN ? AND sm.user_id != ?", spaceID, []string{"editor", "admin", "owner"}, actorID).
		Pluck("sm.user_id", &memberIDs)
	rt, rid := resourceType, resourceID
	for _, uid := range memberIDs {
		_, _ = s.Create(ctx, uid, nType, title, body, &rt, &rid)
	}
}
