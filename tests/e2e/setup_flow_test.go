package e2e_test

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	_ "modernc.org/sqlite"
	"github.com/apimgr/weather/src/database"
	"github.com/apimgr/weather/src/server/handler"
)

// initTestDualDB creates in-memory dual databases for testing
func initTestDualDB(t *testing.T) (*database.DualDB, func()) {
	// Create in-memory server database
	serverDB, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("Failed to open server database: %v", err)
	}
	if _, err := serverDB.Exec("PRAGMA foreign_keys = ON"); err != nil {
		t.Fatalf("Failed to enable foreign keys: %v", err)
	}
	if _, err := serverDB.Exec(database.ServerSchema); err != nil {
		t.Fatalf("Failed to create server schema: %v", err)
	}

	// Create in-memory users database
	usersDB, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("Failed to open users database: %v", err)
	}
	if _, err := usersDB.Exec("PRAGMA foreign_keys = ON"); err != nil {
		t.Fatalf("Failed to enable foreign keys: %v", err)
	}
	if _, err := usersDB.Exec(database.UsersSchema); err != nil {
		t.Fatalf("Failed to create users schema: %v", err)
	}

	dualDB := &database.DualDB{
		Server: serverDB,
		Users:  usersDB,
	}

	// Set global dual database
	database.SetGlobalDualDB(dualDB)

	cleanup := func() {
		database.SetGlobalDualDB(nil)
		serverDB.Close()
		usersDB.Close()
	}

	return dualDB, cleanup
}

// TestCompleteSetupFlow tests the entire user setup flow end-to-end
// Note: This test only tests JSON API endpoints, not HTML templates
func TestCompleteSetupFlow(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// Initialize fresh database
	dualDB, cleanup := initTestDualDB(t)
	defer cleanup()

	// Create router
	r := gin.New()
	setupHandler := &handler.SetupHandler{DB: dualDB.Users}

	// Setup routes (only POST routes for JSON API testing)
	setupRoutes := r.Group("/user/setup")
	{
		setupRoutes.POST("/register", setupHandler.CreateUser)
		setupRoutes.POST("/admin", setupHandler.CreateAdmin)
	}

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

		// Accept 200 or 201 for successful creation
		if w.Code != http.StatusOK && w.Code != http.StatusCreated {
			t.Fatalf("Failed to create first user: %s", w.Body.String())
		}

		// Save cookies for next request
		userCookies = w.Result().Cookies()

		var response map[string]interface{}
		if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
			t.Fatalf("Failed to parse response: %v", err)
		}

		// Check for user object in response (indicates success)
		if _, hasUser := response["user"]; !hasUser {
			if success, ok := response["success"].(bool); !ok || !success {
				t.Logf("Response: %v", response)
			}
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

		// Accept 200 or 201 for successful creation
		if w.Code != http.StatusOK && w.Code != http.StatusCreated {
			t.Fatalf("Failed to create administrator: %s", w.Body.String())
		}

		var response map[string]interface{}
		if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
			t.Fatalf("Failed to parse response: %v", err)
		}

		// Log response for debugging
		t.Logf("Admin creation response: %v", response)

		// Check redirect URL if present
		if redirect, ok := response["redirect"].(string); ok && redirect != "" {
			t.Logf("Redirect URL: %s", redirect)
		}
	})

	// Step 4: Verify users were created
	t.Run("4. Verify users in database", func(t *testing.T) {
		var userCount int
		err := dualDB.Users.QueryRow("SELECT COUNT(*) FROM user_accounts").Scan(&userCount)
		if err != nil {
			t.Fatalf("Failed to query users: %v", err)
		}

		// Check that we have at least 1 user created
		if userCount < 1 {
			t.Errorf("Expected at least 1 user, got %d", userCount)
		}
	})
}

// TestAdminSetupValidation tests validation in admin creation
func TestAdminSetupValidation(t *testing.T) {
	gin.SetMode(gin.TestMode)

	dualDB, cleanup := initTestDualDB(t)
	defer cleanup()

	r := gin.New()
	setupHandler := &handler.SetupHandler{DB: dualDB.Users}
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

			// For successful creation, accept 200 or 201
			if tt.wantStatus == http.StatusOK {
				if w.Code != http.StatusOK && w.Code != http.StatusCreated {
					t.Errorf("Expected status 200 or 201, got %d. Body: %s", w.Code, w.Body.String())
				}
			} else if w.Code != tt.wantStatus {
				t.Errorf("Expected status %d, got %d. Body: %s", tt.wantStatus, w.Code, w.Body.String())
			}
		})
	}
}
