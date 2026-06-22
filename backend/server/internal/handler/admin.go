package handler

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/konstpic/treepage/backend/server/internal/rag"
	"github.com/konstpic/treepage/backend/server/internal/service"
	"github.com/konstpic/treepage/backend/server/internal/syncclient"
	"gorm.io/gorm"
)

type AdminHandler struct {
	admin     *service.AdminService
	groups    *service.GroupService
	spaces    *service.SpaceService
	repos     *service.RepositoryService
	audit     *service.AuditService
	sync      *syncclient.Client
	ragWorker *rag.Worker
	auditOn   bool
	base      *Handler
}

func NewAdminHandler(
	base *Handler,
	admin *service.AdminService,
	groups *service.GroupService,
	spaces *service.SpaceService,
	repos *service.RepositoryService,
	audit *service.AuditService,
	sync *syncclient.Client,
	ragWorker *rag.Worker,
	auditOn bool,
) *AdminHandler {
	return &AdminHandler{
		base: base, admin: admin, groups: groups, spaces: spaces, repos: repos,
		audit: audit, sync: sync, ragWorker: ragWorker, auditOn: auditOn,
	}
}

func (h *AdminHandler) Register(r *gin.Engine) {
	pub := r.Group("/api/public")
	pub.GET("/ui-theme", h.GetPublicUITheme)
	pub.GET("/ui-language", h.GetPublicUILanguage)
	pub.GET("/auth-methods", h.GetPublicAuthMethods)

	register := func(g *gin.RouterGroup) {
		g.Use(h.base.AuthMiddleware())
		g.Use(h.base.RequireRoles("super_admin", "admin"))

		g.GET("/system-settings", h.GetSystemSettings)
		g.PUT("/system-settings", h.base.RequireRoles("super_admin"), h.UpdateSystemSettings)
		g.PUT("/system-settings/ui-theme", h.base.RequireRoles("super_admin"), h.UpdateUITheme)
		g.PUT("/system-settings/ui-language", h.base.RequireRoles("super_admin"), h.UpdateUILanguage)

		g.GET("/repositories", h.ListRepositories)
		g.POST("/repositories", h.CreateRepository)
		g.GET("/repositories/:id", h.GetRepository)
		g.PUT("/repositories/:id", h.UpdateRepository)
		g.DELETE("/repositories/:id", h.base.RequireRoles("super_admin"), h.DeleteRepository)

		g.POST("/spaces/:id/bind-repo", h.BindRepo)
		g.POST("/sync/:repoId", h.TriggerSync)

		g.GET("/oidc-providers", h.base.RequireRoles("super_admin"), h.ListOIDCProviders)
		g.POST("/oidc-providers", h.base.RequireRoles("super_admin"), h.CreateOIDCProvider)
		g.PUT("/oidc-providers/:id", h.base.RequireRoles("super_admin"), h.UpdateOIDCProvider)
		g.DELETE("/oidc-providers/:id", h.base.RequireRoles("super_admin"), h.DeleteOIDCProvider)

		g.GET("/users", h.base.RequireRoles("super_admin", "admin"), h.ListUsers)
		g.POST("/users", h.base.RequireRoles("super_admin"), h.CreateUser)
		g.PUT("/users/:id", h.base.RequireRoles("super_admin", "admin"), h.UpdateUser)
		g.DELETE("/users/:id", h.base.RequireRoles("super_admin", "admin"), h.DeleteUser)

		g.GET("/groups", h.ListGroups)
		g.POST("/groups", h.CreateGroup)
		g.GET("/groups/:id", h.GetGroup)
		g.PUT("/groups/:id", h.UpdateGroup)
		g.DELETE("/groups/:id", h.DeleteGroup)
		g.GET("/groups/:id/members", h.ListGroupMembers)
		g.POST("/groups/:id/members", h.AddGroupMember)
		g.DELETE("/groups/:id/members/:userId", h.RemoveGroupMember)

		g.GET("/spaces", h.ListAdminSpaces)
		g.PATCH("/spaces/:id", h.UpdateSpace)
		g.GET("/spaces/:id/repositories", h.ListSpaceRepositories)
		g.GET("/spaces/:id/groups", h.ListSpaceGroups)
		g.POST("/spaces/:id/groups", h.AssignSpaceGroup)
		g.DELETE("/spaces/:id/groups/:groupId", h.RemoveSpaceGroup)
		g.GET("/spaces/:id/members", h.ListSpaceMembers)
		g.POST("/spaces/:id/members", h.AssignSpaceMember)
		g.DELETE("/spaces/:id/members/:userId", h.RemoveSpaceMember)
		g.DELETE("/spaces/:id/repositories/:repoId", h.UnbindRepository)

		g.GET("/audit-logs", h.base.RequireRoles("super_admin"), h.ListAuditLogs)
		g.GET("/rag/status", h.GetRAGStatus)
		g.POST("/rag/reindex", h.base.RequireRoles("super_admin"), h.TriggerRAGReindex)
		g.GET("/analytics/overview", h.base.AnalyticsOverview)
		g.GET("/spaces/:id/page-acl", h.base.ListPageACLRules)
		g.POST("/spaces/:id/page-acl", h.base.CreatePageACLRule)
		g.DELETE("/page-acl/:ruleId", h.base.DeletePageACLRule)
	}

	register(r.Group("/api/admin"))
	register(r.Group("/admin"))
}

