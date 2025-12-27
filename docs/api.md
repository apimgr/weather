# API Reference

Weather Service provides a RESTful JSON API for accessing weather data, managing locations, and more.

## Base URL

```
http://localhost/api/v1
```

## Authentication

Most endpoints are public and don't require authentication. User-specific endpoints require authentication via session cookie.

### Authentication Flow

1. **Login** - `POST /api/v1/auth/login`
2. **Access Protected Endpoints** - Include session cookie
3. **Logout** - `POST /api/v1/auth/logout`

## API Endpoints

### Weather Endpoints

#### Get Current Weather

Get current weather for client's IP location or specific location.

```http
GET /api/v1/weather
GET /api/v1/weather/:location
```

**Parameters:**

| Parameter | Type | Description |
|-----------|------|-------------|
| `location` | string | Location name, coordinates (lat,lon), or ZIP code |

**Response:**

```json
{
  "location": {
    "name": "New York",
    "lat": 40.7128,
    "lon": -74.0060,
    "country": "US",
    "timezone": "America/New_York"
  },
  "current": {
    "temperature": 72.5,
    "feels_like": 70.2,
    "humidity": 65,
    "pressure": 1013.25,
    "wind_speed": 10.5,
    "wind_direction": 180,
    "description": "Partly cloudy",
    "icon": "partly_cloudy",
    "timestamp": "2025-01-15T14:30:00Z"
  }
}
```

#### Get Weather Forecast

Get 16-day weather forecast.

```http
GET /api/v1/forecast
GET /api/v1/forecast/:location
```

**Response:**

```json
{
  "location": {
    "name": "New York",
    "lat": 40.7128,
    "lon": -74.0060
  },
  "forecast": [
    {
      "date": "2025-01-15",
      "temperature_max": 75.0,
      "temperature_min": 62.0,
      "precipitation": 0.2,
      "humidity": 70,
      "wind_speed": 12.0,
      "description": "Partly cloudy"
    }
  ]
}
```

#### Get Historical Weather

Get historical weather data.

```http
GET /api/v1/history
```

**Query Parameters:**

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `lat` | float | Yes | Latitude |
| `lon` | float | Yes | Longitude |
| `start_date` | string | Yes | Start date (YYYY-MM-DD) |
| `end_date` | string | Yes | End date (YYYY-MM-DD) |

**Example:**

```bash
curl "http://localhost/api/v1/history?lat=40.7128&lon=-74.0060&start_date=2025-01-01&end_date=2025-01-07"
```

### Natural Events

#### Get Earthquakes

Get recent earthquake data from USGS.

```http
GET /api/v1/earthquakes
```

**Query Parameters:**

| Parameter | Type | Default | Description |
|-----------|------|---------|-------------|
| `min_magnitude` | float | 2.5 | Minimum magnitude |
| `days` | int | 7 | Days of history |

**Response:**

```json
{
  "earthquakes": [
    {
      "id": "us6000abcd",
      "magnitude": 5.2,
      "location": "10 km E of Los Angeles, CA",
      "lat": 34.05,
      "lon": -118.24,
      "depth": 15.3,
      "time": "2025-01-15T12:34:56Z",
      "url": "https://earthquake.usgs.gov/earthquakes/eventpage/us6000abcd"
    }
  ],
  "count": 15
}
```

#### Get Hurricanes

Get active hurricane and tropical storm data.

```http
GET /api/v1/hurricanes
```

**Response:**

```json
{
  "active_storms": [
    {
      "name": "Hurricane Example",
      "category": 3,
      "lat": 25.5,
      "lon": -80.0,
      "max_wind": 120,
      "min_pressure": 950,
      "movement": "NNW at 12 mph",
      "forecast_track": [
        {"lat": 25.6, "lon": -80.1, "time": "2025-01-15T18:00:00Z"}
      ]
    }
  ],
  "count": 2
}
```

#### Get Severe Weather Alerts

Get severe weather alerts for a location.

```http
GET /api/v1/severe-weather
```

**Query Parameters:**

| Parameter | Type | Description |
|-----------|------|-------------|
| `lat` | float | Latitude |
| `lon` | float | Longitude |
| `country` | string | Country code (US, CA, UK, AU, JP, MX) |

**Response:**

```json
{
  "alerts": [
    {
      "type": "tornado_warning",
      "severity": "extreme",
      "headline": "Tornado Warning",
      "description": "A tornado warning has been issued...",
      "start": "2025-01-15T14:00:00Z",
      "end": "2025-01-15T16:00:00Z",
      "affected_areas": ["New York County", "Kings County"]
    }
  ],
  "count": 3
}
```

#### Get Moon Phase

Get moon phase and lunar information.

```http
GET /api/v1/moon
```

**Query Parameters:**

| Parameter | Type | Description |
|-----------|------|-------------|
| `date` | string | Date (YYYY-MM-DD), defaults to today |
| `lat` | float | Latitude for rise/set times |
| `lon` | float | Longitude for rise/set times |

**Response:**

```json
{
  "phase": "waxing_gibbous",
  "illumination": 0.73,
  "age": 10.5,
  "distance": 384400,
  "next_full_moon": "2025-01-20T12:00:00Z",
  "next_new_moon": "2025-02-03T08:00:00Z",
  "moonrise": "2025-01-15T15:30:00Z",
  "moonset": "2025-01-16T04:15:00Z"
}
```

### Location Endpoints

#### Search Locations

Search for locations by name.

```http
GET /api/v1/locations/search
```

**Query Parameters:**

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `q` | string | Yes | Search query |
| `limit` | int | No | Result limit (default: 10) |

