package handlers

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"weather-go/src/services"
	"weather-go/src/utils"
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
	cityID := c.Query("city_id")
	nearest := c.Query("nearest")
	unitsParam := c.Query("units")

	var coords *services.Coordinates
	var enhanced *services.EnhancedLocation
	var err error

	// Priority: city_id > lat/lon > nearest > location > IP
	if cityID != "" {
		// Find city by ID
		id, parseErr := strconv.Atoi(cityID)
		if parseErr != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": gin.H{
					"code":    "INVALID_CITY_ID",
					"message": "Invalid city ID format",
				},
			})
			return
		}
		enhanced, err = h.locationEnhancer.FindCityByID(id)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{
				"error": gin.H{
					"code":    "CITY_NOT_FOUND",
					"message": fmt.Sprintf("City ID %d not found", id),
				},
			})
			return
		}
	} else if lat != "" && lon != "" {
		// Use coordinates
		latitude, _ := strconv.ParseFloat(lat, 64)
		longitude, _ := strconv.ParseFloat(lon, 64)

		if nearest == "true" || nearest == "1" {
			// Find nearest city to coordinates
			enhanced, err = h.locationEnhancer.FindNearestCity(latitude, longitude)
			if err != nil {
				// Fallback to coordinate-based location
				coords, err = h.weatherService.GetCoordinates(fmt.Sprintf("%f,%f", latitude, longitude), "")
				if err != nil {
					c.JSON(http.StatusBadRequest, gin.H{
						"error": gin.H{
							"code":    "LOCATION_ERROR",
							"message": err.Error(),
						},
					})
					return
				}
				// Convert to GeocodeResult and enhance
				geocodeResult := &services.GeocodeResult{
					Latitude:    coords.Latitude,
					Longitude:   coords.Longitude,
					Name:        coords.Name,
					Country:     coords.Country,
					CountryCode: coords.CountryCode,
					Admin1:      coords.Admin1,
					Admin2:      coords.Admin2,
					Timezone:    coords.Timezone,
					Population:  coords.Population,
				}
				enhanced, _ = h.locationEnhancer.EnhanceLocationData(geocodeResult)
			}
		} else {
			coords, err = h.weatherService.GetCoordinates(fmt.Sprintf("%f,%f", latitude, longitude), "")
			if err != nil {
				c.JSON(http.StatusBadRequest, gin.H{
					"error": gin.H{
						"code":    "LOCATION_ERROR",
						"message": err.Error(),
					},
				})
				return
			}
			// Convert to GeocodeResult and enhance
			geocodeResult := &services.GeocodeResult{
				Latitude:    coords.Latitude,
				Longitude:   coords.Longitude,
				Name:        coords.Name,
				Country:     coords.Country,
				CountryCode: coords.CountryCode,
				Admin1:      coords.Admin1,
				Admin2:      coords.Admin2,
				Timezone:    coords.Timezone,
				Population:  coords.Population,
			}
			enhanced, _ = h.locationEnhancer.EnhanceLocationData(geocodeResult)
		}
	} else if location != "" {
		clientIP := utils.GetClientIP(c)
		coords, err = h.weatherService.ParseAndResolveLocation(location, clientIP)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": gin.H{
					"code":    "LOCATION_ERROR",
					"message": err.Error(),
				},
			})
			return
		}
		// Convert to GeocodeResult and enhance
		geocodeResult := &services.GeocodeResult{
			Latitude:    coords.Latitude,
			Longitude:   coords.Longitude,
			Name:        coords.Name,
			Country:     coords.Country,
			CountryCode: coords.CountryCode,
			Admin1:      coords.Admin1,
			Admin2:      coords.Admin2,
			Timezone:    coords.Timezone,
			Population:  coords.Population,
		}
		enhanced, _ = h.locationEnhancer.EnhanceLocationData(geocodeResult)
	} else {
		// IP-based location detection
		clientIP := utils.GetClientIP(c)
		coords, err = h.weatherService.GetCoordinatesFromIP(clientIP)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": gin.H{
					"code":    "LOCATION_ERROR",
					"message": err.Error(),
				},
			})
			return
		}
		// Convert to GeocodeResult and enhance
		geocodeResult := &services.GeocodeResult{
			Latitude:    coords.Latitude,
			Longitude:   coords.Longitude,
			Name:        coords.Name,
			Country:     coords.Country,
			CountryCode: coords.CountryCode,
			Admin1:      coords.Admin1,
			Admin2:      coords.Admin2,
			Timezone:    coords.Timezone,
			Population:  coords.Population,
		}
		enhanced, _ = h.locationEnhancer.EnhanceLocationData(geocodeResult)
	}

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

// GetDocsJSON returns API documentation in JSON format (GET /api/v1/docs)
func (h *APIHandler) GetDocsJSON(c *gin.Context) {
	hostInfo := utils.GetHostInfo(c)

	c.JSON(http.StatusOK, gin.H{
		"service":     "Weather API",
		"version":     "2.0.0",
		"description": "Free weather API with no API key required",
		"base_url":    hostInfo.FullHost,
		"endpoints": gin.H{
			"weather": gin.H{
				"GET /api/v1/weather":           "Get weather for current location (IP-based)",
				"GET /api/v1/weather/:location": "Get weather for specific location",
			},
			"forecast": gin.H{
				"GET /api/v1/forecast":           "Get forecast for current location (IP-based)",
				"GET /api/v1/forecast/:location": "Get forecast for specific location",
			},
			"location": gin.H{
				"GET /api/v1/location": "Get current location from IP",
				"GET /api/v1/search":   "Search for locations by name",
			},
			"utility": gin.H{
				"GET /api/v1/ip":   "Get client IP address",
				"GET /api/v1/docs": "API documentation (JSON)",
			},
		},
		"parameters": gin.H{
			"location": "City name, coordinates (lat,lon), or zipcode",
			"units":    "imperial (°F, mph) or metric (°C, km/h). Default: imperial",
			"days":     "Number of forecast days (1-7). Default: 7",
			"q":        "Search query for location search",
		},
		"examples": []string{
			hostInfo.FullHost + "/api/v1/weather?location=London",
			hostInfo.FullHost + "/api/v1/forecast?location=Paris&days=5",
			hostInfo.FullHost + "/api/v1/search?q=New+York",
			hostInfo.FullHost + "/api/v1/ip",
		},
	})
}

// GetDocsHTML returns API documentation as HTML page (GET /docs)
func (h *APIHandler) GetDocsHTML(c *gin.Context) {
	hostInfo := utils.GetHostInfo(c)

	c.HTML(http.StatusOK, "api-docs.html", gin.H{
		"Title":      "API Documentation - Weather",
		"HostInfo":   hostInfo,
		"HideFooter": false,
	})
}
