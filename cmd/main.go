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
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	_ "statistic_service/docs" // Import the generated docs
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
	authService := service.NewAuthService(userRepo, cfg.JWTSecret, logger.SetupLogger(cfg.ServiceLogFile))
	authHandler := handler.NewAuthHandler(authService, logger.SetupLogger(cfg.HandlerLogFile))
    authMiddleware := middleware.JWTAuth(cfg.JWTSecret)

	// Set up Gin router
	r := gin.Default()

	// Define routes
	r.POST("/register", authHandler.Register)
	r.POST("/login", authHandler.Login)
	r.POST("/refresh", authHandler.Refresh)
	r.GET("/me", authMiddleware, authHandler.GetProfile)

	// Swagger routes
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// Start the server
	if err := r.Run(":" + cfg.Port); err != nil {
		appLogger.Fatalf("Failed to start server: %v", err)
	}
}