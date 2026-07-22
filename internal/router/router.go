package router

import (
	"go-sapa-kami/internal/model"
	"go-sapa-kami/internal/config"
	"go-sapa-kami/internal/handler"
	"go-sapa-kami/internal/apperror"
	"go-sapa-kami/internal/middleware"

	"github.com/gin-gonic/gin"
)

func Setup(handler *handler.Handlers, jwt *config.JWT) *gin.Engine {
	r := gin.Default()
	r.Use(middleware.CORS("*"))
	r.Use(middleware.Errors())

	api := r.Group("/api")
	{
		auth := api.Group("/auth")
		{
			auth.POST("/login", handler.Auth.Login)
			auth.GET("/refresh", handler.Auth.RefreshToken)
		}

		protect := api.Group("")
		protect.Use(middleware.Authentication(jwt))
		{
			session := protect.Group("/session")
			{
				session.GET("", handler.User.GetBySession)
				session.POST("/logout", handler.Auth.Logout)
			}

			message := protect.Group("/message")
			{
				message.POST("", handler.Message.Create)
				message.GET("", handler.Message.GetAllBySession)
			}

			admin := protect.Group("/admin")
			admin.Use(middleware.Authorization(model.RoleAdmin))
			{
				admin.GET("/dashboard", handler.Dashboard.GetDashboard)
				message := admin.Group("/message")
				{
					message.GET("", handler.Message.GetAll)
					message.GET("/:id", handler.Message.GetByID)
					message.GET("/export", handler.Message.Export)
				}
			}

			superadmin := protect.Group("/admin")
			superadmin.Use(middleware.Authorization(model.RoleSuperAdmin))
			{
				user := superadmin.Group("/user")
				{
					user.GET("", handler.User.GetAll)
					user.GET("/:id", handler.User.GetByID)
					user.POST("", handler.User.Create)
					user.PUT("/:id", handler.User.Update)
					user.DELETE("/:id", handler.User.Delete)

					user.POST("/import", handler.User.Import)
					user.GET("/export", handler.User.Export)
					user.POST("/:id/reset", handler.User.ResetPassword)
				}

				message := superadmin.Group("/message")
				{
					message.DELETE("/:id", handler.Message.Delete)
				}
			}
		}
	}

	r.HandleMethodNotAllowed = true
	r.NoMethod(func(c *gin.Context) {
		path := c.Request.URL.Path
		method := c.Request.Method
		c.Error(apperror.ErrMethod("Method " + method + " not allowed for path: " + path))
		c.Abort()
	})

	r.NoRoute(func(c *gin.Context) {
		path := c.Request.URL.Path
		c.Error(apperror.ErrNotFound("Route not found: " + path))
		c.Abort()
	})

	return r
}
