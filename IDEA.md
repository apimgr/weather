# Weather Service - Project Idea

## Purpose

A unified weather information platform that aggregates global meteorological data from authoritative sources into a single, accessible service for developers, businesses, and end users. Provides weather forecasts, severe weather alerts, earthquake data, hurricane tracking, and lunar phases through a single API and web interface.

## Target Users

1. **Developers** - Building weather-aware applications, need single API instead of 10+ integrations
2. **System Administrators** - Need self-hosted weather services with full data control
3. **Emergency Services** - Monitoring severe weather and seismic activity in real-time
4. **News Organizations** - Require reliable weather data feeds for broadcast/web
5. **IoT/Smart Home** - Systems needing weather integration for automation
6. **General Public** - Clean, ad-free weather interface with location detection

## Features

| Feature | Description |
|---------|-------------|
| **Weather Forecasts** | 16-day global forecasts with hourly/daily breakdown |
| **Severe Weather Alerts** | Real-time alerts from 6 countries (US, Canada, UK, Australia, Japan, Mexico) |
| **Hurricane Tracking** | Active tropical storm monitoring from NOAA NHC |
| **Earthquake Data** | Real-time seismic activity from USGS |
| **Moon Phases** | Lunar cycles, illumination, rise/set times |
| **Location Detection** | Automatic IP-based geolocation |
| **Multi-Format API** | JSON, text/plain, GraphQL support |
| **WebSocket Alerts** | Real-time push notifications for severe weather |
| **User Locations** | Save favorite locations for quick access |
| **Alert Subscriptions** | Subscribe to alerts for specific regions |

## Data Models

```go
// Weather represents current conditions and forecast
type Weather struct {
    Location    Location    `json:"location"`
    Current     Current     `json:"current"`
    Forecast    []DayForecast `json:"forecast"`
    LastUpdated time.Time   `json:"last_updated"`
}

// Location represents a geographic location
type Location struct {
    Name      string  `json:"name"`
    Region    string  `json:"region"`
    Country   string  `json:"country"`
    Latitude  float64 `json:"lat"`
    Longitude float64 `json:"lon"`
    Timezone  string  `json:"timezone"`
}

// SevereWeatherAlert represents a weather alert
type SevereWeatherAlert struct {
    ID          string    `json:"id"`
    Event       string    `json:"event"`       // tornado, flood, winter_storm, etc.
    Severity    string    `json:"severity"`    // extreme, severe, moderate, minor
    Certainty   string    `json:"certainty"`   // observed, likely, possible
    Urgency     string    `json:"urgency"`     // immediate, expected, future
    Headline    string    `json:"headline"`
    Description string    `json:"description"`
    Instruction string    `json:"instruction"`
    Areas       []string  `json:"areas"`
    Effective   time.Time `json:"effective"`
    Expires     time.Time `json:"expires"`
    Source      string    `json:"source"`      // NOAA, Environment Canada, etc.
}

// Earthquake represents a seismic event
type Earthquake struct {
    ID        string    `json:"id"`
    Magnitude float64   `json:"magnitude"`
    Location  string    `json:"location"`
    Depth     float64   `json:"depth"`       // km
    Time      time.Time `json:"time"`
    Latitude  float64   `json:"lat"`
    Longitude float64   `json:"lon"`
    Tsunami   bool      `json:"tsunami"`
    URL       string    `json:"url"`
}

// Hurricane represents a tropical storm
type Hurricane struct {
    ID          string    `json:"id"`
    Name        string    `json:"name"`
    Category    int       `json:"category"`    // 1-5, 0 for tropical storm
    WindSpeed   int       `json:"wind_speed"`  // mph
    Pressure    int       `json:"pressure"`    // mb
    Movement    string    `json:"movement"`    // direction and speed
    Location    string    `json:"location"`
    Latitude    float64   `json:"lat"`
    Longitude   float64   `json:"lon"`
    Forecast    []HurricaneForecastPoint `json:"forecast"`
}

// MoonPhase represents lunar data
type MoonPhase struct {
    Phase        string    `json:"phase"`       // new, waxing_crescent, first_quarter, etc.
    Illumination float64   `json:"illumination"` // 0.0-1.0
    Age          float64   `json:"age"`         // days since new moon
    Moonrise     time.Time `json:"moonrise"`
    Moonset      time.Time `json:"moonset"`
}
```

