const express = require('express');
const weatherService = require('../services/weatherService');
const moonService = require('../services/moonService');
const asciiWeather = require('../utils/asciiWeather');
const wttrRenderer = require('../utils/wttrRenderer');
const oneLineRenderer = require('../utils/oneLineRenderer');
const locationParser = require('../utils/locationParser');
const hostDetector = require('../utils/hostDetector');
const userAgentDetector = require('../utils/userAgentDetector');
const parameterParser = require('../utils/parameterParser');

const router = express.Router();

// Moon phase endpoints - matches weather pattern exactly
router.get('/moon', async (req, res) => {
  return handleMoonRequest(req, res, 'moon');
});

router.get('/moon/:location(*)', async (req, res) => {
  const location = req.params.location;
  return handleMoonRequest(req, res, `moon@${location}`);
});

router.get('/', async (req, res) => {
  try {
    const clientIp = locationParser.getClientIP(req);
    const parsedLocation = locationParser.parseLocation('', clientIp);
    
    let coords;
    let units = req.query.units || parsedLocation.units || 'imperial';
    
    if (parsedLocation.type === 'coordinates') {
      // For coordinates, use the weatherService to get proper location data with reverse geocoding
      coords = await weatherService.getCoordinates(`${parsedLocation.value.latitude},${parsedLocation.value.longitude}`);
    } else {
      coords = await weatherService.getCoordinates(parsedLocation.value);
      if (!req.query.units && coords.country) {
        units = locationParser.getUnitsForCountry(coords.country);
      }
    }

    const [currentWeather, forecast] = await Promise.all([
      weatherService.getCurrentWeather(coords.latitude, coords.longitude, units),
      weatherService.getForecast(coords.latitude, coords.longitude, 7, units)
    ]);

    const clientInfo = userAgentDetector.getClientInfo(req);
    
    // If it's a browser, serve HTML
    if (clientInfo.isBrowser && !req.query.format) {
      return res.render('weather', {
        title: 'Console Weather Service',
        weatherData: {
          location: coords,
          current: currentWeather,
          forecast: forecast,
          units: units
        },
        hostInfo: hostDetector.getHostInfo(req),
        location: parsedLocation.value, // Auto-fill with detected location
        units,
        hideFooter: req.query.F !== undefined
      });
    }

    // Console clients get ASCII
    const hideFooter = req.query.F !== undefined;
    const hostInfo = hostDetector.getHostInfo(req);
    let output;

    if (req.query.format === 'simple') {
      output = asciiWeather.renderSimple(coords, currentWeather, units, hideFooter);
    } else {
      output = wttrRenderer.renderWttrStyle(coords, currentWeather, forecast, units, hideFooter, null, hostInfo, '');
    }

    res.set('Content-Type', 'text/plain; charset=utf-8');
    res.send(output);

  } catch (error) {
    res.status(500).set('Content-Type', 'text/plain');
    res.send(`Error: ${error.message}`);
  }
});

