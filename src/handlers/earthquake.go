package handlers

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"

	"weather-go/src/middleware"
	"weather-go/src/services"
	"weather-go/src/utils"
)

// EarthquakeHandler handles earthquake-related routes
type EarthquakeHandler struct {
	earthquakeService *services.EarthquakeService
	weatherService    *services.WeatherService
	locationEnhancer  *services.LocationEnhancer
}

// NewEarthquakeHandler creates a new earthquake handler
func NewEarthquakeHandler(es *services.EarthquakeService, ws *services.WeatherService, le *services.LocationEnhancer) *EarthquakeHandler {
	return &EarthquakeHandler{
		earthquakeService: es,
		weatherService:    ws,
		locationEnhancer:  le,
	}
}

// HandleEarthquakes serves the earthquake interface
func (h *EarthquakeHandler) HandleEarthquakes(c *gin.Context) {
	// Check if services are initialized
	if !IsInitialized() {
		ServeLoadingPage(c)
		return
	}

	// Get query parameters
	feedType := c.DefaultQuery("feed", "all_day")
	sortBy := c.DefaultQuery("sort", "newest")
	numberStr := c.DefaultQuery("number", "0")
	number, _ := strconv.Atoi(numberStr)

	// Validate feed type
	validFeeds := map[string]bool{
		"all_hour": true, "all_day": true, "all_week": true, "all_month": true,
		"1.0_hour": true, "1.0_day": true, "1.0_week": true, "1.0_month": true,
		"2.5_hour": true, "2.5_day": true, "2.5_week": true, "2.5_month": true,
		"4.5_hour": true, "4.5_day": true, "4.5_week": true, "4.5_month": true,
		"significant_hour": true, "significant_day": true, "significant_week": true, "significant_month": true,
	}

	if !validFeeds[feedType] {
		feedType = "all_day"
	}

	// Fetch earthquake data
	earthquakes, err := h.earthquakeService.GetEarthquakes(feedType)
	if err != nil {
		c.HTML(http.StatusInternalServerError, "error.html", gin.H{
			"error": "Failed to load earthquake data: " + err.Error(),
		})
		return
	}

	// Sort and limit results
	earthquakes.SortAndLimit(sortBy, number)

	// Get client location for map centering using priority order:
	// 1. Saved cookies
	// 2. IP geolocation (fallback)
	clientIP := utils.GetClientIP(c)
	var centerLat, centerLon float64 = 0.0, 0.0 // Default to world view

	// Check cookies first
	if latStr, err := c.Cookie("user_lat"); err == nil {
		if lonStr, err := c.Cookie("user_lon"); err == nil {
			if lat, err1 := strconv.ParseFloat(latStr, 64); err1 == nil {
				if lon, err2 := strconv.ParseFloat(lonStr, 64); err2 == nil {
					centerLat = lat
					centerLon = lon
				}
			}
		}
	}

	// If no cookies, use IP geolocation as fallback
	if centerLat == 0.0 && centerLon == 0.0 {
		coords, err := h.weatherService.GetCoordinatesFromIP(clientIP)
		if err == nil {
			centerLat = coords.Latitude
			centerLon = coords.Longitude
		}
	}

	// Get host info for console commands
	hostInfo := utils.GetHostInfo(c)

	// Render earthquake page
	c.HTML(http.StatusOK, "earthquake.html", gin.H{
		"Earthquakes": earthquakes.Earthquakes,
		"Metadata":    earthquakes.Metadata,
		"FeedType":    feedType,
		"CenterLat":   centerLat,
		"CenterLon":   centerLon,
		"HostInfo":    hostInfo,
	})
}

