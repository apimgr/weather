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

// Debug and utility endpoints first
app.get('/health', (req, res) => {
  const hostInfo = hostDetector.getHostInfo(req);
  res.json({ 
    status: 'OK', 
    timestamp: new Date().toISOString(),
    host: hostInfo.fullHost
  });
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

// Weather routes last (catch-all)
app.use('/', weatherRoutes);

app.listen(PORT, () => {
  console.log(`Weather API server running on port ${PORT}`);
  console.log(`Visit http://localhost:${PORT}/examples for usage examples`);
  console.log(`API documentation: http://localhost:${PORT}/api/v1/docs`);
});

module.exports = app;