const axios = require('axios');
const NodeCache = require('node-cache');
const unitConverter = require('../utils/unitConverter');
const locationFallbacks = require('../utils/locationFallbacks');
const locationEnhancer = require('./locationEnhancer');

const cache = new NodeCache({ stdTTL: 600 });

class WeatherService {
  constructor() {
    this.openMeteoBaseUrl = 'https://api.open-meteo.com/v1';
    this.geocodingUrl = 'https://geocoding-api.open-meteo.com/v1';
  }

  async getCoordinates(location, country = null) {
    const cacheKey = `coords_${location}_${country || 'any'}`;
    const cached = cache.get(cacheKey);
    if (cached) return cached;

    try {
      // Check if location is coordinates for reverse geocoding
      if (this.isCoordinates(location)) {
        const [lat, lon] = location.split(',').map(coord => parseFloat(coord.trim()));
        return await this.reverseGeocode(lat, lon);
      }
      
      // Parse the location to extract city and state/country info
      const locationParts = this.parseLocationString(location);
      let searchName = location;
      
      // If we have country info from IP, use it unless location already has country/state info
      if (country && !locationParts.hasStateOrCountry) {
        searchName = `${location}, ${country}`;
      }
      
      // Try multiple search approaches
      const searchResults = await this.tryMultipleGeocodingApproaches(searchName, locationParts);
      
      if (searchResults && searchResults.length > 0) {
        // Find the best match based on location context
        const bestMatch = this.selectBestLocationMatch(searchResults, locationParts, country);
        
        // Enhance location data with external sources
        const enhancedData = await locationEnhancer.enhanceLocationData(bestMatch);
        
        const coords = {
          latitude: enhancedData.latitude,
          longitude: enhancedData.longitude,
          name: enhancedData.name,
          country: enhancedData.country,
          timezone: enhancedData.timezone,
          admin1: enhancedData.admin1,
          admin2: enhancedData.admin2,
          countryCode: enhancedData.country_code?.toUpperCase(),
          population: enhancedData.population,
          // Use enhanced location names
          fullName: enhancedData.fullName,
          shortName: enhancedData.shortName
        };
        cache.set(cacheKey, coords);
        return coords;
      }

      // Try fallback coordinates for common cities
      const fallbackLocation = locationFallbacks.createFallbackLocation(location);
      if (fallbackLocation) {
        // console.log(`Using fallback coordinates for "${location}"`);
        cache.set(cacheKey, fallbackLocation);
        return fallbackLocation;
      }

      throw new Error(`Location "${location}" not found`);
    } catch (error) {
      // console.log(`Geocoding error for "${location}":`, error.message);
      
      // Try fallback even on network errors
      const fallbackLocation = locationFallbacks.createFallbackLocation(location);
      if (fallbackLocation) {
        // console.log(`Network error, using fallback coordinates for "${location}"`);
        cache.set(cacheKey, fallbackLocation);
        return fallbackLocation;
      }
      
      // If no fallback available, provide a default location instead of failing
      // console.log(`No fallback found for "${location}", using New York as default`);
      const defaultLocation = {
        latitude: 40.7128,
        longitude: -74.0060,
        name: location, // Keep the requested name
        country: 'United States',
        timezone: 'America/New_York'
      };
      cache.set(cacheKey, defaultLocation);
      return defaultLocation;
    }
  }

