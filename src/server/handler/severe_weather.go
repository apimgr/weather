package handlers

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"

	"github.com/apimgr/weather/src/server/middleware"
	"github.com/apimgr/weather/src/server/service"
	"github.com/apimgr/weather/src/utils"
)

// SevereWeatherHandler handles severe weather tracking requests
type SevereWeatherHandler struct {
	severeWeatherService *services.SevereWeatherService
	locationEnhancer     *services.LocationEnhancer
	weatherService       *services.WeatherService
}

// NewSevereWeatherHandler creates a new severe weather handler
func NewSevereWeatherHandler(severeWeatherService *services.SevereWeatherService, locationEnhancer *services.LocationEnhancer, weatherService *services.WeatherService) *SevereWeatherHandler {
	return &SevereWeatherHandler{
		severeWeatherService: severeWeatherService,
		locationEnhancer:     locationEnhancer,
		weatherService:       weatherService,
	}
}

// HandleSevereWeatherRequest handles severe weather page requests
func (h *SevereWeatherHandler) HandleSevereWeatherRequest(c *gin.Context) {
	// Get location from path parameter or query
	locationParam := c.Param("location")
	if locationParam == "" {
		locationParam = c.Query("location")
	}

	// Get distance filter from query parameter (default 50 miles)
	distanceParam := c.Query("distance")
	// default
	distance := 50.0
	if distanceParam != "" {
		if d, err := strconv.ParseFloat(distanceParam, 64); err == nil {
			distance = d
		}
	}

	// Parse coordinates from location or get from persistent storage
	var latitude, longitude float64
	var locationName string
	var locationCoords *services.Coordinates

	if locationParam != "" {
		// Try to parse as coordinates first
		lat, lon, err := utils.ParseCoordinates(locationParam)
		if err == nil {
			latitude = lat
			longitude = lon
			locationName = fmt.Sprintf("%.4f, %.4f", latitude, longitude)
		} else {
			// Geocode the location using weather service (proper resolution)
			clientIP := utils.GetClientIP(c)
			coords, err := h.weatherService.ParseAndResolveLocation(locationParam, clientIP)
			if err == nil {
				locationCoords = coords
				latitude = coords.Latitude
				longitude = coords.Longitude
				locationName = coords.ShortName
				if locationName == "" {
					locationName = coords.Name
				}
			}
		}
	} else if latitude == 0 && longitude == 0 {
		// Only check cookies if NO location was explicitly provided in URL
		if latStr, err := c.Cookie("user_lat"); err == nil {
			if lat, err := strconv.ParseFloat(latStr, 64); err == nil {
				latitude = lat
			}
		}
		if lonStr, err := c.Cookie("user_lon"); err == nil {
			if lon, err := strconv.ParseFloat(lonStr, 64); err == nil {
				longitude = lon
			}
		}
		if locName, err := c.Cookie("user_location_name"); err == nil {
			locationName = locName
			// Re-resolve the location to get full data
			clientIP := utils.GetClientIP(c)
			coords, err := h.weatherService.ParseAndResolveLocation(locationName, clientIP)
			if err == nil {
				locationCoords = coords
			}
		}
	}

	// If still no location, use IP-based geolocation as fallback
	if latitude == 0 && longitude == 0 {
		clientIP := utils.GetClientIP(c)
		coords, err := h.weatherService.GetCoordinatesFromIP(clientIP)
		if err == nil {
			locationCoords = coords
			enhanced := h.locationEnhancer.EnhanceLocation(coords)
			latitude = enhanced.Latitude
			longitude = enhanced.Longitude
			locationName = enhanced.ShortName
			if locationName == "" {
				locationName = enhanced.Name
			}
		}
	}

	// Fetch severe weather data with distance filter
	data, err := h.severeWeatherService.GetSevereWeatherWithDistance(latitude, longitude, distance)
	if err != nil {
		// Check if user wants JSON
		accept := c.GetHeader("Accept")
		wantsJSON := strings.Contains(accept, "application/json")

		if wantsJSON {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch severe weather data"})
		} else {
			c.String(http.StatusInternalServerError, "Failed to fetch severe weather data: %v", err)
		}
		return
	}

	// Save location to cookies for persistence across navigation
	if latitude != 0 && longitude != 0 && locationName != "" {
		middleware.SaveLocationCookies(c, latitude, longitude, locationName)
	}

	// Calculate totals
	totalAlerts := len(data.TornadoWarnings) + len(data.SevereStorms) + len(data.WinterStorms) + len(data.FloodWarnings) + len(data.OtherAlerts)
	totalStorms := len(data.Hurricanes)

	// Check if user wants JSON
	accept := c.GetHeader("Accept")
	wantsJSON := strings.Contains(accept, "application/json")

	if wantsJSON {
		c.JSON(http.StatusOK, data)
		return
	}

	// Check user agent to determine if browser or console
	isBrowser := utils.IsBrowser(c)

	if isBrowser {
		// Render HTML template
		hostInfo := utils.GetHostInfo(c)

		// Create Location object for uniform display
		var locationData interface{}
		if latitude != 0 || longitude != 0 {
			// Use the full locationCoords if available, otherwise create minimal coords
			var enhanced *services.Coordinates
			if locationCoords != nil {
				enhanced = locationCoords
			} else {
				// Create minimal coords and enhance
				coords := &services.Coordinates{
					Latitude:  latitude,
					Longitude: longitude,
					Name:      locationName,
					ShortName: locationName,
				}
				enhanced = h.locationEnhancer.EnhanceLocation(coords)
			}

			// Format population with commas
			popFormatted := ""
			if enhanced.Population > 0 {
				popFormatted = formatPopulation(enhanced.Population)
			}

			locationData = gin.H{
				"Location": gin.H{
					"Name":                enhanced.FullName,
					"ShortName":           enhanced.ShortName,
					"NameEncoded":         strings.ReplaceAll(enhanced.ShortName, " ", "+"),
					"Country":             enhanced.Country,
					"CountryCode":         enhanced.CountryCode,
					"Latitude":            latitude,
					"Longitude":           longitude,
					"Timezone":            enhanced.Timezone,
					"Population":          enhanced.Population,
					"PopulationFormatted": popFormatted,
				},
			}
		}

		// Get type and distance filters from query params
		typeFilter := c.Query("type")
		if typeFilter == "" {
			typeFilter = "all"
		}
		distanceFilter := fmt.Sprintf("%.0f", distance)

		// Always use full detected location for clarity
		displayLocation := locationName

		c.HTML(http.StatusOK, "pages/severe_weather.tmpl", gin.H{
			"Title":          "Severe Weather Alerts",
			"page":           "severe-weather",
			"Data":           data,
			"TotalAlerts":    totalAlerts,
			"TotalStorms":    totalStorms,
			"LocationName":   displayLocation,
			"LocationData":   locationData,
			"Latitude":       latitude,
			"Longitude":      longitude,
			"Distance":       distance,
			"DistanceFilter": distanceFilter,
			"TypeFilter":     typeFilter,
			"HasLocation":    latitude != 0 && longitude != 0,
			"HostInfo":       hostInfo,
		})
	} else {
		// Render console output
		output := h.renderConsoleOutput(data, locationName)
		c.String(http.StatusOK, output)
	}
}

