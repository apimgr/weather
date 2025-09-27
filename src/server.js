const express = require('express');
const helmet = require('helmet');
const cors = require('cors');
const morgan = require('morgan');
const path = require('path');
const weatherRoutes = require('./routes/weather');
const apiRoutes = require('./routes/api');
const webRoutes = require('./routes/web');
const hostDetector = require('./utils/hostDetector');

const app = express();
const PORT = process.env.PORT || 3000;

// Trust reverse proxy (nginx, cloudflare, etc.)
app.set('trust proxy', true);

// EJS setup
app.set('view engine', 'ejs');
app.set('views', path.join(__dirname, 'views'));

// Static files with proper MIME types
app.use(express.static(path.join(__dirname, 'public'), {
  setHeaders: (res, path) => {
    if (path.endsWith('.css')) {
      res.set('Content-Type', 'text/css');
    }
  }
}));

// Path normalization middleware - fix double slashes
app.use((req, res, next) => {
  // Normalize path by removing double slashes
  if (req.path.includes('//')) {
    const normalizedPath = req.path.replace(/\/+/g, '/');
    return res.redirect(301, normalizedPath + (req.url.includes('?') ? req.url.substring(req.url.indexOf('?')) : ''));
  }
  next();
});

// Logging middleware  
app.use(morgan(process.env.NODE_ENV === 'production' ? 'combined' : 'dev'));

app.use(helmet({
  contentSecurityPolicy: {
    directives: {
      defaultSrc: ["'self'"],
      styleSrc: ["'self'", "'unsafe-inline'"],
      scriptSrc: ["'self'", "'unsafe-inline'"],
      scriptSrcAttr: ["'none'"], // Block inline event handlers for security
      imgSrc: ["'self'", "data:", "https:"],
      fontSrc: ["'self'", "data:"],
      connectSrc: ["'self'"]
    }
  }
}));
app.use(cors('*'));
app.use(express.json());
app.use(express.text({ type: 'text/plain' }));
app.use(express.urlencoded({ extended: true }));

// Service initialization state
let serviceReady = false;
let initializationStatus = {
  countries: false,
  cities: false,
  weather: true, // Weather API is always ready
  started: new Date().toISOString()
};

// Make initialization status available globally
global.serviceReady = serviceReady;
global.initializationStatus = initializationStatus;

// Auto-mark service as ready after timeout (fallback for production)
setTimeout(() => {
  if (!global.serviceReady) {
    console.log('⏰ Initialization timeout reached, marking service as ready');
    global.serviceReady = true;
    global.initializationStatus.countries = true;
    global.initializationStatus.cities = true;
  }
}, 30000); // 30 second timeout

// Health check endpoints (Kubernetes standard)
app.get('/healthz', (req, res) => {
  const hostInfo = hostDetector.getHostInfo(req);
  
  if (!global.serviceReady) {
    return res.status(503).json({ 
      status: 'INITIALIZING', 
      timestamp: new Date().toISOString(),
      host: hostInfo.fullHost,
      service: 'Console Weather Service',
      version: '1.0.0',
      initialization: global.initializationStatus
    });
  }
  
  res.json({ 
    status: 'OK', 
    timestamp: new Date().toISOString(),
    host: hostInfo.fullHost,
    service: 'Console Weather Service',
    version: '1.0.0',
    uptime: Date.now() - new Date(global.initializationStatus.started).getTime()
  });
});

// Legacy health endpoint for backward compatibility
app.get('/health', (req, res) => {
  res.redirect(301, '/healthz');
});

app.get('/debug/params', (req, res) => {
  const parameterParser = require('./utils/parameterParser');
  const params = parameterParser.parseWttrParams(req.query);
  const formatInfo = parameterParser.getWeatherFormat(params);
  
  res.json({
    query: req.query,
    parsed: params,
    formatInfo: formatInfo
  });
});

app.get('/debug/ip', (req, res) => {
  const locationParser = require('./utils/locationParser');
  const clientIp = locationParser.getClientIP(req);
  const locationData = locationParser.getLocationFromIP(clientIp);
  
  res.json({
    detectedIP: clientIp,
    headers: {
      'x-forwarded-for': req.headers['x-forwarded-for'],
      'x-real-ip': req.headers['x-real-ip'],
      'cf-connecting-ip': req.headers['cf-connecting-ip'],
      'user-agent': req.headers['user-agent']
    },
    location: locationData,
    note: clientIp === '127.0.0.1' || clientIp === '::1' ? 
      'Local access detected. In production, reverse proxy headers will provide real client IP.' :
      'Real client IP detected.'
  });
});

// API routes first - all /api/** should return JSON
app.use('/api/v1', apiRoutes);