  async getCurrentWeather(latitude, longitude, units = 'imperial') {
    const cacheKey = `current_${latitude}_${longitude}_${units}`;
    const cached = cache.get(cacheKey);
    if (cached) return cached;

    try {
      const response = await axios.get(`${this.openMeteoBaseUrl}/forecast`, {
        params: {
          latitude,
          longitude,
          current: [
            'temperature_2m',
            'relative_humidity_2m',
            'apparent_temperature',
            'is_day',
            'precipitation',
            'weather_code',
            'cloud_cover',
            'pressure_msl',
            'wind_speed_10m',
            'wind_direction_10m',
            'wind_gusts_10m'
          ].join(','),
          timezone: 'auto'
        }
      });

      const data = response.data;
      let weather = {
        temperature: data.current.temperature_2m,
        feelsLike: data.current.apparent_temperature,
        humidity: data.current.relative_humidity_2m,
        pressure: data.current.pressure_msl,
        windSpeed: data.current.wind_speed_10m,
        windDirection: data.current.wind_direction_10m,
        windGusts: data.current.wind_gusts_10m,
        precipitation: data.current.precipitation,
        cloudCover: data.current.cloud_cover,
        weatherCode: data.current.weather_code,
        isDay: data.current.is_day,
        units: data.current_units,
        timezone: data.timezone
      };

      weather = unitConverter.convertWeatherData(weather, units);
      cache.set(cacheKey, weather);
      return weather;
    } catch (error) {
      // console.error('Weather API error:', error.message, 'for coords:', latitude, longitude);
      throw new Error(`Failed to fetch weather data: ${error.message}`);
    }
  }

  async getForecast(latitude, longitude, days = 7, units = 'imperial') {
    const cacheKey = `forecast_${latitude}_${longitude}_${days}_${units}`;
    const cached = cache.get(cacheKey);
    if (cached) return cached;

    try {
      const response = await axios.get(`${this.openMeteoBaseUrl}/forecast`, {
        params: {
          latitude,
          longitude,
          daily: [
            'weather_code',
            'temperature_2m_max',
            'temperature_2m_min',
            'apparent_temperature_max',
            'apparent_temperature_min',
            'precipitation_sum',
            'precipitation_hours',
            'precipitation_probability_max',
            'wind_speed_10m_max',
            'wind_gusts_10m_max',
            'wind_direction_10m_dominant',
            'shortwave_radiation_sum'
          ].join(','),
          forecast_days: days,
          timezone: 'auto'
        }
      });

      const data = response.data;
      let forecast = {
        days: data.daily.time.map((date, index) => ({
          date,
          weatherCode: data.daily.weather_code[index],
          tempMax: data.daily.temperature_2m_max[index],
          tempMin: data.daily.temperature_2m_min[index],
          feelsLikeMax: data.daily.apparent_temperature_max[index],
          feelsLikeMin: data.daily.apparent_temperature_min[index],
          precipitation: data.daily.precipitation_sum[index],
          precipitationHours: data.daily.precipitation_hours[index],
          precipitationProbability: data.daily.precipitation_probability_max[index],
          windSpeedMax: data.daily.wind_speed_10m_max[index],
          windGustsMax: data.daily.wind_gusts_10m_max[index],
          windDirection: data.daily.wind_direction_10m_dominant[index],
          solarRadiation: data.daily.shortwave_radiation_sum[index]
        })),
        units: data.daily_units,
        timezone: data.timezone
      };

      forecast = unitConverter.convertForecastData(forecast, units);
      cache.set(cacheKey, forecast);
      return forecast;
    } catch (error) {
      throw new Error(`Failed to fetch forecast data: ${error.message}`);
    }
  }

  getWeatherDescription(code) {
    const weatherCodes = {
      0: 'Clear sky',
      1: 'Mainly clear',
      2: 'Partly cloudy',
      3: 'Overcast',
      45: 'Fog',
      48: 'Depositing rime fog',
      51: 'Light drizzle',
      53: 'Moderate drizzle',
      55: 'Dense drizzle',
      56: 'Light freezing drizzle',
      57: 'Dense freezing drizzle',
      61: 'Slight rain',
      63: 'Moderate rain',
      65: 'Heavy rain',
      66: 'Light freezing rain',
      67: 'Heavy freezing rain',
      71: 'Slight snow',
      73: 'Moderate snow',
      75: 'Heavy snow',
      77: 'Snow grains',
      80: 'Slight rain showers',
      81: 'Moderate rain showers',
      82: 'Violent rain showers',
      85: 'Slight snow showers',
      86: 'Heavy snow showers',
      95: 'Thunderstorm',
      96: 'Thunderstorm with slight hail',
      99: 'Thunderstorm with heavy hail'
    };
    return weatherCodes[code] || 'Unknown';
  }

