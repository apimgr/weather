# Weather Service - AI.md Compliance TODO

## Project Info
- **Project Name**: weather
- **Organization**: apimgr
- **Template Version**: Fresh copy from TEMPLATE.md 2026-02-03 (REFRESHED)
- **AI.md Location**: /root/Projects/github/apimgr/weather/AI.md
- **Official Site**: https://wthr.top
- **AI.md Refresh**: 2026-02-03 - Placeholders replaced (weather/apimgr), rules verified

---

## CRITICAL RULES (Committed to Memory)

### NEVER Do These (15 Rules)
1. Use bcrypt → Use Argon2id
2. Put Dockerfile in root → `docker/Dockerfile`
3. Use CGO → CGO_ENABLED=0 always
4. Hardcode dev values → Detect at runtime
5. Use external cron → Internal scheduler (PART 19)
6. Store passwords plaintext → Argon2id (tokens use SHA-256)
7. Create premium tiers → All features free
8. Use Makefile in CI/CD → Explicit commands
9. Guess or assume → Read spec or ask
10. Skip platforms → Build all 8
11. Use strconv.ParseBool() → Use config.ParseBool()
12. Run Go locally → Use containers only (make dev/test/build)
13. Client-side rendering (React/Vue) → Server-side Go templates
14. Require JavaScript for core features → Progressive enhancement only
15. Let long strings break mobile → Use word-break CSS

### MUST Do These (12 Rules)
1. Re-read spec before each task
2. Use containers for all builds/tests
3. Use config.ParseBool() for ALL boolean parsing
4. Use Argon2id for passwords, SHA-256 for tokens
5. Use path normalization/validation (prevent traversal)
6. Support all 4 OSes and 2 architectures (8 binaries)
7. Use server.yml (not .yaml)
8. Keep documentation in sync with code
9. Have admin WebUI for ALL settings
10. Have corresponding API endpoint for every web page
11. Use single static binary (all assets embedded)
12. Detect machine settings at runtime

### COMMIT Rules
1. AI cannot run git add/commit/push - write to .git/COMMIT_MESS instead
2. COMMIT_MESS must accurately reflect actual git status
3. Use emoji prefixes for commit types (✨ feat, 🐛 fix, 📝 docs, etc.)
4. Title max 64 chars including emojis
5. Never commit without verification

---

## Implementation Status

### Core Weather Features (100% Complete)

- [x] **Weather Forecasts** - 16-day global forecasts via Open-Meteo API
  - GET /api/v1/weather - Current weather + today's forecast
  - GET /api/v1/weather/:location - Weather by location path
  - GET /api/v1/forecasts - Multi-day forecast
  - GET /api/v1/weather/history - Historical weather (30 days)
  - 15-minute cache per AI.md spec
  - Supports: city name, ZIP code, coordinates, auto-detect by IP

- [x] **Severe Weather Alerts** - Real-time alerts from 6 countries
  - GET /api/v1/severe-weather - All active alerts (50-mile default filter)
  - GET /api/v1/severe-weather/:id - Specific alert details
  - Sources: NOAA (US), Environment Canada, UK Met Office, BOM (Australia), JMA (Japan), CONAGUA (Mexico)
  - 5-minute polling, deduplication by ID, severity filtering

- [x] **Hurricane Tracking** - Active tropical storms from NOAA NHC
  - GET /api/v1/hurricanes - All active storms
  - GET /api/v1/hurricanes/:id - Specific storm with forecast track
  - Response: name, category, wind speed, pressure, movement, forecast points
  - 15-minute cache

- [x] **Earthquake Data** - Real-time seismic activity from USGS
  - GET /api/v1/earthquakes - Recent earthquakes with magnitude filtering
  - GET /api/v1/earthquakes/:id - Specific earthquake details
  - Feed types: all_hour, all_day, all_week, all_month, magnitude thresholds
  - Response: magnitude, location, depth, tsunami risk, URL

- [x] **Moon Phases** - Lunar cycles with astronomical calculations
  - GET /api/v1/moon - Current moon data for location
  - GET /api/v1/moon/calendar - Month moon phases
  - GET /api/v1/sun - Sunrise/sunset times
  - Pure calculation (Meeus algorithms), no external API

- [x] **Location Services** - Full geolocation support
  - GET /api/v1/locations/search - Search cities by name
  - GET /api/v1/locations/lookup/zip/:code - US ZIP code lookup
  - GET /api/v1/locations/lookup/coords - Reverse geocoding
  - GET /api/v1/ip - Client IP detection
  - Embedded GeoIP database (sapics/ip-location-db)
  - Monthly database updates via scheduler

### API & Integration (100% Complete)

- [x] **REST API** - 50+ endpoints under /api/v1/
- [x] **GraphQL API** - Full query/mutation support
  - Weather, forecast, historicalWeather, moonPhase queries
  - Earthquakes, hurricanes, severeWeather queries
  - Location queries (search, IP geolocation, ZIP lookup, reverse geocode)
  - User preference and saved location mutations
  - Admin mutations (backup, restore, user management)
