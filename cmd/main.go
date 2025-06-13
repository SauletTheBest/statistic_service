package main

import (
	"statistic_service/internal/config"
	"statistic_service/internal/db"
	"statistic_service/internal/handler"
	"statistic_service/internal/logger"
	"statistic_service/internal/middleware"
	"statistic_service/internal/repository"
	"statistic_service/internal/service"
	"statistic_service/internal/middleware"
	"statistic_service/internal/logger"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	_ "statistic_service/docs" // Import the generated docs
	"github.com/gin-gonic/gin"
)
// Package main provides the entry point for the Statistic Service API.
// @title Statistic Service API
// @version 1.0
// @description API for user authentication and profile management.
// @termsOfService http://swagger.io/terms/
// @contact.name API Support
// @contact.email support@example.com
// @license.name Apache 2.0
// @license.url http://www.apache.org/licenses/LICENSE-2.0.html
// @host localhost:8080
// @BasePath /
// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description JWT Authorization header using the Bearer scheme. Example: "Bearer {token}"
func main() {
	// Load configuration
	cfg := config.LoadConfig()
	// Initialize database connection
	database := db.Connect(cfg.DBURL)

	// Initialize logger
	appLogger := logger.SetupLogger(cfg.AppLogFile)
  
	// Initialize repositories, services, handlers, and middleware
	userRepo := repository.NewUserRepository(database)
	txRepo := repository.NewTransactionRepository(database)
  
	authService := service.NewAuthService(userRepo, cfg.JWTSecret, logger.SetupLogger(cfg.ServiceLogFile))
	txService := service.NewTransactionService(txRepo)
  
	authHandler := handler.NewAuthHandler(authService, logger.SetupLogger(cfg.HandlerLogFile))
  
  authMiddleware := middleware.JWTAuth(cfg.JWTSecret)
  
	txHandler := handler.NewTransactionHandler(txService)
  
	statsHandler := handler.NewStatsHandler(txService)

	// Set up Gin router
	r := gin.Default()

	// Auth
	r.POST("/register", authHandler.Register)
	r.POST("/login", authHandler.Login)
	r.POST("/refresh", authHandler.Refresh)
	// Protected
	r.GET("/me", authMiddleware, authHandler.GetProfile)

	// Swagger routes
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// Transactions
	r.POST("/transactions", authMiddleware, txHandler.Create)
	r.GET("/transactions", authMiddleware, txHandler.List)
	r.PUT("/transactions/:id", authMiddleware, txHandler.Update)
	r.DELETE("/transactions/:id", authMiddleware, txHandler.Delete)

	// Statistics
	r.GET("/stats/summary", authMiddleware, statsHandler.Summary)
	r.GET("/stats/categories", authMiddleware, statsHandler.ByCategory)

  //Start the server
	if err := r.Run(":" + cfg.Port); err != nil {
		appLogger.Fatalf("Failed to start server: %v", err)
	}
}