  getWeatherIcon(code, isDay = true) {
    const icons = {
      0: isDay ? '☀️' : '🌙',
      1: isDay ? '🌤️' : '🌙',
      2: isDay ? '⛅' : '☁️',
      3: '☁️',
      45: '🌫️',
      48: '🌫️',
      51: '🌦️',
      53: '🌦️',
      55: '🌧️',
      56: '🌨️',
      57: '🌨️',
      61: '🌧️',
      63: '🌧️',
      65: '⛈️',
      66: '🌨️',
      67: '🌨️',
      71: '🌨️',
      73: '❄️',
      75: '❄️',
      77: '🌨️',
      80: '🌦️',
      81: '🌧️',
      82: '⛈️',
      85: '🌨️',
      86: '❄️',
      95: '⛈️',
      96: '⛈️',
      99: '⛈️'
    };
    return icons[code] || '❓';
  }

  parseLocationString(location) {
    if (!location) return { city: '', state: '', country: '', hasStateOrCountry: false };
    
    const parts = location.split(',').map(p => p.trim());
    const hasStateOrCountry = parts.length > 1;
    
    return {
      city: parts[0] || '',
      state: parts[1] || '',
      country: parts[2] || parts[1] || '',
      hasStateOrCountry,
      originalParts: parts
    };
  }

  async tryMultipleGeocodingApproaches(searchName, locationParts) {
    const approaches = [
      // Try exact search first
      { name: searchName, count: 5 },
      // Try just the city name if we have multiple parts
      ...(locationParts.hasStateOrCountry ? [{ name: locationParts.city, count: 10 }] : [])
    ];

    for (const approach of approaches) {
      try {
        const response = await axios.get(`${this.geocodingUrl}/search`, {
          params: {
            name: approach.name,
            count: approach.count,
            language: 'en',
            format: 'json'
          }
        });

        if (response.data.results && response.data.results.length > 0) {
          // console.log(`Geocoding found ${response.data.results.length} results for "${approach.name}"`);
          return response.data.results;
        }
      } catch (error) {
        // console.log(`Geocoding approach "${approach.name}" failed:`, error.message);
        continue;
      }
    }
    
    return [];
  }

  selectBestLocationMatch(results, locationParts, userCountry) {
    if (results.length === 1) {
      return results[0];
    }

    // If location has state/country info, try to match it
    if (locationParts.hasStateOrCountry) {
      const stateCountryPart = locationParts.state.toLowerCase();
      
      // Look for exact state/country matches
      for (const result of results) {
        const admin1 = (result.admin1 || '').toLowerCase();
        const country = (result.country || '').toLowerCase();
        const countryCode = (result.country_code || '').toLowerCase();
        
        if (admin1.includes(stateCountryPart) || 
            country.includes(stateCountryPart) ||
            countryCode === stateCountryPart ||
            this.matchesStateAbbreviation(stateCountryPart, admin1, countryCode)) {
          // console.log(`Found best match: ${result.name}, ${result.admin1}, ${result.country}`);
          return result;
        }
      }
    }

    // If user's country context available, prefer locations in that country
    if (userCountry) {
      const countryMatches = results.filter(r => 
        r.country_code === userCountry || 
        r.country.toLowerCase().includes(userCountry.toLowerCase())
      );
      if (countryMatches.length > 0) {
        return countryMatches[0];
      }
    }

    // Default to the first result (usually most populated)
    // console.log(`Using default first result: ${results[0].name}, ${results[0].admin1}, ${results[0].country}`);
    return results[0];
  }

