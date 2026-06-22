package service

import (
	"context"
	"time"

	"github.com/konstpic/treepage/backend/pkg/models"
	"github.com/konstpic/treepage/backend/pkg/notify"
	"gorm.io/gorm"
)

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