## Business Rules

### Weather Data
- Cache weather data for 15 minutes (reduce external API calls)
- Accept city name, ZIP code, coordinates, or auto-detect from IP
- Return metric or imperial units based on user preference or region
- Fall back to cached data if external API unavailable
- Include data attribution for all external sources

### Severe Weather Alerts
- Poll alert sources every 5 minutes
- Deduplicate alerts by ID across sources
- Filter alerts by severity threshold (configurable: extreme, severe, moderate, minor)
- Alerts expire automatically based on expiration timestamp
- WebSocket push for new alerts matching user subscriptions

### Location Detection
- IP geolocation using embedded sapics/ip-location-db database
- Database updates monthly via scheduler
- Fallback to default location if IP not found
- User can override detected location

### User Accounts (Multi-User)
- Save up to 10 locations per user
- Subscribe to alerts for saved locations
- Email notifications for severe alerts (optional)
- No authentication required for basic API access

### Rate Limiting
- Anonymous: 20 requests/minute (per AI.md PART 1 defaults)
- Authenticated: 100 requests/minute (per AI.md PART 1 defaults)
- Admin: unlimited

## Endpoints

### Public Weather Endpoints
| Endpoint | Purpose |
|----------|---------|
| Current weather | Get current conditions for location |
| Forecast | Get 16-day forecast for location |
| Hourly forecast | Get 48-hour hourly forecast |
| Historical | Get weather for past date (up to 30 days) |

### Severe Weather Endpoints
| Endpoint | Purpose |
|----------|---------|
| Active alerts | Get all active alerts for region/country |
| Alert by ID | Get specific alert details |
| Alert history | Get past 7 days of alerts for region |

### Natural Events Endpoints
| Endpoint | Purpose |
|----------|---------|
| Earthquakes | Get recent earthquakes (configurable magnitude threshold) |
| Earthquake by ID | Get specific earthquake details |
| Active hurricanes | Get currently active tropical storms |
| Hurricane by ID | Get specific hurricane with forecast track |

### Astronomy Endpoints
| Endpoint | Purpose |
|----------|---------|
| Moon phase | Get current moon phase and illumination |
| Moon calendar | Get moon phases for month |
| Sun times | Get sunrise/sunset for location |

### Location Endpoints
| Endpoint | Purpose |
|----------|---------|
| Detect location | Get location from client IP |
| Search locations | Find cities by name |
| Reverse geocode | Get location from coordinates |

## Data Sources

| Data Type | Source | Update Frequency |
|-----------|--------|------------------|
| Weather Forecasts | Open-Meteo API | 15 minutes |
| US Severe Weather | NOAA Weather API | 5 minutes |
| Canada Alerts | Environment Canada | 5 minutes |
| UK Alerts | UK Met Office | 5 minutes |
| Australia Alerts | Australian BOM | 5 minutes |
| Japan Alerts | JMA (Japan Meteorological Agency) | 5 minutes |
| Mexico Alerts | CONAGUA | 5 minutes |
| Earthquakes | USGS Earthquake API | 1 minute |
| Hurricanes | NOAA National Hurricane Center | 15 minutes |
| Moon Phases | Calculated (no external API) | N/A |
| IP Geolocation | sapics/ip-location-db | Monthly |

## Notes

- All weather data sources are free/public APIs - no API keys required
- GeoIP database embedded in binary, updates downloaded at runtime
- Single binary deployment - no external dependencies
- WebSocket support for real-time alert notifications
- GraphQL endpoint for flexible queries
- Full OpenAPI/Swagger documentation
- i18n support: English, Spanish, French, German, Arabic, Japanese, Chinese
