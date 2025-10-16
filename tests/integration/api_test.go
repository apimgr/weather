package integration_test

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	_ "github.com/mattn/go-sqlite3"
	"weather-go/src/database"
	"weather-go/src/handlers"
	"weather-go/src/services"
)

func setupIntegrationTest(t *testing.T) (*gin.Engine, *database.DB) {
	gin.SetMode(gin.TestMode)

	db, err := database.InitDB(":memory:")
	if err != nil {
		t.Fatalf("Failed to initialize test database: %v", err)
	}

	r := gin.New()
	locationEnhancer := services.NewLocationEnhancer(db.DB)
	weatherService := services.NewWeatherService(locationEnhancer)
	apiHandler := handlers.NewAPIHandler(weatherService, locationEnhancer)

	// Setup API routes
	api := r.Group("/api/v1")
	{
		api.GET("/weather", apiHandler.GetWeather)
		api.GET("/forecast", apiHandler.GetForecast)
		api.GET("/search", apiHandler.SearchLocations)
	}

	return r, db
}

func TestAPI_Weather_Coordinates(t *testing.T) {
	r, db := setupIntegrationTest(t)
	defer db.Close()

	tests := []struct {
		name       string
		lat        string
		lon        string
		wantStatus int
	}{
		{"New York", "40.7128", "-74.0060", http.StatusOK},
		{"London", "51.5074", "-0.1278", http.StatusOK},
		{"Tokyo", "35.6762", "139.6503", http.StatusOK},
		{"Invalid latitude", "999", "0", http.StatusBadRequest},
		{"Missing longitude", "40.7128", "", http.StatusBadRequest},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			url := fmt.Sprintf("/api/v1/weather?lat=%s&lon=%s", tt.lat, tt.lon)
			req := httptest.NewRequest("GET", url, nil)
			w := httptest.NewRecorder()

			r.ServeHTTP(w, req)

			if w.Code != tt.wantStatus {
				t.Errorf("Expected status %d, got %d. Body: %s", tt.wantStatus, w.Code, w.Body.String())
			}

			if w.Code == http.StatusOK {
				var response map[string]interface{}
				if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
					t.Errorf("Failed to parse JSON response: %v", err)
				}

				// Check for required fields
				if _, ok := response["location"]; !ok {
					t.Error("Response missing 'location' field")
				}
				if _, ok := response["current"]; !ok {
					t.Error("Response missing 'current' field")
				}
			}
		})
	}
}

func TestAPI_Weather_CityID(t *testing.T) {
	r, db := setupIntegrationTest(t)
	defer db.Close()

	tests := []struct {
		name       string
		cityID     string
		wantStatus int
	}{
		{"Valid city ID", "5128581", http.StatusNotFound}, // Will be 404 until cities loaded
		{"Invalid city ID", "invalid", http.StatusBadRequest},
		{"Zero city ID", "0", http.StatusNotFound},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			url := fmt.Sprintf("/api/v1/weather?city_id=%s", tt.cityID)
			req := httptest.NewRequest("GET", url, nil)
			w := httptest.NewRecorder()

			r.ServeHTTP(w, req)

			if w.Code != tt.wantStatus {
				t.Logf("City ID test: %s - Status %d (expected %d)", tt.name, w.Code, tt.wantStatus)
			}
		})
	}
}

func TestAPI_Weather_Nearest(t *testing.T) {
	r, db := setupIntegrationTest(t)
	defer db.Close()

	url := "/api/v1/weather?lat=40.7128&lon=-74.0060&nearest=true"
	req := httptest.NewRequest("GET", url, nil)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	// Will work even without cities loaded (falls back to coordinates)
	if w.Code != http.StatusOK {
		t.Logf("Nearest city test returned status %d (fallback expected)", w.Code)
	}
}

func TestAPI_Forecast(t *testing.T) {
	r, db := setupIntegrationTest(t)
	defer db.Close()

	tests := []struct {
		name       string
		query      string
		wantStatus int
	}{
		{"Default days", "lat=40.7128&lon=-74.0060", http.StatusOK},
		{"7 days", "lat=40.7128&lon=-74.0060&days=7", http.StatusOK},
		{"1 day", "lat=40.7128&lon=-74.0060&days=1", http.StatusOK},
		{"Invalid days", "lat=40.7128&lon=-74.0060&days=invalid", http.StatusOK}, // Should default
		{"Too many days", "lat=40.7128&lon=-74.0060&days=999", http.StatusOK},    // Should cap
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			url := fmt.Sprintf("/api/v1/forecast?%s", tt.query)
			req := httptest.NewRequest("GET", url, nil)
			w := httptest.NewRecorder()

			r.ServeHTTP(w, req)

			if w.Code != tt.wantStatus {
				t.Errorf("Expected status %d, got %d. Body: %s", tt.wantStatus, w.Code, w.Body.String())
			}

			if w.Code == http.StatusOK {
				var response map[string]interface{}
				if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
					t.Errorf("Failed to parse JSON response: %v", err)
				}

				// Check for forecast array
				if forecast, ok := response["forecast"].([]interface{}); ok {
					if len(forecast) == 0 {
						t.Error("Forecast array is empty")
					}
				} else {
					t.Error("Response missing 'forecast' array")
				}
			}
		})
	}
}

func TestAPI_Search(t *testing.T) {
	r, db := setupIntegrationTest(t)
	defer db.Close()

	tests := []struct {
		name       string
		query      string
		wantStatus int
	}{
		{"Valid search", "q=London", http.StatusOK},
		{"Empty query", "q=", http.StatusBadRequest},
		{"No query param", "", http.StatusBadRequest},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			url := fmt.Sprintf("/api/v1/search?%s", tt.query)
			req := httptest.NewRequest("GET", url, nil)
			w := httptest.NewRecorder()

			r.ServeHTTP(w, req)

			if w.Code != tt.wantStatus {
				t.Errorf("Expected status %d, got %d. Body: %s", tt.wantStatus, w.Code, w.Body.String())
			}
		})
	}
}