// HandleEarthquakesByLocation serves earthquakes near a specific location
func (h *EarthquakeHandler) HandleEarthquakesByLocation(c *gin.Context) {
	// Check if services are initialized
	if !IsInitialized() {
		ServeLoadingPage(c)
		return
	}

	locationInput := c.Param("location")
	if locationInput == "" {
		c.Redirect(http.StatusFound, "/earthquake")
		return
	}

	// Get query parameters
	feedType := c.DefaultQuery("feed", "all_week")
	radiusStr := c.DefaultQuery("radius", "500")
	sortBy := c.DefaultQuery("sort", "newest")
	numberStr := c.DefaultQuery("number", "0")
	number, _ := strconv.Atoi(numberStr)

	radius, err := strconv.ParseFloat(radiusStr, 64)
	if err != nil || radius <= 0 {
		radius = 500 // Default 500km
	}

	// Parse location
	clientIP := utils.GetClientIP(c)
	coords, err := h.weatherService.ParseAndResolveLocation(locationInput, clientIP)
	if err != nil {
		c.HTML(http.StatusBadRequest, "error.html", gin.H{
			"error": "Location not found: " + locationInput,
		})
		return
	}

	// Enhance location
	enhanced := h.locationEnhancer.EnhanceLocation(coords)

	// Save location to cookies for persistence across navigation
	middleware.SaveLocationCookies(c, enhanced.Latitude, enhanced.Longitude, enhanced.ShortName)

	// Get earthquakes near location
	earthquakes, err := h.earthquakeService.GetEarthquakesByLocation(
		enhanced.Latitude,
		enhanced.Longitude,
		radius,
		feedType,
	)
	if err != nil {
		c.HTML(http.StatusInternalServerError, "error.html", gin.H{
			"error": "Failed to load earthquake data: " + err.Error(),
		})
		return
	}

	// Sort and limit results
	earthquakes.SortAndLimit(sortBy, number)

	// Get host info for console commands
	hostInfo := utils.GetHostInfo(c)

	// Create LocationData for uniform display
	// Format population with commas
	popFormatted := ""
	if enhanced.Population > 0 {
		popFormatted = formatPopulation(enhanced.Population)
	}

	locationData := gin.H{
		"Location": gin.H{
			"Name":                enhanced.FullName,
			"ShortName":           enhanced.ShortName,
			"Country":             enhanced.Country,
			"Latitude":            enhanced.Latitude,
			"Longitude":           enhanced.Longitude,
			"Timezone":            enhanced.Timezone,
			"Population":          enhanced.Population,
			"PopulationFormatted": popFormatted,
		},
	}

	// Render earthquake page
	c.HTML(http.StatusOK, "earthquake.html", gin.H{
		"Earthquakes":  earthquakes.Earthquakes,
		"Metadata":     earthquakes.Metadata,
		"FeedType":     feedType,
		"Location":     enhanced.ShortName,
		"LocationData": locationData,
		"Radius":       radius,
		"CenterLat":    enhanced.Latitude,
		"CenterLon":    enhanced.Longitude,
		"HostInfo":     hostInfo,
	})
}

// HandleEarthquakeAPI serves JSON earthquake data
func (h *EarthquakeHandler) HandleEarthquakeAPI(c *gin.Context) {
	feedType := c.DefaultQuery("feed", "all_day")

	earthquakes, err := h.earthquakeService.GetEarthquakes(feedType)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, earthquakes)
}

// HandleEarthquakeDetail serves detailed information for a specific earthquake
func (h *EarthquakeHandler) HandleEarthquakeDetail(c *gin.Context) {
	earthquakeID := c.Param("id")
	if earthquakeID == "" {
		c.String(http.StatusBadRequest, "Earthquake ID required\n")
		return
	}

	// Fetch all earthquakes and find the one with matching ID
	feedType := c.DefaultQuery("feed", "all_week")
	earthquakes, err := h.earthquakeService.GetEarthquakes(feedType)
	if err != nil {
		if utils.IsBrowser(c) {
			c.HTML(http.StatusInternalServerError, "error.html", gin.H{
				"error": "Failed to load earthquake data: " + err.Error(),
			})
		} else {
			c.String(http.StatusInternalServerError, "Error fetching earthquake data: %v\n", err)
		}
		return
	}

	// Find earthquake by ID
	var earthquake *services.Earthquake
	for i := range earthquakes.Earthquakes {
		if earthquakes.Earthquakes[i].ID == earthquakeID {
			earthquake = &earthquakes.Earthquakes[i]
			break
		}
	}

	if earthquake == nil {
		if utils.IsBrowser(c) {
			c.HTML(http.StatusNotFound, "error.html", gin.H{
				"error": "Earthquake not found",
			})
		} else {
			c.String(http.StatusNotFound, "Earthquake not found\n")
		}
		return
	}

	// Check format parameter
	format := c.DefaultQuery("format", "")
	isBrowser := utils.IsBrowser(c)

	if !isBrowser || format != "" {
		// Console output
		output := h.renderASCIIEarthquakeDetail(earthquake)
		c.String(http.StatusOK, output)
	} else {
		// Browser output
		hostInfo := utils.GetHostInfo(c)
		c.HTML(http.StatusOK, "earthquake_detail.html", gin.H{
			"Earthquake": earthquake,
			"HostInfo":   hostInfo,
		})
	}
}

