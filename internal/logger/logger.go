package logger

import (
	"os"
	"path/filepath"

	"github.com/sirupsen/logrus"
)

func SetupLogger(logFile string) *logrus.Logger {
	logger := logrus.New()
	// Ensure logs directory exists
	logDir := filepath.Dir(logFile)
	if err := os.MkdirAll(logDir, 0755); err != nil {
		logger.Fatalf("Failed to create logs directory: %v", err)
	}

	file, err := os.OpenFile(logFile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		logger.Fatalf("Failed to open log file: %v", err)
	}

	logger.SetOutput(file)
	logger.SetFormatter(&logrus.JSONFormatter{})
	logger.SetLevel(logrus.InfoLevel)

	return logger
}
