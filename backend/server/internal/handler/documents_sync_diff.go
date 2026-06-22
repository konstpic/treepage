package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func (h *Handler) GetDocumentSyncDiff(c *gin.Context) {
	doc, err := h.docs.GetByID(c.Request.Context(), c.Param("id"))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "document not found"})
		return
	}
	space, err := h.spaces.GetByID(c.Request.Context(), doc.SpaceID)
	if err != nil || !h.requireDocumentAccess(c, space, doc) {
		return
	}
	if doc.RepositoryID == nil || !doc.HasPendingChanges {
		c.JSON(http.StatusBadRequest, gin.H{"error": "document has no git sync conflict"})
		return
	}
	diff, err := h.docs.SyncConflictDiff(c.Request.Context(), doc.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, diff)
}