// HandleEarthquakeRequest routes earthquake requests (console vs browser)
func (h *EarthquakeHandler) HandleEarthquakeRequest(c *gin.Context) {
	// Check if services are initialized
	if !IsInitialized() {
		ServeLoadingPage(c)
		return
	}

	// Extract location from path
	location := c.Param("location")
	if location != "" {
		location = strings.TrimPrefix(location, "/")
	}

	// Check if this is a detail request
	if strings.HasPrefix(location, "detail/") {
		earthquakeID := strings.TrimPrefix(location, "detail/")
		c.Params = []gin.Param{{Key: "id", Value: earthquakeID}}
		h.HandleEarthquakeDetail(c)
		return
	}

	isBrowser := utils.IsBrowser(c)

	if isBrowser {
		// Browser users get HTML interface
		if location == "" {
			h.HandleEarthquakes(c)
		} else {
			// Set the param for HandleEarthquakesByLocation
			c.Params = []gin.Param{{Key: "location", Value: location}}
			h.HandleEarthquakesByLocation(c)
		}
	} else {
		// Console users get ASCII text
		h.serveASCIIEarthquakes(c, location)
	}
}

// serveASCIIEarthquakes serves earthquake data as ASCII text for console
func (h *EarthquakeHandler) serveASCIIEarthquakes(c *gin.Context, locationPath string) {
	feedType := c.DefaultQuery("feed", "all_day")
	sortBy := c.DefaultQuery("sort", "newest")
	numberStr := c.DefaultQuery("number", "0")
	number, _ := strconv.Atoi(numberStr)

	var earthquakes *services.EarthquakeCollection
	var err error
	var locationName string

	if locationPath != "" {
		// Get earthquakes near location
		radius := 500.0
		if r := c.Query("radius"); r != "" {
			if parsed, err := strconv.ParseFloat(r, 64); err == nil {
				radius = parsed
			}
		}

		clientIP := utils.GetClientIP(c)
		coords, err := h.weatherService.ParseAndResolveLocation(locationPath, clientIP)
		if err != nil {
			c.String(http.StatusBadRequest, "Location not found: %s\n", locationPath)
			return
		}

		enhanced := h.locationEnhancer.EnhanceLocation(coords)
		locationName = enhanced.ShortName

		earthquakes, err = h.earthquakeService.GetEarthquakesByLocation(
			enhanced.Latitude,
			enhanced.Longitude,
			radius,
			feedType,
		)
	} else {
		earthquakes, err = h.earthquakeService.GetEarthquakes(feedType)
	}

	if err != nil {
		c.String(http.StatusInternalServerError, "Error fetching earthquake data: %v\n", err)
		return
	}

	// Sort and limit results
	earthquakes.SortAndLimit(sortBy, number)

	// Render ASCII output
	output := h.renderASCIIEarthquakes(earthquakes, locationName, feedType)
	c.String(http.StatusOK, output)
}

