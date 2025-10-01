package handlers

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"weather-go/services"
	"weather-go/utils"
)

// APIHandler handles JSON API routes (/api/v1/*)
type APIHandler struct {
	weatherService   *services.WeatherService
	locationEnhancer *services.LocationEnhancer
}

// NewAPIHandler creates a new API handler
func NewAPIHandler(ws *services.WeatherService, le *services.LocationEnhancer) *APIHandler {
	return &APIHandler{
		weatherService:   ws,
		locationEnhancer: le,
	}
}

// GetWeather returns current weather with today's forecast (GET /api/v1/weather)
func (h *APIHandler) GetWeather(c *gin.Context) {
	location := c.Query("location")
	lat := c.Query("lat")
	lon := c.Query("lon")
	unitsParam := c.Query("units")

	var coords *services.Coordinates
	var err error

	// Parse coordinates or location
	if lat != "" && lon != "" {
		latitude, _ := strconv.ParseFloat(lat, 64)
		longitude, _ := strconv.ParseFloat(lon, 64)
		coords, err = h.weatherService.GetCoordinates(fmt.Sprintf("%f,%f", latitude, longitude), "")
	} else if location != "" {
		clientIP := utils.GetClientIP(c)
		coords, err = h.weatherService.ParseAndResolveLocation(location, clientIP)
	} else {
		// IP-based location detection
		clientIP := utils.GetClientIP(c)
		coords, err = h.weatherService.GetCoordinatesFromIP(clientIP)
	}

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": gin.H{
				"code":    "WEATHER_ERROR",
				"message": err.Error(),
			},
		})
		return
	}

	// Enhance location
	enhanced := h.locationEnhancer.EnhanceLocation(coords)

	// Determine units
	units := "imperial"
	if unitsParam != "" {
		units = unitsParam
	} else if enhanced.CountryCode != "US" {
		units = "metric"
	}

	// Get current weather
	current, err := h.weatherService.GetCurrentWeather(enhanced.Latitude, enhanced.Longitude, units)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": gin.H{
				"code":    "WEATHER_ERROR",
				"message": err.Error(),
			},
		})
		return
	}

	// Get today's forecast
	forecast, err := h.weatherService.GetForecast(enhanced.Latitude, enhanced.Longitude, 1, units)
	if err != nil {
		forecast = &services.Forecast{Days: []services.ForecastDay{}}
	}

	// Build response
	response := gin.H{
		"location": gin.H{
			"name":        enhanced.Name,
			"shortName":   enhanced.ShortName,
			"fullName":    enhanced.FullName,
			"country":     enhanced.Country,
			"countryCode": enhanced.CountryCode,
			"latitude":    enhanced.Latitude,
			"longitude":   enhanced.Longitude,
			"timezone":    enhanced.Timezone,
		},
		"current": gin.H{
			"temperature":   current.Temperature,
			"feelsLike":     current.FeelsLike,
			"humidity":      current.Humidity,
			"pressure":      current.Pressure,
			"windSpeed":     current.WindSpeed,
			"windDirection": current.WindDirection,
			"windGusts":     current.WindGusts,
			"precipitation": current.Precipitation,
			"cloudCover":    current.CloudCover,
			"weatherCode":   current.WeatherCode,
			"description":   h.weatherService.GetWeatherDescription(current.WeatherCode),
			"icon":          h.weatherService.GetWeatherIcon(current.WeatherCode, current.IsDay == 1),
			"isDay":         current.IsDay,
		},
		"meta": gin.H{
			"source":      "Open-Meteo",
			"timestamp":   utils.Now(),
			"api_version": "1.0",
			"units":       units,
		},
	}

	// Add today forecast if available
	if len(forecast.Days) > 0 {
		day := forecast.Days[0]
		response["today"] = gin.H{
			"date":        day.Date,
			"weatherCode": day.WeatherCode,
			"description": h.weatherService.GetWeatherDescription(day.WeatherCode),
			"icon":        h.weatherService.GetWeatherIcon(day.WeatherCode, true),
			"temperature": gin.H{
				"min": day.TempMin,
				"max": day.TempMax,
			},
			"precipitation": gin.H{
				"sum":         day.Precipitation,
				"probability": day.PrecipitationProbability,
			},
		}
	}

	c.JSON(http.StatusOK, response)
}

