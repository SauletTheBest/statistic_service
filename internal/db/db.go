package db

import (
	"log"

	"statistic_service/internal/model"

	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func Connect(url string) *gorm.DB {
	var database *gorm.DB
	var err error
	for i := 0; i < 10; i++ {
		database, err = gorm.Open(postgres.Open(url), &gorm.Config{})
		if err == nil {
			break
		}
		log.Printf("DB not ready yet, retrying in 2s... (%d/10)", i+1)
		time.Sleep(2 * time.Second)
	}

	if err != nil {
		log.Fatalf("Could not connect to DB: %v", err)
	}

	err = database.AutoMigrate(&model.User{}, &model.Transaction{}, &model.Category{}, &model.RefreshToken{}, &model.Wallet{}, &model.WalletMember{})
	if err != nil {
		log.Fatalf("Failed to migrate database: %v", err)
	}
	return database
}
