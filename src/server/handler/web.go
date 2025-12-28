package handler

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/apimgr/weather/src/server/middleware"
	"github.com/apimgr/weather/src/server/service"
	"github.com/apimgr/weather/src/utils"
)

// WebHandler handles web interface routes (HTML pages)
type WebHandler struct {
	weatherService   *service.WeatherService
	locationEnhancer *service.LocationEnhancer
}

// NewWebHandler creates a new web handler
func NewWebHandler(ws *service.WeatherService, le *service.LocationEnhancer) *WebHandler {
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
	clientIP := utils.GetClientIP(c)

	// Auto-detect location using priority order:
	// 1. URL parameter (already checked above)
	// 2. Saved cookies
	// 3. IP geolocation (fallback)
	if location == "" {
		// Check cookies first
		if latStr, err := c.Cookie("user_lat"); err == nil {
			if lonStr, err := c.Cookie("user_lon"); err == nil {
				if _, err1 := strconv.ParseFloat(latStr, 64); err1 == nil {
					if _, err2 := strconv.ParseFloat(lonStr, 64); err2 == nil {
						if locationName, err := c.Cookie("user_location_name"); err == nil {
							location = locationName
						}
					}
				}
			}
		}

		// If still no location, use IP geolocation as fallback
		if location == "" {
			coords, err := h.weatherService.GetCoordinatesFromIP(clientIP)
			if err == nil {
				enhanced := h.locationEnhancer.EnhanceLocation(coords)
				location = enhanced.ShortName
			}
		}
	}

	// If location is available (provided or auto-detected), fetch weather
	if location != "" {
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
				forecast, _ := h.weatherService.GetForecast(enhanced.Latitude, enhanced.Longitude, 16, units)

				// Enrich current weather with icon and description
				currentData := gin.H{
					"Temperature":   current.Temperature,
					"FeelsLike":     current.FeelsLike,
					"Humidity":      current.Humidity,
					"Pressure":      current.Pressure,
					"WindSpeed":     current.WindSpeed,
					"WindDirection": current.WindDirection,
					"Precipitation": current.Precipitation,
					"WeatherCode":   current.WeatherCode,
					"Icon":          h.weatherService.GetWeatherIcon(current.WeatherCode, current.IsDay == 1),
					"Description":   h.weatherService.GetWeatherDescription(current.WeatherCode),
				}

				// Format location data
				locationData := gin.H{
					"Name":                enhanced.Name,
					"ShortName":           enhanced.ShortName,
					"FullName":            enhanced.FullName,
					"Latitude":            enhanced.Latitude,
					"Longitude":           enhanced.Longitude,
					"Country":             enhanced.Country,
					"CountryCode":         enhanced.CountryCode,
					"Population":          enhanced.Population,
					"PopulationFormatted": fmt.Sprintf("%d", enhanced.Population),
					"Timezone":            enhanced.Timezone,
				}

				weatherData = gin.H{
					"Location": locationData,
					"Current":  currentData,
					"Forecast": forecast,
					"Units":    units,
				}
			}
		}
	}

	hostInfo := utils.GetHostInfo(c)

	// Format location for URLs (replace spaces with +, keep commas)
	locationFormatted := strings.ReplaceAll(location, " ", "+")

	c.HTML(http.StatusOK, "pages/weather.tmpl", gin.H{
		"Title":             "Weather",
		"WeatherData":       weatherData,
		"HostInfo":          hostInfo,
		"Location":          location,
		"LocationFormatted": locationFormatted,
		"Units":             units,
		"Error":             errorMsg,
		"HideFooter":        false,
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

	// Auto-detect location using priority order:
	// 1. URL parameter (already checked above)
	// 2. Saved cookies
	// 3. IP geolocation (fallback)
	if location == "" {
		// Check cookies first
		if latStr, err := c.Cookie("user_lat"); err == nil {
			if lonStr, err := c.Cookie("user_lon"); err == nil {
				if _, err1 := strconv.ParseFloat(latStr, 64); err1 == nil {
					if _, err2 := strconv.ParseFloat(lonStr, 64); err2 == nil {
						if locationName, err := c.Cookie("user_location_name"); err == nil {
							location = locationName
						}
					}
				}
			}
		}

		// If still no location, use IP geolocation as fallback
		if location == "" {
			coords, err := h.weatherService.GetCoordinatesFromIP(clientIP)
			if err == nil {
				enhanced := h.locationEnhancer.EnhanceLocation(coords)
				location = enhanced.ShortName
			}
		}
	}

	// If still no location, show empty moon page
	if location == "" {
		c.HTML(http.StatusOK, "pages/moon.tmpl", gin.H{
			"Title":      "Moon Phase - Weather",
			"HostInfo":   utils.GetHostInfo(c),
			"Location":   "",
			"Units":      units,
			"HideFooter": false,
		})
		return
	}

	var coords *service.Coordinates
	var err error

	coords, err = h.weatherService.ParseAndResolveLocation(location, clientIP)

	if err != nil {
		c.HTML(http.StatusInternalServerError, "pages/moon.tmpl", gin.H{
			"Title":      "Moon Phase - Weather",
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

	// Save location to cookies for persistence across navigation
	middleware.SaveLocationCookies(c, enhanced.Latitude, enhanced.Longitude, enhanced.ShortName)

	// Get moon data from moon service
	moonService := service.NewMoonService()
	moonDataCalc := moonService.Calculate(enhanced.Latitude, enhanced.Longitude, time.Now())

	moonData := gin.H{
		"Location": gin.H{
			"Name":                enhanced.Name,
			"ShortName":           enhanced.ShortName,
			"FullName":            enhanced.FullName,
			"NameEncoded":         strings.ReplaceAll(enhanced.ShortName, " ", "+"),
			"Latitude":            enhanced.Latitude,
			"Longitude":           enhanced.Longitude,
			"Timezone":            enhanced.Timezone,
			"Country":             enhanced.Country,
			"CountryCode":         enhanced.CountryCode,
			"Population":          enhanced.Population,
			"PopulationFormatted": fmt.Sprintf("%d", enhanced.Population),
		},
		"Moon": gin.H{
			"Phase":         moonDataCalc.Phase,
			"Illumination":  moonDataCalc.Illumination,
			"Icon":          moonDataCalc.Icon,
			"Age":           moonDataCalc.Age,
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

	// Always use full location (ShortName) for clarity
	// This shows "Albany, NY" instead of just "Albany"
	displayLocation := enhanced.ShortName

	c.HTML(http.StatusOK, "pages/moon.tmpl", gin.H{
		"Title":             "Moon Phase - " + enhanced.ShortName,
		"MoonData":          moonData,
		"HostInfo":          utils.GetHostInfo(c),
		"Location":          displayLocation,
		"LocationFormatted": strings.ReplaceAll(enhanced.ShortName, " ", "+"),
		"Units":             units,
		"HideFooter":        false,
	})
}

// ServeExamplesPage serves the examples/documentation page
func (h *WebHandler) ServeExamplesPage(c *gin.Context) {
	hostInfo := utils.GetHostInfo(c)

	c.HTML(http.StatusOK, "examples.tmpl", gin.H{
		"title":    "Examples - Weather",
		"hostInfo": hostInfo,
	})
}
