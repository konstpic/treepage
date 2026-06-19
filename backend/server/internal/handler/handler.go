package handler

import (
	"context"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/konstpic/treepage/backend/pkg/models"
	"github.com/konstpic/treepage/backend/server/internal/book"
	"github.com/konstpic/treepage/backend/server/internal/llm"
	"github.com/konstpic/treepage/backend/server/internal/rag"
	"github.com/konstpic/treepage/backend/server/internal/search"
	"github.com/konstpic/treepage/backend/server/internal/service"
	"github.com/konstpic/treepage/backend/server/internal/syncclient"
	"github.com/konstpic/treepage/backend/server/internal/translate"
	pkgjwt "github.com/konstpic/treepage/backend/pkg/jwt"
	"gorm.io/gorm"
)

type Handler struct {
	spaces        *service.SpaceService
	docs          *service.DocumentService
	repos         *service.RepositoryService
	audit         *service.AuditService
	prefs         *service.UserPrefsService
	notifications *service.NotificationService
	attachments   *service.AttachmentService
	pageACL       *service.PageACLService
	comments      *service.CommentService
	analytics     *service.AnalyticsService
	rag           *rag.Service
	search        search.Searcher
	books         *book.Service
	translate     *translate.Service
	sync          *syncclient.Client
	jwt           *pkgjwt.Manager
	auditOn       bool
}

func New(
	spaces *service.SpaceService,
	docs *service.DocumentService,
	repos *service.RepositoryService,
	audit *service.AuditService,
	prefs *service.UserPrefsService,
	notifications *service.NotificationService,
	attachments *service.AttachmentService,
	pageACL *service.PageACLService,
	comments *service.CommentService,
	analytics *service.AnalyticsService,
	ragSvc *rag.Service,
	searcher search.Searcher,
	llmClient *llm.Client,
	admin *service.AdminService,
	db *gorm.DB,
	jwt *pkgjwt.Manager,
	syncClient *syncclient.Client,
	auditOn bool,
) *Handler {
	return &Handler{
		spaces: spaces, docs: docs, repos: repos,
		audit: audit, prefs: prefs, notifications: notifications, attachments: attachments,
		pageACL: pageACL, comments: comments, analytics: analytics, rag: ragSvc,
		search: searcher, sync: syncClient,
		books: book.NewService(docs, db, llmClient),
		translate: translate.NewService(db, llmClient, admin),
		jwt: jwt, auditOn: auditOn,
	}
}

