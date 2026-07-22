package middleware

import (
	"errors"

	"go-sapa-kami/internal/model"
	"go-sapa-kami/internal/apperror"

	"github.com/gin-gonic/gin"
)

func Errors() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()

		if len(c.Errors) > 0 {
			err := c.Errors.Last().Err

			var appErr *apperror.AppError
			if errors.As(err, &appErr) {
				c.AbortWithStatusJSON(appErr.Code, model.Response{
					Success: false,
					Message: appErr.Message,
					Detail:  appErr.Detail,
				})
				return
			}

			c.AbortWithStatusJSON(500, model.Response{
				Success: false,
				Message: "Internal Server Error",
				Detail:  err.Error(),
			})

			return
		}
	}
}
