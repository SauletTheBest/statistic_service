package main

import (
	"statistic_service/internal/config"
	"statistic_service/internal/db"
	"statistic_service/internal/handler"
	"statistic_service/internal/middleware"
	"statistic_service/internal/repository"
	"statistic_service/internal/service"

	"github.com/gin-gonic/gin"
)

func main() {

	cfg := config.LoadConfig()
	database := db.Connect(cfg.DBURL)

	// REPO
	userRepo := repository.NewUserRepository(database)
	txRepo := repository.NewTransactionRepository(database)

	// Services
	authService := service.NewAuthService(userRepo, cfg.JWTSecret)
	txService := service.NewTransactionService(txRepo)

	// Handlers
	authHandler := handler.NewAuthHandler(authService)
	txHandler := handler.NewTransactionHandler(txService)
	statsHandler := handler.NewStatsHandler(txService)

	// Middleware
	authMiddleware := middleware.JWTAuth(cfg.JWTSecret)

	r := gin.Default()

	// Auth
	r.POST("/register", authHandler.Register)
	r.POST("/login", authHandler.Login)
	// Protected
	r.GET("/me", authMiddleware, authHandler.GetProfile)

	// Transactions
	r.POST("/transactions", authMiddleware, txHandler.Create)
	r.GET("/transactions", authMiddleware, txHandler.List)
	r.PUT("/transactions/:id", authMiddleware, txHandler.Update)
	r.DELETE("/transactions/:id", authMiddleware, txHandler.Delete)

	// Statistics
	r.GET("/stats/summary", authMiddleware, statsHandler.Summary)
	r.GET("/stats/categories", authMiddleware, statsHandler.ByCategory)

	r.Run(":" + cfg.Port)
}