  matchesStateAbbreviation(input, admin1, countryCode) {
    if (countryCode !== 'us') return false;
    
    const stateAbbreviations = {
      'al': 'alabama', 'ak': 'alaska', 'az': 'arizona', 'ar': 'arkansas', 'ca': 'california',
      'co': 'colorado', 'ct': 'connecticut', 'de': 'delaware', 'fl': 'florida', 'ga': 'georgia',
      'hi': 'hawaii', 'id': 'idaho', 'il': 'illinois', 'in': 'indiana', 'ia': 'iowa',
      'ks': 'kansas', 'ky': 'kentucky', 'la': 'louisiana', 'me': 'maine', 'md': 'maryland',
      'ma': 'massachusetts', 'mi': 'michigan', 'mn': 'minnesota', 'ms': 'mississippi', 'mo': 'missouri',
      'mt': 'montana', 'ne': 'nebraska', 'nv': 'nevada', 'nh': 'new hampshire', 'nj': 'new jersey',
      'nm': 'new mexico', 'ny': 'new york', 'nc': 'north carolina', 'nd': 'north dakota', 'oh': 'ohio',
      'ok': 'oklahoma', 'or': 'oregon', 'pa': 'pennsylvania', 'ri': 'rhode island', 'sc': 'south carolina',
      'sd': 'south dakota', 'tn': 'tennessee', 'tx': 'texas', 'ut': 'utah', 'vt': 'vermont',
      'va': 'virginia', 'wa': 'washington', 'wv': 'west virginia', 'wi': 'wisconsin', 'wy': 'wyoming'
    };

    const fullStateName = stateAbbreviations[input.toLowerCase()];
    return fullStateName && admin1.toLowerCase().includes(fullStateName);
  }

  async searchLocations(query, limit = 10) {
    // Search for multiple locations for autocomplete
    const cacheKey = `search_${query}_${limit}`;
    const cached = cache.get(cacheKey);
    if (cached) return cached;

    try {
      const response = await axios.get(`${this.geocodingUrl}/search`, {
        params: {
          name: query,
          count: limit,
          language: 'en',
          format: 'json'
        }
      });

      if (response.data.results && response.data.results.length > 0) {
        // Convert results to simpler format for autocomplete (no enhancement needed for search)
        const results = response.data.results.map(result => ({
          latitude: result.latitude,
          longitude: result.longitude,
          name: result.name,
          country: result.country,
          admin1: result.admin1,
          admin2: result.admin2,
          countryCode: result.country_code?.toUpperCase(),
          population: result.population,
          // Build simple names for search
          fullName: this.buildFullLocationName(result),
          shortName: this.buildShortLocationName(result)
        }));

        cache.set(cacheKey, results, 3600); // Cache for 1 hour
        return results;
      }

      return [];
    } catch (error) {
      console.error('Search error:', error.message);
      return [];
    }
  }

  buildFullLocationName(result) {
    // Build full location name for search results
    const parts = [result.name];
    
    if (result.admin1 && result.admin1 !== result.name) {
      parts.push(result.admin1);
    }
    
    if (result.country) {
      parts.push(result.country);
    }
    
    return parts.join(', ');
  }

  buildShortLocationName(result) {
    // Build short location name for search results
    const city = result.name;
    const countryCode = result.country_code ? result.country_code.toUpperCase() : 'XX';
    
    // For US locations, use state abbreviation if available
    if (countryCode === 'US' && result.admin1) {
      const stateAbbrev = this.getStateAbbreviation(result.admin1);
      return `${city}, ${stateAbbrev || countryCode}`;
    }
    
    // For all other locations, use country code
    return `${city}, ${countryCode}`;
  }

  getStateAbbreviation(stateName) {
    const states = {
      'alabama': 'AL', 'alaska': 'AK', 'arizona': 'AZ', 'arkansas': 'AR', 'california': 'CA',
      'colorado': 'CO', 'connecticut': 'CT', 'delaware': 'DE', 'florida': 'FL', 'georgia': 'GA',
      'hawaii': 'HI', 'idaho': 'ID', 'illinois': 'IL', 'indiana': 'IN', 'iowa': 'IA',
      'kansas': 'KS', 'kentucky': 'KY', 'louisiana': 'LA', 'maine': 'ME', 'maryland': 'MD',
      'massachusetts': 'MA', 'michigan': 'MI', 'minnesota': 'MN', 'mississippi': 'MS', 'missouri': 'MO',
      'montana': 'MT', 'nebraska': 'NE', 'nevada': 'NV', 'new hampshire': 'NH', 'new jersey': 'NJ',
      'new mexico': 'NM', 'new york': 'NY', 'north carolina': 'NC', 'north dakota': 'ND', 'ohio': 'OH',
      'oklahoma': 'OK', 'oregon': 'OR', 'pennsylvania': 'PA', 'rhode island': 'RI', 'south carolina': 'SC',
      'south dakota': 'SD', 'tennessee': 'TN', 'texas': 'TX', 'utah': 'UT', 'vermont': 'VT',
      'virginia': 'VA', 'washington': 'WA', 'west virginia': 'WV', 'wisconsin': 'WI', 'wyoming': 'WY'
    };
    
    return states[stateName.toLowerCase()];
  }

