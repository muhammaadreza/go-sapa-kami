package middleware

import (
	"errors"
	"log"

	"go-sapa-kami/internal/apperror"
	"go-sapa-kami/internal/dto"

	"github.com/gin-gonic/gin"
)

func Errors() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()

		if len(c.Errors) > 0 {
			err := c.Errors.Last().Err

			var appErr *apperror.AppError
			if errors.As(err, &appErr) {
				c.AbortWithStatusJSON(appErr.Code, dto.Response{
					Success: false,
					Message: appErr.Message,
					Detail:  appErr.Detail,
				})
				return
			}

			log.Printf("unhandled error for %s %s: %v", c.Request.Method, c.Request.URL.Path, err)
			c.AbortWithStatusJSON(500, dto.Response{
				Success: false,
				Message: "Internal Server Error",
			})

			return
		}
	}
}
