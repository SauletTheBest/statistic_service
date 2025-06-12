package main

import (
	"github.com/gin-gonic/gin"
	"statistic_service/internal/config"
	"statistic_service/internal/db"
	"statistic_service/internal/handler"
	"statistic_service/internal/repository"
	"statistic_service/internal/service"
	"statistic_service/internal/middleware"
)

func main() {

	cfg := config.LoadConfig()
	database := db.Connect(cfg.DBURL)

	userRepo := repository.NewUserRepository(database)
	authService := service.NewAuthService(userRepo, cfg.JWTSecret)
	authHandler := handler.NewAuthHandler(authService)
    authMiddleware := middleware.JWTAuth(cfg.JWTSecret)

	r := gin.Default()
	r.POST("/register", authHandler.Register)
	r.POST("/login", authHandler.Login)
	r.POST("/refresh", authHandler.Refresh)

	r.GET("/me", authMiddleware, authHandler.GetProfile)
	r.Run(":" + cfg.Port)
}
