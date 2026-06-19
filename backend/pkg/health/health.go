package health

import (
	"context"
	"net/http"
	"sync/atomic"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

type Checker func(ctx context.Context) error

type Handler struct {
	ready atomic.Bool
	check Checker
}

func NewHandler(check Checker) *Handler {
	h := &Handler{check: check}
	h.ready.Store(false)
	return h
}

func (h *Handler) SetReady(v bool) {
	h.ready.Store(v)
}

func (h *Handler) Register(r *gin.Engine) {
	r.GET("/liveness", h.liveness)
	r.GET("/readiness", h.readiness)
	r.GET("/health", h.health)
	r.GET("/metrics", gin.WrapH(promhttp.Handler()))
}

func (h *Handler) health(c *gin.Context) {
	// Alias for tools/Docker expecting /health (same semantics as readiness).
	h.readiness(c)
}

func (h *Handler) liveness(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"status": "alive"})
}

func (h *Handler) readiness(c *gin.Context) {
	if !h.ready.Load() {
		c.JSON(http.StatusServiceUnavailable, gin.H{"status": "not_ready"})
		return
	}
	if h.check != nil {
		if err := h.check(c.Request.Context()); err != nil {
			c.JSON(http.StatusServiceUnavailable, gin.H{"status": "unhealthy", "error": err.Error()})
			return
		}
	}
	c.JSON(http.StatusOK, gin.H{"status": "ready"})
}
