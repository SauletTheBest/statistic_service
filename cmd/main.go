package main

import (
	_ "statistic_service/docs" // Import the generated docs

	"statistic_service/internal/config"
	"statistic_service/internal/db"
	"statistic_service/internal/handler"
	"statistic_service/internal/logger" // <-- Убедись, что этот импорт есть
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
// @description Type "Bearer" + a space + your token
func main() {
	// Инициализация логгера
	// <-- ИСПРАВЛЕНО: Используем NewLogger(), если у тебя есть SetupLogger, проверь, что он не конфликтует.
	appLogger := logger.NewLogger()
	appLogger.Info("Starting Statistic Service API")

	// Загрузка конфигурации
	cfg := config.LoadConfig()

	// Инициализация базы данных
	database, err := db.InitDB(cfg, appLogger)
	if err != nil {
		appLogger.Fatalf("Failed to initialize database: %v", err)
	}

	// Выполнение миграций
	if err := db.MigrateDB(database, appLogger); err != nil {
		appLogger.Fatalf("Failed to run database migrations: %v", err)
	}

	// Инициализация репозиториев
	userRepo := repository.NewUserRepository(database)
	categoryRepo := repository.NewCategoryRepository(database)
	transactionRepo := repository.NewTransactionRepository(database)
	// walletRepo := repository.NewWalletRepository(database) // Пока закомментируем, т.к. репозиторий еще не создан

	// Инициализация сервисов
	authService := service.NewAuthService(userRepo, cfg.JwtSecret, appLogger)
	userService := service.NewUserService(userRepo, appLogger) // <-- ИСПРАВЛЕНО: Раскомментировать и правильно инициализировать
	categoryService := service.NewCategoryService(categoryRepo, appLogger)
	// <-- ИСПРАВЛЕНО: service.NewTransactionService теперь должен принимать логгер
	transactionService := service.NewTransactionService(transactionRepo, categoryRepo, appLogger)
	// walletService := service.NewWalletService(walletRepo, userRepo, appLogger) // Пока закомментируем

	// Инициализация хендлеров
	// <-- ИСПРАВЛЕНО: handler.NewAuthHandler теперь принимает userService и appLogger
	authHandler := handler.NewAuthHandler(authService, userService, appLogger)
	transactionHandler := handler.NewTransactionHandler(transactionService, appLogger)
	statsHandler := handler.NewStatsHandler(transactionService, appLogger)
	predictHandler := handler.NewPredictHandler(transactionService, appLogger)
	timelineHandler := handler.NewTimelineHandler(transactionService, appLogger)
	categoryHandler := handler.NewCategoryHandler(categoryService, appLogger)
	// walletHandler := handler.NewWalletHandler(walletService, appLogger) // Пока закомментируем

	// Настройка Gin
	r := gin.Default()
	// <-- ИСПРАВЛЕНО: middleware.CORSMiddleware
	r.Use(middleware.CORSMiddleware())

	// Группы роутов
	// Auth routes
	r.POST("/register", authHandler.Register)
	r.POST("/login", authHandler.Login)
	r.POST("/refresh", authHandler.Refresh)

	// Swagger UI
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// Protected routes
	protected := r.Group("/")
	// <-- ИСПРАВЛЕНО: middleware.AuthMiddleware
	protected.Use(middleware.AuthMiddleware(cfg.JwtSecret))
	{
		protected.GET("/me", authHandler.GetProfile)

		// Transaction routes
		protected.POST("/transactions", transactionHandler.Create)
		protected.GET("/transactions", transactionHandler.List)
		protected.PUT("/transactions/:id", transactionHandler.Update)
		protected.DELETE("/transactions/:id", transactionHandler.Delete)

		// Category routes
		protected.POST("/categories", categoryHandler.CreateCategory)
		protected.GET("/categories/:id", categoryHandler.GetCategoryByID)
		protected.GET("/categories", categoryHandler.ListCategories)
		protected.PUT("/categories/:id", categoryHandler.UpdateCategory)
		protected.DELETE("/categories/:id", categoryHandler.DeleteCategory)

		// Statistics routes
		protected.GET("/stats/summary", statsHandler.Summary)
		protected.GET("/stats/categories", statsHandler.ByCategory)
		protected.GET("/stats/timeline", timelineHandler.Timeline)
		protected.GET("/predict", predictHandler.Predict)

		// Wallet routes (пока закомментированы, будут добавлены позже)
		// protected.POST("/wallets", walletHandler.Create)
		// protected.PUT("/wallets/Update", walletHandler.UpdateName)
		// protected.DELETE("/wallets/delete", walletHandler.Delete)
		// protected.POST("/wallets/:id/invite", walletHandler.Invite)
		// protected.DELETE("/wallets/:id/invite", walletHandler.DeleteMember) // Изменил имя, чтобы не было конфликта с walletHandler.Delete
		// protected.GET("/wallets/:id/transactions", walletHandler.GetTransactions)
		// protected.GET("/wallets/:id/members", walletHandler.GetMembers)
		// protected.POST("/wallets/:id/transactions", walletHandler.CreateTransaction)
		// protected.PUT("/wallets/:id/transactions/:transaction_id", walletHandler.UpdateTransaction) // Изменил роут
		// protected.DELETE("/wallets/:id/transactions/:transaction_id", walletHandler.DeleteTransaction) // Изменил роут
	}

	// Запуск сервера
	if err := r.Run(":" + cfg.AppPort); err != nil {
		appLogger.Fatalf("Failed to run server: %v", err)
	}
}