- [x] **OpenAPI/Swagger** - GET /openapi, /openapi.json
- [x] **API Autodiscovery** - GET /api/autodiscover (1-hour cache)

### Web Interface (100% Complete)

- [x] **Weather Pages** - / and /:location with auto-detect
- [x] **Moon Pages** - /moon and /moon/:location
- [x] **Earthquake Map** - /earthquakes
- [x] **Hurricane Tracking** - /hurricanes
- [x] **Severe Weather Viewer** - /severe-weather
- [x] **Content Negotiation** - HTML for browsers, text for CLI (Accept header)

### CLI Client (90% Complete)

- [x] **Commands**: current, forecast, alerts, moon, history
- [x] **Auto-detection**: TUI vs CLI mode (no flags per AI.md PART 33)
- [x] **Configuration**: YAML profiles (server, auth, output, color)
- [x] **Cross-platform paths**: Unix XDG, Windows AppData
- [x] **Flags**: --server, --token, --token-file, --user, --config, --output, --debug, --color
- [x] **Shell completions**: --shell completions [bash|zsh]
- [x] **Output formats**: json, table, plain, yaml, csv
- [x] **NO_COLOR support**: Per spec
- [x] **TUI**: Window-aware, vim navigation, Dracula theme

### Infrastructure (100% Complete)

- [x] **Scheduler** - 19 scheduled tasks
  - rotate-logs (24h), cleanup-sessions (15m), cleanup-tokens (15m)
  - cleanup-rate-limits (1h), cleanup-audit-logs (24h)
  - check-weather-alerts (5m), daily-forecast (24h)
  - process-notification-queue (2m), cleanup-notifications (24h)
  - backup-daily (24h), backup-hourly (1h)
  - ssl-renewal (24h), healthcheck-self (5m), tor-health (10m)
  - refresh-weather-cache (15m), update-geoip-database (7d)
  - blocklist-update (24h), cve-update (24h), cluster-heartbeat (30s)

- [x] **Health Checks** - Comprehensive /healthz
  - Database ping + disk space (85% warning, 95% critical)
  - Tor service status monitoring
  - SSL certificate expiry check (7-day warning)
  - Platform-specific disk usage (disk_unix.go, disk_windows.go)

- [x] **Metrics** - Prometheus /metrics endpoint
- [x] **Rate Limiting** - Anonymous 20/min, authenticated 100/min
- [x] **Admin Panel** - Full WebUI for all settings
- [x] **Backup/Restore** - Automated daily/hourly with encryption
- [x] **SSL/TLS** - Let's Encrypt support
- [x] **Tor Hidden Service** - Auto-enabled when Tor found

### Data Services (100% Complete)

- [x] **WeatherService** - Open-Meteo integration with caching
- [x] **SevereWeatherService** - Multi-source alert aggregation
- [x] **EarthquakeService** - USGS GeoJSON API
- [x] **HurricaneService** - NOAA NHC JSON API
- [x] **MoonService** - Pure astronomical calculations
- [x] **LocationEnhancer** - Country/city/timezone enhancement
- [x] **ZipcodeService** - US postal code lookup with Reload()
- [x] **GeoIPService** - IP geolocation with monthly updates

### GraphQL Resolvers (100% Complete)

- [x] Weather, Forecast, HistoricalWeather - use WeatherService
- [x] MoonPhase - uses MoonService
- [x] Earthquakes - uses EarthquakeService
- [x] Hurricanes - uses HurricaneService
- [x] SevereWeather - uses SevereWeatherService
- [x] SearchLocations, LookupCoordinates - use WeatherService
- [x] IPGeolocation - uses GeoIPService via WeatherService.LookupIP()
- [x] CurrentLocation - auto-detects client IP from context
- [x] LookupZipCode - uses ZipcodeService via WeatherService.LookupZipcode()

### Notification System (100% Complete)

- [x] **WebSocket Alerts** - Real-time push notifications
  - Endpoint: /ws/notifications
  - WebSocketHub with client management
  - BroadcastToUser/BroadcastToAdmin methods
  - Ping/pong heartbeat, stale connection cleanup
  - Full integration with NotificationService

- [x] **Email Notifications** - SMTP email delivery
  - SMTPService with 40+ provider presets
  - EmailChannel implementing NotificationChannel interface
  - DeliverySystem with queue, retry logic, dead letter handling
  - Weather alert templates: weather_alert.txt, weather_alert_update.txt, weather_alert_expired.txt
  - Scheduler: check-weather-alerts (5m), process-notification-queue (2m)

- [x] **i18n Support** - Internationalization
  - I18nService with T(), ParseAcceptLanguage()
  - 7 locale files per IDEA.md: en, es, fr, de, ar, ja, zh
  - 170+ translation keys for weather, alerts, locations, UI
  - Integrated via middleware, template {{ t .Lang "key" }}

---

## Remaining Work

### Minor Items

- [x] **Bash Shell Function** - COMPLETE
  - Shell completions exist (--shell completions [bash|zsh])
  - /:bash.function endpoint implemented in src/server/handler/weather.go:536
  - Provides wttr() and w() bash functions for terminal weather

