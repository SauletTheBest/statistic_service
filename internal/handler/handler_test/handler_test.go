package handler_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/glebarez/sqlite"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"

	"statistic_service/internal/handler"
	"statistic_service/internal/middleware"
	"statistic_service/internal/model"
	"statistic_service/internal/repository"
	"statistic_service/internal/service"
)

// setupRouter creates a Gin engine with in-memory DB and routes, manually creating tables
func setupRouter(t *testing.T) *gin.Engine {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("failed to connect database: %v", err)
	}
	// Manually create tables without AutoMigrate to avoid unsupported SQL in SQLite
	if err := db.Exec(`CREATE TABLE users (id TEXT PRIMARY KEY, email TEXT UNIQUE NOT NULL, password_hash TEXT NOT NULL, created_at DATETIME NOT NULL)`).Error; err != nil {
		t.Fatalf("failed to create users table: %v", err)
	}
	if err := db.Exec(`CREATE TABLE transactions (id TEXT PRIMARY KEY, user_id TEXT NOT NULL, amount REAL NOT NULL, type TEXT NOT NULL, category TEXT, comment TEXT, created_at DATETIME NOT NULL)`).Error; err != nil {
		t.Fatalf("failed to create transactions table: %v", err)
	}
	if err := db.Exec(`CREATE TABLE categories (id TEXT PRIMARY KEY, user_id TEXT NOT NULL, name TEXT NOT NULL, type TEXT NOT NULL, created_at DATETIME NOT NULL)`).Error; err != nil {
		t.Fatalf("failed to create categories table: %v", err)
	}

	userRepo := repository.NewUserRepository(db)
	txRepo := repository.NewTransactionRepository(db)
	authSvc := service.NewAuthService(userRepo, "testsecret", logrus.New())
	txSvc := service.NewTransactionService(txRepo)

	authH := handler.NewAuthHandler(authSvc, logrus.New())
	txH := handler.NewTransactionHandler(txSvc, logrus.New())
	statsH := handler.NewStatsHandler(txSvc, logrus.New())

	router := gin.New()
	router.POST("/register", authH.Register)
	router.POST("/login", authH.Login)
	router.POST("/refresh", authH.Refresh)
	authGroup := router.Group("/")
	authGroup.Use(middleware.JWTAuth("testsecret"))
	authGroup.GET("/me", authH.GetProfile)
	authGroup.POST("/transactions", txH.Create)
	authGroup.GET("/transactions", txH.List)
	authGroup.DELETE("/transactions/:id", txH.Delete)
	authGroup.GET("/stats/summary", statsH.Summary)
	authGroup.GET("/stats/categories", statsH.ByCategory)
	return router
}

func TestAuth_RegisterLoginProfile(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := setupRouter(t)

	payload := map[string]string{"email": "user@test.com", "password": "Password123!"}
	body, _ := json.Marshal(payload)
	req := httptest.NewRequest("POST", "/register", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	if w.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d", w.Code)
	}

	req = httptest.NewRequest("POST", "/login", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d, body: %s", w.Code, w.Body.String())
	}
	var resp map[string]string
	json.Unmarshal(w.Body.Bytes(), &resp)
	token := resp["access_token"]

	req = httptest.NewRequest("GET", "/me", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200 on /me, got %d", w.Code)
	}
}

func TestTransactions_CreateDeleteList(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := setupRouter(t)

	payload := map[string]string{"email": "tx@test.com", "password": "Password123!"}
	body, _ := json.Marshal(payload)
	router.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest("POST", "/register", bytes.NewBuffer(body)))
	w := httptest.NewRecorder()
	router.ServeHTTP(w, httptest.NewRequest("POST", "/login", bytes.NewBuffer(body)))
	var resp map[string]string
	json.Unmarshal(w.Body.Bytes(), &resp)
	token := resp["access_token"]

	txPayload := map[string]interface{}{"amount": 50.0, "type": "income", "category": "Test"}
	txBody, _ := json.Marshal(txPayload)
	req := httptest.NewRequest("POST", "/transactions", bytes.NewBuffer(txBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)
	if w.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d", w.Code)
	}

	req = httptest.NewRequest("GET", "/transactions", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)
	var list []model.Transaction
	json.Unmarshal(w.Body.Bytes(), &list)
	if len(list) != 1 || list[0].Amount != 50.0 {
		t.Fatal("transaction list incorrect")
	}

	id := list[0].ID
	req = httptest.NewRequest("DELETE", "/transactions/"+id, nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)
	if w.Code != http.StatusNoContent {
		t.Fatalf("expected 204, got %d", w.Code)
	}
}

func TestStats_SummaryByCategory(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := setupRouter(t)

	payload := map[string]string{"email": "st@test.com", "password": "Password123!"}
	body, _ := json.Marshal(payload)
	router.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest("POST", "/register", bytes.NewBuffer(body)))
	w := httptest.NewRecorder()
	router.ServeHTTP(w, httptest.NewRequest("POST", "/login", bytes.NewBuffer(body)))
	var resp map[string]string
	json.Unmarshal(w.Body.Bytes(), &resp)
	token := resp["access_token"]

	tx1 := map[string]interface{}{"amount": 100.0, "type": "income", "category": "A"}
	tx2 := map[string]interface{}{"amount": 30.0, "type": "expense", "category": "B"}
	for _, tx := range []map[string]interface{}{tx1, tx2} {
		txb, _ := json.Marshal(tx)
		req := httptest.NewRequest("POST", "/transactions", bytes.NewBuffer(txb))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+token)
		router.ServeHTTP(httptest.NewRecorder(), req)
		time.Sleep(10 * time.Millisecond)
	}

	req := httptest.NewRequest("GET", "/stats/summary", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)
	var sumResp map[string]float64
	json.Unmarshal(w.Body.Bytes(), &sumResp)
	if sumResp["income"] != 100.0 || sumResp["expense"] != 30.0 {
		t.Fatalf("summary incorrect: %v", sumResp)
	}

	req = httptest.NewRequest("GET", "/stats/categories", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)
	var catResp map[string]float64
	json.Unmarshal(w.Body.Bytes(), &catResp)
	if catResp["A"] != 100.0 || catResp["B"] != 30.0 {
		t.Fatalf("byCategory incorrect: %v", catResp)
	}
}
