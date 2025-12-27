package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/apimgr/weather/src/server/service"
	"github.com/apimgr/weather/src/utils"
)

// MoonHandler handles moon phase API requests
type MoonHandler struct {
	weatherService   *services.WeatherService
	locationEnhancer *services.LocationEnhancer
}

// NewMoonHandler creates a new moon handler
func NewMoonHandler(ws *services.WeatherService, le *services.LocationEnhancer) *MoonHandler {
	return &MoonHandler{
		weatherService:   ws,
		locationEnhancer: le,
	}
}

// HandleMoonAPI handles GET /api/v1/moon
func (h *MoonHandler) HandleMoonAPI(c *gin.Context) {
	// Get location parameter
	location := c.Query("location")
	lat := c.Query("lat")
	lon := c.Query("lon")
	clientIP := utils.GetClientIP(c)

	var coords *services.Coordinates
	var err error

	// Priority: lat/lon > location > cookie > IP geolocation
	if lat != "" && lon != "" {
		latFloat, err1 := strconv.ParseFloat(lat, 64)
		lonFloat, err2 := strconv.ParseFloat(lon, 64)
		if err1 != nil || err2 != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "Invalid coordinates format",
			})
			return
		}
		coords = &services.Coordinates{
			Latitude:  latFloat,
			Longitude: lonFloat,
		}
	} else if location != "" {
		coords, err = h.weatherService.ParseAndResolveLocation(location, clientIP)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": err.Error(),
			})
			return
		}
	} else {
		// Try cookies first
		if latStr, err := c.Cookie("user_lat"); err == nil {
			if lonStr, err := c.Cookie("user_lon"); err == nil {
				if latFloat, err1 := strconv.ParseFloat(latStr, 64); err1 == nil {
					if lonFloat, err2 := strconv.ParseFloat(lonStr, 64); err2 == nil {
						coords = &services.Coordinates{
							Latitude:  latFloat,
							Longitude: lonFloat,
						}
					}
				}
			}
		}

		// Fallback to IP geolocation
		if coords == nil {
			coords, err = h.weatherService.GetCoordinatesFromIP(clientIP)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{
					"error": "Could not determine location",
				})
				return
			}
		}
	}

	// Enhance location
	enhanced := h.locationEnhancer.EnhanceLocation(coords)

	// TODO: Implement real moon phase calculation
	// For now, return placeholder data matching the web interface
	moonData := gin.H{
		"location": gin.H{
			"name":        enhanced.Name,
			"shortName":   enhanced.ShortName,
			"fullName":    enhanced.FullName,
			"latitude":    enhanced.Latitude,
			"longitude":   enhanced.Longitude,
			"timezone":    enhanced.Timezone,
			"country":     enhanced.Country,
			"countryCode": enhanced.CountryCode,
		},
		"moon": gin.H{
			"phase":        "First Quarter",
			"illumination": 50.0,
			"icon":         "🌓",
			"age":          7.0,
			"rise":         "12:00",
			"set":          "00:00",
		},
		"sun": gin.H{
			"sunrise":   "06:30",
			"sunset":    "18:30",
			"solarNoon": "12:30",
			"dayLength": "12h 0m",
		},
	}

	c.JSON(http.StatusOK, moonData)
}
