package handler

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/konstpic/treepage/backend/server/internal/book"
	"github.com/konstpic/treepage/backend/server/internal/llm"
	"github.com/konstpic/treepage/backend/server/internal/search"
	"github.com/konstpic/treepage/backend/server/internal/service"
	"github.com/konstpic/treepage/backend/server/internal/syncclient"
	"github.com/konstpic/treepage/backend/server/internal/translate"
	pkgjwt "github.com/konstpic/treepage/backend/pkg/jwt"
	"gorm.io/gorm"
)

type Handler struct {
	spaces    *service.SpaceService
	docs      *service.DocumentService
	repos     *service.RepositoryService
	audit     *service.AuditService
	search    search.Searcher
	books     *book.Service
	translate *translate.Service
	sync      *syncclient.Client
	jwt       *pkgjwt.Manager
	auditOn   bool
}

func New(
	spaces *service.SpaceService,
	docs *service.DocumentService,
	repos *service.RepositoryService,
	audit *service.AuditService,
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
		audit: audit, search: searcher, sync: syncClient,
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
	}

	apiAuth := r.Group("/api")
	apiAuth.Use(h.AuthMiddleware())
	{
		apiAuth.POST("/spaces", h.RequireRoles("super_admin", "admin"), h.CreateSpace)
		apiAuth.POST("/spaces/:slug/documents", h.CreateDocument)
		apiAuth.PUT("/documents/:id", h.UpdateDocument)
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
	docs, err := h.docs.ListBySpace(c.Request.Context(), space.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
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
	doc, err := h.docs.GetBySlug(c.Request.Context(), space.ID, c.Param("docSlug"))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "document not found"})
		return
	}
	lang := c.Query("lang")
	if lang == "" {
		lang = acceptLanguage(c.GetHeader("Accept-Language"))
	}
	view, err := h.translate.LocalizeDocument(c.Request.Context(), doc, lang, h.canUseLLMInSpace(c, space))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
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
		Content string `json:"content"`
		Title   string `json:"title"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	doc, err = h.docs.Update(c.Request.Context(), c.Param("id"), c.GetString("userID"), body.Content, body.Title)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "document not found"})
		return
	}
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
