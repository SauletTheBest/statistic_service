package tests

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"statistic_service/internal/config"
	"statistic_service/internal/handler"
	"statistic_service/internal/logger"
	"statistic_service/internal/middleware"
	"statistic_service/internal/model"
	"statistic_service/internal/repository"
	"statistic_service/internal/service"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func setupTxDB(t *testing.T) *gorm.DB {
	dbURL := "postgres://postgres:ernar2005@localhost:5432/statistic_service?sslmode=disable"
	db, err := gorm.Open(postgres.Open(dbURL), &gorm.Config{})
	if err != nil {
		t.Fatalf("connect tx test db: %v", err)
	}
	if err := db.AutoMigrate(&model.User{}, &model.Transaction{}); err != nil {
		t.Fatalf("migrate tx db: %v", err)
	}
	db.Exec("DELETE FROM transactions; DELETE FROM users;")
	return db
}

func setupTxLogger(t *testing.T) *logrus.Logger {
	logDir := "logs"
	if err := os.MkdirAll(logDir, 0755); err != nil {
		t.Fatalf("mkdir logs: %v", err)
	}
	return logger.SetupLogger(filepath.Join(logDir, "tx_tests.log"))
}

func setupTxRouter(t *testing.T, db *gorm.DB, lg *logrus.Logger) *gin.Engine {
	gin.SetMode(gin.TestMode)
	cfg := &config.Config{JWTSecret: "test_secret"}

	userRepo := repository.NewUserRepository(db)
	txRepo := repository.NewTransactionRepository(db)
	categoryRepo := repository.NewCategoryRepository(db)
	authSvc := service.NewAuthService(userRepo, cfg.JWTSecret, lg)
	txSvc := service.NewTransactionService(txRepo, categoryRepo)

	authH := handler.NewAuthHandler(authSvc, lg)
	txH := handler.NewTransactionHandler(txSvc, lg)

	r := gin.Default()
	r.POST("/register", authH.Register)
	r.POST("/login", authH.Login)

	grp := r.Group("/")
	grp.Use(middleware.JWTAuth(cfg.JWTSecret))
	grp.POST("/transactions", txH.Create)
	grp.GET("/transactions", txH.List)
	grp.DELETE("/transactions/:id", txH.Delete)

	return r
}

func TestTransaction_CRUD(t *testing.T) {
	db := setupTxDB(t)
	lg := setupTxLogger(t)
	router := setupTxRouter(t, db, lg)

	// 1. Регистрация
	creds := map[string]string{"email": "a@b.c", "password": "Password1!"}
	jb, _ := json.Marshal(creds)
	req := httptest.NewRequest("POST", "/register", bytes.NewBuffer(jb))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	if w.Code != http.StatusCreated {
		t.Fatalf("want 201 register; got %d", w.Code)
	}

	// 2. Логин → получаем токен
	req = httptest.NewRequest("POST", "/login", bytes.NewBuffer(jb))
	req.Header.Set("Content-Type", "application/json")
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("want 200 login; got %d", w.Code)
	}
	var lr map[string]string
	json.Unmarshal(w.Body.Bytes(), &lr)
	token := lr["access_token"]

	// 3. Создать транзакцию
	tx := map[string]interface{}{"amount": 42.5, "type": "income", "category": "test"}
	tb, _ := json.Marshal(tx)
	req = httptest.NewRequest("POST", "/transactions", bytes.NewBuffer(tb))
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)
	if w.Code != http.StatusCreated {
		t.Errorf("want 201 create; got %d", w.Code)
	}

	// 4. Список
	req = httptest.NewRequest("GET", "/transactions", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("want 200 list; got %d", w.Code)
	}
	var list []model.Transaction
	json.Unmarshal(w.Body.Bytes(), &list)
	if len(list) != 1 || list[0].Amount != 42.5 {
		t.Errorf("unexpected list: %+v", list)
	}

	// 5. Удаление
	id := list[0].ID
	req = httptest.NewRequest("DELETE", "/transactions/"+id, nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)
	if w.Code != http.StatusNoContent {
		t.Errorf("want 204 delete; got %d", w.Code)
	}
}