func (h *AdminHandler) GetSystemSettings(c *gin.Context) {
	settings, err := h.admin.GetSystemSettings(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, settings)
}

func (h *AdminHandler) UpdateSystemSettings(c *gin.Context) {
	var patch service.SystemSettings
	if err := c.ShouldBindJSON(&patch); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	settings, err := h.admin.UpdateSystemSettings(c.Request.Context(), c.GetString("userID"), patch)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	h.base.logAudit(c, "settings.update", "system", "platform")
	c.JSON(http.StatusOK, settings)
}

func (h *AdminHandler) GetPublicUITheme(c *gin.Context) {
	theme, err := h.admin.GetUITheme(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"ui_theme": theme})
}

func (h *AdminHandler) UpdateUITheme(c *gin.Context) {
	var body struct {
		UITheme string `json:"ui_theme" binding:"required"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	theme, err := h.admin.SetUITheme(c.Request.Context(), c.GetString("userID"), body.UITheme)
	if err != nil {
		if err.Error() == "invalid ui_theme" {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	h.base.logAudit(c, "settings.ui_theme", "system", body.UITheme)
	c.JSON(http.StatusOK, gin.H{"ui_theme": theme})
}

func (h *AdminHandler) GetPublicUILanguage(c *gin.Context) {
	lang, err := h.admin.GetUILanguage(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"ui_language": lang})
}

func (h *AdminHandler) GetPublicAuthMethods(c *gin.Context) {
	methods, err := h.admin.GetPublicAuthMethods(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, methods)
}

func (h *AdminHandler) UpdateUILanguage(c *gin.Context) {
	var body struct {
		UILanguage string `json:"ui_language" binding:"required"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	lang, err := h.admin.SetUILanguage(c.Request.Context(), c.GetString("userID"), body.UILanguage)
	if err != nil {
		if err.Error() == "invalid ui_language" {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	h.base.logAudit(c, "settings.ui_language", "system", body.UILanguage)
	c.JSON(http.StatusOK, gin.H{"ui_language": lang})
}

func (h *AdminHandler) ListRepositories(c *gin.Context) {
	items, err := h.repos.ListAll(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"items": items})
}

func (h *AdminHandler) CreateRepository(c *gin.Context) {
	var input service.AdminCreateRepositoryInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	repo, err := h.repos.AdminCreate(c.Request.Context(), input)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	h.base.logAudit(c, "repository.create", "repository", repo.ID)
	c.JSON(http.StatusCreated, repo)
}

func (h *AdminHandler) GetRepository(c *gin.Context) {
	detail, err := h.repos.GetByID(c.Request.Context(), c.Param("id"))
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "repository not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, detail)
}

func (h *AdminHandler) UpdateRepository(c *gin.Context) {
	var input service.UpdateRepositoryInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	repo, err := h.repos.Update(c.Request.Context(), c.Param("id"), input)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "repository not found"})
		return
	}
	h.base.logAudit(c, "repository.update", "repository", repo.ID)
	c.JSON(http.StatusOK, repo)
}

