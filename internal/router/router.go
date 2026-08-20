package router

import (
	"go-sapa-kami/internal/model"
	"go-sapa-kami/internal/config"
	"go-sapa-kami/internal/handler"
	"go-sapa-kami/internal/apperror"
	"go-sapa-kami/internal/middleware"

	"github.com/gin-gonic/gin"
)

func Setup(handlers *handler.Handlers, app *config.App, jwt *config.JWT) *gin.Engine {
	r := gin.Default()
	r.Use(middleware.CORS(app.AllowedOrigins))
	r.Use(middleware.CSRF(app.AllowedOrigins))
	r.Use(middleware.Errors())
	r.Static("/uploads", "./uploads")

	api := r.Group("/api")
	{
		auth := api.Group("/auth")
		{
			auth.POST("/login", handlers.Auth.Login)
			auth.POST("/refresh", handlers.Auth.RefreshToken)
		}

		protect := api.Group("")
		protect.Use(middleware.Authentication(jwt))
		{
			session := protect.Group("/session")
			{
				session.GET("", handlers.User.GetBySession)
				session.POST("/logout", handlers.Auth.Logout)
			}

			aspiration := protect.Group("/aspiration")
			{
				aspiration.POST("", handlers.Aspiration.Create)
				aspiration.GET("", handlers.Aspiration.GetAllBySession)
				aspiration.GET("/:id", handlers.Aspiration.GetByID)
				aspiration.POST("/:id/close", handlers.Aspiration.Close)
				aspiration.POST("/:id/vote", handlers.Aspiration.Vote)
				aspiration.POST("/:id/discussion", handlers.Aspiration.Discussion)
				aspiration.GET("/:id/stream", handlers.Aspiration.Stream)
			}

			admin := protect.Group("/admin")
			admin.Use(middleware.RequireRoles(model.RoleAdmin, model.RoleSuperAdmin))
			{
				admin.GET("/dashboard", handlers.Dashboard.GetDashboard)
				aspiration := admin.Group("/aspiration")
				{
					aspiration.GET("", handlers.Aspiration.GetAll)
					aspiration.GET("/export", handlers.Aspiration.Export)
					aspiration.GET("/:id", handlers.Aspiration.GetByID)
				}
			}

			superadmin := protect.Group("/admin")
			superadmin.Use(middleware.RequireRoles(model.RoleSuperAdmin))
			{
				user := superadmin.Group("/user")
				{
					user.GET("", handlers.User.GetAll)
					user.POST("", handlers.User.Create)
					user.POST("/import", handlers.User.Import)
					user.GET("/export", handlers.User.Export)
					user.GET("/:id", handlers.User.GetByID)
					user.PUT("/:id", handlers.User.Update)
					user.DELETE("/:id", handlers.User.Delete)
					user.POST("/:id/reset", handlers.User.ResetPassword)
				}

				aspiration := superadmin.Group("/aspiration")
				{
					aspiration.DELETE("/:id", handlers.Aspiration.Delete)
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
