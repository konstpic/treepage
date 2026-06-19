package handler

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/konstpic/treepage/backend/server/internal/book"
	"github.com/konstpic/treepage/backend/server/internal/service"
	"github.com/konstpic/treepage/backend/server/internal/translate"
	"gorm.io/gorm"
)

func (h *Handler) ListBooks(c *gin.Context) {
	space, err := h.spaces.GetBySlug(c.Request.Context(), c.Param("slug"))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "space not found"})
		return
	}
	if !h.requireSpaceAccess(c, space) {
		return
	}
	resp, err := h.books.List(c.Request.Context(), space.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	effective, err := h.getEffectiveSpaceRole(c, space)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if !service.CanManageBooksInSpace(effective, getRoles(c)) {
		c.JSON(http.StatusOK, gin.H{"saved": resp.Saved})
		return
	}
	c.JSON(http.StatusOK, resp)
}

func (h *Handler) GetSavedBook(c *gin.Context) {
	space, err := h.spaces.GetBySlug(c.Request.Context(), c.Param("slug"))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "space not found"})
		return
	}
	if !h.requireSpaceAccess(c, space) {
		return
	}
	withContent := strings.EqualFold(c.Query("content"), "true") ||
		c.Query("format") == "md" ||
		strings.EqualFold(c.Query("markdown"), "true")
	view, err := h.books.GetSaved(c.Request.Context(), space.ID, c.Param("bookSlug"), withContent)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "book not found"})
		return
	}
	if c.Query("format") == "md" {
		lang := c.Query("lang")
		if lang == "" {
			lang = acceptLanguage(c.GetHeader("Accept-Language"))
		}
		md := view.ContentMarkdown
	allowLLM := h.canUseLLMInSpace(c, space)
		if lang != "" && md != "" && allowLLM {
			if loc, err := h.translate.LocalizeBook(c.Request.Context(), view.ID, view.Title, view.Description, md, lang, true); err == nil && loc.Translated {
				md = loc.Content
			}
		}
		c.Header("Content-Type", "text/markdown; charset=utf-8")
		c.Header("Content-Disposition", "attachment; filename=\""+view.Slug+".md\"")
		c.String(http.StatusOK, md)
		return
	}

	lang := c.Query("lang")
	if lang == "" {
		lang = acceptLanguage(c.GetHeader("Accept-Language"))
	}
	allowLLM := h.canUseLLMInSpace(c, space)
	var tr *translate.BookLocalization
	if withContent && view.ContentMarkdown != "" && lang != "" && allowLLM {
		loc, err := h.translate.LocalizeBook(c.Request.Context(), view.ID, view.Title, view.Description, view.ContentMarkdown, lang, true)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		if loc.Translated {
			view.Title = loc.Title
			view.Description = loc.Description
			view.ContentMarkdown = loc.Content
			tr = loc
		} else if loc.DisplayLanguage != "" {
			tr = loc
		}
	}
	c.JSON(http.StatusOK, bookDetailJSON(view, allowLLM && h.books.LLMAvailable(), tr))
}

func bookDetailJSON(view *book.StoredBookView, llmAvailable bool, tr *translate.BookLocalization) gin.H {
	out := gin.H{
		"id":                  view.ID,
		"slug":                view.Slug,
		"title":               view.Title,
		"description":         view.Description,
		"root_path":           view.RootPath,
		"audience":            view.Audience,
		"focus":               view.Focus,
		"status":              view.Status,
		"source_hash":         view.SourceHash,
		"current_source_hash": view.CurrentHash,
		"sources_stale":       view.SourcesStale,
		"chapter_count":       view.ChapterCount,
		"enhanced":            view.Enhanced,
		"error_message":       view.ErrorMessage,
		"content_markdown":    view.ContentMarkdown,
		"created_at":            view.CreatedAt,
		"updated_at":            view.UpdatedAt,
		"generated_at":        view.GeneratedAt,
		"llm_available":       llmAvailable,
	}
	if tr != nil {
		if tr.Translated {
			out["translated"] = true
		}
		if tr.SourceLanguage != "" {
			out["source_language"] = tr.SourceLanguage
		}
		if tr.DisplayLanguage != "" {
			out["display_language"] = tr.DisplayLanguage
		}
	}
	return out
}

func (h *Handler) CreateBook(c *gin.Context) {
	space, err := h.spaces.GetBySlug(c.Request.Context(), c.Param("slug"))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "space not found"})
		return
	}
	if !h.requireSpaceRole(c, space, service.SpaceRoleEditor) {
		return
	}
	var input book.CreateBookInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	userID := c.GetString("userID")
	var uid *string
	if userID != "" {
		uid = &userID
	}
	view, err := h.books.Create(c.Request.Context(), space.ID, uid, input)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, view)
}

func (h *Handler) GenerateBook(c *gin.Context) {
	space, err := h.spaces.GetBySlug(c.Request.Context(), c.Param("slug"))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "space not found"})
		return
	}
	if !h.requireSpaceRole(c, space, service.SpaceRoleEditor) {
		return
	}
	force := strings.EqualFold(c.Query("force"), "true")
	view, err := h.books.Generate(c.Request.Context(), space.ID, c.Param("bookSlug"), force)
	if err != nil {
		if err.Error() == "generation already in progress" {
			c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusBadGateway, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, bookDetailJSON(view, h.books.LLMAvailable(), nil))
}

func (h *Handler) DeleteBook(c *gin.Context) {
	space, err := h.spaces.GetBySlug(c.Request.Context(), c.Param("slug"))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "space not found"})
		return
	}
	if !h.requireSpaceRole(c, space, service.SpaceRoleEditor) {
		return
	}
	if err := h.books.Delete(c.Request.Context(), space.ID, c.Param("bookSlug")); err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "book not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "deleted"})
}
