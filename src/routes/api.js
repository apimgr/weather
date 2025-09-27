const express = require('express');
const weatherService = require('../services/weatherService');
const locationParser = require('../utils/locationParser');
const hostDetector = require('../utils/hostDetector');

const router = express.Router();

// Weather endpoint with path parameter support
router.get('/weather/:location(*)', async (req, res) => {
  try {
    const locationFromPath = req.params.location;
    const { units } = req.query;
    
    // Decode location from path
    const location = decodeURIComponent(locationFromPath).replace(/\+/g, ' ');
    
    // Use same logic as query parameter route
    let selectedUnits = units || 'imperial';
    let coords;
    
    const clientIp = locationParser.getClientIP(req);
    const parsedLocation = locationParser.parseLocation(location, clientIp);
    
    if (parsedLocation.type === 'coordinates') {
      // For coordinates, use the weatherService to get proper location data with reverse geocoding
      coords = await weatherService.getCoordinates(`${parsedLocation.value.latitude},${parsedLocation.value.longitude}`);
    } else {
      coords = await weatherService.getCoordinates(parsedLocation.value);
      if (!units && coords.country) {
        selectedUnits = locationParser.getUnitsForCountry(coords.country);
      }
    }

    const currentWeather = await weatherService.getCurrentWeather(coords.latitude, coords.longitude, selectedUnits);
    
    const response = {
      location: {
        name: coords.name,
        country: coords.country,
        latitude: coords.latitude,
        longitude: coords.longitude,
        timezone: coords.timezone || currentWeather.timezone
      },
      current: {
        temperature: currentWeather.temperature,
        feelsLike: currentWeather.feelsLike,
        humidity: currentWeather.humidity,
        pressure: currentWeather.pressure,
        windSpeed: currentWeather.windSpeed,
        windDirection: currentWeather.windDirection,
        windGusts: currentWeather.windGusts,
        precipitation: currentWeather.precipitation,
        cloudCover: currentWeather.cloudCover,
        weatherCode: currentWeather.weatherCode,
        description: weatherService.getWeatherDescription(currentWeather.weatherCode),
        icon: weatherService.getWeatherIcon(currentWeather.weatherCode, currentWeather.isDay),
        isDay: currentWeather.isDay,
        units: currentWeather.units
      },
      meta: {
        source: 'Open-Meteo',
        timestamp: new Date().toISOString(),
        api_version: '1.0',
        units: selectedUnits
      }
    };

    res.json(response);

  } catch (error) {
    res.status(400).json({
      error: {
        code: 'WEATHER_ERROR',
        message: error.message
      }
    });
  }
});

router.get('/weather', async (req, res) => {
  try {
    const { location, lat, lon, format = 'json', units } = req.query;
    
    let coords;
    let selectedUnits = units || 'imperial';
    
    if (lat && lon) {
      coords = {
        latitude: parseFloat(lat),
        longitude: parseFloat(lon),
        name: `${lat}, ${lon}`,
        country: ''
      };
    } else {
      const clientIp = locationParser.getClientIP(req);
      const parsedLocation = locationParser.parseLocation(location, clientIp);
      
      if (parsedLocation.type === 'coordinates') {
        coords = parsedLocation.value;
        coords.name = `${coords.latitude}, ${coords.longitude}`;
        coords.country = '';
      } else {
        coords = await weatherService.getCoordinates(parsedLocation.value);
        if (!units && coords.country) {
          selectedUnits = locationParser.getUnitsForCountry(coords.country);
        }
      }
    }

    const currentWeather = await weatherService.getCurrentWeather(coords.latitude, coords.longitude, selectedUnits);
    
    const response = {
      location: {
        name: coords.name,
        country: coords.country,
        latitude: coords.latitude,
        longitude: coords.longitude,
        timezone: coords.timezone || currentWeather.timezone
      },
      current: {
        temperature: currentWeather.temperature,
        feelsLike: currentWeather.feelsLike,
        humidity: currentWeather.humidity,
        pressure: currentWeather.pressure,
        windSpeed: currentWeather.windSpeed,
        windDirection: currentWeather.windDirection,
        windGusts: currentWeather.windGusts,
        precipitation: currentWeather.precipitation,
        cloudCover: currentWeather.cloudCover,
        weatherCode: currentWeather.weatherCode,
        description: weatherService.getWeatherDescription(currentWeather.weatherCode),
        icon: weatherService.getWeatherIcon(currentWeather.weatherCode, currentWeather.isDay),
        isDay: currentWeather.isDay,
        units: currentWeather.units
      },
      meta: {
        source: 'Open-Meteo',
        timestamp: new Date().toISOString(),
        api_version: '1.0',
        units: selectedUnits
      }
    };

    res.json(response);

  } catch (error) {
    res.status(400).json({
      error: {
        code: 'WEATHER_ERROR',
        message: error.message
      }
    });
  }
});