---

## Uncommitted Changes

**Status**: Ready to commit - full AI.md compliance update

New files (AI.md compliance):
- LICENSE - MIT license file (AI.md required root file)
- CLAUDE.md - Project memory file (AI.md PART 0 required)
- .claude/settings.json - Shared team settings (AI.md PART 0)
- .claude/rules/ai-rules.md - PART 0, 1
- .claude/rules/project-rules.md - PART 2, 3, 4
- .claude/rules/config-rules.md - PART 5, 6, 12
- .claude/rules/binary-rules.md - PART 7, 8, 33
- .claude/rules/backend-rules.md - PART 9, 10, 11, 32
- .claude/rules/api-rules.md - PART 13, 14, 15
- .claude/rules/frontend-rules.md - PART 16, 17
- .claude/rules/features-rules.md - PART 18-23
- .claude/rules/service-rules.md - PART 24, 25
- .claude/rules/makefile-rules.md - PART 26
- .claude/rules/docker-rules.md - PART 27
- .claude/rules/cicd-rules.md - PART 28
- .claude/rules/testing-rules.md - PART 29, 30, 31
- .claude/rules/optional-rules.md - PART 34, 35, 36

CI/CD fixes (AI.md PART 28 compliance):
- .github/workflows/daily.yml - Added all 8 platforms + .exe extension + OfficialSite LDFLAGS
- .github/workflows/beta.yml - Added OfficialSite LDFLAGS
- .github/workflows/release.yml - Added OfficialSite LDFLAGS
- .gitea/workflows/daily.yml - Added all 8 platforms + .exe extension + OfficialSite LDFLAGS
- .gitea/workflows/beta.yml - Added OfficialSite LDFLAGS
- .gitea/workflows/release.yml - Added OfficialSite LDFLAGS

Files modified:
- TODO.AI.md - Updated to reflect completed work
- .gitignore - Fixed to track .claude/rules/ and .claude/settings.json
- Makefile - Added OfficialSite to LDFLAGS
- src/main.go - Fixed duplicate smtpService declaration (line 632: := to =)
- src/scheduler/scheduler.go - Implemented UpdateBlocklist() and UpdateCVEDatabase() (removed TODO stubs)

PART 34 User Settings Routes (AI.md PART 34 compliance):
- src/server/handler/user_settings.go - NEW - User settings handler with all API endpoints
- src/server/handler/user_public.go - NEW - Public profile, avatar, password change handler
- src/server/templates/pages/user/settings.tmpl - NEW - Account settings page
- src/server/templates/pages/user/settings-privacy.tmpl - NEW - Privacy settings page
- src/server/templates/pages/user/settings-notifications.tmpl - NEW - Notification settings page
- src/server/templates/pages/user/settings-appearance.tmpl - NEW - Appearance settings page
- src/server/templates/pages/user/settings-tokens.tmpl - NEW - API tokens page
- src/main.go - Added userSettingsHandler + userPublicHandler + routes
  - Web routes: /users/settings, /users/settings/privacy, /users/settings/notifications, /users/settings/appearance, /users/tokens
  - API routes: GET/PATCH /api/v1/users/settings, GET/POST/DELETE /api/v1/users/tokens
  - Public profile: GET /api/v1/public/users/:username
  - Avatar: GET/POST/PATCH/DELETE /api/v1/users/avatar
  - Password change: POST /api/v1/users/security/password

Previously created:
- src/locales/*.json - 7 locale files (en, es, fr, de, ar, ja, zh)
- src/server/service/email_channel.go - EmailChannel implementation
- src/email/templates/weather_alert*.txt - 3 email templates
- src/main.go - EmailChannel registration
- src/server/service/smtp.go - IsEnabled() method
- src/graphql/schema.resolvers.go - All resolvers
- src/scheduler/*.go - Enhanced tasks, disk monitoring
- src/server/handler/health*.go - Tor status, disk monitoring
- src/client/*_test.go - Updated tests

---

## Working Notes

- Container-only development - NEVER run go locally
- Use `make dev` for quick builds, `make test` for tests
- Test binaries in Docker/Incus, never on host
- AI.md is SOURCE OF TRUTH - always re-read relevant PART before implementing
- Build: PASSING | Tests: ALL PASSING
- Last verification: 2026-02-03
  - All 14 .claude/rules/*.md files present
  - No TODO/FIXME/XXX comments in src/
  - No bcrypt usage (Argon2id correctly used)
  - No strconv.ParseBool (config.ParseBool used)
  - modernc.org/sqlite used (not mattn/go-sqlite3)
  - CI/CD workflows: all 8 platforms + OfficialSite LDFLAGS
  - Platform-specific disk monitoring (disk_unix.go, disk_windows.go)
  - PART 34 (Multi-User) compliance verified and updated:
    - Database schema updated with profile fields (visibility, avatar, bio, website, etc.)
    - User model updated with all PART 34 fields
    - Username validation and blocklist implemented
    - Registration modes (public/private/disabled) implemented
    - 2FA, passkeys, OIDC, LDAP support in schema
