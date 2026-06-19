package handler

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/coreos/go-oidc/v3/oidc"
	"github.com/gin-gonic/gin"
	"github.com/konstpic/treepage/backend/auth/internal/service"
	pkgjwt "github.com/konstpic/treepage/backend/pkg/jwt"
	"go.uber.org/zap"
	"golang.org/x/oauth2"
)

type StateStore interface {
	Create() (string, error)
	Validate(state string) bool
}

type memoryStateStore struct {
	mu   sync.Mutex
	data map[string]time.Time
}

func NewMemoryStateStore() *memoryStateStore {
	s := &memoryStateStore{data: make(map[string]time.Time)}
	go func() {
		for {
			time.Sleep(5 * time.Minute)
			s.cleanup()
		}
	}()
	return s
}

func (s *memoryStateStore) Create() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	state := base64.URLEncoding.EncodeToString(b)
	s.mu.Lock()
	s.data[state] = time.Now().Add(10 * time.Minute)
	s.mu.Unlock()
	return state, nil
}

func (s *memoryStateStore) Validate(state string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	exp, ok := s.data[state]
	if !ok || time.Now().After(exp) {
		return false
	}
	delete(s.data, state)
	return true
}

func (s *memoryStateStore) cleanup() {
	s.mu.Lock()
	defer s.mu.Unlock()
	now := time.Now()
	for k, exp := range s.data {
		if now.After(exp) {
			delete(s.data, k)
		}
	}
}

type Handler struct {
	svc          *service.AuthService
	jwt          *pkgjwt.Manager
	states       StateStore
	oidcProvider *oidc.Provider
	oauth2Cfg    *oauth2.Config
	frontendURL  string
	devLogin     bool
	logger       *zap.Logger
}

func New(
	svc *service.AuthService,
	jwt *pkgjwt.Manager,
	states StateStore,
	oidcProvider *oidc.Provider,
	oauth2Cfg *oauth2.Config,
	frontendURL string,
	devLogin bool,
	logger *zap.Logger,
) *Handler {
	if states == nil {
		states = NewMemoryStateStore()
	}
	return &Handler{
		svc: svc, jwt: jwt, states: states,
		oidcProvider: oidcProvider, oauth2Cfg: oauth2Cfg,
		frontendURL: frontendURL, devLogin: devLogin, logger: logger,
	}
}

func (h *Handler) Register(r *gin.Engine) {
	h.registerAuthRoutes(r.Group("/api/auth"))
	h.registerAuthRoutes(r.Group("/auth"))
}

func (h *Handler) registerAuthRoutes(api *gin.RouterGroup) {
	api.GET("/login", h.LoginOIDC)
	if h.devLogin {
		api.POST("/login", h.LoginLocal)
	}
	api.GET("/callback", h.Callback)
	api.POST("/refresh", h.Refresh)
	api.POST("/logout", h.Logout)
	api.GET("/me", h.AuthMiddleware(), h.Me)
	api.GET("/users", h.AuthMiddleware(), h.RequireRole("super_admin"), h.ListUsers)
}

func (h *Handler) LoginOIDC(c *gin.Context) {
	if h.oauth2Cfg == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "OIDC not configured"})
		return
	}
	state, err := h.states.Create()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create state"})
		return
	}
	url := h.oauth2Cfg.AuthCodeURL(state, oauth2.AccessTypeOffline)
	c.Redirect(http.StatusFound, url)
}

type localLoginRequest struct {
	Email    string `json:"email" binding:"required"`
	Password string `json:"password" binding:"required"`
}

func (h *Handler) LoginLocal(c *gin.Context) {
	if !h.devLogin {
		c.JSON(http.StatusForbidden, gin.H{"error": "local login is disabled"})
		return
	}
	var req localLoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	profile, pair, err := h.svc.LoginLocal(c.Request.Context(), req.Email, req.Password)
	if err != nil {
		if errors.Is(err, service.ErrInvalidCredentials) {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid email or password"})
			return
		}
		if errors.Is(err, service.ErrDevLoginDisabled) {
			c.JSON(http.StatusForbidden, gin.H{"error": "local login is disabled"})
			return
		}
		h.logger.Error("local login failed", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "login failed"})
		return
	}
	h.svc.AuditLogin(c.Request.Context(), profile.ID, "local", c.ClientIP(), c.GetHeader("User-Agent"))
	c.JSON(http.StatusOK, gin.H{
		"access_token":  pair.AccessToken,
		"refresh_token": pair.RefreshToken,
		"expires_at":    pair.ExpiresAt,
		"user":          profile,
	})
}

