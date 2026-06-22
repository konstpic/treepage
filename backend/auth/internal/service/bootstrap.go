package service

import (
	"context"
	"errors"

	"github.com/konstpic/treepage/backend/pkg/models"
	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

const (
	BootstrapEmail       = "admin@local"
	BootstrapPassword    = "admin"
	BootstrapDisplayName = "Local Admin"
)

var (
	ErrLocalLoginDisabled = errors.New("local login is disabled")
	ErrInvalidCredentials = errors.New("invalid email or password")
)

func (s *AuthService) EnsureLocalAuthSchema(ctx context.Context) error {
	return s.db.WithContext(ctx).Exec(
		"ALTER TABLE users ADD COLUMN IF NOT EXISTS password_hash VARCHAR(255)",
	).Error
}

// BootstrapLocalAdmin creates the default local admin on first startup if missing.
// Used for initial setup and as a fallback when the OIDC provider is unavailable.
func (s *AuthService) BootstrapLocalAdmin(ctx context.Context) error {
	if err := s.EnsureLocalAuthSchema(ctx); err != nil {
		return err
	}

	var existing int64
	if err := s.db.WithContext(ctx).Model(&models.User{}).Where("email = ?", BootstrapEmail).Count(&existing).Error; err != nil {
		return err
	}
	if existing > 0 {
		return nil
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(BootstrapPassword), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	user := models.User{
		Email:        BootstrapEmail,
		DisplayName:  BootstrapDisplayName,
		PasswordHash: string(hash),
		IsActive:     true,
	}
	if err := s.db.WithContext(ctx).Create(&user).Error; err != nil {
		return err
	}

	s.syncUserRoles(ctx, user.ID, []string{"super_admin"})
	s.logger.Warn("bootstrap local admin created",
		zap.String("email", BootstrapEmail),
		zap.String("role", "super_admin"),
	)
	return nil
}

type LocalLoginInput struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

func (s *AuthService) LoginLocal(ctx context.Context, email, password string) (*UserProfile, *TokenPair, error) {
	var user models.User
	err := s.db.WithContext(ctx).Where("email = ? AND is_active = ?", email, true).First(&user).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil, ErrInvalidCredentials
		}
		return nil, nil, err
	}
	if user.PasswordHash == "" {
		return nil, nil, ErrInvalidCredentials
	}
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		return nil, nil, ErrInvalidCredentials
	}

	roles, err := s.loadRoleNames(ctx, user.ID)
	if err != nil {
		return nil, nil, err
	}

	pair, err := s.issueTokens(ctx, user.ID, user.Email, roles)
	if err != nil {
		return nil, nil, err
	}

	return &UserProfile{
		ID:          user.ID,
		Email:       user.Email,
		DisplayName: user.DisplayName,
		AvatarURL:   user.AvatarURL,
		Roles:       roles,
	}, pair, nil
}
