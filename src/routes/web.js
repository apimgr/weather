const express = require('express');
const weatherService = require('../services/weatherService');
const locationParser = require('../utils/locationParser');
const hostDetector = require('../utils/hostDetector');

const router = express.Router();

// Web interface routes
router.get('/', async (req, res) => {
  try {
    const location = req.query.location;
    const units = req.query.units || 'imperial';
    const hideFooter = req.query.F !== undefined;
    
    let weatherData = null;
    let coords = null;
    
    if (location) {
      const clientIp = locationParser.getClientIP(req);
      const parsedLocation = locationParser.parseLocation(location, clientIp);
      
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

      weatherData = {
        location: coords,
        current: currentWeather,
        forecast: forecast,
        units: units
      };
    }

    const hostInfo = hostDetector.getHostInfo(req);
    
    res.render('weather', {
      title: 'Console Weather Service',
      weatherData,
      hostInfo,
      location: location || '',
      units,
      hideFooter
    });

  } catch (error) {
    res.render('weather', {
      title: 'Weather Service - Error',
      weatherData: null,
      hostInfo: hostDetector.getHostInfo(req),
      location: req.query.location || '',
      units: req.query.units || 'imperial',
      error: error.message
    });
  }
});

module.exports = router;