  isCoordinates(location) {
    // Check if location string is coordinates (lat,lon format)
    const coordRegex = /^-?\d+\.?\d*,-?\d+\.?\d*$/;
    return coordRegex.test(location.trim());
  }

  async reverseGeocode(latitude, longitude) {
    // Reverse geocode coordinates to get nearest city name
    const cacheKey = `reverse_${latitude}_${longitude}`;
    const cached = cache.get(cacheKey);
    if (cached) {
      if (process.env.NODE_ENV !== 'production') {
        console.log(`🔄 Using cached reverse geocode for ${latitude}, ${longitude}`);
      }
      return cached;
    }

    if (process.env.NODE_ENV !== 'production') {
      console.log(`🌍 Starting reverse geocoding for ${latitude}, ${longitude}`);
    }

    try {
      // First try Open-Meteo geocoding to get proper admin1 state data
      const geocodingUrl = `${this.geocodingUrl}/search?latitude=${latitude}&longitude=${longitude}&count=1&language=en&format=json`;

      try {
        const response = await fetch(geocodingUrl);
        if (response.ok) {
          const data = await response.json();
          if (data.results && data.results.length > 0) {
            const result = data.results[0];
            const geocodingData = {
              latitude: latitude,
              longitude: longitude,
              name: result.name,
              country: result.country,
              country_code: result.country_code,
              admin1: result.admin1,
              admin2: result.admin2,
              timezone: result.timezone
            };

            // Enhance with external database (cities, population, etc.)
            const enhancedData = await locationEnhancer.enhanceLocationData(geocodingData);

            cache.set(cacheKey, enhancedData, 3600); // Cache for 1 hour
            return enhancedData;
          }
        }
      } catch (geocodingError) {
        console.log('⚠️ Open-Meteo geocoding failed, trying cities database:', geocodingError.message);
      }

      // Fallback: Find nearest city using our external city database
      const nearestCity = await locationEnhancer.findNearestCity(latitude, longitude);

      if (nearestCity) {
        const coords = {
          latitude: latitude, // Use original coordinates
          longitude: longitude,
          name: nearestCity.name,
          country: nearestCity.country || 'Unknown',
          timezone: nearestCity.timezone || 'UTC',
          admin1: nearestCity.admin1,
          admin2: nearestCity.admin2,
          countryCode: nearestCity.countryCode,
          population: nearestCity.population,
          fullName: nearestCity.fullName,
          shortName: nearestCity.shortName
        };

        cache.set(cacheKey, coords, 3600); // Cache for 1 hour
        return coords;
      }

      // Fallback: return coordinates with minimal info
      if (process.env.NODE_ENV !== 'production') {
        console.log(`⚠️ No nearest city found for ${latitude}, ${longitude}, using fallback data`);
      }
      return {
        latitude: latitude,
        longitude: longitude,
        name: `${latitude}, ${longitude}`,
        country: 'Unknown',
        timezone: 'UTC',
        admin1: null,
        admin2: null,
        countryCode: 'XX',
        population: null,
        fullName: `Coordinates ${latitude}, ${longitude}`,
        shortName: `${latitude}, ${longitude}`
      };

    } catch (error) {
      console.error('Reverse geocoding error:', error.message);
      
      // Fallback: return coordinates with minimal info
      if (process.env.NODE_ENV !== 'production') {
        console.log(`⚠️ No nearest city found for ${latitude}, ${longitude}, using fallback data`);
      }
      return {
        latitude: latitude,
        longitude: longitude,
        name: `${latitude}, ${longitude}`,
        country: 'Unknown',
        timezone: 'UTC',
        admin1: null,
        admin2: null,
        countryCode: 'XX',
        population: null,
        fullName: `Coordinates ${latitude}, ${longitude}`,
        shortName: `${latitude}, ${longitude}`
      };
    }
  }
}

module.exports = new WeatherService();