func (h *Handler) Register(r *gin.Engine) {
	pub := r.Group("/api/public")
	{
		pub.GET("/branding", h.Branding)
		pub.GET("/spaces", h.ListPublicSpaces)
	}

	api := r.Group("/api")
	api.Use(h.OptionalAuthMiddleware())
	{
		api.GET("/spaces", h.ListSpaces)
		api.GET("/spaces/:slug", h.GetSpace)
		api.GET("/spaces/:slug/documents", h.ListDocuments)
		api.GET("/spaces/:slug/documents/:docSlug", h.GetDocument)
		api.GET("/spaces/:slug/books", h.ListBooks)
		api.GET("/spaces/:slug/books/:bookSlug", h.GetSavedBook)
		api.GET("/search", h.Search)
		api.GET("/attachments/:id/download", h.DownloadAttachment)
	}

	apiAuth := r.Group("/api")
	apiAuth.Use(h.AuthMiddleware())
	{
		apiAuth.POST("/spaces", h.RequireRoles("super_admin", "admin"), h.CreateSpace)
		apiAuth.POST("/spaces/:slug/documents", h.CreateDocument)
		apiAuth.PUT("/documents/:id", h.UpdateDocument)
		apiAuth.DELETE("/documents/:id", h.DeleteDocument)
		apiAuth.POST("/documents/:id/publish", h.PublishDocument)
		apiAuth.POST("/documents/:id/publish-local", h.PublishDocumentLocal)
		apiAuth.POST("/documents/:id/revert/:version", h.RevertDocumentVersion)
		apiAuth.GET("/documents/:id/attachments", h.ListDocumentAttachments)
		apiAuth.POST("/documents/:id/attachments", h.UploadDocumentAttachment)
		apiAuth.DELETE("/attachments/:id", h.DeleteAttachment)
		apiAuth.GET("/me/favorites", h.ListFavorites)
		apiAuth.POST("/me/favorites/:documentId", h.AddFavorite)
		apiAuth.DELETE("/me/favorites/:documentId", h.RemoveFavorite)
		apiAuth.GET("/me/recent", h.ListRecent)
		apiAuth.GET("/notifications", h.ListNotifications)
		apiAuth.GET("/notifications/unread-count", h.NotificationUnreadCount)
		apiAuth.POST("/notifications/:id/read", h.MarkNotificationRead)
		apiAuth.POST("/notifications/read-all", h.MarkAllNotificationsRead)
		apiAuth.POST("/documents/:id/publish-workflow", h.PublishDocumentWorkflow)
		apiAuth.GET("/documents/:id/comments", h.ListDocumentComments)
		apiAuth.POST("/documents/:id/comments", h.CreateDocumentComment)
		apiAuth.DELETE("/comments/:id", h.DeleteComment)
		apiAuth.POST("/documents/:id/submit-review", h.SubmitDocumentReview)
		apiAuth.POST("/documents/:id/approve", h.ApproveDocumentReview)
		apiAuth.POST("/documents/:id/reject-review", h.RejectDocumentReview)
		apiAuth.POST("/rag/ask", h.RAGAsk)
		apiAuth.POST("/rag/feedback", h.RAGFeedback)
		apiAuth.GET("/documents/:id/versions", h.ListDocumentVersions)
		apiAuth.GET("/documents/:id/versions/:version/diff", h.DiffDocumentVersions)
		apiAuth.GET("/documents/:id/versions/:version", h.GetDocumentVersion)
		apiAuth.GET("/spaces/:slug/repositories", h.ListRepositories)
		apiAuth.POST("/spaces/:slug/repositories/:repoId/sync", h.TriggerSpaceSync)
		apiAuth.POST("/spaces/:slug/books", h.CreateBook)
		apiAuth.POST("/spaces/:slug/books/:bookSlug/generate", h.GenerateBook)
		apiAuth.DELETE("/spaces/:slug/books/:bookSlug", h.DeleteBook)
	}
}

func (h *Handler) ListPublicSpaces(c *gin.Context) {
	spaces, err := h.spaces.List(c.Request.Context(), "", false)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"items": spaces})
}

func (h *Handler) Branding(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"project_name": "TreePage",
		"project_code": "treepage",
	})
}

func (h *Handler) ListSpaces(c *gin.Context) {
	roles := getRoles(c)
	spaces, err := h.spaces.List(c.Request.Context(), c.GetString("userID"), service.HasRole(roles, "super_admin"))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"items": spaces})
}

func (h *Handler) CreateSpace(c *gin.Context) {
	var input service.CreateSpaceInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	space, err := h.spaces.Create(c.Request.Context(), input, c.GetString("userID"))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if h.auditOn {
		h.audit.Log(c.Request.Context(), c.GetString("userID"), "space.create", "space", space.ID, c.ClientIP(), c.GetHeader("User-Agent"))
	}
	c.JSON(http.StatusCreated, space)
}

func (h *Handler) GetSpace(c *gin.Context) {
	space, err := h.spaces.GetBySlug(c.Request.Context(), c.Param("slug"))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "space not found"})
		return
	}
	if !h.requireSpaceAccess(c, space) {
		return
	}
	out := gin.H{
		"id":          space.ID,
		"slug":        space.Slug,
		"name":        space.Name,
		"description": space.Description,
		"is_public":   space.IsPublic,
		"created_at":  space.CreatedAt,
		"updated_at":  space.UpdatedAt,
	}
	if c.GetString("userID") != "" {
		if role, err := h.getEffectiveSpaceRole(c, space); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		} else if role != "" {
			out["my_role"] = role
		}
		canEdit, err := h.canEditInSpace(c, space)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		out["can_edit"] = canEdit
	}
	c.JSON(http.StatusOK, out)
}

