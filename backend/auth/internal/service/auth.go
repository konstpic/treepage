package service

import (
	"context"
	"time"

	pkgjwt "github.com/konstpic/treepage/backend/pkg/jwt"
	"github.com/konstpic/treepage/backend/pkg/models"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

type AuthService struct {
	db     *gorm.DB
	jwt    *pkgjwt.Manager
	logger *zap.Logger
}

func NewAuthService(db *gorm.DB, jwt *pkgjwt.Manager, logger *zap.Logger) *AuthService {
	return &AuthService{db: db, jwt: jwt, logger: logger}
}

type TokenPair struct {
	AccessToken  string    `json:"access_token"`
	RefreshToken string    `json:"refresh_token"`
	ExpiresAt    time.Time `json:"expires_at"`
}

type UserProfile struct {
	ID          string   `json:"id"`
	Email       string   `json:"email"`
	DisplayName string   `json:"display_name"`
	AvatarURL   string   `json:"avatar_url,omitempty"`
	Roles       []string `json:"roles"`
}

func (s *AuthService) UpsertOIDCUser(ctx context.Context, email, displayName, avatarURL, externalID string, roleNames, groupNames []string, syncGroups bool) (*UserProfile, *TokenPair, error) {
	var user models.User
	err := s.db.WithContext(ctx).Where("email = ?", email).First(&user).Error
	if err == gorm.ErrRecordNotFound {
		user = models.User{
			Email:       email,
			DisplayName: displayName,
			AvatarURL:   avatarURL,
			ExternalID:  externalID,
			IsActive:    true,
		}
		if err := s.db.WithContext(ctx).Create(&user).Error; err != nil {
			return nil, nil, err
		}
	} else if err != nil {
		return nil, nil, err
	} else {
		now := time.Now()
		user.DisplayName = displayName
		user.AvatarURL = avatarURL
		user.ExternalID = externalID
		user.LastLoginAt = &now
		if err := s.db.WithContext(ctx).Save(&user).Error; err != nil {
			return nil, nil, err
		}
	}

	if len(roleNames) > 0 {
		s.syncUserRoles(ctx, user.ID, roleNames)
	}
	if syncGroups {
		s.syncUserGroupsFromOIDC(ctx, user.ID, groupNames)
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

func (s *AuthService) syncUserRoles(ctx context.Context, userID string, roleNames []string) {
	s.db.WithContext(ctx).Exec("DELETE FROM user_roles WHERE user_id = ?", userID)
	for _, name := range roleNames {
		var role models.Role
		if err := s.db.WithContext(ctx).Where("name = ?", name).First(&role).Error; err != nil {
			continue
		}
		s.db.WithContext(ctx).Exec(
			"INSERT INTO user_roles (user_id, role_id) VALUES (?, ?) ON CONFLICT DO NOTHING",
			userID, role.ID,
		)
	}
}

type OIDCClaimSettings struct {
	RoleClaim  string
	GroupClaim string
	SyncGroups bool
}

func (s *AuthService) GetOIDCClaimSettings(ctx context.Context) OIDCClaimSettings {
	settings := OIDCClaimSettings{RoleClaim: "roles", GroupClaim: "groups"}
	var provider models.OIDCProvider
	if err := s.db.WithContext(ctx).Where("enabled = ?", true).Order("created_at ASC").First(&provider).Error; err == nil {
		if provider.RoleClaim != "" {
			settings.RoleClaim = provider.RoleClaim
		}
		if provider.GroupClaim != "" {
			settings.GroupClaim = provider.GroupClaim
		}
		settings.SyncGroups = provider.SyncGroups
	}
	return settings
}

func (s *AuthService) syncUserGroupsFromOIDC(ctx context.Context, userID string, groupNames []string) {
	var oidcGroupIDs []string
	s.db.WithContext(ctx).Model(&models.Group{}).
		Where("external_id != ''").
		Pluck("id", &oidcGroupIDs)

	if len(oidcGroupIDs) > 0 {
		s.db.WithContext(ctx).
			Where("user_id = ? AND group_id IN ?", userID, oidcGroupIDs).
			Delete(&models.GroupMember{})
	}

	for _, name := range groupNames {
		if name == "" {
			continue
		}
		var grp models.Group
		err := s.db.WithContext(ctx).
			Where("external_id = ? OR name = ?", name, name).
			First(&grp).Error
		if err == gorm.ErrRecordNotFound {
			grp = models.Group{Name: name, ExternalID: name}
			if err := s.db.WithContext(ctx).Create(&grp).Error; err != nil {
				continue
			}
		} else if err != nil {
			continue
		}
		if grp.ExternalID == "" {
			s.db.WithContext(ctx).Model(&grp).Update("external_id", name)
		}
		s.db.WithContext(ctx).Exec(
			"INSERT INTO group_members (group_id, user_id) VALUES (?, ?) ON CONFLICT DO NOTHING",
			grp.ID, userID,
		)
	}
}

func (s *AuthService) loadRoleNames(ctx context.Context, userID string) ([]string, error) {
	var roles []models.Role
	err := s.db.WithContext(ctx).
		Joins("JOIN user_roles ON user_roles.role_id = roles.id").
		Where("user_roles.user_id = ?", userID).
		Find(&roles).Error
	if err != nil {
		return nil, err
	}
	if len(roles) == 0 {
		var viewer models.Role
		if err := s.db.WithContext(ctx).Where("name = ?", "viewer").First(&viewer).Error; err == nil {
			s.db.WithContext(ctx).Exec(
				"INSERT INTO user_roles (user_id, role_id) VALUES (?, ?) ON CONFLICT DO NOTHING",
				userID, viewer.ID,
			)
			return []string{"viewer"}, nil
		}
		return []string{"viewer"}, nil
	}
	names := make([]string, len(roles))
	for i, r := range roles {
		names[i] = r.Name
	}
	return names, nil
}

func (s *AuthService) issueTokens(ctx context.Context, userID, email string, roles []string) (*TokenPair, error) {
	access, exp, err := s.jwt.IssueAccess(userID, email, roles)
	if err != nil {
		return nil, err
	}
	refresh, refreshExp, err := s.jwt.IssueRefresh(userID)
	if err != nil {
		return nil, err
	}
	rt := models.RefreshToken{
		UserID:    userID,
		TokenHash: pkgjwt.HashToken(refresh),
		ExpiresAt: refreshExp,
	}
	if err := s.db.WithContext(ctx).Create(&rt).Error; err != nil {
		return nil, err
	}
	return &TokenPair{AccessToken: access, RefreshToken: refresh, ExpiresAt: exp}, nil
}

func (s *AuthService) Refresh(ctx context.Context, refreshToken string) (*TokenPair, error) {
	hash := pkgjwt.HashToken(refreshToken)
	var rt models.RefreshToken
	if err := s.db.WithContext(ctx).Where("token_hash = ? AND revoked_at IS NULL", hash).First(&rt).Error; err != nil {
		return nil, err
	}
	if time.Now().After(rt.ExpiresAt) {
		return nil, gorm.ErrRecordNotFound
	}
	var user models.User
	if err := s.db.WithContext(ctx).First(&user, "id = ?", rt.UserID).Error; err != nil {
		return nil, err
	}
	roles, err := s.loadRoleNames(ctx, user.ID)
	if err != nil {
		return nil, err
	}
	now := time.Now()
	s.db.WithContext(ctx).Model(&rt).Update("revoked_at", now)
	return s.issueTokens(ctx, user.ID, user.Email, roles)
}

func (s *AuthService) Logout(ctx context.Context, refreshToken string) error {
	hash := pkgjwt.HashToken(refreshToken)
	now := time.Now()
	return s.db.WithContext(ctx).Model(&models.RefreshToken{}).
		Where("token_hash = ?", hash).
		Update("revoked_at", now).Error
}

func (s *AuthService) Me(ctx context.Context, userID string) (*UserProfile, error) {
	var user models.User
	if err := s.db.WithContext(ctx).First(&user, "id = ?", userID).Error; err != nil {
		return nil, err
	}
	roles, err := s.loadRoleNames(ctx, user.ID)
	if err != nil {
		return nil, err
	}
	return &UserProfile{
		ID:          user.ID,
		Email:       user.Email,
		DisplayName: user.DisplayName,
		AvatarURL:   user.AvatarURL,
		Roles:       roles,
	}, nil
}

func (s *AuthService) ListUsers(ctx context.Context, limit, offset int) ([]UserProfile, error) {
	var users []models.User
	if limit <= 0 {
		limit = 50
	}
	if err := s.db.WithContext(ctx).Limit(limit).Offset(offset).Find(&users).Error; err != nil {
		return nil, err
	}
	result := make([]UserProfile, 0, len(users))
	for _, u := range users {
		roles, _ := s.loadRoleNames(ctx, u.ID)
		result = append(result, UserProfile{
			ID: u.ID, Email: u.Email, DisplayName: u.DisplayName,
			AvatarURL: u.AvatarURL, Roles: roles,
		})
	}
	return result, nil
}
