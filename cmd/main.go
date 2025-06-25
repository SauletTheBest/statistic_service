package main

import (
	_ "statistic_service/docs"
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

	authMiddleware := middleware.JWTAuth(cfg.JWTSecret)
	// Initialize logger
	appLogger := logger.SetupLogger(cfg.AppLogFile)

	// Initialize repositories, services, handlers, and middleware
	userRepo := repository.NewUserRepository(database)
	txRepo := repository.NewTransactionRepository(database)
	categoryRepo := repository.NewCategoryRepository(database)
	walletRepo := repository.NewWalletRepository(database)

	authService := service.NewAuthService(userRepo, cfg.JWTSecret, logger.SetupLogger(cfg.ServiceLogFile))
	txService := service.NewTransactionService(txRepo, categoryRepo)
	categoryService := service.NewCategoryService(categoryRepo, appLogger) // <-- ИСПРАВЛЕНО: appLogger.Logger (если appLogger - это ваша обертка) ИЛИ appLogger напрямую (если appLogger уже *logrus.Logger)
	walletService := service.NewWalletService(walletRepo, userRepo, txRepo, categoryRepo, appLogger)

	authHandler := handler.NewAuthHandler(authService, logger.SetupLogger(cfg.HandlerLogFile))
	txHandler := handler.NewTransactionHandler(txService, logger.SetupLogger(cfg.HandlerLogFile))
	statsHandler := handler.NewStatsHandler(txService, logger.SetupLogger(cfg.HandlerLogFile))
	predictHandler := handler.NewPredictHandler(txService, logger.SetupLogger(cfg.HandlerLogFile))
	timelineHandler := handler.NewTimelineHandler(txService, logger.SetupLogger(cfg.HandlerLogFile))
	categoryHandler := handler.NewCategoryHandler(categoryService, appLogger)
	walletHandler := handler.NewWalletHandler(walletService, appLogger)

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
	r.GET("/predict", authMiddleware, predictHandler.Predict)
	r.GET("/stats/timeline", authMiddleware, timelineHandler.Timeline)

	// Categories routes
	r.POST("/categories", authMiddleware, categoryHandler.CreateCategory)
	r.GET("/categories/:id", authMiddleware, categoryHandler.GetCategoryByID)
	r.GET("/categories", authMiddleware, categoryHandler.ListCategories)
	r.PUT("/categories/:id", authMiddleware, categoryHandler.UpdateCategory)
	r.DELETE("/categories/:id", authMiddleware, categoryHandler.DeleteCategory)

	r.POST("/wallets", authMiddleware, walletHandler.Create)        // Создать кошелек
	r.GET("/wallets", authMiddleware, walletHandler.List)           // Список своих кошельков
	r.GET("/wallets/:id", authMiddleware, walletHandler.GetByID)    // Информация о кошельке
	r.PUT("/wallets/:id", authMiddleware, walletHandler.UpdateName) // Обновить имя
	r.DELETE("/wallets/:id", authMiddleware, walletHandler.Delete)  // Удалить кошелек

	// Управление участниками
	r.GET("/wallets/:id/members", authMiddleware, walletHandler.GetMembers)
	r.POST("/wallets/:id/invite", authMiddleware, walletHandler.InviteMember)
	r.DELETE("/wallets/:id/members/:userID", authMiddleware, walletHandler.RemoveMember)
	r.PUT("/wallets/:id/members/:userID/role", authMiddleware, walletHandler.UpdateMemberRole)

	// Транзакции внутри кошелька
	r.GET("/wallets/:id/transactions", authMiddleware, walletHandler.GetTransactions)
	r.POST("/wallets/:id/transactions", authMiddleware, walletHandler.CreateTransaction)
	// ... PUT и DELETE для транзакций кошелька

	//Start the server
	if err := r.Run(":" + cfg.Port); err != nil {
		appLogger.Fatalf("Failed to start server: %v", err)
	}
}
