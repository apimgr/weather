const geoip = require('geoip-lite');

class LocationParser {
  parseLocation(input, clientIp) {
    if (!input || input.trim() === '') {
      return this.getLocationFromIP(clientIp);
    }

    const location = input.trim();
    
    if (this.isCoordinates(location)) {
      return this.parseCoordinates(location);
    }

    return { type: 'name', value: location };
  }

  isCoordinates(input) {
    const coordRegex = /^-?\d+\.?\d*,-?\d+\.?\d*$/;
    return coordRegex.test(input);
  }

  parseCoordinates(input) {
    const [lat, lon] = input.split(',').map(coord => parseFloat(coord.trim()));
    
    if (lat >= -90 && lat <= 90 && lon >= -180 && lon <= 180) {
      return {
        type: 'coordinates',
        value: { latitude: lat, longitude: lon }
      };
    }
    
    throw new Error('Invalid coordinates format');
  }

  getLocationFromIP(ip) {
    // console.log('Getting location for IP:', ip);
    
    if (!ip || ip === '127.0.0.1' || ip === '::1' || ip === '::ffff:127.0.0.1') {
      // console.log('Localhost detected, using New York City as default');
      return { 
        type: 'default', 
        value: 'New York City, NY',
        country: 'US',
        units: 'imperial'
      };
    }

    const geo = geoip.lookup(ip);
    console.log('🗺️ GeoIP lookup result for', ip + ':', geo);

    if (geo) {
      const location = geo.city || this.getLocationFromCountry(geo.country);
      console.log('📍 Location resolved to:', location, 'Country:', geo.country);

      // Include coordinates if available from geoip-lite
      const result = {
        type: 'name',
        value: location,
        country: geo.country,
        units: this.getUnitsForCountry(geo.country)
      };

      // Add coordinates if available (geo.ll = [latitude, longitude])
      if (geo.ll && geo.ll.length === 2) {
        result.coordinates = {
          latitude: geo.ll[0],
          longitude: geo.ll[1]
        };
        console.log('✅ Coordinates from geoip-lite:', geo.ll[0], geo.ll[1]);
      } else {
        console.log('⚠️ No coordinates available from geoip-lite');
      }

      return result;
    }

    // console.log('GeoIP lookup failed, using New York as fallback');
    return { 
      type: 'default', 
      value: 'New York City, NY',
      country: 'US',
      units: 'imperial'
    };
  }

  getLocationFromCountry(countryCode) {
    const countryCapitals = {
      'US': 'New York City, NY',
      'GB': 'London',
      'CA': 'Toronto',
      'AU': 'Sydney',
      'DE': 'Berlin',
      'FR': 'Paris',
      'JP': 'Tokyo',
      'CN': 'Beijing',
      'IN': 'Mumbai',
      'BR': 'São Paulo',
      'RU': 'Moscow',
      'IT': 'Rome',
      'ES': 'Madrid',
      'MX': 'Mexico City',
      'AR': 'Buenos Aires',
      'ZA': 'Cape Town'
    };
    return countryCapitals[countryCode] || 'New York';
  }

  getUnitsForCountry(countryCode) {
    const metricCountries = ['GB', 'CA', 'AU', 'DE', 'FR', 'JP', 'CN', 'IN', 'BR', 'RU', 'IT', 'ES', 'AR', 'ZA'];
    const imperialCountries = ['US', 'BS', 'BZ', 'KY', 'PW'];
    
    if (imperialCountries.includes(countryCode)) {
      return 'imperial';
    }
    
    if (metricCountries.includes(countryCode)) {
      return 'metric';
    }
    
    return 'imperial';
  }

  detectUnitsFromLocation(locationData) {
    if (locationData.country) {
      return this.getUnitsForCountry(locationData.country);
    }
    return 'imperial';
  }

  extractLocationFromPath(path) {
    const parts = path.split('/').filter(part => part.length > 0);
    
    if (parts.length === 0) {
      return null;
    }

    let location = parts[0];
    
    try {
      // Handle URL encoding properly
      // First replace + with spaces (common in URL encoding)
      location = location.replace(/\+/g, ' ');
      
      // Then decode URL encoding (%20, %2C, etc.)
      location = decodeURIComponent(location);
      
      // Handle any remaining encoded characters
      location = location.replace(/%20/g, ' ');
      location = location.replace(/%2C/g, ',');
      location = location.replace(/%2c/g, ',');
      
      // Remove query parameters if any
      const queryIndex = location.indexOf('?');
      if (queryIndex !== -1) {
        location = location.substring(0, queryIndex);
      }
      
      // Trim whitespace
      location = location.trim();
      
    } catch (error) {
      // If decoding fails, use original location with basic cleanup
      location = parts[0].replace(/\+/g, ' ').trim();
    }

    return location;
  }

  isIPAddress(input) {
    // Check if input is an IPv4 or IPv6 address
    const ipv4Regex = /^(?:(?:25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)\.){3}(?:25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)$/;
    const ipv6Regex = /^(?:[0-9a-fA-F]{1,4}:){7}[0-9a-fA-F]{1,4}$/;
    
    return ipv4Regex.test(input) || ipv6Regex.test(input);
  }

  encodeLocationForURL(location) {
    if (!location) return '';
    
    // Encode spaces and special characters for URLs
    return encodeURIComponent(location)
      .replace(/'/g, '%27')
      .replace(/"/g, '%22');
  }

  normalizeLocationInput(location) {
    if (!location) return null;
    
    // Normalize common variations
    return location
      .trim()
      .replace(/\s+/g, ' ') // Multiple spaces to single space
      .replace(/,\s*,/g, ',') // Remove duplicate commas
      .replace(/^\s*,|,\s*$/g, ''); // Remove leading/trailing commas
  }

  getClientIP(req) {
    const forwarded = req.headers['x-forwarded-for'];
    const realIP = req.headers['x-real-ip'];
    const cfIP = req.headers['cf-connecting-ip']; // Cloudflare
    const trueIP = req.headers['true-client-ip']; // Akamai
    
    let ip = forwarded || realIP || cfIP || trueIP ||
             req.connection?.remoteAddress ||
             req.socket?.remoteAddress ||
             req.connection?.socket?.remoteAddress ||
             req.ip ||
             '127.0.0.1';

    // Handle comma-separated list in x-forwarded-for
    if (ip && ip.includes(',')) {
      ip = ip.split(',')[0].trim();
    }

    // Remove IPv6 prefix for IPv4-mapped addresses
    if (ip && ip.startsWith('::ffff:')) {
      ip = ip.substring(7);
    }

    // Debug logging disabled for production

    return ip;
  }
}

module.exports = new LocationParser();