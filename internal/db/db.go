package db

import (
	"fmt"
	"statistic_service/internal/config"
	"statistic_service/internal/model" // Убедись, что модель импортирована
	"time"

	"github.com/sirupsen/logrus"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// InitDB устанавливает соединение с базой данных
func InitDB(cfg *config.Config, log *logrus.Logger) (*gorm.DB, error) {
	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=disable",
		cfg.DBHost, cfg.DBUser, cfg.DBPassword, cfg.DBName, cfg.DBPort)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info), // Для логирования SQL-запросов
	})
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("failed to get database instance: %w", err)
	}

	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetMaxOpenConns(100)
	sqlDB.SetConnMaxLifetime(time.Hour)

	log.Info("Successfully connected to database")
	return db, nil
}

// MigrateDB выполняет автоматические миграции базы данных
func MigrateDB(db *gorm.DB, log *logrus.Logger) error {
	log.Info("Starting database migration...")
	err := db.AutoMigrate(
		&model.User{},
		&model.Category{},
		&model.Transaction{},
		&model.Wallet{},       // <-- НОВОЕ: Добавляем модель Wallet
		&model.WalletMember{}, // <-- НОВОЕ: Добавляем модель WalletMember
	)
	if err != nil {
		log.WithError(err).Error("Failed to migrate database")
		return fmt.Errorf("failed to migrate database: %w", err)
	}
	log.Info("Database migration completed successfully.")

	// После миграции, если есть старые транзакции, которые не привязаны к кошельку,
	// тебе может потребоваться назначить им WalletID.
	// Например, создать для каждого пользователя "персональный" кошелек
	// и переместить все его существующие транзакции в этот кошелек.
	// Это можно сделать отдельным скриптом или здесь, но в рамках первой миграции это может быть сложно.
	// Пока что просто убедимся, что GORM создает поле WalletID.
	// Если у тебя уже есть данные в таблице `transactions` без `WalletID`,
	// и ты сделал `WalletID` `not null`, то миграция, скорее всего, снова выдаст ошибку.
	// В таком случае, тебе нужно будет:
	// 1. Временно изменить `WalletID` в `internal/model/Transaction.go` на `gorm:"type:uuid;index"`,
	//    т.е. убрать `not null` для первой миграции, чтобы столбец создался.
	// 2. После успешной миграции, выполнить SQL-запрос для заполнения `WalletID`
	//    для существующих транзакций (например, создать дефолтный кошелек для каждого пользователя и присвоить его ID).
	// 3. Затем вернуть `gorm:"type:uuid;not null;index"` в `internal/model/Transaction.go`.

	return nil
}
