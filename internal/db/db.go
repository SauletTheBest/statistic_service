package db

import (
	"log"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"statistic_service/internal/model"
)

func Connect(url string) *gorm.DB {
	db, err := gorm.Open(postgres.Open(url), &gorm.Config{})
	if err != nil {
		log.Fatalf("Could not connect to DB: %v", err)
	}

	err = db.AutoMigrate(&model.User{}, &model.RefreshToken{})
	if err != nil {
		log.Fatalf("Failed to migrate database: %v", err)
	}

	return db
}
