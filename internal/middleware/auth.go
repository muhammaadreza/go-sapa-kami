package middleware

import (
	"go-sapa-kami/internal/model"
	"go-sapa-kami/internal/utils"
	"go-sapa-kami/internal/config"
	"go-sapa-kami/internal/apperror"

	"github.com/gin-gonic/gin"
)

func Authentication(jwt *config.JWT) gin.HandlerFunc {
	return func(c *gin.Context) {
		token, err := c.Cookie("access_token")
		if token == "" || err != nil {
			c.Error(apperror.ErrUnauthorized("Authentication required"))
			c.Abort()
			return
		}

		claims, err := utils.ValidateAccessToken(token, jwt.Secret, jwt.Issuer)
		if err != nil {
			c.Error(apperror.ErrUnauthorized("Invalid or expired token"))
			c.Abort()
			return
		}

		c.Set("userID", claims.UserID)
		c.Set("userRole", claims.UserRole)
		c.Next()
	}
}

func RequireRoles(allowedRoles ...model.RoleEnum) gin.HandlerFunc {
	allowed := make(map[model.RoleEnum]struct{}, len(allowedRoles))
	for _, role := range allowedRoles { allowed[role] = struct{}{} }

	return func(ctx *gin.Context) {
		value, ok := ctx.Get("userRole")
		if !ok {
			ctx.Error(apperror.ErrForbidden("Role not found in token"))
			ctx.Abort()
			return
		}

		role, ok := value.(model.RoleEnum)
		if !ok {
			ctx.Error(apperror.ErrForbidden("Invalid role in token"))
			ctx.Abort()
			return
		}

		_, ok = allowed[role];
		if ok { ctx.Next(); return }

		ctx.Error(apperror.ErrForbidden("You don't have permission to access this resource"))
		ctx.Abort()
	}
}