// HandleSevereWeatherByType handles severe weather requests filtered by type
func (h *SevereWeatherHandler) HandleSevereWeatherByType(c *gin.Context) {
	alertType := c.Param("type")
	locationParam := c.Param("location")

	// Get distance filter from query parameter (default 50 miles)
	distanceParam := c.Query("distance")
	distance := 50.0
	if distanceParam != "" {
		if d, err := strconv.ParseFloat(distanceParam, 64); err == nil {
			distance = d
		}
	}

	// Parse location if provided
	var latitude, longitude float64
	var locationName string
	var locationCoords *services.Coordinates

	if locationParam != "" {
		// Try to parse as coordinates first
		lat, lon, err := utils.ParseCoordinates(locationParam)
		if err == nil {
			latitude = lat
			longitude = lon
			locationName = fmt.Sprintf("%.4f, %.4f", latitude, longitude)
		} else {
			// Geocode the location using weather service (proper resolution)
			clientIP := utils.GetClientIP(c)
			coords, err := h.weatherService.ParseAndResolveLocation(locationParam, clientIP)
			if err == nil {
				locationCoords = coords
				latitude = coords.Latitude
				longitude = coords.Longitude
				locationName = coords.ShortName
				if locationName == "" {
					locationName = coords.Name
				}
			}
		}
	} else if latitude == 0 && longitude == 0 {
		// Only check cookies if NO location was explicitly provided in URL
		if latStr, err := c.Cookie("user_lat"); err == nil {
			if lat, err := strconv.ParseFloat(latStr, 64); err == nil {
				latitude = lat
			}
		}
		if lonStr, err := c.Cookie("user_lon"); err == nil {
			if lon, err := strconv.ParseFloat(lonStr, 64); err == nil {
				longitude = lon
			}
		}
		if locName, err := c.Cookie("user_location_name"); err == nil {
			locationName = locName
			// Re-resolve the location to get full data
			clientIP := utils.GetClientIP(c)
			coords, err := h.weatherService.ParseAndResolveLocation(locationName, clientIP)
			if err == nil {
				locationCoords = coords
			}
		}
	}

	// If still no location, use IP-based fallback
	if latitude == 0 && longitude == 0 {
		clientIP := utils.GetClientIP(c)
		coords, err := h.weatherService.GetCoordinatesFromIP(clientIP)
		if err == nil {
			locationCoords = coords
			enhanced := h.locationEnhancer.EnhanceLocation(coords)
			latitude = enhanced.Latitude
			longitude = enhanced.Longitude
			locationName = enhanced.ShortName
			if locationName == "" {
				locationName = enhanced.Name
			}
		}
	}

	// Fetch severe weather data with distance filter
	data, err := h.severeWeatherService.GetSevereWeatherWithDistance(latitude, longitude, distance)
	if err != nil {
		accept := c.GetHeader("Accept")
		wantsJSON := strings.Contains(accept, "application/json")
		if wantsJSON {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch severe weather data"})
		} else {
			c.String(http.StatusInternalServerError, "Failed to fetch severe weather data: %v", err)
		}
		return
	}

	// Save location to cookies for persistence across navigation
	if latitude != 0 && longitude != 0 && locationName != "" {
		middleware.SaveLocationCookies(c, latitude, longitude, locationName)
	}

	// Filter data by type
	filteredData := &services.SevereWeatherData{
		LastUpdate: data.LastUpdate,
	}

	switch strings.ToLower(alertType) {
	case "hurricanes":
		filteredData.Hurricanes = data.Hurricanes
	case "tornadoes":
		filteredData.TornadoWarnings = data.TornadoWarnings
	case "floods":
		filteredData.FloodWarnings = data.FloodWarnings
	case "winter":
		filteredData.WinterStorms = data.WinterStorms
	case "storms":
		filteredData.SevereStorms = data.SevereStorms
	default:
		c.String(http.StatusBadRequest, "Invalid alert type. Use: hurricanes, tornadoes, floods, winter, or storms")
		return
	}

	// Calculate totals for filtered data
	totalAlerts := len(filteredData.TornadoWarnings) + len(filteredData.SevereStorms) +
		len(filteredData.WinterStorms) + len(filteredData.FloodWarnings)
	totalStorms := len(filteredData.Hurricanes)

	// Check if user wants JSON
	accept := c.GetHeader("Accept")
	wantsJSON := strings.Contains(accept, "application/json")

	if wantsJSON {
		c.JSON(http.StatusOK, filteredData)
		return
	}

	// Check user agent to determine if browser or console
	isBrowser := utils.IsBrowser(c)

	if isBrowser {
		// Render HTML template
		hostInfo := utils.GetHostInfo(c)

		// Create Location object for uniform display
		var locationData interface{}
		if latitude != 0 || longitude != 0 {
			// Use the full locationCoords if available, otherwise create minimal coords
			var enhanced *services.Coordinates
			if locationCoords != nil {
				enhanced = locationCoords
			} else {
				// Create minimal coords and enhance
				coords := &services.Coordinates{
					Latitude:  latitude,
					Longitude: longitude,
					Name:      locationName,
					ShortName: locationName,
				}
				enhanced = h.locationEnhancer.EnhanceLocation(coords)
			}

			// Format population with commas
			popFormatted := ""
			if enhanced.Population > 0 {
				popFormatted = formatPopulation(enhanced.Population)
			}

			locationData = gin.H{
				"Location": gin.H{
					"Name":                enhanced.FullName,
					"ShortName":           enhanced.ShortName,
					"NameEncoded":         strings.ReplaceAll(enhanced.ShortName, " ", "+"),
					"Country":             enhanced.Country,
					"CountryCode":         enhanced.CountryCode,
					"Latitude":            latitude,
					"Longitude":           longitude,
					"Timezone":            enhanced.Timezone,
					"Population":          enhanced.Population,
					"PopulationFormatted": popFormatted,
				},
			}
		}

		// Get distance filter
		distanceFilter := fmt.Sprintf("%.0f", distance)

		// Always use full detected location for clarity
		displayLocation := locationName

		c.HTML(http.StatusOK, "pages/severe_weather.tmpl", gin.H{
			"Title":          fmt.Sprintf("%s - Severe Weather Alerts", strings.Title(alertType)),
			"page":           "severe-weather",
			"Data":           filteredData,
			"TotalAlerts":    totalAlerts,
			"TotalStorms":    totalStorms,
			"LocationName":   displayLocation,
			"LocationData":   locationData,
			"Latitude":       latitude,
			"Longitude":      longitude,
			"Distance":       distance,
			"DistanceFilter": distanceFilter,
			"TypeFilter":     strings.ToLower(alertType),
			"HasLocation":    latitude != 0 && longitude != 0,
			"HostInfo":       hostInfo,
		})
	} else {
		// Render console output
		output := h.renderConsoleOutput(filteredData, locationName)
		c.String(http.StatusOK, output)
	}
}