// renderASCIIEarthquakes formats earthquake data as ASCII text
func (h *EarthquakeHandler) renderASCIIEarthquakes(earthquakes *services.EarthquakeCollection, location, feedType string) string {
	var sb strings.Builder

	// ANSI color codes (Dracula theme)
	cyan := "\x1b[38;2;139;233;253m"
	green := "\x1b[38;2;80;250;123m"
	yellow := "\x1b[38;2;241;250;140m"
	orange := "\x1b[38;2;255;184;108m"
	pink := "\x1b[38;2;255;121;198m"
	purple := "\x1b[38;2;189;147;249m"
	red := "\x1b[38;2;255;85;85m"
	comment := "\x1b[38;2;98;114;164m"
	bold := "\x1b[1m"
	reset := "\x1b[0m"

	// Header
	if location != "" {
		sb.WriteString(fmt.Sprintf("%s%s🌍 Earthquakes near %s%s\n\n", bold, yellow, location, reset))
	} else {
		sb.WriteString(fmt.Sprintf("%s%s🌍 %s%s\n\n", bold, yellow, earthquakes.Metadata.Title, reset))
	}

	sb.WriteString(fmt.Sprintf("%sTotal: %s%d%s earthquakes\n", comment, cyan, earthquakes.Metadata.Count, reset))
	sb.WriteString(fmt.Sprintf("%sUpdated: %s%s\n\n", comment, time.UnixMilli(earthquakes.Metadata.Generated).Format(time.RFC1123), reset))

	if len(earthquakes.Earthquakes) == 0 {
		sb.WriteString(fmt.Sprintf("%sNo earthquakes found.%s\n", comment, reset))
		return sb.String()
	}

	// Table header with box drawing
	sb.WriteString(fmt.Sprintf("%s┌─────────┬────────────────────────────────────────────────────┬──────────────────────┬─────────────┐%s\n",
		purple, reset))
	sb.WriteString(fmt.Sprintf("%s│%s %s%-7s%s %s│%s %s%-50s%s %s│%s %s%-20s%s %s│%s %s%-11s%s %s│%s\n",
		purple, reset, orange, "MAG", reset,
		purple, reset, orange, "LOCATION", reset,
		purple, reset, orange, "TIME", reset,
		purple, reset, orange, "DEPTH", reset,
		purple, reset))
	sb.WriteString(fmt.Sprintf("%s├─────────┼────────────────────────────────────────────────────┼──────────────────────┼─────────────┤%s\n",
		purple, reset))

	// Earthquake list
	for i, eq := range earthquakes.Earthquakes {
		tsunamiWarning := ""
		if eq.Tsunami == 1 {
			tsunamiWarning = " 🌊"
		}

		// Color code by magnitude
		var magColor string
		if eq.Magnitude < 2.0 {
			magColor = green
		} else if eq.Magnitude < 4.0 {
			magColor = yellow
		} else if eq.Magnitude < 5.0 {
			magColor = orange
		} else if eq.Magnitude < 6.0 {
			magColor = pink
		} else {
			magColor = red
		}

		sb.WriteString(fmt.Sprintf("%s│%s %s%-7.1f%s %s│%s %-50s %s│%s %s%-20s%s %s│%s %s%-7.1fkm%s%s %s│%s\n",
			purple, reset,
			magColor, eq.Magnitude, reset,
			purple, reset, truncateString(eq.Place, 50),
			purple, reset,
			cyan, eq.Time.Format("2006-01-02 15:04:05"), reset,
			purple, reset,
			comment, eq.Depth, reset, tsunamiWarning,
			purple, reset))

		// Don't add separator after last item
		if i < len(earthquakes.Earthquakes)-1 {
			sb.WriteString(fmt.Sprintf("%s├─────────┼────────────────────────────────────────────────────┼──────────────────────┼─────────────┤%s\n",
				purple, reset))
		}
	}

	sb.WriteString(fmt.Sprintf("%s└─────────┴────────────────────────────────────────────────────┴──────────────────────┴─────────────┘%s\n",
		purple, reset))

	sb.WriteString(fmt.Sprintf("\n%s🌊 = Tsunami warning%s\n", cyan, reset))
	sb.WriteString(fmt.Sprintf("\n%sView details: https://earthquake.usgs.gov/earthquakes/map/%s\n", comment, reset))

	return sb.String()
}

func truncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}