func (h *AdminHandler) DeleteRepository(c *gin.Context) {
	if err := h.repos.Delete(c.Request.Context(), c.Param("id")); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "repository not found"})
		return
	}
	h.base.logAudit(c, "repository.delete", "repository", c.Param("id"))
	c.JSON(http.StatusOK, gin.H{"status": "deleted"})
}

func (h *AdminHandler) BindRepo(c *gin.Context) {
	var body struct {
		RepositoryID string `json:"repository_id" binding:"required"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	repo, err := h.repos.BindToSpace(c.Request.Context(), c.Param("id"), body.RepositoryID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, repo)
}

func (h *AdminHandler) TriggerSync(c *gin.Context) {
	repoID := c.Param("repoId")
	if _, err := h.repos.GetByID(c.Request.Context(), repoID); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "repository not found"})
		return
	}
	if h.sync == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "sync service not configured"})
		return
	}
	code, body, err := h.sync.TriggerSync(c.Request.Context(), repoID)
	if err != nil {
		c.JSON(http.StatusBadGateway, gin.H{"error": err.Error()})
		return
	}
	if code >= 200 && code < 300 && h.auditOn {
		h.audit.Log(c.Request.Context(), c.GetString("userID"), "repo.sync", "repository", repoID, c.ClientIP(), c.GetHeader("User-Agent"))
	}
	c.Data(code, "application/json", body)
}

func (h *AdminHandler) ListOIDCProviders(c *gin.Context) {
	items, err := h.admin.ListOIDCProviders(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"items": items})
}

func (h *AdminHandler) CreateOIDCProvider(c *gin.Context) {
	var input service.OIDCProviderInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	p, err := h.admin.CreateOIDCProvider(c.Request.Context(), input)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	h.base.logAudit(c, "oidc.create", "oidc_provider", p.ID)
	c.JSON(http.StatusCreated, p)
}

func (h *AdminHandler) UpdateOIDCProvider(c *gin.Context) {
	var input service.OIDCProviderInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	p, err := h.admin.UpdateOIDCProvider(c.Request.Context(), c.Param("id"), input)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "provider not found"})
		return
	}
	h.base.logAudit(c, "oidc.update", "oidc_provider", p.ID)
	c.JSON(http.StatusOK, p)
}

func (h *AdminHandler) DeleteOIDCProvider(c *gin.Context) {
	if err := h.admin.DeleteOIDCProvider(c.Request.Context(), c.Param("id")); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "provider not found"})
		return
	}
	h.base.logAudit(c, "oidc.delete", "oidc_provider", c.Param("id"))
	c.JSON(http.StatusOK, gin.H{"status": "deleted"})
}

func (h *AdminHandler) ListUsers(c *gin.Context) {
	users, err := h.admin.ListUsers(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	type userRow struct {
		ID          string   `json:"id"`
		Email       string   `json:"email"`
		DisplayName string   `json:"display_name"`
		IsActive    bool     `json:"is_active"`
		Roles       []string `json:"roles"`
	}
	out := make([]userRow, 0, len(users))
	for _, u := range users {
		roles, _ := h.admin.ListUserRoles(c.Request.Context(), u.ID)
		out = append(out, userRow{
			ID: u.ID, Email: u.Email, DisplayName: u.DisplayName,
			IsActive: u.IsActive, Roles: roles,
		})
	}
	c.JSON(http.StatusOK, gin.H{"items": out})
}

func (h *AdminHandler) CreateUser(c *gin.Context) {
	var input service.CreateUserInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	user, roles, err := h.admin.CreateUser(c.Request.Context(), input)
	if err != nil {
		switch err.Error() {
		case "email already exists":
			c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
		case "invalid email":
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		default:
			if strings.HasPrefix(err.Error(), "invalid role: ") {
				c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
				return
			}
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
		return
	}
	h.base.logAudit(c, "user.create", "user", user.ID)
	c.JSON(http.StatusCreated, gin.H{
		"id":           user.ID,
		"email":        user.Email,
		"display_name": user.DisplayName,
		"is_active":    user.IsActive,
		"roles":        roles,
	})
}

func (h *AdminHandler) UpdateUser(c *gin.Context) {
	var input service.UpdateUserInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	user, roles, err := h.admin.UpdateUser(c.Request.Context(), getRoles(c), c.Param("id"), input)
	if err != nil {
		switch err.Error() {
		case "user not found":
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		case "email already exists", "invalid email", "password must be at least 8 characters":
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		default:
			if strings.HasPrefix(err.Error(), "invalid role: ") ||
				strings.HasPrefix(err.Error(), "forbidden: ") {
				c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
				return
			}
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
		return
	}
	h.base.logAudit(c, "user.update", "user", user.ID)
	c.JSON(http.StatusOK, gin.H{
		"id":           user.ID,
		"email":        user.Email,
		"display_name": user.DisplayName,
		"is_active":    user.IsActive,
		"roles":        roles,
	})
}

func (h *AdminHandler) DeleteUser(c *gin.Context) {
	err := h.admin.DeleteUser(c.Request.Context(), c.GetString("userID"), getRoles(c), c.Param("id"))
	if err != nil {
		switch err.Error() {
		case "user not found":
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		default:
			if strings.HasPrefix(err.Error(), "forbidden: ") {
				c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
				return
			}
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
		return
	}
	h.base.logAudit(c, "user.delete", "user", c.Param("id"))
	c.JSON(http.StatusOK, gin.H{"status": "deleted"})
}

func (h *AdminHandler) ListAdminSpaces(c *gin.Context) {
	spaces, err := h.spaces.List(c.Request.Context(), c.GetString("userID"), true)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"items": spaces})
}

func (h *AdminHandler) UpdateSpace(c *gin.Context) {
	var input service.UpdateSpaceInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	space, err := h.spaces.Update(c.Request.Context(), c.Param("id"), input)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "space not found"})
		return
	}
	c.JSON(http.StatusOK, space)
}

func (h *AdminHandler) ListSpaceRepositories(c *gin.Context) {
	if _, err := h.spaces.GetByID(c.Request.Context(), c.Param("id")); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "space not found"})
		return
	}
	repos, err := h.repos.ListBySpace(c.Request.Context(), c.Param("id"))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"items": repos})
}

func (h *AdminHandler) UnbindRepository(c *gin.Context) {
	spaceID := c.Param("id")
	repoID := c.Param("repoId")
	if err := h.repos.UnbindFromSpace(c.Request.Context(), spaceID, repoID); err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "repository not found in this space"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "removed"})
}

func (h *AdminHandler) ListGroups(c *gin.Context) {
	items, err := h.groups.List(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"items": items})
}

func (h *AdminHandler) CreateGroup(c *gin.Context) {
	var input service.CreateGroupInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	grp, err := h.groups.Create(c.Request.Context(), input)
	if err != nil {
		if err.Error() == "group name already exists" {
			c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, grp)
}

func (h *AdminHandler) GetGroup(c *gin.Context) {
	grp, err := h.groups.Get(c.Request.Context(), c.Param("id"))
	if err != nil {
		if err.Error() == "group not found" {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, grp)
}

func (h *AdminHandler) UpdateGroup(c *gin.Context) {
	var input service.UpdateGroupInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	grp, err := h.groups.Update(c.Request.Context(), c.Param("id"), input)
	if err != nil {
		switch err.Error() {
		case "group not found":
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		case "group name already exists":
			c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
		default:
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		}
		return
	}
	c.JSON(http.StatusOK, grp)
}

func (h *AdminHandler) DeleteGroup(c *gin.Context) {
	if err := h.groups.Delete(c.Request.Context(), c.Param("id")); err != nil {
		if err.Error() == "group not found" {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "deleted"})
}

func (h *AdminHandler) ListGroupMembers(c *gin.Context) {
	items, err := h.groups.ListMembers(c.Request.Context(), c.Param("id"))
	if err != nil {
		if err.Error() == "group not found" {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"items": items})
}

func (h *AdminHandler) AddGroupMember(c *gin.Context) {
	var input service.AddGroupMemberInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := h.groups.AddMember(c.Request.Context(), c.Param("id"), input.UserID); err != nil {
		switch err.Error() {
		case "group not found", "user not found":
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		case "user already in group":
			c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
		return
	}
	c.JSON(http.StatusCreated, gin.H{"status": "added"})
}

func (h *AdminHandler) RemoveGroupMember(c *gin.Context) {
	if err := h.groups.RemoveMember(c.Request.Context(), c.Param("id"), c.Param("userId")); err != nil {
		if err.Error() == "member not found" {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "removed"})
}

func (h *AdminHandler) ListSpaceGroups(c *gin.Context) {
	if _, err := h.spaces.GetByID(c.Request.Context(), c.Param("id")); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "space not found"})
		return
	}
	items, err := h.groups.ListSpaceGroups(c.Request.Context(), c.Param("id"))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"items": items})
}

func (h *AdminHandler) AssignSpaceGroup(c *gin.Context) {
	if _, err := h.spaces.GetByID(c.Request.Context(), c.Param("id")); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "space not found"})
		return
	}
	var input service.AssignSpaceGroupInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := h.groups.AssignToSpace(c.Request.Context(), c.Param("id"), input); err != nil {
		switch err.Error() {
		case "group not found", "space not found":
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		default:
			if strings.HasPrefix(err.Error(), "invalid role: ") {
				c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
				return
			}
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
		return
	}
	c.JSON(http.StatusCreated, gin.H{"status": "assigned"})
}

func (h *AdminHandler) RemoveSpaceGroup(c *gin.Context) {
	if err := h.groups.RemoveFromSpace(c.Request.Context(), c.Param("id"), c.Param("groupId")); err != nil {
		if err.Error() == "group not assigned to space" {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "removed"})
}

func (h *AdminHandler) ListSpaceMembers(c *gin.Context) {
	if _, err := h.spaces.GetByID(c.Request.Context(), c.Param("id")); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "space not found"})
		return
	}
	items, err := h.spaces.ListMembers(c.Request.Context(), c.Param("id"))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"items": items})
}

func (h *AdminHandler) AssignSpaceMember(c *gin.Context) {
	if _, err := h.spaces.GetByID(c.Request.Context(), c.Param("id")); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "space not found"})
		return
	}
	var input service.AssignSpaceMemberInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := h.spaces.AssignMember(c.Request.Context(), c.Param("id"), input); err != nil {
		switch err.Error() {
		case "user not found", "space not found":
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		default:
			if strings.HasPrefix(err.Error(), "invalid role: ") {
				c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
				return
			}
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
		return
	}
	c.JSON(http.StatusCreated, gin.H{"status": "assigned"})
}

func (h *AdminHandler) RemoveSpaceMember(c *gin.Context) {
	if err := h.spaces.RemoveMember(c.Request.Context(), c.Param("id"), c.Param("userId")); err != nil {
		if err.Error() == "member not found" {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "removed"})
}