// GetWeatherByLocation returns weather for specific location (GET /api/v1/weather/:location)
func (h *APIHandler) GetWeatherByLocation(c *gin.Context) {
	location := c.Param("location")
	unitsParam := c.Query("units")

	clientIP := utils.GetClientIP(c)
	coords, err := h.weatherService.ParseAndResolveLocation(location, clientIP)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": gin.H{
				"code":    "WEATHER_ERROR",
				"message": err.Error(),
			},
		})
		return
	}

	// Enhance location
	enhanced := h.locationEnhancer.EnhanceLocation(coords)

	// Determine units
	units := "imperial"
	if unitsParam != "" {
		units = unitsParam
	} else if enhanced.CountryCode != "US" {
		units = "metric"
	}

	// Get current weather
	current, err := h.weatherService.GetCurrentWeather(enhanced.Latitude, enhanced.Longitude, units)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": gin.H{
				"code":    "WEATHER_ERROR",
				"message": err.Error(),
			},
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"location": gin.H{
			"name":        enhanced.Name,
			"shortName":   enhanced.ShortName,
			"fullName":    enhanced.FullName,
			"country":     enhanced.Country,
			"countryCode": enhanced.CountryCode,
			"latitude":    enhanced.Latitude,
			"longitude":   enhanced.Longitude,
			"timezone":    enhanced.Timezone,
		},
		"current": gin.H{
			"temperature":   current.Temperature,
			"feelsLike":     current.FeelsLike,
			"humidity":      current.Humidity,
			"pressure":      current.Pressure,
			"windSpeed":     current.WindSpeed,
			"windDirection": current.WindDirection,
			"windGusts":     current.WindGusts,
			"precipitation": current.Precipitation,
			"cloudCover":    current.CloudCover,
			"weatherCode":   current.WeatherCode,
			"description":   h.weatherService.GetWeatherDescription(current.WeatherCode),
			"icon":          h.weatherService.GetWeatherIcon(current.WeatherCode, current.IsDay == 1),
			"isDay":         current.IsDay,
		},
		"meta": gin.H{
			"source":      "Open-Meteo",
			"timestamp":   utils.Now(),
			"api_version": "1.0",
			"units":       units,
		},
	})
}

// GetForecast returns multi-day forecast (GET /api/v1/forecast)
func (h *APIHandler) GetForecast(c *gin.Context) {
	location := c.Query("location")
	lat := c.Query("lat")
	lon := c.Query("lon")
	daysParam := c.Query("days")
	unitsParam := c.Query("units")

	// Parse days
	days := 7
	if daysParam != "" {
		if d, err := strconv.Atoi(daysParam); err == nil && d > 0 && d <= 16 {
			days = d
		}
	}

	var coords *services.Coordinates
	var err error

	// Parse coordinates or location
	if lat != "" && lon != "" {
		latitude, _ := strconv.ParseFloat(lat, 64)
		longitude, _ := strconv.ParseFloat(lon, 64)
		coords, err = h.weatherService.GetCoordinates(fmt.Sprintf("%f,%f", latitude, longitude), "")
	} else if location != "" {
		clientIP := utils.GetClientIP(c)
		coords, err = h.weatherService.ParseAndResolveLocation(location, clientIP)
	} else {
		clientIP := utils.GetClientIP(c)
		coords, err = h.weatherService.GetCoordinatesFromIP(clientIP)
	}

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": gin.H{
				"code":    "FORECAST_ERROR",
				"message": err.Error(),
			},
		})
		return
	}

	// Enhance location
	enhanced := h.locationEnhancer.EnhanceLocation(coords)

	// Determine units
	units := "imperial"
	if unitsParam != "" {
		units = unitsParam
	} else if enhanced.CountryCode != "US" {
		units = "metric"
	}

	// Get forecast
	forecast, err := h.weatherService.GetForecast(enhanced.Latitude, enhanced.Longitude, days, units)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": gin.H{
				"code":    "FORECAST_ERROR",
				"message": err.Error(),
			},
		})
		return
	}

	// Build forecast response
	forecastDays := make([]gin.H, len(forecast.Days))
	for i, day := range forecast.Days {
		forecastDays[i] = gin.H{
			"date":        day.Date,
			"weatherCode": day.WeatherCode,
			"description": h.weatherService.GetWeatherDescription(day.WeatherCode),
			"icon":        h.weatherService.GetWeatherIcon(day.WeatherCode, true),
			"temperature": gin.H{
				"min": day.TempMin,
				"max": day.TempMax,
			},
			"feelsLike": gin.H{
				"min": day.FeelsLikeMin,
				"max": day.FeelsLikeMax,
			},
			"precipitation": gin.H{
				"sum":         day.Precipitation,
				"hours":       day.PrecipitationHours,
				"probability": day.PrecipitationProbability,
			},
			"wind": gin.H{
				"speedMax":  day.WindSpeedMax,
				"gustsMax":  day.WindGustsMax,
				"direction": day.WindDirection,
			},
			"solarRadiation": day.SolarRadiation,
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"location": gin.H{
			"name":        enhanced.Name,
			"shortName":   enhanced.ShortName,
			"fullName":    enhanced.FullName,
			"country":     enhanced.Country,
			"countryCode": enhanced.CountryCode,
			"latitude":    enhanced.Latitude,
			"longitude":   enhanced.Longitude,
			"timezone":    enhanced.Timezone,
		},
		"forecast": gin.H{
			"days": forecastDays,
		},
		"meta": gin.H{
			"source":      "Open-Meteo",
			"timestamp":   utils.Now(),
			"api_version": "1.0",
			"units":       units,
		},
	})
}