// HandleSevereWeatherAPI handles JSON API requests for severe weather data
func (h *SevereWeatherHandler) HandleSevereWeatherAPI(c *gin.Context) {
	locationParam := c.Query("location")

	var latitude, longitude float64

	if locationParam != "" {
		// Try to parse as coordinates
		lat, lon, err := utils.ParseCoordinates(locationParam)
		if err == nil {
			latitude = lat
			longitude = lon
		} else {
			// Geocode using proper resolution
			clientIP := utils.GetClientIP(c)
			coords, err := h.weatherService.ParseAndResolveLocation(locationParam, clientIP)
			if err == nil {
				latitude = coords.Latitude
				longitude = coords.Longitude
			}
		}
	}

	data, err := h.severeWeatherService.GetSevereWeather(latitude, longitude)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch severe weather data"})
		return
	}

	c.JSON(http.StatusOK, data)
}

// renderConsoleOutput renders severe weather data for console/terminal
func (h *SevereWeatherHandler) renderConsoleOutput(data *services.SevereWeatherData, locationName string) string {
	output := "⚠️  Severe Weather Alerts\n"
	output += "═══════════════════════════════════════\n\n"

	if locationName != "" {
		output += "📍 Location: " + locationName + "\n"
		output += "───────────────────────────────────────\n\n"
	}

	// Hurricanes
	if len(data.Hurricanes) > 0 {
		output += "🌀 HURRICANES & TROPICAL STORMS (" + fmt.Sprintf("%d", len(data.Hurricanes)) + ")\n"
		output += "───────────────────────────────────────\n"
		for _, storm := range data.Hurricanes {
			icon := h.severeWeatherService.GetStormIcon(storm.Classification, storm.WindSpeed)
			category := h.severeWeatherService.GetStormCategory(storm.WindSpeed)

			output += icon + " " + storm.Name + "\n"
			output += "   Category: " + category + "\n"
			output += "   Wind: " + formatInt(storm.WindSpeed) + " mph\n"
			output += "   Pressure: " + formatInt(storm.Pressure) + " mb\n"
			output += "   Location: " + formatFloat(storm.Latitude) + ", " + formatFloat(storm.Longitude) + "\n"

			if storm.DistanceMiles > 0 {
				output += "   Distance: " + formatFloat(storm.DistanceMiles) + " miles\n"
			}

			if storm.MovementSpeed > 0 {
				output += "   Movement: " + storm.MovementDir + " at " + formatInt(storm.MovementSpeed) + " mph\n"
			}

			output += "\n"
		}
		output += "\n"
	}

	// Tornado Warnings
	if len(data.TornadoWarnings) > 0 {
		output += "🌪️  TORNADO WARNINGS (" + fmt.Sprintf("%d", len(data.TornadoWarnings)) + ")\n"
		output += "───────────────────────────────────────\n"
		for _, alert := range data.TornadoWarnings {
			output += h.formatAlert(alert)
		}
		output += "\n"
	}

	// Severe Thunderstorms
	if len(data.SevereStorms) > 0 {
		output += "⛈️  SEVERE THUNDERSTORM WARNINGS (" + fmt.Sprintf("%d", len(data.SevereStorms)) + ")\n"
		output += "───────────────────────────────────────\n"
		for _, alert := range data.SevereStorms {
			output += h.formatAlert(alert)
		}
		output += "\n"
	}

	// Winter Storms
	if len(data.WinterStorms) > 0 {
		output += "❄️  WINTER STORM WARNINGS (" + fmt.Sprintf("%d", len(data.WinterStorms)) + ")\n"
		output += "───────────────────────────────────────\n"
		for _, alert := range data.WinterStorms {
			output += h.formatAlert(alert)
		}
		output += "\n"
	}

	// Flood Warnings
	if len(data.FloodWarnings) > 0 {
		output += "🌊 FLOOD WARNINGS (" + fmt.Sprintf("%d", len(data.FloodWarnings)) + ")\n"
		output += "───────────────────────────────────────\n"
		for _, alert := range data.FloodWarnings {
			output += h.formatAlert(alert)
		}
		output += "\n"
	}

	// Other Alerts
	if len(data.OtherAlerts) > 0 {
		output += "⚠️  OTHER ALERTS (" + fmt.Sprintf("%d", len(data.OtherAlerts)) + ")\n"
		output += "───────────────────────────────────────\n"
		for _, alert := range data.OtherAlerts {
			output += h.formatAlert(alert)
		}
		output += "\n"
	}

	// No alerts
	if len(data.Hurricanes) == 0 && len(data.TornadoWarnings) == 0 && len(data.SevereStorms) == 0 &&
		len(data.WinterStorms) == 0 && len(data.FloodWarnings) == 0 && len(data.OtherAlerts) == 0 {
		output += "✅ No active severe weather alerts at this time.\n\n"
	}

	output += "Data from NOAA & National Weather Service\n"
	output += "Updates every 5 minutes\n\n"

	return output
}

// formatAlert formats an alert for console output
func (h *SevereWeatherHandler) formatAlert(alert services.Alert) string {
	icon := h.severeWeatherService.GetAlertIcon(alert.Event)
	output := icon + " " + alert.Event + "\n"
	output += "   Area: " + alert.AreaDesc + "\n"
	output += "   Severity: " + alert.Severity + " | Urgency: " + alert.Urgency + "\n"

	if alert.Headline != "" {
		// Truncate headline if too long
		headline := alert.Headline
		if len(headline) > 80 {
			headline = headline[:77] + "..."
		}
		output += "   " + headline + "\n"
	}

	if alert.Expires != "" {
		output += "   Expires: " + alert.Expires + "\n"
	}

	output += "\n"
	return output
}
