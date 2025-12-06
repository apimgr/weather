package handlers

import (
	"database/sql"
	"net/http"
	"strconv"
	"strings"

	"weather-go/src/middleware"
	"weather-go/src/models"
	"weather-go/src/services"
	"weather-go/src/utils"

	"github.com/gin-gonic/gin"
)

type LocationHandler struct {
	DB               *sql.DB
	WeatherService   *services.WeatherService
	LocationEnhancer *services.LocationEnhancer
}

// ListLocations returns all saved locations for the current user
func (h *LocationHandler) ListLocations(c *gin.Context) {
	user, ok := middleware.GetCurrentUser(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Not authenticated"})
		return
	}

	locationModel := &models.LocationModel{DB: h.DB}
	locations, err := locationModel.GetByUserID(user.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch locations"})
		return
	}

	c.JSON(http.StatusOK, locations)
}

// GetLocation returns a specific saved location
func (h *LocationHandler) GetLocation(c *gin.Context) {
	user, ok := middleware.GetCurrentUser(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Not authenticated"})
		return
	}

	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid location ID"})
		return
	}

	locationModel := &models.LocationModel{DB: h.DB}
	location, err := locationModel.GetByID(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Location not found"})
		return
	}

	// Verify ownership
	if location.UserID != user.ID {
		c.JSON(http.StatusForbidden, gin.H{"error": "Access denied"})
		return
	}

	c.JSON(http.StatusOK, location)
}

// CreateLocation creates a new saved location
func (h *LocationHandler) CreateLocation(c *gin.Context) {
	user, ok := middleware.GetCurrentUser(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Not authenticated"})
		return
	}

	var req struct {
		Name      string  `json:"name" binding:"required"`
		Latitude  float64 `json:"latitude" binding:"required"`
		Longitude float64 `json:"longitude" binding:"required"`
		Timezone  string  `json:"timezone"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Validate coordinates
	if req.Latitude < -90 || req.Latitude > 90 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Latitude must be between -90 and 90"})
		return
	}
	if req.Longitude < -180 || req.Longitude > 180 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Longitude must be between -180 and 180"})
		return
	}

	locationModel := &models.LocationModel{DB: h.DB}
	location, err := locationModel.Create(user.ID, req.Name, req.Latitude, req.Longitude, req.Timezone)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create location"})
		return
	}

	c.JSON(http.StatusCreated, location)
}

// UpdateLocation updates a saved location
func (h *LocationHandler) UpdateLocation(c *gin.Context) {
	user, ok := middleware.GetCurrentUser(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Not authenticated"})
		return
	}

	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid location ID"})
		return
	}

	var req struct {
		Name          string  `json:"name" binding:"required"`
		Latitude      float64 `json:"latitude" binding:"required"`
		Longitude     float64 `json:"longitude" binding:"required"`
		Timezone      string  `json:"timezone"`
		AlertsEnabled bool    `json:"alerts_enabled"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	locationModel := &models.LocationModel{DB: h.DB}

	// Verify ownership
	location, err := locationModel.GetByID(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Location not found"})
		return
	}
	if location.UserID != user.ID {
		c.JSON(http.StatusForbidden, gin.H{"error": "Access denied"})
		return
	}

	// Update location
	if err := locationModel.Update(id, req.Name, req.Latitude, req.Longitude, req.Timezone, req.AlertsEnabled); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update location"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Location updated successfully"})
}

// DeleteLocation deletes a saved location
func (h *LocationHandler) DeleteLocation(c *gin.Context) {
	user, ok := middleware.GetCurrentUser(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Not authenticated"})
		return
	}

	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid location ID"})
		return
	}

	locationModel := &models.LocationModel{DB: h.DB}

	// Verify ownership
	location, err := locationModel.GetByID(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Location not found"})
		return
	}
	if location.UserID != user.ID {
		c.JSON(http.StatusForbidden, gin.H{"error": "Access denied"})
		return
	}

	// Delete location
	if err := locationModel.Delete(id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete location"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Location deleted successfully"})
}

