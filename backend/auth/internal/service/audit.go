package service

import (
	"context"

	"github.com/konstpic/treepage/backend/pkg/models"
)

func (s *AuthService) AuditLogin(ctx context.Context, userID, method, ip, ua string) {
	if userID == "" {
		return
	}
	entry := models.AuditLog{
		Action:       "auth.login." + method,
		ResourceType: "user",
		IPAddress:    ip,
		UserAgent:    ua,
	}
	entry.UserID = &userID
	entry.ResourceID = &userID
	_ = s.db.WithContext(ctx).Create(&entry).Error
}
