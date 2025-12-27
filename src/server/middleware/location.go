package middleware

import (
	"fmt"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

// SaveLocationCookies saves user location to persistent cookies
func SaveLocationCookies(c *gin.Context, latitude, longitude float64, locationName string) {
	// Cookie settings: 30 days expiration, accessible to all paths
	// 30 days in seconds
	maxAge := 30 * 24 * 60 * 60

	c.SetCookie("user_lat", fmt.Sprintf("%.6f", latitude), maxAge, "/", "", false, false)
	c.SetCookie("user_lon", fmt.Sprintf("%.6f", longitude), maxAge, "/", "", false, false)
	c.SetCookie("user_location_name", locationName, maxAge, "/", "", false, false)
	c.SetCookie("user_location_set_at", time.Now().Format(time.RFC3339), maxAge, "/", "", false, false)
}

// GetLocationFromCookies retrieves user location from cookies
func GetLocationFromCookies(c *gin.Context) (latitude, longitude float64, locationName string, found bool) {
	latStr, err1 := c.Cookie("user_lat")
	lonStr, err2 := c.Cookie("user_lon")
	locName, err3 := c.Cookie("user_location_name")

	if err1 != nil || err2 != nil || err3 != nil {
		return 0, 0, "", false
	}

	lat, err := strconv.ParseFloat(latStr, 64)
	if err != nil {
		return 0, 0, "", false
	}

	lon, err := strconv.ParseFloat(lonStr, 64)
	if err != nil {
		return 0, 0, "", false
	}

	return lat, lon, locName, true
}

// ClearLocationCookies removes location cookies
func ClearLocationCookies(c *gin.Context) {
	c.SetCookie("user_lat", "", -1, "/", "", false, false)
	c.SetCookie("user_lon", "", -1, "/", "", false, false)
	c.SetCookie("user_location_name", "", -1, "/", "", false, false)
	c.SetCookie("user_location_set_at", "", -1, "/", "", false, false)
}
