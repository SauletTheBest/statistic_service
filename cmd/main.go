package main

import (
	"github.com/gin-gonic/gin"
	"statistic_service/internal/config"
	"statistic_service/internal/db"
	"statistic_service/internal/handler"
	"statistic_service/internal/repository"
	"statistic_service/internal/service"
	"statistic_service/internal/middleware"
	"statistic_service/internal/logger"
)

func main() {

	cfg := config.LoadConfig()
	database := db.Connect(cfg.DBURL)

	appLogger := logger.SetupLogger(cfg.AppLogFile)

	userRepo := repository.NewUserRepository(database)
	authService := service.NewAuthService(userRepo, cfg.JWTSecret, logger.SetupLogger(cfg.ServiceLogFile))
	authHandler := handler.NewAuthHandler(authService, logger.SetupLogger(cfg.HandlerLogFile))
    authMiddleware := middleware.JWTAuth(cfg.JWTSecret)

	r := gin.Default()
	
	r.POST("/register", authHandler.Register)
	r.POST("/login", authHandler.Login)
	r.POST("/refresh", authHandler.Refresh)
	r.GET("/me", authMiddleware, authHandler.GetProfile)

	if err := r.Run(":" + cfg.Port); err != nil {
		appLogger.Fatalf("Failed to start server: %v", err)
	}
}