// Forecast endpoint with path parameter support  
router.get('/forecast/:location(*)', async (req, res) => {
  try {
    const locationFromPath = req.params.location;
    const { days = 7, units } = req.query;
    
    // Decode location from path
    const location = decodeURIComponent(locationFromPath).replace(/\+/g, ' ');
    const forecastDays = Math.min(parseInt(days), 16);
    
    let selectedUnits = units || 'imperial';
    let coords;
    
    const clientIp = locationParser.getClientIP(req);
    const parsedLocation = locationParser.parseLocation(location, clientIp);
    
    if (parsedLocation.type === 'coordinates') {
      // For coordinates, use the weatherService to get proper location data with reverse geocoding
      coords = await weatherService.getCoordinates(`${parsedLocation.value.latitude},${parsedLocation.value.longitude}`);
    } else {
      coords = await weatherService.getCoordinates(parsedLocation.value);
      if (!units && coords.country) {
        selectedUnits = locationParser.getUnitsForCountry(coords.country);
      }
    }

    const forecast = await weatherService.getForecast(coords.latitude, coords.longitude, forecastDays, selectedUnits);
    
    const response = {
      location: {
        name: coords.name,
        country: coords.country,
        latitude: coords.latitude,
        longitude: coords.longitude,
        timezone: coords.timezone || forecast.timezone
      },
      forecast: {
        days: forecast.days.map(day => ({
          date: day.date,
          weatherCode: day.weatherCode,
          description: weatherService.getWeatherDescription(day.weatherCode),
          icon: weatherService.getWeatherIcon(day.weatherCode),
          temperature: {
            min: day.tempMin,
            max: day.tempMax
          },
          feelsLike: {
            min: day.feelsLikeMin,
            max: day.feelsLikeMax
          },
          precipitation: {
            sum: day.precipitation,
            hours: day.precipitationHours,
            probability: day.precipitationProbability
          },
          wind: {
            speedMax: day.windSpeedMax,
            gustsMax: day.windGustsMax,
            direction: day.windDirection
          },
          solarRadiation: day.solarRadiation
        })),
        units: forecast.units
      },
      meta: {
        source: 'Open-Meteo',
        timestamp: new Date().toISOString(),
        api_version: '1.0',
        units: selectedUnits
      }
    };

    res.json(response);

  } catch (error) {
    res.status(400).json({
      error: {
        code: 'FORECAST_ERROR',
        message: error.message
      }
    });
  }
});

router.get('/forecast', async (req, res) => {
  try {
    const { location, lat, lon, days = 7, units } = req.query;
    const forecastDays = Math.min(parseInt(days), 16);
    
    let coords;
    let selectedUnits = units || 'imperial';
    
    if (lat && lon) {
      coords = {
        latitude: parseFloat(lat),
        longitude: parseFloat(lon),
        name: `${lat}, ${lon}`,
        country: ''
      };
    } else {
      const clientIp = locationParser.getClientIP(req);
      const parsedLocation = locationParser.parseLocation(location, clientIp);
      
      if (parsedLocation.type === 'coordinates') {
        coords = parsedLocation.value;
        coords.name = `${coords.latitude}, ${coords.longitude}`;
        coords.country = '';
      } else {
        coords = await weatherService.getCoordinates(parsedLocation.value);
        if (!units && coords.country) {
          selectedUnits = locationParser.getUnitsForCountry(coords.country);
        }
      }
    }

    const forecast = await weatherService.getForecast(coords.latitude, coords.longitude, forecastDays, selectedUnits);
    
    const response = {
      location: {
        name: coords.name,
        country: coords.country,
        latitude: coords.latitude,
        longitude: coords.longitude,
        timezone: coords.timezone || forecast.timezone
      },
      forecast: {
        days: forecast.days.map(day => ({
          date: day.date,
          weatherCode: day.weatherCode,
          description: weatherService.getWeatherDescription(day.weatherCode),
          icon: weatherService.getWeatherIcon(day.weatherCode),
          temperature: {
            min: day.tempMin,
            max: day.tempMax
          },
          feelsLike: {
            min: day.feelsLikeMin,
            max: day.feelsLikeMax
          },
          precipitation: {
            sum: day.precipitation,
            hours: day.precipitationHours,
            probability: day.precipitationProbability
          },
          wind: {
            speedMax: day.windSpeedMax,
            gustsMax: day.windGustsMax,
            direction: day.windDirection
          },
          solarRadiation: day.solarRadiation
        })),
        units: forecast.units
      },
      meta: {
        source: 'Open-Meteo',
        timestamp: new Date().toISOString(),
        api_version: '1.0',
        units: selectedUnits
      }
    };

    res.json(response);

  } catch (error) {
    res.status(400).json({
      error: {
        code: 'FORECAST_ERROR',
        message: error.message
      }
    });
  }
});

