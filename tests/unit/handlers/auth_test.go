package handlers_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	_ "modernc.org/sqlite"
	"github.com/apimgr/weather/src/database"
	"github.com/apimgr/weather/src/server/handler"
)

func setupTestRouter(t *testing.T) (*gin.Engine, *database.DB) {
	gin.SetMode(gin.TestMode)

	db, err := database.InitDB(":memory:")
	if err != nil {
		t.Fatalf("Failed to initialize test database: %v", err)
	}

	r := gin.New()
	return r, db
}

func TestAuthHandler_Register(t *testing.T) {
	r, db := setupTestRouter(t)
	defer db.Close()

	authHandler := &handlers.AuthHandler{DB: db.DB}

	r.POST("/register", authHandler.HandleRegister)

	tests := []struct {
		name       string
		payload    map[string]string
		wantStatus int
	}{
		{
			name: "Valid registration",
			payload: map[string]string{
				"username": "testuser",
				"email":    "test@example.com",
				"password": "ValidPass123",
			},
			wantStatus: http.StatusOK,
		},
		{
			name: "Missing username",
			payload: map[string]string{
				"email":    "test@example.com",
				"password": "ValidPass123",
			},
			wantStatus: http.StatusBadRequest,
		},
		{
			name: "Missing email",
			payload: map[string]string{
				"username": "testuser2",
				"password": "ValidPass123",
			},
			wantStatus: http.StatusBadRequest,
		},
		{
			name: "Short password",
			payload: map[string]string{
				"username": "testuser3",
				"email":    "test3@example.com",
				"password": "short",
			},
			wantStatus: http.StatusBadRequest,
		},
		{
			name: "Invalid email",
			payload: map[string]string{
				"username": "testuser4",
				"email":    "invalid-email",
				"password": "ValidPass123",
			},
			wantStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			jsonData, _ := json.Marshal(tt.payload)
			req := httptest.NewRequest("POST", "/register", bytes.NewBuffer(jsonData))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			r.ServeHTTP(w, req)

			if w.Code != tt.wantStatus {
				t.Errorf("Expected status %d, got %d. Body: %s", tt.wantStatus, w.Code, w.Body.String())
			}
		})
	}
}

func TestAuthHandler_Login(t *testing.T) {
	r, db := setupTestRouter(t)
	defer db.Close()

	authHandler := &handlers.AuthHandler{DB: db.DB}

	// Setup routes
	r.POST("/register", authHandler.HandleRegister)
	r.POST("/login", authHandler.HandleLogin)

	// First, create a user
	registerPayload := map[string]string{
		"username": "logintest",
		"email":    "login@example.com",
		"password": "ValidPass123",
	}
	jsonData, _ := json.Marshal(registerPayload)
	req := httptest.NewRequest("POST", "/register", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("Failed to create test user: %s", w.Body.String())
	}

	tests := []struct {
		name       string
		payload    map[string]string
		wantStatus int
	}{
		{
			name: "Valid login",
			payload: map[string]string{
				"username": "logintest",
				"password": "ValidPass123",
			},
			wantStatus: http.StatusOK,
		},
		{
			name: "Wrong password",
			payload: map[string]string{
				"username": "logintest",
				"password": "WrongPassword",
			},
			wantStatus: http.StatusUnauthorized,
		},
		{
			name: "Non-existent user",
			payload: map[string]string{
				"username": "nonexistent",
				"password": "ValidPass123",
			},
			wantStatus: http.StatusUnauthorized,
		},
		{
			name: "Missing password",
			payload: map[string]string{
				"username": "logintest",
			},
			wantStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			jsonData, _ := json.Marshal(tt.payload)
			req := httptest.NewRequest("POST", "/login", bytes.NewBuffer(jsonData))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			r.ServeHTTP(w, req)

			if w.Code != tt.wantStatus {
				t.Errorf("Expected status %d, got %d. Body: %s", tt.wantStatus, w.Code, w.Body.String())
			}
		})
	}
}
