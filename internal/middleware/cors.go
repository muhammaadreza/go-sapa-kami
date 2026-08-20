package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func CORS(allowedOrigins []string) gin.HandlerFunc {
	allowed := originAllowlist(allowedOrigins)
	return func(c *gin.Context) {
		origin, hasOrigin := normalizeOrigin(c.GetHeader("Origin"))
		_, originAllowed := allowed[origin]
		if hasOrigin && originAllowed {
			c.Header("Access-Control-Allow-Origin", origin)
			c.Header("Access-Control-Allow-Credentials", "true")
			c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
			c.Header("Access-Control-Allow-Headers", "Content-Type, Authorization")
			c.Header("Vary", "Origin")
		}

		if c.Request.Method == http.MethodOptions {
			if !hasOrigin || !originAllowed {
				c.AbortWithStatus(http.StatusForbidden)
				return
			}

			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		c.Next()
	}
}