router.get('/current', async (req, res) => {
  try {
    const { location, lat, lon, units } = req.query;
    
    let coords;
    let selectedUnits = units || 'imperial';
    
    if (lat && lon) {
      coords = {
        latitude: parseFloat(lat),
        longitude: parseFloat(lon),
        name: `${lat}, ${lon}`,
        country: ''
      };
    } else {
      const clientIp = locationParser.getClientIP(req);
      const parsedLocation = locationParser.parseLocation(location, clientIp);
      
      if (parsedLocation.type === 'coordinates') {
        coords = parsedLocation.value;
        coords.name = `${coords.latitude}, ${coords.longitude}`;
        coords.country = '';
      } else {
        coords = await weatherService.getCoordinates(parsedLocation.value);
        if (!units && coords.country) {
          selectedUnits = locationParser.getUnitsForCountry(coords.country);
        }
      }
    }

    const [currentWeather, forecast] = await Promise.all([
      weatherService.getCurrentWeather(coords.latitude, coords.longitude, selectedUnits),
      weatherService.getForecast(coords.latitude, coords.longitude, 1, selectedUnits)
    ]);
    
    const response = {
      location: {
        name: coords.name,
        country: coords.country,
        latitude: coords.latitude,
        longitude: coords.longitude,
        timezone: coords.timezone || currentWeather.timezone
      },
      current: {
        temperature: currentWeather.temperature,
        feelsLike: currentWeather.feelsLike,
        humidity: currentWeather.humidity,
        pressure: currentWeather.pressure,
        windSpeed: currentWeather.windSpeed,
        windDirection: currentWeather.windDirection,
        windGusts: currentWeather.windGusts,
        precipitation: currentWeather.precipitation,
        cloudCover: currentWeather.cloudCover,
        weatherCode: currentWeather.weatherCode,
        description: weatherService.getWeatherDescription(currentWeather.weatherCode),
        icon: weatherService.getWeatherIcon(currentWeather.weatherCode, currentWeather.isDay),
        isDay: currentWeather.isDay,
        units: currentWeather.units
      },
      today: forecast.days[0] ? {
        date: forecast.days[0].date,
        weatherCode: forecast.days[0].weatherCode,
        description: weatherService.getWeatherDescription(forecast.days[0].weatherCode),
        icon: weatherService.getWeatherIcon(forecast.days[0].weatherCode),
        temperature: {
          min: forecast.days[0].tempMin,
          max: forecast.days[0].tempMax
        },
        precipitation: {
          sum: forecast.days[0].precipitation,
          probability: forecast.days[0].precipitationProbability
        }
      } : null,
      meta: {
        source: 'Open-Meteo',
        timestamp: new Date().toISOString(),
        api_version: '1.0',
        units: selectedUnits
      }
    };

    res.json(response);

  } catch (error) {
    res.status(400).json({
      error: {
        code: 'CURRENT_ERROR',
        message: error.message
      }
    });
  }
});

router.get('/search', async (req, res) => {
  try {
    const { q: query } = req.query;
    
    if (!query) {
      return res.status(400).json({
        error: {
          code: 'MISSING_QUERY',
          message: 'Query parameter "q" is required'
        }
      });
    }

    // Get multiple results for autocomplete
    const searchResults = await weatherService.searchLocations(query, 10);
    
    res.json({
      query,
      results: searchResults,
      meta: {
        source: 'Open-Meteo Geocoding',
        timestamp: new Date().toISOString(),
        api_version: '1.0'
      }
    });

  } catch (error) {
    res.status(400).json({
      error: {
        code: 'SEARCH_ERROR',
        message: error.message
      }
    });
  }
});

// IP address endpoint
router.get('/ip', (req, res) => {
  const clientIp = locationParser.getClientIP(req);
  
  const response = {
    ip: clientIp,
    timestamp: new Date().toISOString(),
    headers: {
      'x-forwarded-for': req.headers['x-forwarded-for'],
      'x-real-ip': req.headers['x-real-ip'],
      'cf-connecting-ip': req.headers['cf-connecting-ip']
    }
  };
  
  res.set('Content-Type', 'application/json; charset=utf-8');
  res.send(JSON.stringify(response, null, 2) + '\n');
});

