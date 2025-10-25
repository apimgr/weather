# API Reference

Complete API documentation for the Weather service.

## Base URL

```
https://wthr.top
https://your-domain.com
```

## Authentication

No authentication required. Rate limiting applies (see Rate Limiting section).

## Response Formats

All API endpoints return JSON by default. HTML views are available for web endpoints.

### Success Response

```json
{
  "status": "success",
  "data": {...}
}
```

### Error Response

```json
{
  "status": "error",
  "message": "Error description"
}
```

## Weather Endpoints

### Get Weather Forecast

Retrieve weather forecast for a location.

**Web Endpoint:**
```
GET /
GET /weather/{location}
GET /{location}
```

**API Endpoint:**
```
GET /api/weather?location={location}
```

**Parameters:**

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| location | string | No | Location (city, ZIP, coordinates) |

**Location Formats:**

- City, State: `Brooklyn, NY` or `Brooklyn,NY`
- City, Country: `London, UK`
- ZIP Code: `10001`
- Coordinates: `40.7128,-74.0060` (lat,lon)
- Airport Code: `JFK`

If no location specified, uses IP-based geolocation.

**Response:**

```json
{
  "location": "Brooklyn, NY",
  "latitude": 40.7128,
  "longitude": -74.0060,
  "timezone": "America/New_York",
  "elevation": 15,
  "current": {
    "time": "2025-10-16T12:00:00-04:00",
    "temperature": 72.5,
    "apparent_temperature": 71.2,
    "humidity": 65,
    "precipitation": 0,
    "rain": 0,
    "showers": 0,
    "snowfall": 0,
    "weather_code": 2,
    "cloud_cover": 40,
    "pressure_msl": 1013.2,
    "surface_pressure": 1010.5,
    "wind_speed": 10.5,
    "wind_direction": 180,
    "wind_gusts": 15.2
  },
  "daily": [
    {
      "date": "2025-10-16",
      "temperature_max": 75.2,
      "temperature_min": 62.1,
      "apparent_temperature_max": 74.5,
      "apparent_temperature_min": 61.8,
      "sunrise": "06:42:00",
      "sunset": "18:15:00",
      "uv_index_max": 5.2,
      "precipitation_sum": 0,
      "rain_sum": 0,
      "showers_sum": 0,
      "snowfall_sum": 0,
      "precipitation_hours": 0,
      "precipitation_probability_max": 10,
      "wind_speed_max": 12.5,
      "wind_gusts_max": 18.3,
      "wind_direction_dominant": 180,
      "weather_code": 2
    }
  ],
  "hourly": [...]
}
```

**Weather Codes:**

| Code | Description |
|------|-------------|
| 0 | Clear sky |
| 1, 2, 3 | Mainly clear, partly cloudy, and overcast |
| 45, 48 | Fog and depositing rime fog |
| 51, 53, 55 | Drizzle: Light, moderate, and dense intensity |
| 61, 63, 65 | Rain: Slight, moderate and heavy intensity |
| 71, 73, 75 | Snow fall: Slight, moderate, and heavy intensity |
| 77 | Snow grains |
| 80, 81, 82 | Rain showers: Slight, moderate, and violent |
| 85, 86 | Snow showers slight and heavy |
| 95 | Thunderstorm: Slight or moderate |
| 96, 99 | Thunderstorm with slight and heavy hail |

**Example:**

```bash
curl -q -LSsf "https://wthr.top/api/weather?location=Brooklyn,NY"
```

## Severe Weather Endpoints

### Get Severe Weather Alerts

Retrieve severe weather alerts for a location.

**Web Endpoints:**
```
GET /severe-weather
GET /severe-weather/{location}
GET /severe/{type}
GET /severe/{type}/{location}
```

**API Endpoint:**
```
GET /api/severe?location={location}&distance={miles}&type={type}
```

**Parameters:**

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| location | string | No | Location (city, coordinates, etc.) |
| distance | integer | No | Filter radius in miles (default: 50) |
| type | string | No | Alert type filter |

**Alert Types:**

- `hurricanes` - Tropical storms and hurricanes
- `tornadoes` - Tornado warnings and watches
- `storms` - Severe thunderstorm warnings
- `winter` - Winter storm warnings and advisories
- `floods` - Flood warnings and watches

**Response:**

