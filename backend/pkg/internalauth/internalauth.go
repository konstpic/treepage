package internalauth

import (
	"crypto/subtle"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
)

const HeaderName = "X-Internal-Token"

func TokenFromEnv() string {
	return os.Getenv("INTERNAL_SERVICE_TOKEN")
}

func Middleware(token string) gin.HandlerFunc {
	return func(c *gin.Context) {
		if token == "" {
			c.AbortWithStatusJSON(http.StatusServiceUnavailable, gin.H{"error": "internal service token not configured"})
			return
		}
		got := c.GetHeader(HeaderName)
		if subtle.ConstantTimeCompare([]byte(got), []byte(token)) != 1 {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
			return
		}
		c.Next()
	}
}

func ClientHeader() (string, string) {
	return HeaderName, TokenFromEnv()
}
