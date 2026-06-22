package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/konstpic/treepage/backend/server/internal/service"
)

func (h *Handler) DeleteDocument(c *gin.Context) {
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
	if err := h.docs.Delete(c.Request.Context(), doc.ID); err != nil {
		if err == service.ErrDocumentNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "document not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	h.logAudit(c, "document.delete", "document", doc.ID)
	h.deleteDocumentFromIndex(doc.ID)
	c.JSON(http.StatusOK, gin.H{"status": "deleted"})
}

func (h *Handler) RevertDocumentVersion(c *gin.Context) {
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
	versionNum, err := strconv.Atoi(c.Param("version"))
	if err != nil || versionNum < 1 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid version"})
		return
	}
	doc, err = h.docs.RevertToVersion(c.Request.Context(), doc.ID, c.GetString("userID"), versionNum)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	h.logAudit(c, "document.revert", "document", doc.ID)
	c.JSON(http.StatusOK, doc)
}

func (h *AdminHandler) ListAuditLogs(c *gin.Context) {
	limit := 50
	offset := 0
	if v := c.Query("limit"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 && n <= 200 {
			limit = n
		}
	}
	if v := c.Query("offset"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n >= 0 {
			offset = n
		}
	}
	items, total, err := h.audit.List(c.Request.Context(), limit, offset, c.Query("action"))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"items": items, "total": total})
}