```json
{
  "alerts": [
    {
      "id": "urn:oid:2.49.0.1.840.0.12345",
      "type": "Tornado Warning",
      "headline": "Tornado Warning issued for Kings County",
      "severity": "Extreme",
      "urgency": "Immediate",
      "certainty": "Observed",
      "description": "A confirmed tornado was located...",
      "instruction": "Take shelter immediately...",
      "event": "Tornado Warning",
      "effective": "2025-10-16T12:00:00-04:00",
      "expires": "2025-10-16T12:45:00-04:00",
      "sender_name": "National Weather Service New York",
      "area_desc": "Kings County, NY",
      "distance_miles": 5.2,
      "country": "US"
    }
  ],
  "location": {
    "name": "Brooklyn, NY",
    "latitude": 40.7128,
    "longitude": -74.0060
  },
  "filter": {
    "distance_miles": 50,
    "type": "all"
  }
}
```

**Example:**

```bash
# All alerts within 50 miles
curl -q -LSsf "https://wthr.top/api/severe?location=Miami,FL"

# Tornado warnings within 25 miles
curl -q -LSsf "https://wthr.top/api/severe?location=Oklahoma City,OK&distance=25&type=tornadoes"

# Hurricane alerts within 100 miles
curl -q -LSsf "https://wthr.top/api/severe?location=New Orleans,LA&distance=100&type=hurricanes"
```

## Moon Phase Endpoint

### Get Moon Phase

Retrieve moon phase information for a location and date.

**Web Endpoints:**
```
GET /moon
GET /moon/{location}
```

**API Endpoint:**
```
GET /api/moon?location={location}&date={date}
```

**Parameters:**

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| location | string | No | Location for rise/set times |
| date | string | No | Date (YYYY-MM-DD format) |

**Response:**

```json
{
  "phase": "Waxing Gibbous",
  "illumination": 0.73,
  "age": 10.5,
  "distance": 384400,
  "angular_diameter": 0.524,
  "sun_distance": 149600000,
  "sun_angular_diameter": 0.533,
  "moonrise": "15:30:00",
  "moonset": "02:15:00",
  "location": {
    "name": "Brooklyn, NY",
    "latitude": 40.7128,
    "longitude": -74.0060,
    "timezone": "America/New_York"
  },
  "date": "2025-10-16"
}
```

**Moon Phases:**

- New Moon (0-0.03 illumination)
- Waxing Crescent (0.03-0.47)
- First Quarter (0.47-0.53)
- Waxing Gibbous (0.53-0.97)
- Full Moon (0.97-1.0)
- Waning Gibbous (1.0-0.53)
- Last Quarter (0.53-0.47)
- Waning Crescent (0.47-0.03)

**Example:**

```bash
curl -q -LSsf "https://wthr.top/api/moon?location=Brooklyn,NY&date=2025-10-16"
```

## Earthquake Endpoint

### Get Earthquake Data

Retrieve recent earthquakes near a location.

**Web Endpoints:**
```
GET /earthquake
GET /earthquake/{location}
```

**API Endpoint:**
```
GET /api/earthquake?location={location}&radius={km}&days={days}
```

**Parameters:**

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| location | string | No | Center location |
| radius | integer | No | Search radius in km (default: 500) |
| days | integer | No | Days of history (default: 7, max: 30) |

**Response:**

```json
{
  "earthquakes": [
    {
      "id": "us7000abc123",
      "magnitude": 5.2,
      "place": "10 km NE of Los Angeles, CA",
      "time": "2025-10-16T08:30:45.123Z",
      "latitude": 34.1,
      "longitude": -118.2,
      "depth": 12.5,
      "type": "earthquake",
      "url": "https://earthquake.usgs.gov/earthquakes/eventpage/us7000abc123",
      "distance_km": 15.2,
      "felt": 120,
      "cdi": 4.5,
      "mmi": 5.2,
      "alert": "green",
      "status": "reviewed",
      "tsunami": false
    }
  ],
  "location": {
    "name": "Los Angeles, CA",
    "latitude": 34.0522,
    "longitude": -118.2437
  },
  "filter": {
    "radius_km": 500,
    "days": 7
  }
}
```

**Magnitude Scale:**

| Magnitude | Description |
|-----------|-------------|
| < 2.0 | Micro (not felt) |
| 2.0-3.9 | Minor (rarely felt) |
| 4.0-4.9 | Light (often felt) |
| 5.0-5.9 | Moderate (damage possible) |
| 6.0-6.9 | Strong (damaging) |
| 7.0-7.9 | Major (serious damage) |
| 8.0+ | Great (catastrophic) |

