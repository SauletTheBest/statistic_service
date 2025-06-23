package main

import (
	_ "statistic_service/docs" // Import the generated docs
	"statistic_service/internal/config"
	"statistic_service/internal/db"
	"statistic_service/internal/handler"
	"statistic_service/internal/logger"
	"statistic_service/internal/middleware"
	"statistic_service/internal/repository"
	"statistic_service/internal/service"

	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
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
	walletRepo := repository.NewWalletRepository(database)
	categoryRepo := repository.NewCategoryRepository(database)

	authService := service.NewAuthService(userRepo, cfg.JWTSecret, logger.SetupLogger(cfg.ServiceLogFile))
	txService := service.NewTransactionService(txRepo)
	walletService := service.NewWalletService(walletRepo)
	authMiddleware := middleware.JWTAuth(cfg.JWTSecret)
	categoryService := service.NewCategoryService(categoryRepo, logger.SetupLogger(cfg.HandlerLogFile))

	authHandler := handler.NewAuthHandler(authService, logger.SetupLogger(cfg.HandlerLogFile))
	txHandler := handler.NewTransactionHandler(txService, logger.SetupLogger(cfg.HandlerLogFile))
	statsHandler := handler.NewStatsHandler(txService, logger.SetupLogger(cfg.HandlerLogFile))
	predictHandler := handler.NewPredictHandler(txService, logger.SetupLogger(cfg.HandlerLogFile))
	timelineHandler := handler.NewTimelineHandler(txService, logger.SetupLogger(cfg.HandlerLogFile))
	walletHandler := handler.NewWalletHandler(walletService)
	categoryHandler := handler.NewCategoryHandler(categoryService, logger.SetupLogger(cfg.HandlerLogFile))

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
	r.GET("/transactions", authMiddleware, txHandler.List)
	r.PUT("/transactions/:id", authMiddleware, txHandler.Update)
	r.DELETE("/transactions/:id", authMiddleware, txHandler.Delete)

	// Statistics
	r.GET("/stats/summary", authMiddleware, statsHandler.Summary)
	r.GET("/stats/categories", authMiddleware, statsHandler.ByCategory)
	r.GET("/predict", authMiddleware, predictHandler.Predict)

	// Graph
	r.GET("/stats/timeline", authMiddleware, timelineHandler.Timeline)

	//Category
	r.POST("/categories", authMiddleware, categoryHandler.CreateCategory)
	r.GET("/categories/:id", authMiddleware, categoryHandler.GetCategoryByID)
	r.GET("/categories", authMiddleware, categoryHandler.ListCategories)
	r.PUT("/categories/:id", authMiddleware, categoryHandler.UpdateCategory)
	r.DELETE("/categories/:id", authMiddleware, categoryHandler.DeleteCategory)

	//Wallets
	r.POST("/wallets", authMiddleware, walletHandler.Create)
	r.POST("/wallets/:id/invite", authMiddleware, walletHandler.Invite)
	r.GET("/wallets/:id/transactions", authMiddleware, walletHandler.GetTransactions)
	r.GET("/wallets/:id/members", authMiddleware, walletHandler.GetMembers)
	r.POST("/wallets/:id/transactions", authMiddleware, walletHandler.CreateTransaction)

	//Start the server
	if err := r.Run(":" + cfg.Port); err != nil {
		appLogger.Fatalf("Failed to start server: %v", err)
	}
}