func (h *Handler) ListDocuments(c *gin.Context) {
	space, err := h.spaces.GetBySlug(c.Request.Context(), c.Param("slug"))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "space not found"})
		return
	}
	if !h.requireSpaceAccess(c, space) {
		return
	}
	canEdit, err := h.canEditInSpace(c, space)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	var docs []models.Document
	if canEdit {
		docs, err = h.docs.ListBySpaceIncludingDrafts(c.Request.Context(), space.ID)
	} else {
		docs, err = h.docs.ListBySpace(c.Request.Context(), space.ID)
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	effective, _ := h.getEffectiveSpaceRole(c, space)
	docs = h.pageACL.FilterDocuments(c.Request.Context(), c.GetString("userID"), effective,
		service.HasRole(getRoles(c), "super_admin"), docs)
	if !canEdit {
		var visible []models.Document
		for _, d := range docs {
			if service.DocumentVisibleToViewer(&d) {
				visible = append(visible, d)
			}
		}
		docs = visible
	}
	c.JSON(http.StatusOK, gin.H{"items": docs})
}

func (h *Handler) GetDocument(c *gin.Context) {
	space, err := h.spaces.GetBySlug(c.Request.Context(), c.Param("slug"))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "space not found"})
		return
	}
	if !h.requireSpaceAccess(c, space) {
		return
	}
	canEdit, err := h.canEditInSpace(c, space)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	doc, err := h.docs.GetBySlugVisible(c.Request.Context(), space.ID, c.Param("docSlug"), canEdit)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "document not found"})
		return
	}
	if !canEdit && !service.DocumentVisibleToViewer(doc) {
		c.JSON(http.StatusNotFound, gin.H{"error": "document not found"})
		return
	}
	if !h.requireDocumentAccess(c, space, doc) {
		return
	}
	userID := c.GetString("userID")
	if userID != "" {
		_ = h.prefs.RecordView(c.Request.Context(), userID, doc.ID, space.ID)
	}
	h.analytics.RecordView(c.Request.Context(), doc.ID)
	lang := c.Query("lang")
	if lang == "" {
		lang = acceptLanguage(c.GetHeader("Accept-Language"))
	}
	view, err := h.translate.LocalizeDocument(c.Request.Context(), doc, lang, h.canUseLLMInSpace(c, space))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if userID != "" {
		if fav, err := h.prefs.IsFavorite(c.Request.Context(), userID, doc.ID); err == nil {
			view.IsFavorite = fav
		}
	}
	c.JSON(http.StatusOK, view)
}

func (h *Handler) CreateDocument(c *gin.Context) {
	space, err := h.spaces.GetBySlug(c.Request.Context(), c.Param("slug"))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "space not found"})
		return
	}
	if !h.requireSpaceRole(c, space, service.SpaceRoleEditor) {
		return
	}
	var input service.CreateDocumentInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	doc, err := h.docs.Create(c.Request.Context(), space.ID, c.GetString("userID"), input)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if doc.IsPublished {
		h.notifications.NotifySpaceEditors(
			c.Request.Context(), space.ID, c.GetString("userID"),
			"document.created", "New document", doc.Title,
			"document", doc.ID,
		)
	}
	h.logAudit(c, "document.create", "document", doc.ID)
	go func(d models.Document) { _ = h.rag.ReindexDocument(context.Background(), &d) }(*doc)
	c.JSON(http.StatusCreated, doc)
}