**Example:**

```bash
curl -q -LSsf "https://wthr.top/api/earthquake?location=San Francisco,CA&radius=200&days=7"
```

## Health Check Endpoint

### Check Service Health

Verify the service is running and healthy.

**Endpoint:**
```
GET /api/health
```

**Response:**

```json
{
  "status": "healthy",
  "timestamp": "2025-10-16T12:00:00Z",
  "version": "1.0.0",
  "uptime_seconds": 86400
}
```

**Example:**

```bash
curl -q -LSsf "https://wthr.top/api/health"
```

## Rate Limiting

Rate limiting is enforced per IP address:

- **Default**: 100 requests per 15 minutes
- **Burst**: Up to 20 requests in rapid succession

When rate limited, the response will be:

```json
{
  "status": "error",
  "message": "Rate limit exceeded. Try again in X seconds."
}
```

HTTP Status Code: `429 Too Many Requests`

## Error Codes

| Code | Description |
|------|-------------|
| 400 | Bad Request - Invalid parameters |
| 404 | Not Found - Endpoint or resource not found |
| 429 | Too Many Requests - Rate limit exceeded |
| 500 | Internal Server Error - Server-side error |
| 503 | Service Unavailable - External API unavailable |

## Caching

Responses are cached to improve performance:

- Weather data: 15 minutes
- Severe weather alerts: 5 minutes
- Moon data: 1 hour
- Earthquake data: 10 minutes

Cache headers are included in responses:

```
Cache-Control: public, max-age=900
ETag: "abc123..."
```

## CORS

CORS is enabled for all origins. Cross-origin requests are supported.

## Compression

All responses support gzip compression. Include the header:

```
Accept-Encoding: gzip
```

## Best Practices

1. **Use Location Persistence**: The web interface saves your location in cookies
2. **Specify Distance**: For severe weather, specify appropriate distance to reduce response size
3. **Cache Responses**: Respect cache headers to reduce load
4. **Handle Errors**: Always check the `status` field in responses
5. **Use Health Check**: Monitor service availability with `/api/health`

## Code Examples

### Python

```python
import requests

# Get weather
response = requests.get('https://wthr.top/api/weather', params={
    'location': 'Brooklyn, NY'
})
data = response.json()
print(f"Temperature: {data['current']['temperature']}°F")

# Get severe weather alerts
response = requests.get('https://wthr.top/api/severe', params={
    'location': 'Miami, FL',
    'distance': 50,
    'type': 'hurricanes'
})
alerts = response.json()['alerts']
for alert in alerts:
    print(f"{alert['type']}: {alert['headline']}")
```

### JavaScript

```javascript
// Get weather
fetch('https://wthr.top/api/weather?location=Brooklyn,NY')
  .then(response => response.json())
  .then(data => {
    console.log(`Temperature: ${data.current.temperature}°F`);
  });

// Get moon phase
fetch('https://wthr.top/api/moon?location=Brooklyn,NY')
  .then(response => response.json())
  .then(data => {
    console.log(`Moon Phase: ${data.phase} (${Math.round(data.illumination * 100)}%)`);
  });
```

### Go

```go
package main

import (
    "encoding/json"
    "fmt"
    "net/http"
)

type WeatherResponse struct {
    Location string `json:"location"`
    Current struct {
        Temperature float64 `json:"temperature"`
    } `json:"current"`
}

func main() {
    resp, _ := http.Get("https://wthr.top/api/weather?location=Brooklyn,NY")
    defer resp.Body.Close()

    var data WeatherResponse
    json.NewDecoder(resp.Body).Decode(&data)

    fmt.Printf("Temperature in %s: %.1f°F\n", data.Location, data.Current.Temperature)
}
```

### cURL

```bash
# Get weather
curl -q -LSsf "https://wthr.top/api/weather?location=Brooklyn,NY" | jq .

# Get severe weather with pretty output
curl -q -LSsf "https://wthr.top/api/severe?location=Miami,FL&distance=50" | jq '.alerts[] | {type, severity, headline}'

# Get earthquakes
curl -q -LSsf "https://wthr.top/api/earthquake?location=San Francisco,CA&radius=200" | jq '.earthquakes[] | {magnitude, place, time}'

# Health check
curl -q -LSsf "https://wthr.top/api/health"
```
