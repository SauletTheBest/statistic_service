package tests

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"statistic_service/internal/config"
	"statistic_service/internal/handler"
	"statistic_service/internal/logger"
	"statistic_service/internal/model"
	"statistic_service/internal/repository"
	"statistic_service/internal/service"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func setupTestDB(t *testing.T) *gorm.DB {
	// Use a test-specific PostgreSQL database
	dbURL := "postgres://postgres:myStrongTestPassword123@localhost:5432/mydatabase?sslmode=disable"

	db, err := gorm.Open(postgres.Open(dbURL), &gorm.Config{})
	if err != nil {
		t.Fatalf("Failed to connect to test database: %v", err)
	}

	// Auto-migrate the User and RefreshToken models
	err = db.AutoMigrate(&model.User{}, &model.RefreshToken{})
	if err != nil {
		t.Fatalf("Failed to migrate test database: %v", err)
	}

	// Clean up existing data to ensure test isolation
	db.Exec("DELETE FROM users")
	db.Exec("DELETE FROM refresh_tokens")

	return db
}

func setupTestLogger(t *testing.T) *logrus.Logger {
	logDir := "logs"
	logFile := filepath.Join(logDir, "test_handler.log")
	if err := os.MkdirAll(logDir, 0755); err != nil {
		t.Fatalf("Failed to create logs directory: %v", err)
	}
	return logger.SetupLogger(logFile)
}

func setupRouter(t *testing.T, db *gorm.DB, logger *logrus.Logger) (*gin.Engine, *handler.AuthHandler) {
	gin.SetMode(gin.TestMode)

	cfg := &config.Config{
		JWTSecret: "test_secret",
	}
	userRepo := repository.NewUserRepository(db)
	authService := service.NewAuthService(userRepo, cfg.JWTSecret, logger)
	authHandler := handler.NewAuthHandler(authService, logger)

	r := gin.Default()
	r.POST("/register", authHandler.Register)
	r.POST("/login", authHandler.Login)

	return r, authHandler
}

func TestRegister(t *testing.T) {
	db := setupTestDB(t)
	logger := setupTestLogger(t)
	router, _ := setupRouter(t, db, logger)

	tests := []struct {
		name           string
		body           map[string]string
		expectedStatus int
		expectedBody   map[string]interface{}
	}{
		{
			name: "Successful registration",
			body: map[string]string{
				"email":    "test@example.com",
				"password": "Password123!",
			},
			expectedStatus: http.StatusCreated,
			expectedBody: map[string]interface{}{
				"status":  "success",
				"message": "user registered successfully",
			},
		},
		{
			name: "Duplicate email",
			body: map[string]string{
				"email":    "test@example.com",
				"password": "Password123!",
			},
			expectedStatus: http.StatusConflict,
			expectedBody: map[string]interface{}{
				"error": "user already exists",
			},
		},
		{
			name: "Invalid email",
			body: map[string]string{
				"email":    "invalid",
				"password": "Password123!",
			},
			expectedStatus: http.StatusBadRequest,
			expectedBody: map[string]interface{}{
				"error": "validation failed",
			},
		},
		{
			name: "Weak password",
			body: map[string]string{
				"email":    "test2@example.com",
				"password": "weak",
			},
			expectedStatus: http.StatusBadRequest,
			expectedBody: map[string]interface{}{
				"error": "validation failed",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bodyBytes, err := json.Marshal(tt.body)
			if err != nil {
				t.Fatalf("Failed to marshal request body: %v", err)
			}
			req, err := http.NewRequest("POST", "/register", bytes.NewBuffer(bodyBytes))
			if err != nil {
				t.Fatalf("Failed to create request: %v", err)
			}
			req.Header.Set("Content-Type", "application/json")

			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			if w.Code != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, w.Code)
			}

			var response map[string]interface{}
			if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
				t.Fatalf("Failed to unmarshal response: %v", err)
			}

			for key, expectedValue := range tt.expectedBody {
				if _, ok := response[key]; !ok {
					t.Errorf("Response missing key: %s", key)
				}
				if key == "error" {
					if !bytes.Contains([]byte(response[key].(string)), []byte(expectedValue.(string))) {
						t.Errorf("Expected error containing %v, got %v", expectedValue, response[key])
					}
				} else {
					if response[key] != expectedValue {
						t.Errorf("Expected %s: %v, got %v", key, expectedValue, response[key])
					}
				}
			}
		})
	}
}
func TestLogin(t *testing.T) {
	db := setupTestDB(t)
	logger := setupTestLogger(t)
	router, _ := setupRouter(t, db, logger)

	// Pre-register a user for login tests
	registerBody := map[string]string{
		"email":    "test@example.com",
		"password": "Password123!",
	}
	registerBodyBytes, _ := json.Marshal(registerBody)
	registerReq, _ := http.NewRequest("POST", "/register", bytes.NewBuffer(registerBodyBytes))
	registerReq.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, registerReq)
	if w.Code != http.StatusCreated {
		t.Fatalf("Failed to pre-register user, status: %d", w.Code)
	}

	tests := []struct {
		name           string
		body           map[string]string
		expectedStatus int
		expectedBody   map[string]interface{}
	}{
		{
			name: "Successful login",
			body: map[string]string{
				"email":    "test@example.com",
				"password": "Password123!",
			},
			expectedStatus: http.StatusOK,
			expectedBody: map[string]interface{}{
				"access_token":  "",
				"refresh_token": "",
			},
		},
		{
			name: "Invalid password",
			body: map[string]string{
				"email":    "test@example.com",
				"password": "WrongPassword!",
			},
			expectedStatus: http.StatusUnauthorized,
			expectedBody: map[string]interface{}{
				"error": "invalid email or password",
			},
		},
		{
			name: "Invalid email",
			body: map[string]string{
				"email":    "invalid@example.com",
				"password": "Password123!",
			},
			expectedStatus: http.StatusUnauthorized,
			expectedBody: map[string]interface{}{
				"error": "invalid email or password",
			},
		},
		{
			name: "Invalid request format",
			body: map[string]string{
				"email":    "invalid",
				"password": "weak",
			},
			expectedStatus: http.StatusBadRequest,
			expectedBody: map[string]interface{}{
				"error": "validation failed",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bodyBytes, err := json.Marshal(tt.body)
			if err != nil {
				t.Fatalf("Failed to marshal request body: %v", err)
			}
			req, err := http.NewRequest("POST", "/login", bytes.NewBuffer(bodyBytes))
			if err != nil {
				t.Fatalf("Failed to create request: %v", err)
			}
			req.Header.Set("Content-Type", "application/json")

			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			if w.Code != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, w.Code)
			}

			var response map[string]interface{}
			if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
				t.Fatalf("Failed to unmarshal response: %v", err)
			}

			for key, expectedValue := range tt.expectedBody {
				if _, ok := response[key]; !ok {
					t.Errorf("Response missing key: %s", key)
				}
				if key == "error" {
					if !bytes.Contains([]byte(response[key].(string)), []byte(expectedValue.(string))) {
						t.Errorf("Expected error containing %v, got %v", expectedValue, response[key])
					}
				} else if key == "access_token" || key == "refresh_token" {
					if tt.expectedStatus == http.StatusOK {
						if response[key] == "" {
							t.Errorf("Expected non-empty %s, got empty", key)
						}
					}
				} else {
					if response[key] != expectedValue {
						t.Errorf("Expected %s: %v, got %v", key, expectedValue, response[key])
					}
				}
			}
		})
	}
}
