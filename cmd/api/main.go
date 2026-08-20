package main

import (
	"log"

	"go-sapa-kami/internal/config"
	"go-sapa-kami/internal/router"
	"go-sapa-kami/internal/handler"
	"go-sapa-kami/internal/service"
	"go-sapa-kami/internal/repository"
)

func main() {
	cfg := config.LoadConfig()
	db := config.PostgresConfig(cfg.Postgres, cfg.Defaults)

	sseManager := handler.NewSSEManager()

	userRepository := repository.NewUserRepository(db)
	userService := service.NewUserService(userRepository, cfg.Defaults)
	userHandler := handler.NewUserHandler(userService)

	aspirationRepository := repository.NewAspirationRepository(db)
	aspirationService := service.NewAspirationService(aspirationRepository, userService, sseManager)
	aspirationHandler := handler.NewAspirationHandler(aspirationService, sseManager)

	authService := service.NewAuthService(userRepository, cfg.Defaults, cfg.JWT)
	authHandler := handler.NewAuthHandler(authService, cfg)

	dashboardService := service.NewDashboardService(userService, aspirationService)
	dashboardHandler := handler.NewDashboardHandler(dashboardService)

	handlers := &handler.Handlers{
		Auth:       authHandler,
		User:       userHandler,
		Aspiration: aspirationHandler,
		Dashboard:  dashboardHandler,
	}

	r := router.Setup(handlers, cfg.App, cfg.JWT)

	err := r.Run(":" + cfg.App.Port)
	if err != nil { log.Fatalf("Failed to start server: %v", err) }
}
