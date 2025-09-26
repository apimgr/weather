// Hardcoded coordinates for major cities as fallbacks
const cityCoordinates = {
  'New York City, NY': { latitude: 40.7128, longitude: -74.0060, country: 'United States', timezone: 'America/New_York' },
  'New York': { latitude: 40.7128, longitude: -74.0060, country: 'United States', timezone: 'America/New_York' },
  'London': { latitude: 51.5074, longitude: -0.1278, country: 'United Kingdom', timezone: 'Europe/London' },
  'Paris': { latitude: 48.8566, longitude: 2.3522, country: 'France', timezone: 'Europe/Paris' },
  'Tokyo': { latitude: 35.6762, longitude: 139.6503, country: 'Japan', timezone: 'Asia/Tokyo' },
  'Berlin': { latitude: 52.5200, longitude: 13.4050, country: 'Germany', timezone: 'Europe/Berlin' },
  'Sydney': { latitude: -33.8688, longitude: 151.2093, country: 'Australia', timezone: 'Australia/Sydney' },
  'Toronto': { latitude: 43.6532, longitude: -79.3832, country: 'Canada', timezone: 'America/Toronto' },
  'Mumbai': { latitude: 19.0760, longitude: 72.8777, country: 'India', timezone: 'Asia/Kolkata' },
  'São Paulo': { latitude: -23.5505, longitude: -46.6333, country: 'Brazil', timezone: 'America/Sao_Paulo' },
  'Moscow': { latitude: 55.7558, longitude: 37.6173, country: 'Russia', timezone: 'Europe/Moscow' },
  'Rome': { latitude: 41.9028, longitude: 12.4964, country: 'Italy', timezone: 'Europe/Rome' },
  'Madrid': { latitude: 40.4168, longitude: -3.7038, country: 'Spain', timezone: 'Europe/Madrid' },
  'Mexico City': { latitude: 19.4326, longitude: -99.1332, country: 'Mexico', timezone: 'America/Mexico_City' },
  'Buenos Aires': { latitude: -34.6037, longitude: -58.3816, country: 'Argentina', timezone: 'America/Argentina/Buenos_Aires' },
  'Cape Town': { latitude: -33.9249, longitude: 18.4241, country: 'South Africa', timezone: 'Africa/Johannesburg' },
  'Beijing': { latitude: 39.9042, longitude: 116.4074, country: 'China', timezone: 'Asia/Shanghai' },
  'Los Angeles': { latitude: 34.0522, longitude: -118.2437, country: 'United States', timezone: 'America/Los_Angeles' },
  'Chicago': { latitude: 41.8781, longitude: -87.6298, country: 'United States', timezone: 'America/Chicago' },
  'Miami': { latitude: 25.7617, longitude: -80.1918, country: 'United States', timezone: 'America/New_York' }
};

class LocationFallbacks {
  getFallbackCoordinates(locationName) {
    const normalized = this.normalizeLocationName(locationName);
    return cityCoordinates[normalized] || null;
  }

  normalizeLocationName(name) {
    if (!name) return null;
    
    // Handle common variations
    const normalizations = {
      'nyc': 'New York',
      'ny': 'New York',
      'new york city': 'New York',
      'la': 'Los Angeles',
      'san francisco': 'San Francisco',
      'sf': 'San Francisco'
    };

    const lower = name.toLowerCase().trim();
    if (normalizations[lower]) {
      return normalizations[lower];
    }

    // Title case the input
    return name.split(' ')
      .map(word => word.charAt(0).toUpperCase() + word.slice(1).toLowerCase())
      .join(' ');
  }

  createFallbackLocation(locationName) {
    const coords = this.getFallbackCoordinates(locationName);
    if (!coords) return null;

    return {
      latitude: coords.latitude,
      longitude: coords.longitude,
      name: locationName,
      country: coords.country,
      timezone: coords.timezone
    };
  }

  getAllFallbackLocations() {
    return Object.keys(cityCoordinates);
  }
}

module.exports = new LocationFallbacks();