func (h *Handler) Callback(c *gin.Context) {
	if h.oauth2Cfg == nil || h.oidcProvider == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "OIDC not configured"})
		return
	}
	state := c.Query("state")
	if !h.states.Validate(state) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid state"})
		return
	}
	code := c.Query("code")
	if code == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "missing code"})
		return
	}

	ctx := c.Request.Context()
	token, err := h.oauth2Cfg.Exchange(ctx, code)
	if err != nil {
		h.logger.Error("oidc exchange failed", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": "token exchange failed"})
		return
	}

	rawIDToken, ok := token.Extra("id_token").(string)
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{"error": "missing id_token"})
		return
	}
	verifier := h.oidcProvider.Verifier(&oidc.Config{ClientID: h.oauth2Cfg.ClientID})
	idToken, err := verifier.Verify(ctx, rawIDToken)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id_token"})
		return
	}

	var claims map[string]interface{}
	if err := idToken.Claims(&claims); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid claims"})
		return
	}

	email := claimString(claims, "email")
	if email == "" {
		email = claimString(claims, "preferred_username")
	}
	displayName := claimString(claims, "name")
	if displayName == "" {
		displayName = email
	}
	picture := claimString(claims, "picture")

	oidcSettings := h.svc.GetOIDCClaimSettings(ctx)
	roleNames := extractStringSlice(claims[oidcSettings.RoleClaim])
	groupNames := extractStringSlice(claims[oidcSettings.GroupClaim])

	profile, pair, err := h.svc.UpsertOIDCUser(ctx, email, displayName, picture, idToken.Subject, roleNames, groupNames, oidcSettings.SyncGroups)
	if err != nil {
		h.logger.Error("user upsert failed", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "user sync failed"})
		return
	}

	h.svc.AuditLogin(ctx, profile.ID, "oidc", c.ClientIP(), c.GetHeader("User-Agent"))

	redirect := fmt.Sprintf("%s/auth/callback?access_token=%s&refresh_token=%s",
		strings.TrimRight(h.frontendURL, "/"),
		pair.AccessToken, pair.RefreshToken,
	)
	c.Redirect(http.StatusFound, redirect)
}

func extractStringSlice(v any) []string {
	switch t := v.(type) {
	case []any:
		out := make([]string, 0, len(t))
		for _, item := range t {
			if s, ok := item.(string); ok {
				out = append(out, s)
			}
		}
		return out
	case []string:
		return t
	case string:
		if t != "" {
			return []string{t}
		}
	}
	return nil
}

func claimString(claims map[string]interface{}, key string) string {
	if claims == nil || key == "" {
		return ""
	}
	v, ok := claims[key]
	if !ok {
		return ""
	}
	if s, ok := v.(string); ok {
		return s
	}
	return ""
}

type refreshRequest struct {
	RefreshToken string `json:"refresh_token" binding:"required"`
}

func (h *Handler) Refresh(c *gin.Context) {
	var req refreshRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	pair, err := h.svc.Refresh(c.Request.Context(), req.RefreshToken)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid refresh token"})
		return
	}
	c.JSON(http.StatusOK, pair)
}

func (h *Handler) Logout(c *gin.Context) {
	var req refreshRequest
	_ = c.ShouldBindJSON(&req)
	if req.RefreshToken != "" {
		_ = h.svc.Logout(c.Request.Context(), req.RefreshToken)
	}
	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}

func (h *Handler) Me(c *gin.Context) {
	userID := c.GetString("userID")
	profile, err := h.svc.Me(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
		return
	}
	c.JSON(http.StatusOK, profile)
}

func (h *Handler) ListUsers(c *gin.Context) {
	users, err := h.svc.ListUsers(c.Request.Context(), 50, 0)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"items": users})
}

func (h *Handler) AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		auth := c.GetHeader("Authorization")
		if !strings.HasPrefix(auth, "Bearer ") {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "missing token"})
			return
		}
		token := strings.TrimPrefix(auth, "Bearer ")
		claims, err := h.jwt.ParseAccess(token)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid token"})
			return
		}
		c.Set("userID", claims.UserID)
		c.Set("email", claims.Email)
		c.Set("roles", claims.Roles)
		c.Next()
	}
}

func (h *Handler) RequireRole(role string) gin.HandlerFunc {
	return func(c *gin.Context) {
		roles, _ := c.Get("roles")
		roleList, ok := roles.([]string)
		if !ok {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "forbidden"})
			return
		}
		for _, r := range roleList {
			if r == role || r == "super_admin" {
				c.Next()
				return
			}
		}
		c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "forbidden"})
	}
}

// Ensure context import used in tests
var _ = context.Background
