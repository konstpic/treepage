package service

import (
	"context"
	"encoding/json"
	"errors"
	"strings"

	"github.com/konstpic/treepage/backend/pkg/models"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type AdminService struct {
	db *gorm.DB
}

func NewAdminService(db *gorm.DB) *AdminService {
	return &AdminService{db: db}
}

type SystemSettings struct {
	Auth       map[string]any `json:"auth"`
	Git        map[string]any `json:"git"`
	Platform   map[string]any `json:"platform"`
	UITheme    string         `json:"ui_theme"`
	UILanguage string         `json:"ui_language"`
}

const defaultUITheme = "fox_white"

var validUIThemes = map[string]bool{
	"coral_night": true,
	"fox_white":   true,
	"light":       true,
	"dark":        true,
}

var validUILanguages = map[string]bool{
	"en": true,
	"ru": true,
}

func (a *AdminService) GetUITheme(ctx context.Context) (string, error) {
	var row models.SystemSetting
	if err := a.db.WithContext(ctx).First(&row, "key = ?", "ui_theme").Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return defaultUITheme, nil
		}
		return "", err
	}
	var theme string
	if err := json.Unmarshal([]byte(row.Value), &theme); err != nil {
		return defaultUITheme, nil
	}
	if !validUIThemes[theme] {
		return defaultUITheme, nil
	}
	return theme, nil
}

func (a *AdminService) SetUITheme(ctx context.Context, userID, theme string) (string, error) {
	if !validUIThemes[theme] {
		return "", errors.New("invalid ui_theme")
	}
	raw, err := json.Marshal(theme)
	if err != nil {
		return "", err
	}
	var uid *string
	if userID != "" {
		uid = &userID
	}
	row := models.SystemSetting{Key: "ui_theme", Value: string(raw), UpdatedBy: uid}
	if err := a.db.WithContext(ctx).
		Where("key = ?", "ui_theme").
		Assign(models.SystemSetting{Value: string(raw), UpdatedBy: uid}).
		FirstOrCreate(&row).Error; err != nil {
		return "", err
	}
	return theme, nil
}

func (a *AdminService) GetUILanguage(ctx context.Context) (string, error) {
	var row models.SystemSetting
	if err := a.db.WithContext(ctx).First(&row, "key = ?", "ui_language").Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return "en", nil
		}
		return "", err
	}
	var lang string
	if err := json.Unmarshal([]byte(row.Value), &lang); err != nil {
		return "en", nil
	}
	if !validUILanguages[lang] {
		return "en", nil
	}
	return lang, nil
}

func (a *AdminService) SetUILanguage(ctx context.Context, userID, lang string) (string, error) {
	if !validUILanguages[lang] {
		return "", errors.New("invalid ui_language")
	}
	raw, err := json.Marshal(lang)
	if err != nil {
		return "", err
	}
	var uid *string
	if userID != "" {
		uid = &userID
	}
	row := models.SystemSetting{Key: "ui_language", Value: string(raw), UpdatedBy: uid}
	if err := a.db.WithContext(ctx).
		Where("key = ?", "ui_language").
		Assign(models.SystemSetting{Value: string(raw), UpdatedBy: uid}).
		FirstOrCreate(&row).Error; err != nil {
		return "", err
	}
	return lang, nil
}

func (a *AdminService) GetSystemSettings(ctx context.Context) (*SystemSettings, error) {
	out := &SystemSettings{
		Auth:     map[string]any{},
		Git:      map[string]any{},
		Platform: map[string]any{},
	}
	theme, err := a.GetUITheme(ctx)
	if err != nil {
		return nil, err
	}
	out.UITheme = theme
	lang, err := a.GetUILanguage(ctx)
	if err != nil {
		return nil, err
	}
	out.UILanguage = lang
	for _, key := range []string{"auth", "git", "platform"} {
		var row models.SystemSetting
		if err := a.db.WithContext(ctx).First(&row, "key = ?", key).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				continue
			}
			return nil, err
		}
		var m map[string]any
		if err := json.Unmarshal([]byte(row.Value), &m); err != nil {
			return nil, err
		}
		switch key {
		case "auth":
			out.Auth = m
		case "git":
			out.Git = m
		case "platform":
			out.Platform = m
		}
	}
	return out, nil
}

func (a *AdminService) UpdateSystemSettings(ctx context.Context, userID string, patch SystemSettings) (*SystemSettings, error) {
	updates := map[string]map[string]any{
		"auth":     patch.Auth,
		"git":      patch.Git,
		"platform": patch.Platform,
	}
	for key, val := range updates {
		if val == nil {
			continue
		}
		raw, err := json.Marshal(val)
		if err != nil {
			return nil, err
		}
		var uid *string
		if userID != "" {
			uid = &userID
		}
		row := models.SystemSetting{Key: key, Value: string(raw), UpdatedBy: uid}
		if err := a.db.WithContext(ctx).
			Where("key = ?", key).
			Assign(models.SystemSetting{Value: string(raw), UpdatedBy: uid}).
			FirstOrCreate(&row).Error; err != nil {
			return nil, err
		}
	}
	return a.GetSystemSettings(ctx)
}

