package handler

import (
	"context"
	"time"

	"github.com/gin-gonic/gin"
)

func (h *AdminHandler) GetRAGStatus(c *gin.Context) {
	if h.ragWorker == nil {
		c.JSON(200, gin.H{"phase": "unavailable"})
		return
	}
	c.JSON(200, h.ragWorker.Status())
}

func (h *AdminHandler) TriggerRAGReindex(c *gin.Context) {
	if h.ragWorker == nil {
		c.JSON(503, gin.H{"error": "rag worker unavailable"})
		return
	}
	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()
	h.ragWorker.TriggerReindex(ctx)
	h.base.logAudit(c, "rag.reindex", "system", "rag")
	c.JSON(202, gin.H{"ok": true, "status": h.ragWorker.Status()})
}