// Special endpoints
router.get('/:help', (req, res) => {
  if (req.params.help === ':help') {
    const baseUrl = hostDetector.getBaseUrl(req);
    const examples = hostDetector.generateExampleUrls(req);
    
    res.set('Content-Type', 'text/plain; charset=utf-8');
    res.send(`Console Weather Service

USAGE:
    curl -q -LSs ${baseUrl}/
    curl -q -LSs ${baseUrl}/London,GB
    
PARAMETERS:
    F                 Remove footer
    format=1          Icon + temperature: 🌦 +11⁰C
    format=2          Icon + temp + wind: 🌦 🌡️+11°C 🌬️↓4km/h
    format=3          Location + weather: London, GB: 🌦 +11⁰C
    format=4          Location + detailed: London, GB: 🌦 🌡️+11°C 🌬️↓4km/h
    u                 Imperial units (°F, mph)
    m                 Metric units (°C, km/h)
    
LOCATION FORMATS:
    /London,GB        City with country code
    /Albany,NY        City with state code
    /New+York,NY      Spaces as + symbols
    /33.0392,-80.1805 GPS coordinates (resolves to nearest city)
    /moon             Moon phase (current location)
    /moon/{location}  Moon phase for specific location
    
EXAMPLES:
    curl -q -LSs ${baseUrl}/London,GB?format=3
    curl -q -LSs ${baseUrl}/Albany,NY?u&format=4
    curl -q -LSs ${baseUrl}/33.0392,-80.1805
    curl -q -LSs ${baseUrl}/moon
    curl -q -LSs ${baseUrl}/moon/Tokyo,JP
    
SPECIAL ENDPOINTS:
    /:help            This help message
    /:bash.function   Bash integration function
    
JSON API:
    curl -q -LSs ${baseUrl}/api/v1/weather?location=paris
    curl -q -LSs ${baseUrl}/api/v1/search?q=alb
    
WEB INTERFACE:
    ${baseUrl} (browser interface with autocomplete)
    
More info: ${baseUrl}/api/v1/docs
`);
    return;
  }
  
  if (req.params.help === ':bash.function') {
    const fs = require('fs');
    const path = require('path');
    const bashFunction = fs.readFileSync(path.join(__dirname, '../share/bash.function'), 'utf8');
    const baseUrl = hostDetector.getBaseUrl(req);
    const customizedFunction = bashFunction.replace('wttr.in/', baseUrl.replace('http://', '').replace('https://', '') + '/');
    
    res.set('Content-Type', 'text/plain; charset=utf-8');
    res.send(customizedFunction);
    return;
  }
  
  // Continue to regular location handling
  handleLocationRequest(req, res);
});

async function getLocationFromExternalIP(ip) {
  // Use external IP geolocation for accurate location detection
  const axios = require('axios');

  // Try multiple services for better coordinate availability
  const services = [
    {
      name: 'ipapi.co',
      url: `https://ipapi.co/${ip}/json/`,
      timeout: 5000,
      extract: (data) => data.city ? {
        city: data.city,
        region: data.region,
        country_name: data.country_name,
        country_code: data.country_code,
        latitude: data.latitude,
        longitude: data.longitude
      } : null
    },
    {
      name: 'ip-api.com',
      url: `http://ip-api.com/json/${ip}?fields=status,city,regionName,country,countryCode,lat,lon`,
      timeout: 4000,
      extract: (data) => data.status === 'success' ? {
        city: data.city,
        region: data.regionName,
        country_name: data.country,
        country_code: data.countryCode,
        latitude: data.lat,
        longitude: data.lon
      } : null
    },
    {
      name: 'ipwho.is',
      url: `http://ipwho.is/${ip}`,
      timeout: 4000,
      extract: (data) => data.success ? {
        city: data.city,
        region: data.region,
        country_name: data.country,
        country_code: data.country_code,
        latitude: data.latitude,
        longitude: data.longitude
      } : null
    }
  ];

  for (const service of services) {
    try {
      console.log(`🌐 Trying ${service.name} for IP ${ip}...`);
      const response = await axios.get(service.url, { timeout: service.timeout });

      if (response.data) {
        const result = service.extract(response.data);
        if (result) {
          console.log(`✅ ${service.name} success:`, result);
          return result;
        }
      }
    } catch (error) {
      console.log(`❌ ${service.name} failed:`, error.message);
    }
  }

  console.log('❌ All external IP services failed');
  return null;
}