type PublicOIDCProvider struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type PublicAuthMethods struct {
	LocalEnabled bool                 `json:"local_enabled"`
	OIDC         *PublicOIDCProviders `json:"oidc,omitempty"`
}

type PublicOIDCProviders struct {
	Providers []PublicOIDCProvider `json:"providers"`
}

func (a *AdminService) GetPublicAuthMethods(ctx context.Context) (*PublicAuthMethods, error) {
	out := &PublicAuthMethods{LocalEnabled: true}
	settings, err := a.GetSystemSettings(ctx)
	if err != nil {
		return nil, err
	}
	oidcEnabled, _ := settings.Auth["oidc_enabled"].(bool)
	if !oidcEnabled {
		return out, nil
	}
	providers, err := a.ListOIDCProviders(ctx)
	if err != nil {
		return nil, err
	}
	var public []PublicOIDCProvider
	for _, p := range providers {
		if p.Enabled {
			public = append(public, PublicOIDCProvider{ID: p.ID, Name: p.Name})
		}
	}
	if len(public) == 0 {
		public = []PublicOIDCProvider{{ID: "default", Name: "OIDC"}}
	}
	out.OIDC = &PublicOIDCProviders{Providers: public}
	return out, nil
}

func (a *AdminService) ListOIDCProviders(ctx context.Context) ([]models.OIDCProvider, error) {
	var items []models.OIDCProvider
	err := a.db.WithContext(ctx).Order("name ASC").Find(&items).Error
	return items, err
}

type OIDCProviderInput struct {
	Name         string `json:"name" binding:"required"`
	ProviderType string `json:"provider_type"`
	IssuerURL    string `json:"issuer_url" binding:"required"`
	ClientID     string `json:"client_id" binding:"required"`
	RedirectURL  string `json:"redirect_url" binding:"required"`
	Scopes       string `json:"scopes"`
	RoleClaim    string `json:"role_claim"`
	GroupClaim   string `json:"group_claim"`
	SyncGroups   *bool  `json:"sync_groups"`
	Enabled      *bool  `json:"enabled"`
}

func (a *AdminService) CreateOIDCProvider(ctx context.Context, input OIDCProviderInput) (*models.OIDCProvider, error) {
	enabled := true
	if input.Enabled != nil {
		enabled = *input.Enabled
	}
	pt := input.ProviderType
	if pt == "" {
		pt = "generic"
	}
	scopes := input.Scopes
	if scopes == "" {
		scopes = "openid profile email"
	}
	roleClaim := input.RoleClaim
	if roleClaim == "" {
		roleClaim = "roles"
	}
	groupClaim := input.GroupClaim
	if groupClaim == "" {
		groupClaim = "groups"
	}
	syncGroups := false
	if input.SyncGroups != nil {
		syncGroups = *input.SyncGroups
	}
	p := models.OIDCProvider{
		Name: input.Name, ProviderType: pt, IssuerURL: input.IssuerURL,
		ClientID: input.ClientID, ClientSecretRef: "OIDC_CLIENT_SECRET",
		RedirectURL: input.RedirectURL, Scopes: scopes,
		RoleClaim: roleClaim, GroupClaim: groupClaim, SyncGroups: syncGroups,
		Enabled: enabled,
	}
	if err := a.db.WithContext(ctx).Create(&p).Error; err != nil {
		return nil, err
	}
	return &p, nil
}

func (a *AdminService) UpdateOIDCProvider(ctx context.Context, id string, input OIDCProviderInput) (*models.OIDCProvider, error) {
	var p models.OIDCProvider
	if err := a.db.WithContext(ctx).First(&p, "id = ?", id).Error; err != nil {
		return nil, err
	}
	p.Name = input.Name
	p.IssuerURL = input.IssuerURL
	p.ClientID = input.ClientID
	p.RedirectURL = input.RedirectURL
	if input.ProviderType != "" {
		p.ProviderType = input.ProviderType
	}
	if input.Scopes != "" {
		p.Scopes = input.Scopes
	}
	if input.RoleClaim != "" {
		p.RoleClaim = input.RoleClaim
	}
	if input.GroupClaim != "" {
		p.GroupClaim = input.GroupClaim
	}
	if input.SyncGroups != nil {
		p.SyncGroups = *input.SyncGroups
	}
	if input.Enabled != nil {
		p.Enabled = *input.Enabled
	}
	if err := a.db.WithContext(ctx).Save(&p).Error; err != nil {
		return nil, err
	}
	return &p, nil
}