// renderASCIIEarthquakeDetail renders detailed information for a single earthquake
func (h *EarthquakeHandler) renderASCIIEarthquakeDetail(eq *services.Earthquake) string {
	var sb strings.Builder

	// ANSI color codes (Dracula theme)
	cyan := "\x1b[38;2;139;233;253m"
	green := "\x1b[38;2;80;250;123m"
	yellow := "\x1b[38;2;241;250;140m"
	orange := "\x1b[38;2;255;184;108m"
	pink := "\x1b[38;2;255;121;198m"
	purple := "\x1b[38;2;189;147;249m"
	red := "\x1b[38;2;255;85;85m"
	comment := "\x1b[38;2;98;114;164m"
	bold := "\x1b[1m"
	reset := "\x1b[0m"

	// Magnitude color
	var magColor string
	if eq.Magnitude < 2.0 {
		magColor = green
	} else if eq.Magnitude < 4.0 {
		magColor = yellow
	} else if eq.Magnitude < 5.0 {
		magColor = orange
	} else if eq.Magnitude < 6.0 {
		magColor = pink
	} else {
		magColor = red
	}

	// Header
	sb.WriteString(fmt.Sprintf("%s%s🌍 Earthquake Details%s\n\n", bold, yellow, reset))

	// Main info box
	sb.WriteString(fmt.Sprintf("%s┌─────────────────────────────────────────────────────────────────────┐%s\n", purple, reset))
	sb.WriteString(fmt.Sprintf("%s│%s %s%-67s%s %s│%s\n",
		purple, reset, bold, eq.Place, reset, purple, reset))
	sb.WriteString(fmt.Sprintf("%s├─────────────────────────────────────────────────────────────────────┤%s\n", purple, reset))

	// Magnitude and tsunami
	tsunamiIndicator := ""
	if eq.Tsunami == 1 {
		tsunamiIndicator = fmt.Sprintf("  %s🌊 TSUNAMI WARNING%s", cyan, reset)
	}
	sb.WriteString(fmt.Sprintf("%s│%s %sMagnitude:%s %s%.1f%s %s(%s)%s%s %s│%s\n",
		purple, reset,
		orange, reset,
		magColor, eq.Magnitude, reset,
		comment, eq.MagnitudeType, reset,
		tsunamiIndicator,
		purple, reset))

	// Time
	sb.WriteString(fmt.Sprintf("%s│%s %sTime:%s     %s%s%s %s│%s\n",
		purple, reset,
		orange, reset,
		cyan, eq.Time.Format("2006-01-02 15:04:05 MST"), reset,
		purple, reset))

	// Location
	sb.WriteString(fmt.Sprintf("%s│%s %sLocation:%s  %sLat %.4f°, Lon %.4f°%s %s│%s\n",
		purple, reset,
		orange, reset,
		cyan, eq.Latitude, eq.Longitude, reset,
		purple, reset))

	// Depth
	sb.WriteString(fmt.Sprintf("%s│%s %sDepth:%s     %s%.1f km%s %s│%s\n",
		purple, reset,
		orange, reset,
		cyan, eq.Depth, reset,
		purple, reset))

	// Status
	sb.WriteString(fmt.Sprintf("%s│%s %sStatus:%s    %s%s%s %s│%s\n",
		purple, reset,
		orange, reset,
		green, eq.Status, reset,
		purple, reset))

	// Optional fields
	if eq.Felt != nil && *eq.Felt > 0 {
		sb.WriteString(fmt.Sprintf("%s│%s %sFelt:%s      %s%d reports%s %s│%s\n",
			purple, reset,
			orange, reset,
			pink, *eq.Felt, reset,
			purple, reset))
	}

	if eq.CDI != nil {
		sb.WriteString(fmt.Sprintf("%s│%s %sCDI:%s       %s%.1f%s %s│%s\n",
			purple, reset,
			orange, reset,
			pink, *eq.CDI, reset,
			purple, reset))
	}

	if eq.MMI != nil {
		sb.WriteString(fmt.Sprintf("%s│%s %sMMI:%s       %s%.1f%s %s│%s\n",
			purple, reset,
			orange, reset,
			pink, *eq.MMI, reset,
			purple, reset))
	}

	// Network and ID
	sb.WriteString(fmt.Sprintf("%s│%s %sNetwork:%s   %s%s%s %s│%s\n",
		purple, reset,
		orange, reset,
		comment, eq.Network, reset,
		purple, reset))

	sb.WriteString(fmt.Sprintf("%s│%s %sID:%s        %s%s%s %s│%s\n",
		purple, reset,
		orange, reset,
		comment, eq.ID, reset,
		purple, reset))

	sb.WriteString(fmt.Sprintf("%s└─────────────────────────────────────────────────────────────────────┘%s\n", purple, reset))

	// External link
	sb.WriteString(fmt.Sprintf("\n%sUSGS Details: %s%s%s\n", comment, cyan, eq.URL, reset))

	return sb.String()
}
