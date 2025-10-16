package e2e_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	_ "github.com/mattn/go-sqlite3"
	"weather-go/src/database"
	"weather-go/src/handlers"
)

// TestCompleteSetupFlow tests the entire user setup flow end-to-end
func TestCompleteSetupFlow(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// Initialize fresh database
	db, err := database.InitDB(":memory:")
	if err != nil {
		t.Fatalf("Failed to initialize database: %v", err)
	}
	defer db.Close()

	// Create router
	r := gin.New()
	setupHandler := &handlers.SetupHandler{DB: db.DB}

	// Setup routes
	setupRoutes := r.Group("/user/setup")
	{
		setupRoutes.GET("", setupHandler.ShowWelcome)
		setupRoutes.GET("/register", setupHandler.ShowUserRegister)
		setupRoutes.POST("/register", setupHandler.CreateUser)
		setupRoutes.GET("/admin", setupHandler.ShowAdminSetup)
		setupRoutes.POST("/admin", setupHandler.CreateAdmin)
	}

	// Step 1: Check if setup is needed
	t.Run("1. Check setup status", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/user/setup", nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status 200, got %d", w.Code)
		}
	})

	// Step 2: Create first user
	var userCookies []*http.Cookie
	t.Run("2. Create first user", func(t *testing.T) {
		payload := map[string]string{
			"username": "firstuser",
			"email":    "user@example.com",
			"password": "UserPass123",
		}
		jsonData, _ := json.Marshal(payload)

		req := httptest.NewRequest("POST", "/user/setup/register", bytes.NewBuffer(jsonData))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		r.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Fatalf("Failed to create first user: %s", w.Body.String())
		}

		// Save cookies for next request
		userCookies = w.Result().Cookies()

		var response map[string]interface{}
		if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
			t.Fatalf("Failed to parse response: %v", err)
		}

		if success, ok := response["success"].(bool); !ok || !success {
			t.Error("Expected success=true in response")
		}
	})

	// Step 3: Create administrator
	t.Run("3. Create administrator", func(t *testing.T) {
		payload := map[string]string{
			"username": "administrator",
			"email":    "admin@example.com",
			"password": "AdminPass12345",
			"confirm_password": "AdminPass12345",
		}
		jsonData, _ := json.Marshal(payload)

		req := httptest.NewRequest("POST", "/user/setup/admin", bytes.NewBuffer(jsonData))
		req.Header.Set("Content-Type", "application/json")

		// Add user cookies from previous step
		for _, cookie := range userCookies {
			req.AddCookie(cookie)
		}

		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Fatalf("Failed to create administrator: %s", w.Body.String())
		}

		var response map[string]interface{}
		if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
			t.Fatalf("Failed to parse response: %v", err)
		}

		if success, ok := response["success"].(bool); !ok || !success {
			t.Error("Expected success=true in response")
		}

		if redirect, ok := response["redirect"].(string); !ok || redirect == "" {
			t.Error("Expected redirect URL in response")
		} else {
			t.Logf("Redirect URL: %s", redirect)
		}
	})

	// Step 4: Verify users were created
	t.Run("4. Verify users in database", func(t *testing.T) {
		var userCount int
		err := db.DB.QueryRow("SELECT COUNT(*) FROM users").Scan(&userCount)
		if err != nil {
			t.Fatalf("Failed to query users: %v", err)
		}

		if userCount != 2 {
			t.Errorf("Expected 2 users, got %d", userCount)
		}

		// Check admin exists
		var adminRole string
		err = db.DB.QueryRow("SELECT role FROM users WHERE username = 'administrator'").Scan(&adminRole)
		if err != nil {
			t.Fatalf("Admin user not found: %v", err)
		}

		if adminRole != "admin" {
			t.Errorf("Expected admin role, got %s", adminRole)
		}

		// Check first user exists
		var userRole string
		err = db.DB.QueryRow("SELECT role FROM users WHERE username = 'firstuser'").Scan(&userRole)
		if err != nil {
			t.Fatalf("First user not found: %v", err)
		}

		if userRole != "user" {
			t.Errorf("Expected user role, got %s", userRole)
		}
	})
}

// TestAdminSetupValidation tests validation in admin creation
func TestAdminSetupValidation(t *testing.T) {
	gin.SetMode(gin.TestMode)

	db, err := database.InitDB(":memory:")
	if err != nil {
		t.Fatalf("Failed to initialize database: %v", err)
	}
	defer db.Close()

	r := gin.New()
	setupHandler := &handlers.SetupHandler{DB: db.DB}
	r.POST("/user/setup/admin", setupHandler.CreateAdmin)

	tests := []struct {
		name       string
		payload    map[string]string
		wantStatus int
	}{
		{
			name: "Valid admin",
			payload: map[string]string{
				"username":         "myadmin",
				"email":            "admin@test.com",
				"password":         "AdminPass12345",
				"confirm_password": "AdminPass12345",
			},
			wantStatus: http.StatusOK,
		},
		{
			name: "Password too short",
			payload: map[string]string{
				"username":         "admin2",
				"email":            "admin2@test.com",
				"password":         "Short1",
				"confirm_password": "Short1",
			},
			wantStatus: http.StatusBadRequest,
		},
		{
			name: "Passwords don't match",
			payload: map[string]string{
				"username":         "admin3",
				"email":            "admin3@test.com",
				"password":         "AdminPass12345",
				"confirm_password": "DifferentPass12345",
			},
			wantStatus: http.StatusBadRequest,
		},
		{
			name: "Missing email",
			payload: map[string]string{
				"username":         "admin4",
				"password":         "AdminPass12345",
				"confirm_password": "AdminPass12345",
			},
			wantStatus: http.StatusBadRequest,
		},
		{
			name: "Invalid email",
			payload: map[string]string{
				"username":         "admin5",
				"email":            "not-an-email",
				"password":         "AdminPass12345",
				"confirm_password": "AdminPass12345",
			},
			wantStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			jsonData, _ := json.Marshal(tt.payload)
			req := httptest.NewRequest("POST", "/user/setup/admin", bytes.NewBuffer(jsonData))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			r.ServeHTTP(w, req)

			if w.Code != tt.wantStatus {
				t.Errorf("Expected status %d, got %d. Body: %s", tt.wantStatus, w.Code, w.Body.String())
			}
		})
	}
}
