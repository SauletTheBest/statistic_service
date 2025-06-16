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

func setupStatsDB(t *testing.T) *gorm.DB {
	dbURL := "postgres://postgres:0000@localhost:5432/mydatabase?sslmode=disable"
	db, err := gorm.Open(postgres.Open(dbURL), &gorm.Config{})
	if err != nil {
		t.Fatalf("connect stats db: %v", err)
	}
	if err := db.AutoMigrate(&model.User{}, &model.Transaction{}); err != nil {
		t.Fatalf("migrate stats db: %v", err)
	}
	db.Exec("DELETE FROM transactions; DELETE FROM users;")
	return db
}

func setupStatsLogger(t *testing.T) *logrus.Logger {
	logDir := "logs"
	if err := os.MkdirAll(logDir, 0755); err != nil {
		t.Fatalf("mkdir logs: %v", err)
	}
	return logger.SetupLogger(filepath.Join(logDir, "stats_tests.log"))
}

func setupStatsRouter(t *testing.T, db *gorm.DB, lg *logrus.Logger) *gin.Engine {
	gin.SetMode(gin.TestMode)
	cfg := &config.Config{JWTSecret: "test_secret"}

	userRepo := repository.NewUserRepository(db)
	txRepo := repository.NewTransactionRepository(db)
	authSvc := service.NewAuthService(userRepo, cfg.JWTSecret, lg)
	txSvc := service.NewTransactionService(txRepo)

	authH := handler.NewAuthHandler(authSvc, lg)
	txH := handler.NewTransactionHandler(txSvc, lg)
	statsH := handler.NewStatsHandler(txSvc, lg)

	r := gin.Default()
	r.POST("/register", authH.Register)
	r.POST("/login", authH.Login)

	grp := r.Group("/")
	grp.Use(middleware.JWTAuth(cfg.JWTSecret))
	grp.POST("/transactions", txH.Create)
	grp.GET("/stats/summary", statsH.Summary)
	grp.GET("/stats/categories", statsH.ByCategory)

	return r
}

func TestStats_SummaryAndByCategory(t *testing.T) {
	db := setupStatsDB(t)
	lg := setupStatsLogger(t)
	router := setupStatsRouter(t, db, lg)

	// 1) Регистрация + логин
	creds := map[string]string{"email": "s@t.c", "password": "Password1!"}
	jb, _ := json.Marshal(creds)
	router.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest("POST", "/register", bytes.NewBuffer(jb)))
	w := httptest.NewRecorder()
	router.ServeHTTP(w, httptest.NewRequest("POST", "/login", bytes.NewBuffer(jb)))
	var lr map[string]string
	json.Unmarshal(w.Body.Bytes(), &lr)
	token := lr["access_token"]

	// 2) Создаём две транзакции
	for _, tx := range []map[string]interface{}{
		{"amount": 10.0, "type": "income", "category": "X"},
		{"amount": 3.0, "type": "expense", "category": "Y"},
	} {
		b, _ := json.Marshal(tx)
		req := httptest.NewRequest("POST", "/transactions", bytes.NewBuffer(b))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+token)
		router.ServeHTTP(httptest.NewRecorder(), req)
	}

	// 3) Проверяем summary
	req := httptest.NewRequest("GET", "/stats/summary", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("want 200 summary; got %d", w.Code)
	}
	var sum map[string]float64
	json.Unmarshal(w.Body.Bytes(), &sum)
	if sum["income"] != 10.0 || sum["expense"] != 3.0 {
		t.Errorf("unexpected summary: %+v", sum)
	}

	// 4) Проверяем by-category
	req = httptest.NewRequest("GET", "/stats/categories", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("want 200 categories; got %d", w.Code)
	}
	var cats map[string]float64
	json.Unmarshal(w.Body.Bytes(), &cats)
	if cats["X"] == 10.0 || cats["Y"] == 3.0 {
		t.Errorf("unexpected categories: %+v", cats)
	}
}