// ToggleAlerts toggles weather alerts for a location
func (h *LocationHandler) ToggleAlerts(c *gin.Context) {
	user, ok := middleware.GetCurrentUser(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Not authenticated"})
		return
	}

	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid location ID"})
		return
	}

	var req struct {
		Enabled bool `json:"enabled"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	locationModel := &models.LocationModel{DB: h.DB}

	// Verify ownership
	location, err := locationModel.GetByID(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Location not found"})
		return
	}
	if location.UserID != user.ID {
		c.JSON(http.StatusForbidden, gin.H{"error": "Access denied"})
		return
	}

	// Toggle alerts
	if err := locationModel.ToggleAlerts(id, req.Enabled); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to toggle alerts"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Alerts updated successfully", "enabled": req.Enabled})
}

// ShowAddLocationPage renders the add location page
func (h *LocationHandler) ShowAddLocationPage(c *gin.Context) {
	user, ok := middleware.GetCurrentUser(c)
	if !ok {
		c.Redirect(http.StatusFound, "/login")
		return
	}

	c.HTML(http.StatusOK, "add-location.tmpl", utils.TemplateData(c, gin.H{
		"title": "Add Location - Weather Service",
		"user":  user,
		"page":  "locations",
	}))
}

// ShowEditLocationPage renders the edit location page
func (h *LocationHandler) ShowEditLocationPage(c *gin.Context) {
	user, ok := middleware.GetCurrentUser(c)
	if !ok {
		c.Redirect(http.StatusFound, "/login")
		return
	}

	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.HTML(http.StatusBadRequest, "error.tmpl", gin.H{"error": "Invalid location ID"})
		return
	}

	locationModel := &models.LocationModel{DB: h.DB}
	location, err := locationModel.GetByID(id)
	if err != nil {
		c.HTML(http.StatusNotFound, "error.tmpl", gin.H{"error": "Location not found"})
		return
	}

	if location.UserID != user.ID {
		c.HTML(http.StatusForbidden, "error.tmpl", gin.H{"error": "Access denied"})
		return
	}

	c.HTML(http.StatusOK, "edit-location.tmpl", utils.TemplateData(c, gin.H{
		"title":    "Edit Location - Weather Service",
		"user":     user,
		"location": location,
		"page":     "locations",
	}))
}

// SearchLocations searches for cities by name
func (h *LocationHandler) SearchLocations(c *gin.Context) {
	query := strings.TrimSpace(c.Query("q"))
	if len(query) < 2 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Query must be at least 2 characters"})
		return
	}

	// Search in cities database
	results := h.searchCities(query, 10)

	c.JSON(http.StatusOK, results)
}

// LookupZipCode looks up a location by ZIP/postal code
func (h *LocationHandler) LookupZipCode(c *gin.Context) {
	zipCode := strings.TrimSpace(c.Param("code"))
	if zipCode == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ZIP code is required"})
		return
	}

	// Use weather service to geocode the ZIP code
	coords, err := h.WeatherService.ParseAndResolveLocation(zipCode, "")
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "ZIP code not found"})
		return
	}

	// Enhance the location with full details
	enhanced := h.LocationEnhancer.EnhanceLocation(coords)

	// Return location data
	result := gin.H{
		"name":      enhanced.Name,
		"latitude":  enhanced.Latitude,
		"longitude": enhanced.Longitude,
		"timezone":  enhanced.Timezone,
		"country":   enhanced.Country,
		"admin1":    enhanced.Admin1,
	}

	c.JSON(http.StatusOK, result)
}

// LookupCoordinates reverse geocodes coordinates to get location info
func (h *LocationHandler) LookupCoordinates(c *gin.Context) {
	latStr := c.Query("lat")
	lonStr := c.Query("lon")

	lat, err := strconv.ParseFloat(latStr, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid latitude"})
		return
	}

	lon, err := strconv.ParseFloat(lonStr, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid longitude"})
		return
	}

	// Validate coordinate ranges
	if lat < -90 || lat > 90 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Latitude must be between -90 and 90"})
		return
	}
	if lon < -180 || lon > 180 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Longitude must be between -180 and 180"})
		return
	}

	// Use weather service to reverse geocode
	coords, err := h.WeatherService.ReverseGeocode(lat, lon)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Location not found"})
		return
	}

	// Enhance the location with full details
	enhanced := h.LocationEnhancer.EnhanceLocation(coords)

	// Return location data
	result := gin.H{
		"name":      enhanced.Name,
		"latitude":  enhanced.Latitude,
		"longitude": enhanced.Longitude,
		"timezone":  enhanced.Timezone,
		"country":   enhanced.Country,
		"admin1":    enhanced.Admin1,
	}

	c.JSON(http.StatusOK, result)
}

// searchCities searches the cities database for matching names
func (h *LocationHandler) searchCities(query string, limit int) []gin.H {
	if h.LocationEnhancer == nil {
		return []gin.H{}
	}

	queryLower := strings.ToLower(query)
	var results []gin.H

	// Search through cities data
	for _, city := range h.LocationEnhancer.GetCitiesData() {
		// Check if city name starts with query (higher priority)
		if strings.HasPrefix(strings.ToLower(city.Name), queryLower) {
			results = append(results, h.formatCityResult(city))
			if len(results) >= limit {
				break
			}
			continue
		}

		// Check if city name contains query
		if strings.Contains(strings.ToLower(city.Name), queryLower) {
			results = append(results, h.formatCityResult(city))
			if len(results) >= limit {
				break
			}
		}
	}

	return results
}

// formatCityResult formats a city for the search results
func (h *LocationHandler) formatCityResult(city services.City) gin.H {
	// Format the display name
	displayParts := []string{city.Name}
	if city.State != "" {
		displayParts = append(displayParts, city.State)
	}
	if city.Country != "" {
		displayParts = append(displayParts, city.Country)
	}

	return gin.H{
		"name":      city.Name,
		"latitude":  city.Coord.Lat,
		"longitude": city.Coord.Lon,
		"country":   city.Country,
		"admin1":    city.State,
		"timezone":  "",
		"display":   strings.Join(displayParts, ", "),
	}
}
