package handlers

import (
	"net/http"
	"strings"

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

	// Get location from path parameter or query
	location := c.Param("location")
	if location == "" {
		location = c.Query("location")
	}

	// Get units from query parameter (default to imperial)
	units := c.Query("units")
	if units == "" {
		units = "imperial"
	}

	clientIP := utils.GetClientIP(c)

	// If no location specified, show empty moon page
	if location == "" {
		c.HTML(http.StatusOK, "moon.html", gin.H{
			"Title":      "Moon Phase - Console Weather Service",
			"HostInfo":   utils.GetHostInfo(c),
			"Location":   "",
			"Units":      units,
			"HideFooter": false,
		})
		return
	}

	var coords *services.Coordinates
	var err error

	coords, err = h.weatherService.ParseAndResolveLocation(location, clientIP)

	if err != nil {
		c.HTML(http.StatusInternalServerError, "moon.html", gin.H{
			"Title":      "Moon Phase - Console Weather Service",
			"Error":      err.Error(),
			"HostInfo":   utils.GetHostInfo(c),
			"Location":   location,
			"Units":      units,
			"HideFooter": false,
		})
		return
	}

	// Enhance location
	enhanced := h.locationEnhancer.EnhanceLocation(coords)

	// TODO: Get moon data from moon service
	// For now, provide basic structure
	moonData := gin.H{
		"Location": gin.H{
			"Name":         enhanced.FullName,
			"NameEncoded":  strings.ReplaceAll(enhanced.ShortName, " ", "+"),
			"Latitude":     coords.Latitude,
			"Longitude":    coords.Longitude,
			"Timezone":     coords.Timezone,
			"Country":      coords.Country,
			"CountryCode":  coords.CountryCode,
		},
		"Moon": gin.H{
			"Phase":         "First Quarter",
			"Illumination":  50,
			"Icon":          "🌓",
			"Age":           7,
			"Rise":          "12:00",
			"Set":           "00:00",
			"RiseFormatted": "12:00 PM",
			"SetFormatted":  "12:00 AM",
		},
		"Sun": gin.H{
			"SunriseFormatted":   "6:30 AM",
			"SunsetFormatted":    "6:30 PM",
			"SolarNoonFormatted": "12:30 PM",
			"DayLengthFormatted": "12h 0m",
		},
	}

	c.HTML(http.StatusOK, "moon.html", gin.H{
		"Title":      "Moon Phase - " + enhanced.ShortName,
		"MoonData":   moonData,
		"HostInfo":   utils.GetHostInfo(c),
		"Location":   enhanced.ShortName,
		"Units":      units,
		"HideFooter": false,
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
