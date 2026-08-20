package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func CSRF(trustedOrigins []string) gin.HandlerFunc {
	trusted := originAllowlist(trustedOrigins)
	return func(c *gin.Context) {
		if isUnsafeMethod(c.Request.Method) && hasAuthCookie(c.Request) {
			origin, valid := requestOrigin(c.Request)
			if _, ok := trusted[origin]; !valid || !ok {
				c.AbortWithStatus(http.StatusForbidden)
				return
			}
		}

		c.Next()
	}
}

func isUnsafeMethod(method string) bool {
	return method != http.MethodGet && method != http.MethodHead && method != http.MethodOptions && method != http.MethodTrace
}

func hasAuthCookie(r *http.Request) bool {
	_, accessErr := r.Cookie("access_token")
	_, refreshErr := r.Cookie("refresh_token")
	return accessErr == nil || refreshErr == nil
}