async function handleLocationRequest(req, res) {
  try {
    const locationInput = locationParser.extractLocationFromPath(req.path);
    const clientInfo = userAgentDetector.getClientInfo(req);
    
    // Parse all wttr.in-style parameters
    const params = parameterParser.parseWttrParams(req.query);
    // console.log('Format requested:', req.query.format);
    
    // Check for moon queries - support both /moon and /moon@location formats
    if (locationInput && locationInput.toLowerCase() === 'moon') {
      return handleMoonRequest(req, res, 'moon');
    }
    if (locationInput && locationInput.toLowerCase().startsWith('moon@')) {
      return handleMoonRequest(req, res, locationInput);
    }
    
    // If it's a browser, serve HTML (unless explicitly requesting ASCII format)
    const isAsciiExplicitlyRequested = req.query.format === 'ascii' || req.query.format === 'text';
    
    if (clientInfo.isBrowser && !isAsciiExplicitlyRequested) {
      // console.log('Serving HTML to browser for location:', locationInput);
      
      const clientIp = locationParser.getClientIP(req);
      const parsedLocation = locationParser.parseLocation(locationInput, clientIp);
      const units = req.query.units || 'imperial';

      let coords;
      if (parsedLocation.type === 'coordinates') {
        // For coordinates, use the weatherService to get proper location data with reverse geocoding
        coords = await weatherService.getCoordinates(`${parsedLocation.value.latitude},${parsedLocation.value.longitude}`);
      } else {
        const country = parsedLocation.country || null;
        coords = await weatherService.getCoordinates(parsedLocation.value, country);
      }

      const [currentWeather, forecast] = await Promise.all([
        weatherService.getCurrentWeather(coords.latitude, coords.longitude, units),
        weatherService.getForecast(coords.latitude, coords.longitude, 7, units)
      ]);

      return res.render('weather', {
        title: 'Console Weather Service',
        weatherData: {
          location: coords,
          current: currentWeather,
          forecast: forecast,
          units: units
        },
        hostInfo: hostDetector.getHostInfo(req),
        location: locationInput || coords.name, // Auto-fill with detected or requested location
        units,
        hideFooter: req.query.F !== undefined
      });
    }
    
    // Console clients get ASCII
    const clientIp = locationParser.getClientIP(req);
    let parsedLocation, coords, units;
    
    // Check if locationInput is an IP address
    if (locationInput && locationParser.isIPAddress(locationInput)) {
      // Handle IP address lookup
      try {
        const ipLocationData = await getLocationFromExternalIP(locationInput);
        coords = await weatherService.getCoordinates(`${ipLocationData.city}, ${ipLocationData.region}`);
        coords.sourceIP = locationInput;
        units = locationParser.getUnitsForCountry(ipLocationData.country_code);
      } catch (error) {
        // Fallback to default if IP lookup fails
        parsedLocation = locationParser.parseLocation('', clientIp);
        coords = await weatherService.getCoordinates(parsedLocation.value);
        units = parsedLocation.units || 'imperial';
      }
    } else {
      // Handle regular location input
      parsedLocation = locationParser.parseLocation(locationInput, clientIp);
      units = parameterParser.getUnitsFromParams(params, parsedLocation.units);
      
      if (parsedLocation.type === 'coordinates') {
        // For coordinates, use the weatherService to get proper location data with reverse geocoding
        coords = await weatherService.getCoordinates(`${parsedLocation.value.latitude},${parsedLocation.value.longitude}`);
      } else {
        const country = parsedLocation.country || null;
        coords = await weatherService.getCoordinates(parsedLocation.value, country);
        if (params.units === 'auto' && coords.country) {
          units = locationParser.getUnitsForCountry(coords.country);
        }
      }
    }

    // For formats 1-4, we only need current weather (no forecast needed)
    const needsForecast = !req.query.format || (req.query.format !== '1' && req.query.format !== '2' && req.query.format !== '3' && req.query.format !== '4');
    
    let currentWeather, forecast;
    if (needsForecast) {
      [currentWeather, forecast] = await Promise.all([
        weatherService.getCurrentWeather(coords.latitude, coords.longitude, units),
        weatherService.getForecast(coords.latitude, coords.longitude, 7, units)
      ]);
    } else {
      // Only get current weather for simple formats
      currentWeather = await weatherService.getCurrentWeather(coords.latitude, coords.longitude, units);
      forecast = null;
    }

    let output;

    // Simple direct format handling
    if (req.query.format === '1') {
      output = oneLineRenderer.renderFormat1(currentWeather, units);
    } else if (req.query.format === '2') {
      output = oneLineRenderer.renderFormat2(currentWeather, units);
    } else if (req.query.format === '3') {
      output = oneLineRenderer.renderFormat3(coords, currentWeather, units);
    } else if (req.query.format === '4') {
      output = oneLineRenderer.renderFormat4(coords, currentWeather, units);
    } else if (req.query.format === '0') {
      // Format 0: Current weather only (no forecast)
      const hostInfo = hostDetector.getHostInfo(req);
      output = wttrRenderer.renderCurrentOnly(coords, currentWeather, units, !params.showFooter, hostInfo, locationInput);
    } else if (req.query.format === 'simple') {
      output = asciiWeather.renderSimple(coords, currentWeather, units, !params.showFooter);
    } else {
      // Default full ASCII art weather report
      const hostInfo = hostDetector.getHostInfo(req);
      output = wttrRenderer.renderWttrStyle(coords, currentWeather, forecast, units, !params.showFooter, null, hostInfo, locationInput);
    }

    res.set('Content-Type', 'text/plain; charset=utf-8');
    res.send(output);

  } catch (error) {
    console.error('Weather route error:', error.message);
    console.error('Query:', req.query);
    console.error('Stack:', error.stack);
    
    // Provide helpful error messages based on error type
    let errorMessage = `Error: ${error.message}`;
    const hostInfo = hostDetector.getHostInfo(req);
    
    if (error.message.includes('Failed to fetch weather data')) {
      errorMessage = `☁️ Weather service temporarily unavailable. Please try again in a moment.

The weather API might be experiencing issues. Try:
• A different location: curl -q -LSs ${hostInfo.fullHost}/London,GB
• Our JSON API: curl -q -LSs ${hostInfo.fullHost}/api/v1/weather?location=london
• Check service status: curl -q -LSs ${hostInfo.fullHost}/healthz

Thank you for your patience! ☁️`;
    } else if (error.message.includes('Location') && error.message.includes('not found')) {
      const requestedLocation = req.path.substring(1) || 'unknown';
      errorMessage = `📍 Location not found: "${requestedLocation}"

Try these alternatives:
• Use city with country: London,GB or Albany,NY
• Try coordinates: 40.7128,-74.0060
• Search for locations: ${hostInfo.fullHost}/api/v1/search?q=your-city
• Current location: ${hostInfo.fullHost}/

📍 Need help? ${hostInfo.fullHost}/:help`;

      res.status(404).set('Content-Type', 'text/plain; charset=utf-8');
      res.send(errorMessage + '\n\n');
      return;
    } else if (error.message.includes('timeout') || error.message.includes('ETIMEDOUT')) {
      errorMessage = `⏰ Request timeout - services are busy. Please try again.

Quick alternatives:
• Try again: curl -q -LSs ${hostInfo.fullHost}${req.path}
• Different location: curl -q -LSs ${hostInfo.fullHost}/London,GB  
• Service status: curl -q -LSs ${hostInfo.fullHost}/healthz

⏰ Usually resolves within a few moments.`;
    } else if (error.message.includes('network') || error.message.includes('ENOTFOUND')) {
      errorMessage = `🌐 Network connectivity issue. External services may be unavailable.

Try these options:
• Wait a moment and retry: curl -q -LSs ${hostInfo.fullHost}${req.path}
• Use a major city: curl -q -LSs ${hostInfo.fullHost}/London,GB
• Check our status: curl -q -LSs ${hostInfo.fullHost}/healthz

🌐 Network issues usually resolve quickly.`;
    }
    
    res.status(500).set('Content-Type', 'text/plain; charset=utf-8');
    res.send(errorMessage + '\n\n');
  }
}

