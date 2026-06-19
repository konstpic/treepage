package handler

import "github.com/gin-gonic/gin"

func (h *Handler) logAudit(c *gin.Context, action, resourceType, resourceID string) {
	if !h.auditOn {
		return
	}
	h.audit.Log(c.Request.Context(), c.GetString("userID"), action, resourceType, resourceID, c.ClientIP(), c.GetHeader("User-Agent"))
}