// GetForecastByLocation returns forecast for specific location (GET /api/v1/forecast/:location)
func (h *APIHandler) GetForecastByLocation(c *gin.Context) {
	location := c.Param("location")
	daysParam := c.Query("days")
	unitsParam := c.Query("units")

	// Parse days
	days := 7
	if daysParam != "" {
		if d, err := strconv.Atoi(daysParam); err == nil && d > 0 && d <= 16 {
			days = d
		}
	}

	clientIP := utils.GetClientIP(c)
	coords, err := h.weatherService.ParseAndResolveLocation(location, clientIP)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": gin.H{
				"code":    "FORECAST_ERROR",
				"message": err.Error(),
			},
		})
		return
	}

	// Enhance location
	enhanced := h.locationEnhancer.EnhanceLocation(coords)

	// Determine units
	units := "imperial"
	if unitsParam != "" {
		units = unitsParam
	} else if enhanced.CountryCode != "US" {
		units = "metric"
	}

	// Get forecast
	forecast, err := h.weatherService.GetForecast(enhanced.Latitude, enhanced.Longitude, days, units)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": gin.H{
				"code":    "FORECAST_ERROR",
				"message": err.Error(),
			},
		})
		return
	}

	// Build forecast response
	forecastDays := make([]gin.H, len(forecast.Days))
	for i, day := range forecast.Days {
		forecastDays[i] = gin.H{
			"date":        day.Date,
			"weatherCode": day.WeatherCode,
			"description": h.weatherService.GetWeatherDescription(day.WeatherCode),
			"icon":        h.weatherService.GetWeatherIcon(day.WeatherCode, true),
			"temperature": gin.H{
				"min": day.TempMin,
				"max": day.TempMax,
			},
			"feelsLike": gin.H{
				"min": day.FeelsLikeMin,
				"max": day.FeelsLikeMax,
			},
			"precipitation": gin.H{
				"sum":         day.Precipitation,
				"hours":       day.PrecipitationHours,
				"probability": day.PrecipitationProbability,
			},
			"wind": gin.H{
				"speedMax":  day.WindSpeedMax,
				"gustsMax":  day.WindGustsMax,
				"direction": day.WindDirection,
			},
			"solarRadiation": day.SolarRadiation,
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"location": gin.H{
			"name":        enhanced.Name,
			"shortName":   enhanced.ShortName,
			"fullName":    enhanced.FullName,
			"country":     enhanced.Country,
			"countryCode": enhanced.CountryCode,
			"latitude":    enhanced.Latitude,
			"longitude":   enhanced.Longitude,
			"timezone":    enhanced.Timezone,
		},
		"forecast": gin.H{
			"days": forecastDays,
		},
		"meta": gin.H{
			"source":      "Open-Meteo",
			"timestamp":   utils.Now(),
			"api_version": "1.0",
			"units":       units,
		},
	})
}

// SearchLocations searches for locations (GET /api/v1/search)
func (h *APIHandler) SearchLocations(c *gin.Context) {
	query := c.Query("q")

	if query == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": gin.H{
				"code":    "MISSING_QUERY",
				"message": "Query parameter 'q' is required",
			},
		})
		return
	}

	// Search for locations
	results, err := h.weatherService.SearchLocations(query, 10)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": gin.H{
				"code":    "SEARCH_ERROR",
				"message": err.Error(),
			},
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"query":   query,
		"results": results,
		"meta": gin.H{
			"source":      "Open-Meteo Geocoding",
			"timestamp":   utils.Now(),
			"api_version": "1.0",
		},
	})
}

// GetIP returns client IP information (GET /api/v1/ip)
func (h *APIHandler) GetIP(c *gin.Context) {
	clientIP := utils.GetClientIP(c)

	c.JSON(http.StatusOK, gin.H{
		"ip":        clientIP,
		"timestamp": utils.Now(),
		"headers": gin.H{
			"x-forwarded-for": c.GetHeader("X-Forwarded-For"),
			"x-real-ip":       c.GetHeader("X-Real-IP"),
			"cf-connecting-ip": c.GetHeader("CF-Connecting-IP"),
		},
	})
}

