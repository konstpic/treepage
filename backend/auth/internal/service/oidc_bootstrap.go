package service

import (
	"context"
	"errors"
	"strings"

	"github.com/konstpic/treepage/backend/pkg/config"
	"github.com/konstpic/treepage/backend/pkg/models"
	"gorm.io/gorm"
)

const configOIDCProviderName = "Authentik (config)"

// BootstrapOIDCFromConfig mirrors backend-auth YAML/env OIDC into oidc_providers so the admin UI lists the active IdP.
func (s *AuthService) BootstrapOIDCFromConfig(ctx context.Context, cfg config.OIDCConfig) error {
	if !cfg.Enabled || strings.TrimSpace(cfg.IssuerURL) == "" {
		return nil
	}

	scopes := strings.TrimSpace(cfg.Scopes)
	if scopes == "" {
		scopes = "openid profile email"
	}
	roleClaim := strings.TrimSpace(cfg.RoleClaim)
	if roleClaim == "" {
		roleClaim = "roles"
	}
	groupClaim := strings.TrimSpace(cfg.GroupClaim)
	if groupClaim == "" {
		groupClaim = "groups"
	}

	var provider models.OIDCProvider
	err := s.db.WithContext(ctx).Where("name = ?", configOIDCProviderName).First(&provider).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		provider = models.OIDCProvider{
			Name:            configOIDCProviderName,
			ProviderType:    "authentik",
			IssuerURL:       cfg.IssuerURL,
			ClientID:        cfg.ClientID,
			ClientSecretRef: "OIDC_CLIENT_SECRET",
			RedirectURL:     cfg.RedirectURL,
			Scopes:          scopes,
			RoleClaim:       roleClaim,
			GroupClaim:      groupClaim,
			SyncGroups:      cfg.SyncGroups,
			Enabled:         true,
		}
		return s.db.WithContext(ctx).Create(&provider).Error
	}
	if err != nil {
		return err
	}

	provider.ProviderType = "authentik"
	provider.IssuerURL = cfg.IssuerURL
	provider.ClientID = cfg.ClientID
	provider.ClientSecretRef = "OIDC_CLIENT_SECRET"
	provider.RedirectURL = cfg.RedirectURL
	provider.Scopes = scopes
	provider.RoleClaim = roleClaim
	provider.GroupClaim = groupClaim
	provider.SyncGroups = cfg.SyncGroups
	provider.Enabled = true
	return s.db.WithContext(ctx).Save(&provider).Error
}
