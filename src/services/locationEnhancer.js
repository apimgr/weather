const axios = require('axios');
const NodeCache = require('node-cache');

// Cache for external location data (one week TTL since this data doesn't change often)
const cache = new NodeCache({ stdTTL: 604800 }); // 1 week (7 days * 24 hours * 60 minutes * 60 seconds)

class LocationEnhancer {
  constructor() {
    this.countriesUrl = 'https://github.com/apimgr/countries/raw/refs/heads/main/countries.json';
    this.citiesUrl = 'https://github.com/apimgr/citylist/raw/refs/heads/master/api/city.list.json';
    this.countriesData = null;
    this.citiesData = null;
  }

  async loadCountriesData() {
    if (this.countriesData) return this.countriesData;
    
    const cached = cache.get('countries_data');
    if (cached) {
      this.countriesData = cached;
      return cached;
    }

    try {
      console.log('Loading countries data from GitHub...');
      const response = await axios.get(this.countriesUrl);
      this.countriesData = response.data;
      cache.set('countries_data', this.countriesData);
      console.log(`Loaded ${Object.keys(this.countriesData).length} countries`);
      return this.countriesData;
    } catch (error) {
      console.error('Failed to load countries data:', error.message);
      return {};
    }
  }

  async loadCitiesData() {
    if (this.citiesData) return this.citiesData;
    
    const cached = cache.get('cities_data');
    if (cached) {
      this.citiesData = cached;
      return cached;
    }

    try {
      console.log('Loading cities data from GitHub...');
      const response = await axios.get(this.citiesUrl);
      this.citiesData = response.data;
      cache.set('cities_data', this.citiesData);
      console.log(`Loaded ${this.citiesData.length} cities`);
      return this.citiesData;
    } catch (error) {
      console.error('Failed to load cities data:', error.message);
      return [];
    }
  }

  async enhanceLocationData(geocodingResult) {
    const [countries, cities] = await Promise.all([
      this.loadCountriesData(),
      this.loadCitiesData()
    ]);

    const enhanced = { ...geocodingResult };

    // Enhance country information using real database
    if (enhanced.country_code && countries) {
      const countryInfo = countries.find(c => c.country_code === enhanced.country_code.toUpperCase());
      if (countryInfo) {
        enhanced.countryName = countryInfo.name;
        enhanced.countryCode = countryInfo.country_code;
        enhanced.capital = countryInfo.capital;
        enhanced.timezones = countryInfo.timezones;
      }
    }

    // Find matching city data for better population/coordinates
    if (cities && cities.length > 0) {
      const cityMatch = cities.find(city => 
        city.name.toLowerCase() === enhanced.name.toLowerCase() ||
        (Math.abs(city.coord.lat - enhanced.latitude) < 0.1 && 
         Math.abs(city.coord.lon - enhanced.longitude) < 0.1)
      );

      if (cityMatch) {
        enhanced.cityId = cityMatch.id;
        enhanced.population = cityMatch.population || enhanced.population;
        enhanced.cityCountryCode = cityMatch.country;
      }
    }

    // Build enhanced location names with real data
    enhanced.fullName = this.buildEnhancedFullName(enhanced);
    enhanced.shortName = this.buildEnhancedShortName(enhanced);

    return enhanced;
  }

  buildEnhancedFullName(data) {
    // Build full location name: "London, England, United Kingdom"
    const parts = [data.name];
    
    if (data.admin1 && data.admin1 !== data.name) {
      parts.push(data.admin1);
    }
    
    if (data.country) {
      parts.push(data.country);
    }
    
    return parts.join(', ');
  }

  buildEnhancedShortName(data) {
    // Build short location name using real country code from database
    const city = data.name;
    const countryCode = data.countryCode || data.country_code?.toUpperCase() || 'XX';
    
    // For US locations, use state abbreviation if available
    if (countryCode === 'US' && data.admin1) {
      const stateAbbrev = this.lookupStateAbbreviation(data.admin1);
      return `${city}, ${stateAbbrev || countryCode}`;
    }
    
    // For all other locations, use country code from database
    return `${city}, ${countryCode}`;
  }

  lookupStateAbbreviation(stateName) {
    // Common state abbreviations (could be enhanced with external US states database)
    const commonStates = {
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
    
    return commonStates[stateName.toLowerCase()];
  }

  async findBestCityMatch(searchTerm) {
    const cities = await this.loadCitiesData();
    const term = searchTerm.toLowerCase();
    
    // Find exact matches first
    const exactMatches = cities.filter(city => 
      city.name.toLowerCase() === term
    );
    
    if (exactMatches.length > 0) {
      return exactMatches[0];
    }
    
    // Find partial matches
    const partialMatches = cities.filter(city => 
      city.name.toLowerCase().includes(term)
    );
    
    // Sort by population (descending) to get most important cities first
    return partialMatches.sort((a, b) => (b.population || 0) - (a.population || 0))[0];
  }
}

module.exports = new LocationEnhancer();