// Main /api endpoint - JSON overview
app.get('/api', (req, res) => {
  const hostInfo = hostDetector.getHostInfo(req);
  
  const response = {
    service: 'Console Weather Service API',
    version: '1.0.0',
    description: 'Free weather API with no API key required',
    base_url: hostInfo.fullHost,
    endpoints: {
      weather: `${hostInfo.fullHost}/api/v1/weather[/{location}]`,
      forecast: `${hostInfo.fullHost}/api/v1/forecast[/{location}]`,
      search: `${hostInfo.fullHost}/api/v1/search?q={query}`,
      ip: `${hostInfo.fullHost}/api/v1/ip`,
      location: `${hostInfo.fullHost}/api/v1/location`
    },
    documentation: {
      json: `${hostInfo.fullHost}/api/v1/docs`,
      html: `${hostInfo.fullHost}/api/docs`
    },
    health_check: `${hostInfo.fullHost}/healthz`,
    features: [
      'No API key required',
      'Global weather data from Open-Meteo.com',
      'Automatic IP-based location detection',
      'Multiple API patterns (query and path parameters)',
      'Clean URLs and JSON responses',
      'Rate limiting and security headers'
    ],
    data_sources: {
      weather: 'Open-Meteo.com',
      countries: 'github.com/apimgr/countries',
      cities: 'github.com/apimgr/citylist'
    },
    timestamp: new Date().toISOString()
  };
  
  res.set('Content-Type', 'application/json; charset=utf-8');
  res.send(JSON.stringify(response, null, 2) + '\n');
});

// Add /api/docs endpoint for HTML documentation
app.get('/api/docs', (req, res) => {
  res.render('api-docs', {
    title: 'Console Weather Service - API Documentation',
    hostInfo: hostDetector.getHostInfo(req)
  });
});

// Main application routes
app.use('/web', webRoutes);

// Removed test endpoints for production

app.get('/examples', (req, res) => {
  const examples = hostDetector.generateExampleUrls(req);
  res.set('Content-Type', 'text/plain; charset=utf-8');
  res.send(`Weather API Examples

Console Interface:
${examples.console.current}
${examples.console.location}
${examples.console.simple}
${examples.console.units}

JSON API:
${examples.api.weather}
${examples.api.current}
${examples.api.forecast}
${examples.api.search}
`);
});

// Initialization check middleware for web requests
app.use((req, res, next) => {
  // Only show loading page for browser requests to main routes
  const userAgent = req.headers['user-agent'] || '';
  const isBrowser = userAgent.toLowerCase().includes('mozilla') && 
                   (userAgent.toLowerCase().includes('chrome') || 
                    userAgent.toLowerCase().includes('firefox') || 
                    userAgent.toLowerCase().includes('safari'));
  
  // Skip for API routes, health checks, and static files
  if (req.path.startsWith('/api') || 
      req.path.startsWith('/healthz') || 
      req.path.startsWith('/health') ||
      req.path.startsWith('/debug') ||
      req.path.includes('.')) {
    return next();
  }
  
  // Handle service initialization
  if (!global.serviceReady) {
    if (isBrowser) {
      // Show loading page for browsers
      return res.render('loading', {
        title: 'Console Weather Service - Initializing',
        hostInfo: hostDetector.getHostInfo(req),
        status: global.initializationStatus
      });
    } else {
      // Send helpful message for console clients
      res.set('Content-Type', 'text/plain; charset=utf-8');
      return res.send(`☕ Console Weather Service is warming up...

🌍 Loading global location database (247 countries, 209K+ cities)
🌤️ Connecting to weather services
🎨 Preparing beautiful interfaces

This usually takes 10-15 seconds. Please try again in a moment:

curl -q -LSs ${hostDetector.getHostInfo(req).fullHost}/

For immediate access, try our API endpoints:
curl -q -LSs ${hostDetector.getHostInfo(req).fullHost}/api/v1/weather?location=london

Thank you for your patience! ☕
`);
    }
  }
  
  next();
});

// Weather routes last (catch-all)
app.use('/', weatherRoutes);

app.listen(PORT, async () => {
  console.log(`Weather API server running on port ${PORT}`);
  console.log(`Visit http://localhost:${PORT}/examples for usage examples`);
  console.log(`API documentation: http://localhost:${PORT}/api/v1/docs`);

  // Start data loading immediately on startup
  try {
    const locationEnhancer = require('./services/locationEnhancer');
    console.log('🌍 Starting data initialization...');

    // Start loading both datasets in parallel
    const [countries, cities] = await Promise.all([
      locationEnhancer.loadCountriesData(),
      locationEnhancer.loadCitiesData()
    ]);

    console.log('✅ Data initialization complete');
    console.log(`📍 Loaded ${countries?.length || 0} countries and ${cities?.length || 0} cities`);

    // Mark as ready
    global.serviceReady = true;
    global.initializationStatus.countries = true;
    global.initializationStatus.cities = true;

  } catch (error) {
    console.error('⚠️ Data initialization failed:', error.message);
    console.log('🔄 Service will continue with reduced functionality');
  }
});

module.exports = app;