// GetLocation returns detected location from IP (GET /api/v1/location)
func (h *APIHandler) GetLocation(c *gin.Context) {
	clientIP := utils.GetClientIP(c)

	coords, err := h.weatherService.GetCoordinatesFromIP(clientIP)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": gin.H{
				"code":    "LOCATION_ERROR",
				"message": err.Error(),
			},
		})
		return
	}

	// Enhance location
	enhanced := h.locationEnhancer.EnhanceLocation(coords)

	// Determine units based on country
	units := "metric"
	if enhanced.CountryCode == "US" {
		units = "imperial"
	}

	c.JSON(http.StatusOK, gin.H{
		"ip": clientIP,
		"location": gin.H{
			"city":        enhanced.Name,
			"shortName":   enhanced.ShortName,
			"fullName":    enhanced.FullName,
			"country":     enhanced.Country,
			"countryCode": enhanced.CountryCode,
			"state":       enhanced.Admin1,
			"units":       units,
			"coordinates": gin.H{
				"latitude":  enhanced.Latitude,
				"longitude": enhanced.Longitude,
			},
			"population": enhanced.Population,
			"timezone":   enhanced.Timezone,
		},
		"timestamp": utils.Now(),
		"source":    "IP Geolocation with Coordinates",
	})
}

// GetDocs returns API documentation (GET /api/v1/docs)
func (h *APIHandler) GetDocs(c *gin.Context) {
	hostInfo := utils.GetHostInfo(c)
	baseURL := hostInfo.FullHost

	c.JSON(http.StatusOK, gin.H{
		"api":         "Weather API v1",
		"version":     "1.0.0",
		"description": "A free weather API using open data sources",
		"base_url":    baseURL,
		"endpoints": gin.H{
			"GET /api/v1/weather": gin.H{
				"description": "Get current weather with today forecast",
				"parameters": gin.H{
					"location": "Location name (optional)",
					"lat":      "Latitude (optional, use with lon)",
					"lon":      "Longitude (optional, use with lat)",
					"units":    "Unit system: imperial or metric (auto-detected by location if not specified)",
				},
				"example": fmt.Sprintf("%s/api/v1/weather?location=paris", baseURL),
			},
			"GET /api/v1/forecast": gin.H{
				"description": "Get weather forecast",
				"parameters": gin.H{
					"location": "Location name (optional)",
					"lat":      "Latitude (optional, use with lon)",
					"lon":      "Longitude (optional, use with lat)",
					"days":     "Number of forecast days (1-16, default: 7)",
					"units":    "Unit system: imperial or metric (auto-detected by location if not specified)",
				},
				"example": fmt.Sprintf("%s/api/v1/forecast?location=london&days=5", baseURL),
			},
			"GET /api/v1/search": gin.H{
				"description": "Search for locations",
				"parameters": gin.H{
					"q": "Search query (required)",
				},
				"example": fmt.Sprintf("%s/api/v1/search?q=alb", baseURL),
			},
			"GET /api/v1/ip": gin.H{
				"description": "Get client IP information",
				"parameters":  gin.H{},
				"example":     fmt.Sprintf("%s/api/v1/ip", baseURL),
			},
			"GET /api/v1/location": gin.H{
				"description": "Get detected location from IP",
				"parameters":  gin.H{},
				"example":     fmt.Sprintf("%s/api/v1/location", baseURL),
			},
		},
		"console_endpoints": gin.H{
			"GET /": gin.H{
				"description": "Console weather for current location (IP-based detection)",
				"parameters": gin.H{
					"format": "Display format (0-4)",
					"units":  "Unit system: imperial or metric",
					"F":      "Hide footer",
				},
				"example": fmt.Sprintf("curl -q -LSs %s/", baseURL),
			},
			"GET /:location": gin.H{
				"description": "Console weather for specific location",
				"parameters": gin.H{
					"format": "Display format (0-4)",
					"units":  "Unit system: imperial or metric",
					"F":      "Hide footer",
				},
				"example": fmt.Sprintf("curl -q -LSs %s/London,GB", baseURL),
			},
		},
		"data_source": "Open-Meteo.com",
		"features": []string{
			"No API key required",
			"Free for non-commercial use",
			"Global weather data",
			"Up to 16-day forecasts",
			"Imperial and metric units",
			"Automatic location and unit detection from IP",
			"Dynamic host and protocol detection",
			"Caching for better performance",
			"Console-friendly ASCII output",
			"JSON API responses",
		},
	})
}