**Response:**

```json
{
  "results": [
    {
      "name": "New York",
      "lat": 40.7128,
      "lon": -74.0060,
      "country": "US",
      "state": "NY",
      "population": 8336817
    }
  ]
}
```

#### Lookup by ZIP Code

```http
GET /api/v1/locations/lookup/zip/:code
```

**Example:**

```bash
curl http://localhost/api/v1/locations/lookup/zip/10001
```

#### Lookup by Coordinates

```http
GET /api/v1/locations/lookup/coords
```

**Query Parameters:**

| Parameter | Type | Required |
|-----------|------|----------|
| `lat` | float | Yes |
| `lon` | float | Yes |

### User Location Management

Requires authentication.

#### List Saved Locations

```http
GET /api/v1/locations
```

**Response:**

```json
{
  "locations": [
    {
      "id": 1,
      "name": "Home",
      "lat": 40.7128,
      "lon": -74.0060,
      "alerts_enabled": true,
      "created_at": "2025-01-15T10:00:00Z"
    }
  ]
}
```

#### Save Location

```http
POST /api/v1/locations
```

**Request Body:**

```json
{
  "name": "Home",
  "lat": 40.7128,
  "lon": -74.0060,
  "alerts_enabled": true
}
```

#### Update Location

```http
PUT /api/v1/locations/:id
```

#### Delete Location

```http
DELETE /api/v1/locations/:id
```

### Authentication Endpoints

#### Login

```http
POST /api/v1/auth/login
```

**Request Body:**

```json
{
  "username": "user@example.com",
  "password": "your-password"
}
```

**Response:**

```json
{
  "success": true,
  "user": {
    "id": 1,
    "username": "user@example.com",
    "email": "user@example.com"
  }
}
```

#### Logout

```http
POST /api/v1/auth/logout
```

#### Register

```http
POST /api/v1/auth/register
```

**Request Body:**

```json
{
  "username": "newuser",
  "email": "newuser@example.com",
  "password": "SecurePassword123!"
}
```

### Utility Endpoints

#### Get Client IP

```http
GET /api/v1/ip
```

**Response:**

```json
{
  "ip": "203.0.113.45",
  "country": "US",
  "city": "New York",
  "latitude": 40.7128,
  "longitude": -74.0060
}
```

#### Get GeoIP Location

```http
GET /api/v1/location
```

Returns location for client's IP address.

#### Health Check

```http
GET /healthz
GET /api/v1/healthz
```

**Response:**

```json
{
  "status": "healthy",
  "version": "1.0.0",
  "uptime": 86400,
  "timestamp": "2025-01-15T14:30:00Z"
}
```

## Rate Limiting

API endpoints are rate-limited:

| Endpoint Type | Limit |
|---------------|-------|
| Public endpoints | 60 requests/minute |
| Authenticated endpoints | 100 requests/minute |
| Admin endpoints | 200 requests/minute |

Rate limit headers:

```http
X-RateLimit-Limit: 60
X-RateLimit-Remaining: 45
X-RateLimit-Reset: 1642252800
```

When rate limited:

```json
{
  "error": "Rate limit exceeded",
  "retry_after": 30
}
```

## Error Responses

Standard error format:

```json
{
  "error": "Error message",
  "code": "ERROR_CODE",
  "details": {}
}
```

HTTP status codes:

| Code | Meaning |
|------|---------|
| 200 | Success |
| 400 | Bad Request |
| 401 | Unauthorized |
| 403 | Forbidden |
| 404 | Not Found |
| 429 | Rate Limited |
| 500 | Server Error |

## Examples

### Using cURL

```bash
# Get current weather
curl http://localhost/api/v1/weather

# Get weather for New York
curl http://localhost/api/v1/weather/New%20York

# Get forecast
curl http://localhost/api/v1/forecast?lat=40.7128&lon=-74.0060

# Get earthquakes
curl http://localhost/api/v1/earthquakes?min_magnitude=4.0

# Login
curl -X POST http://localhost/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"username":"user","password":"pass"}' \
  -c cookies.txt

# Use authenticated endpoint
curl http://localhost/api/v1/locations \
  -b cookies.txt
```

### Using JavaScript

```javascript
// Get current weather
fetch('/api/v1/weather')
  .then(res => res.json())
  .then(data => console.log(data));

// Get forecast
fetch('/api/v1/forecast/New York')
  .then(res => res.json())
  .then(data => console.log(data.forecast));

// Login
fetch('/api/v1/auth/login', {
  method: 'POST',
  headers: {'Content-Type': 'application/json'},
  body: JSON.stringify({username: 'user', password: 'pass'})
})
  .then(res => res.json())
  .then(data => console.log(data));
```

### Using Python

```python
import requests

# Get current weather
response = requests.get('http://localhost/api/v1/weather')
weather = response.json()
print(weather)

# Get forecast
response = requests.get('http://localhost/api/v1/forecast',
                       params={'lat': 40.7128, 'lon': -74.0060})
forecast = response.json()
print(forecast)

# Login
session = requests.Session()
response = session.post('http://localhost/api/v1/auth/login',
                       json={'username': 'user', 'password': 'pass'})
print(response.json())

# Use authenticated endpoint
response = session.get('http://localhost/api/v1/locations')
locations = response.json()
print(locations)
```

## API Versioning

Current API version: `v1`

The API version is included in the URL path: `/api/v1/`

Future versions will be available at `/api/v2/`, etc.

## Next Steps

- [Admin Panel](admin.md) - Manage the service via web UI
- [CLI Reference](cli.md) - Use command-line tools
- [Configuration](configuration.md) - Configure API settings