func (h *Handler) UpdateDocument(c *gin.Context) {
	doc, err := h.docs.GetByID(c.Request.Context(), c.Param("id"))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "document not found"})
		return
	}
	space, err := h.spaces.GetByID(c.Request.Context(), doc.SpaceID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "space not found"})
		return
	}
	if !h.requireSpaceRole(c, space, service.SpaceRoleEditor) {
		return
	}
	var body struct {
		Content     string `json:"content"`
		Title       string `json:"title"`
		IsPublished *bool  `json:"is_published"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	wasPublished := doc.IsPublished
	doc, err = h.docs.UpdateFull(c.Request.Context(), c.Param("id"), c.GetString("userID"), body.Content, body.Title, body.IsPublished)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "document not found"})
		return
	}
	if body.IsPublished != nil && *body.IsPublished && !wasPublished {
		h.notifications.NotifySpaceEditors(
			c.Request.Context(), space.ID, c.GetString("userID"),
			"document.published", "Document published", doc.Title,
			"document", doc.ID,
		)
	}
	h.logAudit(c, "document.update", "document", doc.ID)
	go func(d models.Document) { _ = h.rag.ReindexDocument(context.Background(), &d) }(*doc)
	c.JSON(http.StatusOK, doc)
}

func (h *Handler) ListRepositories(c *gin.Context) {
	space, err := h.spaces.GetBySlug(c.Request.Context(), c.Param("slug"))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "space not found"})
		return
	}
	if !h.requireSpaceAccess(c, space) {
		return
	}
	repos, err := h.repos.ListBySpace(c.Request.Context(), space.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"items": repos})
}

func (h *Handler) CreateRepository(c *gin.Context) {
	space, err := h.spaces.GetBySlug(c.Request.Context(), c.Param("slug"))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "space not found"})
		return
	}
	var input service.CreateRepositoryInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	repo, err := h.repos.Create(c.Request.Context(), space.ID, input)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, repo)
}

func (h *Handler) Search(c *gin.Context) {
	q := search.Query{
		Text:    c.Query("q"),
		SpaceID: c.Query("space_id"),
		Author:  c.Query("author"),
		Limit:   20,
	}
	if tags := c.Query("tags"); tags != "" {
		q.Tags = strings.Split(tags, ",")
	}
	if spaceSlug := c.Query("space_slug"); spaceSlug != "" {
		if sp, err := h.spaces.GetBySlug(c.Request.Context(), spaceSlug); err == nil {
			q.SpaceID = sp.ID
		}
	}
	allowed, err := h.spaces.AccessibleSpaceIDs(
		c.Request.Context(),
		c.GetString("userID"),
		service.HasRole(getRoles(c), "super_admin"),
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	q.AllowedSpaceIDs = allowed
	results, total, err := h.search.Search(c.Request.Context(), q)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	h.analytics.LogSearch(c.Request.Context(), c.GetString("userID"), q.Text, int(total))
	c.JSON(http.StatusOK, gin.H{"items": results, "total": total})
}

func (h *Handler) AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		auth := c.GetHeader("Authorization")
		if !strings.HasPrefix(auth, "Bearer ") {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "missing token"})
			return
		}
		claims, err := h.jwt.ParseAccess(strings.TrimPrefix(auth, "Bearer "))
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid token"})
			return
		}
		c.Set("userID", claims.UserID)
		c.Set("email", claims.Email)
		c.Set("roles", claims.Roles)
		c.Set("authenticated", true)
		c.Next()
	}
}

func (h *Handler) OptionalAuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		auth := c.GetHeader("Authorization")
		if strings.HasPrefix(auth, "Bearer ") {
			claims, err := h.jwt.ParseAccess(strings.TrimPrefix(auth, "Bearer "))
			if err == nil {
				c.Set("userID", claims.UserID)
				c.Set("email", claims.Email)
				c.Set("roles", claims.Roles)
				c.Set("authenticated", true)
			}
		}
		c.Next()
	}
}

func (h *Handler) RequireRoles(roles ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		if !service.HasRole(getRoles(c), roles...) {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "forbidden"})
			return
		}
		c.Next()
	}
}

func acceptLanguage(header string) string {
	header = strings.TrimSpace(strings.ToLower(header))
	if header == "" {
		return ""
	}
	part := header
	if i := strings.IndexAny(part, ",;"); i >= 0 {
		part = part[:i]
	}
	part = strings.TrimSpace(part)
	if i := strings.Index(part, "-"); i >= 0 {
		part = part[:i]
	}
	if part == "ru" || part == "en" {
		return part
	}
	return ""
}
