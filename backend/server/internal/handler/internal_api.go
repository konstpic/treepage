package handler

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/konstpic/treepage/backend/pkg/models"
)

func (h *Handler) RegisterInternal(r *gin.RouterGroup) {
	r.POST("/documents/:id/reindex", h.InternalReindexDocument)
	r.DELETE("/documents/:id/search-index", h.InternalDeleteSearchIndex)
}

func (h *Handler) InternalReindexDocument(c *gin.Context) {
	doc, err := h.docs.GetByID(c.Request.Context(), c.Param("id"))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "document not found"})
		return
	}
	go func(d models.Document) {
		_ = h.rag.ReindexDocument(context.Background(), &d)
	}(*doc)
	h.indexDocumentAsync(doc)
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

func (h *Handler) InternalDeleteSearchIndex(c *gin.Context) {
	h.deleteDocumentFromIndex(c.Param("id"))
	c.JSON(http.StatusOK, gin.H{"ok": true})
}
