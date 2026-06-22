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
	status := h.ragWorker.Status()
	if h.base.rag != nil {
		if stats, err := h.base.rag.IndexStats(c.Request.Context()); err == nil {
			status.PublishedDocuments = stats.PublishedDocuments
			status.DocumentsWithChunks = stats.DocumentsWithChunks
			status.ChunksTotal = stats.ChunksTotal
			status.ChunksEmbedded = stats.ChunksEmbedded
			status.ChunksPending = stats.ChunksPending
			status.EmbeddingsEnabled = stats.EmbeddingsEnabled
			if status.DocumentsTotal == 0 {
				status.DocumentsTotal = stats.PublishedDocuments
			}
			if status.DocumentsDone == 0 && !status.Running {
				status.DocumentsDone = stats.DocumentsWithChunks
			}
		}
	}
	c.JSON(200, status)
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
