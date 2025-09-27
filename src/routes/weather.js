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

router.get('/', async (req, res) => {
  try {
    const clientIp = locationParser.getClientIP(req);
    const parsedLocation = locationParser.parseLocation('', clientIp);
    
    let coords;
    let units = req.query.units || parsedLocation.units || 'imperial';
    
    if (parsedLocation.type === 'coordinates') {
      coords = parsedLocation.value;
      coords.name = `${coords.latitude}, ${coords.longitude}`;
      coords.country = '';
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
    ${examples.console.current}
    ${examples.console.location}
    
PARAMETERS:
    F                 Remove footer
    format=simple     Simple output format  
    units=imperial    Imperial units (°F, mph, inHg, in)
    units=metric      Metric units (°C, km/h, hPa, mm)
    
LOCATION FORMATS:
    /London           City name
    /New+York         City with spaces
    /40.7128,-74.0060 GPS coordinates
    /moon             Moon phase (current location)
    /moon@London      Moon phase for specific location
    
EXAMPLES:
    ${examples.console.simple}
    ${examples.console.units}
    ${baseUrl}/moon
    ${baseUrl}/moon@tokyo
    
SPECIAL ENDPOINTS:
    /:help            This help message
    /:bash.function   Bash integration function
    
JSON API:
    ${examples.api.weather}
    ${examples.api.forecast}
    
WEB INTERFACE:
    ${baseUrl}/web
    
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

async function handleLocationRequest(req, res) {
  try {
    const locationInput = locationParser.extractLocationFromPath(req.path);
    const clientInfo = userAgentDetector.getClientInfo(req);
    
    // Parse all wttr.in-style parameters
    const params = parameterParser.parseWttrParams(req.query);
    // console.log('Format requested:', req.query.format);
    
    // Check for moon queries
    if (locationInput && (locationInput.toLowerCase() === 'moon' || locationInput.toLowerCase().startsWith('moon@'))) {
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
        coords = parsedLocation.value;
        coords.name = `${coords.latitude}, ${coords.longitude}`;
        coords.country = '';
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
    const parsedLocation = locationParser.parseLocation(locationInput, clientIp);

    let coords;
    let units = parameterParser.getUnitsFromParams(params, parsedLocation.units);
    
    if (parsedLocation.type === 'coordinates') {
      coords = parsedLocation.value;
      coords.name = `${coords.latitude}, ${coords.longitude}`;
      coords.country = '';
    } else {
      const country = parsedLocation.country || null;
      coords = await weatherService.getCoordinates(parsedLocation.value, country);
      if (params.units === 'auto' && coords.country) {
        units = locationParser.getUnitsForCountry(coords.country);
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
    res.status(500).set('Content-Type', 'text/plain');
    res.send(`Error: ${error.message}`);
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
    }

    // Console clients get ASCII
    const moonReport = moonService.renderMoonReport(coords);
    
    res.set('Content-Type', 'text/plain; charset=utf-8');
    res.send(moonReport);

  } catch (error) {
    res.status(500).set('Content-Type', 'text/plain');
    res.send(`Error: ${error.message}`);
  }
}

router.get('/:location(*)', handleLocationRequest);

module.exports = router;