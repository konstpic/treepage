package middleware

import (
	"github.com/gin-gonic/gin"
	"github.com/konstpic/treepage/backend/pkg/metrics"
)

// PrometheusHTTP records treepage_http_requests_total for each request.
func PrometheusHTTP(serviceName string) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()
		path := c.FullPath()
		if path == "" {
			path = c.Request.URL.Path
		}
		metrics.ObserveHTTP(serviceName, c.Request.Method, path, c.Writer.Status())
	}
}
