package config

import (
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	DBURL          string
	JWTSecret      string
	Port           string
	AppLogFile     string
	ServiceLogFile string
	HandlerLogFile string
}

func LoadConfig() *Config {
	err := godotenv.Load(".env")
	if err != nil {
		_ = godotenv.Load("../.env")
	}

	return &Config{
		DBURL:          os.Getenv("DB_URL"),
		JWTSecret:      os.Getenv("JWT_SECRET"),
		Port:           os.Getenv("PORT"),
		AppLogFile:     os.Getenv("APP_LOG_FILE"),
		ServiceLogFile: os.Getenv("SERVICE_LOG_FILE"),
		HandlerLogFile: os.Getenv("HANDLER_LOG_FILE"),
	}
}
