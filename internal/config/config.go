package config

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

// Config хранит конфигурацию приложения
type Config struct {
	AppPort   string
	JwtSecret string
	// Конфигурация базы данных
	DBHost     string
	DBUser     string
	DBPassword string
	DBName     string
	DBPort     string
}

// LoadConfig загружает конфигурацию из переменных окружения или .env файла
func LoadConfig() *Config {
	err := godotenv.Load() // Загружает .env файл, если он есть
	if err != nil && !os.IsNotExist(err) {
		log.Printf("Error loading .env file: %v, continuing without it.", err)
	}

	return &Config{
		AppPort:    getEnv("APP_PORT", "8080"),
		JwtSecret:  getEnv("JWT_SECRET", "supersecretjwtkey"),
		DBHost:     getEnv("DB_HOST", "localhost"),
		DBUser:     getEnv("DB_USER", "postgres"),
		DBPassword: getEnv("DB_PASSWORD", "postgres"),
		DBName:     getEnv("DB_NAME", "statistic_service"),
		DBPort:     getEnv("DB_PORT", "5432"),
	}
}

// getEnv получает значение переменной окружения или возвращает значение по умолчанию
func getEnv(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultValue
}
