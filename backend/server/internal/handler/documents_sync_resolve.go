package handler

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/konstpic/treepage/backend/pkg/models"
	"github.com/konstpic/treepage/backend/server/internal/service"
)

func (h *Handler) ResolveDocumentSyncConflict(c *gin.Context) {
	doc, err := h.docs.GetByID(c.Request.Context(), c.Param("id"))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "document not found"})
		return
	}
	space, err := h.spaces.GetByID(c.Request.Context(), doc.SpaceID)
	if err != nil || !h.requireDocumentAccess(c, space, doc) {
		return
	}
	if !h.requireSpaceRole(c, space, service.SpaceRoleEditor) {
		return
	}
	var body struct {
		Strategy string `json:"strategy" binding:"required"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	updated, err := h.docs.ResolveSyncConflict(c.Request.Context(), doc.ID, c.GetString("userID"), body.Strategy)
	if err != nil {
		if err == service.ErrInvalidSyncStrategy {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	h.logAudit(c, "document.sync_resolve", "document", doc.ID)
	go func(d models.Document) { _ = h.rag.ReindexDocument(context.Background(), &d) }(*updated)
	h.indexDocumentAsync(updated)
	c.JSON(http.StatusOK, updated)
}