// Location detection endpoint  
router.get('/location', (req, res) => {
  try {
    const clientIp = locationParser.getClientIP(req);
    const locationData = locationParser.getLocationFromIP(clientIp);
    
    // Enhanced location response based on country
    const response = {
      ip: clientIp,
      location: {
        city: locationData.value,
        country: locationData.country,
        countryCode: locationData.country === 'United States' ? 'US' :
                     locationData.country === 'United Kingdom' ? 'GB' :
                     locationData.country === 'Canada' ? 'CA' : 'XX',
        units: locationData.units,
        coordinates: locationData.coordinates || null // Use coordinates from IP geolocation if available
      },
      timestamp: new Date().toISOString(),
      source: locationData.coordinates ? 'IP Geolocation with Coordinates' : 'IP Geolocation'
    };
    
    // Helper function to send formatted JSON
    const sendFormattedJSON = (data) => {
      res.set('Content-Type', 'application/json; charset=utf-8');
      res.send(JSON.stringify(data, null, 2) + '\n');
    };
    
    // Try to get precise coordinates if possible
    if (locationData.value && locationData.value !== 'New York City, NY') {
      weatherService.getCoordinates(locationData.value)
        .then(coords => {
          response.location.coordinates = {
            latitude: coords.latitude,
            longitude: coords.longitude
          };
          response.location.city = coords.name;
          response.location.state = coords.admin1;
          response.location.country = coords.country;
          response.location.countryCode = coords.countryCode;
          response.location.fullName = coords.fullName;
          response.location.shortName = coords.shortName;
          response.location.population = coords.population;
          sendFormattedJSON(response);
        })
        .catch(() => {
          // Return basic response if geocoding fails
          sendFormattedJSON(response);
        });
    } else {
      sendFormattedJSON(response);
    }
    
  } catch (error) {
    console.error('Location API error:', error.message);
    console.error('Stack:', error.stack);
    
    const errorResponse = {
      error: {
        code: 'LOCATION_ERROR',
        message: error.message,
        suggestions: [
          'Check your internet connection',
          'Try again in a few moments',
          'Use IP endpoint: /api/v1/ip for basic info'
        ],
        timestamp: new Date().toISOString()
      }
    };
    
    res.status(500).set('Content-Type', 'application/json; charset=utf-8');
    res.send(JSON.stringify(errorResponse, null, 2) + '\n');
  }
});

router.get('/docs', (req, res) => {
  const baseUrl = hostDetector.getBaseUrl(req);
  const examples = hostDetector.generateExampleUrls(req);
  
  res.json({
    api: 'Weather API v1',
    version: '1.0.0',
    description: 'A free weather API using open data sources',
    base_url: baseUrl,
    endpoints: {
      'GET /api/v1/weather': {
        description: 'Get current weather with today forecast',
        parameters: {
          location: 'Location name (optional)',
          lat: 'Latitude (optional, use with lon)',
          lon: 'Longitude (optional, use with lat)',
          format: 'Response format (default: json)',
          units: 'Unit system: imperial or metric (auto-detected by location if not specified)'
        },
        example: examples.api.weather
      },
      'GET /api/v1/current': {
        description: 'Get detailed current weather',
        parameters: {
          location: 'Location name (optional)',
          lat: 'Latitude (optional, use with lon)',
          lon: 'Longitude (optional, use with lat)',
          units: 'Unit system: imperial or metric (auto-detected by location if not specified)'
        },
        example: examples.api.current
      },
      'GET /api/v1/forecast': {
        description: 'Get weather forecast',
        parameters: {
          location: 'Location name (optional)',
          lat: 'Latitude (optional, use with lon)',
          lon: 'Longitude (optional, use with lat)',
          days: 'Number of forecast days (1-16, default: 7)',
          units: 'Unit system: imperial or metric (auto-detected by location if not specified)'
        },
        example: examples.api.forecast
      },
      'GET /api/v1/search': {
        description: 'Search for locations',
        parameters: {
          q: 'Search query (required)'
        },
        example: examples.api.search
      }
    },
    console_endpoints: {
      'GET /': {
        description: 'Console weather for current location (IP-based detection)',
        parameters: {
          format: 'Display format (full, simple)',
          units: 'Unit system: imperial or metric (auto-detected by location if not specified)'
        },
        example: examples.console.current
      },
      'GET /:location': {
        description: 'Console weather for specific location',
        parameters: {
          format: 'Display format (full, simple)',
          units: 'Unit system: imperial or metric (auto-detected by location if not specified)'
        },
        example: examples.console.location
      }
    },
    data_source: 'Open-Meteo.com',
    features: [
      'No API key required',
      'Free for non-commercial use',
      'Global weather data',
      'Up to 16-day forecasts',
      'Imperial units by default with metric support',
      'Automatic location and unit detection from IP',
      'Dynamic host and protocol detection',
      'Caching for better performance',
      'Console-friendly ASCII output',
      'JSON API responses'
    ]
  });
});

module.exports = router;