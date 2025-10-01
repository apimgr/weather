package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"weather-go/services"
	"weather-go/utils"
)

// WebHandler handles web interface routes (HTML pages)
type WebHandler struct {
	weatherService   *services.WeatherService
	locationEnhancer *services.LocationEnhancer
}

// NewWebHandler creates a new web handler
func NewWebHandler(ws *services.WeatherService, le *services.LocationEnhancer) *WebHandler {
	return &WebHandler{
		weatherService:   ws,
		locationEnhancer: le,
	}
}

// ServeWebInterface serves the main web interface
func (h *WebHandler) ServeWebInterface(c *gin.Context) {
	// Check if services are initialized
	if !IsInitialized() {
		ServeLoadingPage(c)
		return
	}

	location := c.Query("location")
	units := c.Query("units")
	if units == "" {
		units = "imperial"
	}

	var weatherData interface{}
	var errorMsg string

	// If location is provided, fetch weather
	if location != "" {
		clientIP := utils.GetClientIP(c)
		coords, err := h.weatherService.ParseAndResolveLocation(location, clientIP)
		if err != nil {
			errorMsg = err.Error()
		} else {
			// Enhance location
			enhanced := h.locationEnhancer.EnhanceLocation(coords)

			// Get current weather and forecast
			current, err := h.weatherService.GetCurrentWeather(enhanced.Latitude, enhanced.Longitude, units)
			if err != nil {
				errorMsg = err.Error()
			} else {
				forecast, _ := h.weatherService.GetForecast(enhanced.Latitude, enhanced.Longitude, 7, units)

				weatherData = gin.H{
					"location": enhanced,
					"current":  current,
					"forecast": forecast,
					"units":    units,
				}
			}
		}
	}

	hostInfo := utils.GetHostInfo(c)

	c.HTML(http.StatusOK, "weather.html", gin.H{
		"title":       "Console Weather Service",
		"weatherData": weatherData,
		"hostInfo":    hostInfo,
		"location":    location,
		"units":       units,
		"error":       errorMsg,
	})
}

// ServeMoonInterface serves the moon phase interface
func (h *WebHandler) ServeMoonInterface(c *gin.Context) {
	// Check if services are initialized
	if !IsInitialized() {
		ServeLoadingPage(c)
		return
	}

	location := c.Query("location")
	clientIP := utils.GetClientIP(c)

	var coords *services.Coordinates
	var err error

	if location != "" {
		coords, err = h.weatherService.ParseAndResolveLocation(location, clientIP)
	} else {
		coords, err = h.weatherService.GetCoordinatesFromIP(clientIP)
	}

	if err != nil {
		c.HTML(http.StatusInternalServerError, "moon.html", gin.H{
			"title":    "Moon Phase Report - Console Weather Service",
			"error":    err.Error(),
			"hostInfo": utils.GetHostInfo(c),
		})
		return
	}

	// Enhance location
	enhanced := h.locationEnhancer.EnhanceLocation(coords)

	// TODO: Get moon data from moon service
	// For now, provide basic structure
	moonData := gin.H{
		"location": enhanced,
		"moon": gin.H{
			"phase":         "First Quarter",
			"illumination":  0.5,
			"icon":          "🌓",
			"age":           7.4,
		},
	}

	c.HTML(http.StatusOK, "moon.html", gin.H{
		"title":    "Moon Phase Report - Console Weather Service",
		"moonData": moonData,
		"hostInfo": utils.GetHostInfo(c),
		"location": location,
	})
}

// ServeExamplesPage serves the examples/documentation page
func (h *WebHandler) ServeExamplesPage(c *gin.Context) {
	hostInfo := utils.GetHostInfo(c)

	c.HTML(http.StatusOK, "examples.html", gin.H{
		"title":    "Examples - Console Weather Service",
		"hostInfo": hostInfo,
	})
}
