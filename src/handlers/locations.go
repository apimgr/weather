package handlers

import (
	"database/sql"
	"net/http"
	"strconv"

	"weather-go/src/middleware"
	"weather-go/src/models"

	"github.com/gin-gonic/gin"
)

type LocationHandler struct {
	DB *sql.DB
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

	c.HTML(http.StatusOK, "add-location.html", gin.H{
		"title": "Add Location - Weather Service",
		"user":  user,
	})
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
		c.HTML(http.StatusBadRequest, "error.html", gin.H{"error": "Invalid location ID"})
		return
	}

	locationModel := &models.LocationModel{DB: h.DB}
	location, err := locationModel.GetByID(id)
	if err != nil {
		c.HTML(http.StatusNotFound, "error.html", gin.H{"error": "Location not found"})
		return
	}

	if location.UserID != user.ID {
		c.HTML(http.StatusForbidden, "error.html", gin.H{"error": "Access denied"})
		return
	}

	c.HTML(http.StatusOK, "edit-location.html", gin.H{
		"title":    "Edit Location - Weather Service",
		"user":     user,
		"location": location,
	})
}