func (a *AdminService) DeleteOIDCProvider(ctx context.Context, id string) error {
	return a.db.WithContext(ctx).Delete(&models.OIDCProvider{}, "id = ?", id).Error
}

func (a *AdminService) ListUsers(ctx context.Context) ([]models.User, error) {
	var users []models.User
	err := a.db.WithContext(ctx).Order("email ASC").Find(&users).Error
	return users, err
}

func (a *AdminService) ListUserRoles(ctx context.Context, userID string) ([]string, error) {
	var roles []models.Role
	err := a.db.WithContext(ctx).
		Joins("JOIN user_roles ON user_roles.role_id = roles.id").
		Where("user_roles.user_id = ?", userID).
		Find(&roles).Error
	if err != nil {
		return nil, err
	}
	names := make([]string, len(roles))
	for i, r := range roles {
		names[i] = r.Name
	}
	return names, nil
}

var validAssignRoles = map[string]bool{
	"super_admin": true,
	"admin":       true,
	"editor":      true,
	"viewer":      true,
}

var elevatedUserRoles = map[string]bool{
	"admin":       true,
	"super_admin": true,
}

func HasElevatedUserRole(roles []string) bool {
	for _, r := range roles {
		if elevatedUserRoles[r] {
			return true
		}
	}
	return false
}

func CanManageUser(actorRoles, targetRoles []string) bool {
	if HasRole(actorRoles, "super_admin") {
		return true
	}
	return HasRole(actorRoles, "admin") && !HasElevatedUserRole(targetRoles)
}

func AssignableRolesFor(actorRoles []string) []string {
	if HasRole(actorRoles, "super_admin") {
		return []string{"viewer", "editor", "admin", "super_admin"}
	}
	if HasRole(actorRoles, "admin") {
		return []string{"viewer", "editor"}
	}
	return nil
}

type CreateUserInput struct {
	Email       string   `json:"email" binding:"required"`
	Password    string   `json:"password" binding:"required,min=8"`
	DisplayName string   `json:"display_name"`
	Roles       []string `json:"roles"`
	IsActive    *bool    `json:"is_active"`
}

func normalizeAccountEmail(email string) string {
	return strings.TrimSpace(strings.ToLower(email))
}

func isValidAccountEmail(email string) bool {
	local, domain, ok := strings.Cut(email, "@")
	return ok && local != "" && domain != "" && !strings.ContainsAny(local, " \t")
}

func (a *AdminService) EnsurePasswordHashColumn(ctx context.Context) error {
	return a.db.WithContext(ctx).Exec(
		"ALTER TABLE users ADD COLUMN IF NOT EXISTS password_hash VARCHAR(255)",
	).Error
}