async function handleMoonRequest(req, res, locationInput) {
  try {
    const clientIp = locationParser.getClientIP(req);
    const clientInfo = userAgentDetector.getClientInfo(req);
    let coords;
    
    if (locationInput === 'moon') {
      const parsedLocation = locationParser.parseLocation('', clientIp);
      if (parsedLocation.type === 'coordinates') {
        coords = parsedLocation.value;
        coords.name = `${coords.latitude}, ${coords.longitude}`;
        coords.country = '';
      } else {
        const country = parsedLocation.country || null;
        coords = await weatherService.getCoordinates(parsedLocation.value, country);
      }
    } else {
      const location = locationInput.substring(5);
      const parsedLocation = locationParser.parseLocation(location, clientIp);
      
      if (parsedLocation.type === 'coordinates') {
        coords = parsedLocation.value;
        coords.name = `${coords.latitude}, ${coords.longitude}`;
        coords.country = '';
      } else {
        const country = parsedLocation.country || null;
        coords = await weatherService.getCoordinates(parsedLocation.value, country);
      }
    }

    // If it's a browser, serve HTML moon interface
    if (clientInfo.isBrowser && !req.query.format) {
      try {
        const moonData = moonService.getMoonPhase(new Date(), coords.latitude, coords.longitude);
        const sunTimes = require('suncalc').getTimes(new Date(), coords.latitude, coords.longitude);
        
        return res.render('moon', {
          title: 'Moon Phase Report - Console Weather Service',
          moonData: {
            location: coords,
            moon: moonData,
            sun: sunTimes
          },
          hostInfo: hostDetector.getHostInfo(req),
          location: locationInput === 'moon' ? '' : locationInput.substring(5),
          hideFooter: req.query.F !== undefined
        });
      } catch (renderError) {
        console.error('Moon HTML render error:', renderError.message);
        // Fall through to ASCII output if HTML fails
      }
    }

    // Console clients get ASCII
    const moonReport = moonService.renderMoonReport(coords);
    
    res.set('Content-Type', 'text/plain; charset=utf-8');
    res.send(moonReport);

  } catch (error) {
    console.error('Moon route error:', error.message);
    console.error('Stack:', error.stack);
    
    const hostInfo = hostDetector.getHostInfo(req);
    let errorMessage = `🌙 Moon phase service error: ${error.message}`;
    
    if (error.message.includes('Location') && error.message.includes('not found')) {
      const requestedLocation = locationInput.replace('moon@', '') || 'unknown';
      errorMessage = `🌙 Moon phase location not found: "${requestedLocation}"

Try these alternatives:
• Use city with country: ${hostInfo.fullHost}/moon/London,GB
• Try coordinates: ${hostInfo.fullHost}/moon/40.7128,-74.0060
• Current location: ${hostInfo.fullHost}/moon

🌙 Need help? ${hostInfo.fullHost}/:help`;

      res.status(404).set('Content-Type', 'text/plain; charset=utf-8');
      res.send(errorMessage + '\n\n');
      return;
    }

    res.status(500).set('Content-Type', 'text/plain; charset=utf-8');
    res.send(errorMessage + '\n\n');
  }
}

// Filter out invalid paths that should return 404
router.get('/:location(*)', (req, res, next) => {
  const path = req.path;

  // Block common invalid paths that should not be processed as locations
  const invalidPaths = [
    /^\/wp-/,           // WordPress paths
    /^\/admin/,         // Admin panels
    /^\/\.well-known/,  // Well-known URIs
    /^\/favicon\.ico/,  // Favicon
    /^\/robots\.txt/,   // Robots.txt
    /^\/sitemap/,       // Sitemaps
    /\.(php|asp|jsp|cgi)/, // Server-side scripts
    /\.(js|css|png|jpg|gif|ico|svg)$/, // Static files
    /^\/[a-z0-9]+\.(html?|xml|json)$/ // Static documents
  ];

  // Check if path matches any invalid pattern
  if (invalidPaths.some(pattern => pattern.test(path))) {
    return res.status(404).set('Content-Type', 'text/plain; charset=utf-8').send('404 Not Found\n');
  }

  // Continue to location handler for valid paths
  handleLocationRequest(req, res, next);
});

module.exports = router;