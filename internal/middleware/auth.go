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

		claims, err := utils.ValidateAccessToken(token, jwt.Secret)
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

func Authorization(permission model.RoleEnum) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		role, ok := ctx.Get("userRole")
        if !ok {
            ctx.Error(apperror.ErrForbidden("Role not found in token"))
            ctx.Abort()
            return
        }

		if role == model.RoleSuperAdmin || role == permission {
            ctx.Next()
            return
        }

        ctx.Error(apperror.ErrForbidden("You don't have permission to access this resource"))
        ctx.Abort()
	}
}