func (a *AdminService) CreateUser(ctx context.Context, input CreateUserInput) (*models.User, []string, error) {
	if err := a.EnsurePasswordHashColumn(ctx); err != nil {
		return nil, nil, err
	}

	input.Email = normalizeAccountEmail(input.Email)
	if !isValidAccountEmail(input.Email) {
		return nil, nil, errors.New("invalid email")
	}

	var existing int64
	if err := a.db.WithContext(ctx).Model(&models.User{}).Where("email = ?", input.Email).Count(&existing).Error; err != nil {
		return nil, nil, err
	}
	if existing > 0 {
		return nil, nil, errors.New("email already exists")
	}

	roleNames := input.Roles
	if len(roleNames) == 0 {
		roleNames = []string{"viewer"}
	}
	for _, name := range roleNames {
		if !validAssignRoles[name] {
			return nil, nil, errors.New("invalid role: " + name)
		}
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(input.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, nil, err
	}

	displayName := input.DisplayName
	if displayName == "" {
		if at := strings.Index(input.Email, "@"); at > 0 {
			displayName = input.Email[:at]
		} else {
			displayName = input.Email
		}
	}

	isActive := true
	if input.IsActive != nil {
		isActive = *input.IsActive
	}

	user := models.User{
		Email:        input.Email,
		DisplayName:  displayName,
		PasswordHash: string(hash),
		IsActive:     isActive,
	}
	if err := a.db.WithContext(ctx).Create(&user).Error; err != nil {
		return nil, nil, err
	}
	if err := a.assignUserRoles(ctx, user.ID, roleNames); err != nil {
		return nil, nil, err
	}
	return &user, roleNames, nil
}

type UpdateUserInput struct {
	Email       *string  `json:"email"`
	Password    *string  `json:"password"`
	DisplayName *string  `json:"display_name"`
	Roles       []string `json:"roles"`
	IsActive    *bool    `json:"is_active"`
}

func (a *AdminService) UpdateUser(ctx context.Context, actorRoles []string, userID string, input UpdateUserInput) (*models.User, []string, error) {
	if err := a.EnsurePasswordHashColumn(ctx); err != nil {
		return nil, nil, err
	}

	var user models.User
	if err := a.db.WithContext(ctx).First(&user, "id = ?", userID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil, errors.New("user not found")
		}
		return nil, nil, err
	}

	targetRoles, err := a.ListUserRoles(ctx, userID)
	if err != nil {
		return nil, nil, err
	}
	if !CanManageUser(actorRoles, targetRoles) {
		return nil, nil, errors.New("forbidden: cannot manage this user")
	}

	allowedSet := make(map[string]bool, len(validAssignRoles))
	for _, name := range AssignableRolesFor(actorRoles) {
		allowedSet[name] = true
	}

	updates := map[string]any{}

	if input.Email != nil {
		email := normalizeAccountEmail(*input.Email)
		if !isValidAccountEmail(email) {
			return nil, nil, errors.New("invalid email")
		}
		if email != user.Email {
			var existing int64
			if err := a.db.WithContext(ctx).Model(&models.User{}).Where("email = ? AND id != ?", email, userID).Count(&existing).Error; err != nil {
				return nil, nil, err
			}
			if existing > 0 {
				return nil, nil, errors.New("email already exists")
			}
			updates["email"] = email
		}
	}

	if input.DisplayName != nil {
		updates["display_name"] = strings.TrimSpace(*input.DisplayName)
	}

	if input.IsActive != nil {
		updates["is_active"] = *input.IsActive
	}

	if input.Password != nil && strings.TrimSpace(*input.Password) != "" {
		password := *input.Password
		if len(password) < 8 {
			return nil, nil, errors.New("password must be at least 8 characters")
		}
		hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
		if err != nil {
			return nil, nil, err
		}
		updates["password_hash"] = string(hash)
	}

	if len(updates) > 0 {
		if err := a.db.WithContext(ctx).Model(&user).Updates(updates).Error; err != nil {
			return nil, nil, err
		}
	}

	roleNames := targetRoles
	if input.Roles != nil {
		roleNames = input.Roles
		if len(roleNames) == 0 {
			roleNames = []string{"viewer"}
		}
		for _, name := range roleNames {
			if !validAssignRoles[name] {
				return nil, nil, errors.New("invalid role: " + name)
			}
			if !allowedSet[name] {
				return nil, nil, errors.New("forbidden: cannot assign role " + name)
			}
		}
		if err := a.assignUserRoles(ctx, userID, roleNames); err != nil {
			return nil, nil, err
		}
	}

	if err := a.db.WithContext(ctx).First(&user, "id = ?", userID).Error; err != nil {
		return nil, nil, err
	}
	return &user, roleNames, nil
}

func (a *AdminService) DeleteUser(ctx context.Context, actorID string, actorRoles []string, userID string) error {
	if actorID == userID {
		return errors.New("forbidden: cannot delete your own account")
	}

	var user models.User
	if err := a.db.WithContext(ctx).First(&user, "id = ?", userID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("user not found")
		}
		return err
	}

	targetRoles, err := a.ListUserRoles(ctx, userID)
	if err != nil {
		return err
	}
	if !CanManageUser(actorRoles, targetRoles) {
		return errors.New("forbidden: cannot manage this user")
	}

	if HasRole(targetRoles, "super_admin") {
		count, err := a.countUsersWithRole(ctx, "super_admin")
		if err != nil {
			return err
		}
		if count <= 1 {
			return errors.New("forbidden: cannot delete the last super admin")
		}
	}

	return a.db.WithContext(ctx).Delete(&models.User{}, "id = ?", userID).Error
}

func (a *AdminService) countUsersWithRole(ctx context.Context, roleName string) (int64, error) {
	var count int64
	err := a.db.WithContext(ctx).
		Model(&models.User{}).
		Joins("JOIN user_roles ON user_roles.user_id = users.id").
		Joins("JOIN roles ON roles.id = user_roles.role_id").
		Where("roles.name = ?", roleName).
		Count(&count).Error
	return count, err
}

func (a *AdminService) assignUserRoles(ctx context.Context, userID string, roleNames []string) error {
	if err := a.db.WithContext(ctx).Exec("DELETE FROM user_roles WHERE user_id = ?", userID).Error; err != nil {
		return err
	}
	for _, name := range roleNames {
		var role models.Role
		if err := a.db.WithContext(ctx).Where("name = ?", name).First(&role).Error; err != nil {
			return err
		}
		if err := a.db.WithContext(ctx).Exec(
			"INSERT INTO user_roles (user_id, role_id) VALUES (?, ?) ON CONFLICT DO NOTHING",
			userID, role.ID,
		).Error; err != nil {
			return err
		}
	}
	return nil
}
