package middleware

import (
	"log/slog"
	"net/http"
	"os"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/new-marty/health-connect/internal/apperror"
)

const authEnvVar = "HEALTH_CONNECT_API_TOKEN"

// BearerAuth requires Authorization: Bearer <token> matching HEALTH_CONNECT_API_TOKEN.
// If the env var is unset, the middleware is a no-op and a startup warning is logged
// so local dev stays frictionless.
func BearerAuth() gin.HandlerFunc {
	token := os.Getenv(authEnvVar)
	if token == "" {
		slog.Warn("api auth disabled: " + authEnvVar + " not set; all routes are public")
		return func(c *gin.Context) { c.Next() }
	}
	slog.Info("api auth enabled (bearer token)")

	return func(c *gin.Context) {
		header := c.GetHeader("Authorization")
		const prefix = "Bearer "
		if !strings.HasPrefix(header, prefix) || header[len(prefix):] != token {
			c.AbortWithStatusJSON(http.StatusUnauthorized, apperror.Response{
				Error: apperror.Body{
					Code:    "UNAUTHORIZED",
					Message: "missing or invalid bearer token",
				},
			})
			return
		}
		c.Next()
	}
}
