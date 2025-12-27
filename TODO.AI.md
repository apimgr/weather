# TODO.AI.md - Weather API Implementation Tasks

This file tracks all remaining implementation tasks for the weather API project based on AI.md specification.

---

## PART 2: LICENSE & ATTRIBUTION

### Embedded License Attribution
- [ ] Run `go-licenses csv ./...` to identify all dependencies
- [ ] Add embedded licenses section to LICENSE.md with all third-party library licenses
- [ ] Ensure all MIT/Apache/BSD licenses are properly attributed

---

## PART 3: PROJECT STRUCTURE

### Directory Structure
- [x] src/ directory created
- [x] docker/ directory created
- [x] docs/ directory created (ReadTheDocs)
- [ ] Verify .gitignore includes binaries/, releases/, rootfs/
- [ ] Verify .dockerignore excludes .github/, docker/, binaries/, releases/
- [ ] Create binaries/ directory (gitignored)
- [ ] Ensure all paths work relative to project root

---

## PART 4: OS-SPECIFIC PATHS

### Path Detection & Handling
- [ ] Implement paths package (src/paths/paths.go)
- [ ] Add OS-specific path resolution (Linux, macOS, Windows, BSD)
- [ ] Support XDG Base Directory on Unix systems
- [ ] Support Windows AppData paths
- [ ] Add root vs user detection
- [ ] Add path validation and directory creation
- [ ] Test all path combinations on all platforms

---

## PART 5: CONFIGURATION

### Config Package Implementation
- [x] Basic config.go structure exists
- [ ] Add config.yml support with YAML parser
- [ ] Implement ParseBool with all truthy/falsy values (1/0, yes/no, true/false, on/off, enabled/disabled, etc.)
- [ ] Add environment variable override (WEATHER_* prefix)
- [ ] Add CLI flag override (highest priority)
- [ ] Implement live reload with file watcher
- [ ] Add configuration validation on load
- [ ] Support nested config keys (server.address, weather.api_key, etc.)
- [ ] Add default configuration file generation on first run

### Weather-Specific Configuration
- [ ] Add weather service settings to config
  - [ ] forecast_days (default: 7)
  - [ ] update_interval (default: 3600)
  - [ ] location_search_enabled (default: true)
  - [ ] multiuser_enabled (default: true)
  - [ ] open_registration (default: true)
- [ ] Add API integration settings
- [ ] Add external weather service credentials

---

## PART 6: APPLICATION MODES

### Mode Package
- [ ] Implement mode package (src/mode/mode.go)
- [ ] Add production mode (strict validation, caching, minimal logging)
- [ ] Add development mode (relaxed validation, verbose logging, no caching)
- [ ] Add mode detection (MODE env var, --mode flag, config)
- [ ] Add debug flag support (--debug, DEBUG env var)
- [ ] Implement mode-specific behaviors across app

---

## PART 7: SERVER BINARY CLI

### CLI Flags Implementation
- [ ] Implement --help flag (show usage)
- [ ] Implement --version flag (show version, commit, build date)
- [ ] Implement --mode flag (production/development)
- [ ] Implement --config flag (config directory path)
- [ ] Implement --data flag (data directory path)
- [ ] Implement --log flag (log directory path)
- [ ] Implement --pid flag (PID file path)
- [ ] Implement --address flag (listen address)
- [ ] Implement --port flag (listen port)
- [ ] Implement --status flag (show server status)
- [ ] Implement --daemon flag (daemonize server - Unix only)
- [ ] Implement --debug flag (enable debug mode)

### Directory & PID File Handling
- [ ] Add directory validation and creation
- [ ] Implement proper permissions (0755 root, 0700 user)
- [ ] Add PID file handling with stale detection
- [ ] Implement process verification (Unix: /proc, macOS: ps, Windows: API)
- [ ] Add platform-specific PID checking (pid_unix.go, pid_windows.go)

### Daemonization
- [ ] Implement Unix daemonization (fork, setsid, detach)
- [ ] Add service manager detection (systemd, launchd, sysv, runit, s6)
- [ ] Add container detection (Docker, Podman)
- [ ] Implement Windows service warning (use --service instead)
- [ ] Add parent process name detection

### Signal Handling
- [ ] Implement graceful shutdown on SIGTERM/SIGINT/SIGQUIT
- [ ] Add SIGRTMIN+3 handling for Docker
- [ ] Implement SIGUSR1 for log rotation (Unix only)
- [ ] Implement SIGUSR2 for status dump (Unix only)
- [ ] Add Windows os.Interrupt handling
- [ ] Implement cleanup sequence (stop accepting, wait for requests, close DB, flush logs)

---

## PART 8: UPDATE COMMAND

### Self-Update Functionality
- [ ] Implement --update check (check for updates)
- [ ] Implement --update yes (download and install update)
- [ ] Implement --update branch (switch release branch: stable/beta/daily)
- [ ] Add GitHub releases API integration
- [ ] Add binary verification (checksums, signatures)
- [ ] Add backup of current binary before update
- [ ] Add rollback capability if update fails
- [ ] Implement platform-specific binary download (OS + arch detection)

---

## PART 9: PRIVILEGE ESCALATION & SERVICE

### Service Support
- [ ] Implement --service install (install system service)
- [ ] Implement --service start (start service)
- [ ] Implement --service stop (stop service)
- [ ] Implement --service restart (restart service)
- [ ] Implement --service uninstall (uninstall service)
- [ ] Implement --service disable (disable service)
- [ ] Implement --service help (show service help)

### Service Manager Integration
- [ ] Add systemd service file generation (Linux)
- [ ] Add launchd plist generation (macOS)
- [ ] Add SysV init script generation (legacy Linux)
- [ ] Add rc.d script generation (BSD)
- [ ] Add Windows SCM integration (Windows Service)
- [ ] Add runit/s6 run script generation
- [ ] Auto-detect service manager on install

### Privilege Escalation
- [ ] Implement sudo/doas elevation for service operations
- [ ] Add privilege check before service operations
- [ ] Implement safe privilege drop after binding privileged ports

---

## PART 10: SERVICE SUPPORT (CONTINUED)

### Service Files
- [ ] Test systemd service creation and management
- [ ] Test launchd service on macOS
- [ ] Test Windows Service creation
- [ ] Verify service auto-start on boot
- [ ] Test service restart on failure

---

## PART 11: BINARY REQUIREMENTS

### Build Configuration
- [x] Makefile exists
- [ ] Verify CGO_ENABLED=0 for all builds
- [ ] Ensure static binaries (no external deps)
- [ ] Add version/commit/builddate ldflags
- [ ] Test builds for all platforms (Linux, macOS, Windows, BSD)
- [ ] Test builds for all architectures (amd64, arm64)
- [ ] Verify binary size optimization (-s -w flags)

---

## PART 12: MAKEFILE

### Makefile Targets
- [x] Basic Makefile exists
- [ ] Verify `make dev` target (quick build to temp dir)
- [ ] Verify `make build` target (all platforms)
- [ ] Verify `make release` target (with version tagging)
- [ ] Verify `make docker` target (build Docker image)
- [ ] Verify `make test` target (run all tests)
- [ ] Verify `make clean` target (remove build artifacts)
- [ ] Add `make install` target (install to system)
- [ ] Add `make uninstall` target (remove from system)

---

## PART 13: TESTING & DEVELOPMENT

### Test Infrastructure
- [x] tests/ directory exists
- [ ] Implement tests/run_tests.sh (auto-detect and run tests)
- [ ] Implement tests/docker.sh (beta testing with Docker)
- [ ] Implement tests/incus.sh (beta testing with Incus)
- [ ] Add unit tests for all packages
- [ ] Add integration tests
- [ ] Add e2e tests
- [ ] Ensure tests run in Docker (no Go on host)
- [ ] Add test coverage reporting
- [ ] Add benchmarks for critical paths

---

## PART 14: DOCKER

### Docker Configuration
- [x] docker/Dockerfile exists
- [x] docker/docker-compose.yml exists
- [x] docker/docker-compose.dev.yml exists
- [x] docker/docker-compose.test.yml exists
- [ ] Verify multi-stage Dockerfile (builder + runtime)
- [ ] Ensure all OCI labels are present
- [ ] Verify tor package is installed in image
- [ ] Check STOPSIGNAL SIGRTMIN+3
- [ ] Verify HEALTHCHECK configuration
- [ ] Verify tini init system
- [ ] Verify MODE=development in Dockerfile

### Entrypoint Script
- [ ] Verify docker/rootfs/usr/local/bin/entrypoint.sh exists
- [ ] Test signal handling (SIGTERM, SIGINT, SIGRTMIN+3)
- [ ] Test Tor auto-start when tor binary detected
- [ ] Test multi-service orchestration
- [ ] Test graceful shutdown sequence
- [ ] Test timezone configuration (TZ env var)
- [ ] Test DEBUG flag functionality
- [ ] Verify directory creation on startup

### Docker Compose
- [ ] Verify production compose has MODE=production
- [ ] Verify dev compose has DEBUG commented
- [ ] Verify test compose has DEBUG=true
- [ ] Test volume mounts (./rootfs/config, ./rootfs/data)
- [ ] Test port mapping configuration
- [ ] Verify custom network configuration
- [ ] Test SELinux :z labels on volumes

---

## PART 15: CI/CD WORKFLOWS

### GitHub Actions
- [ ] Create .github/workflows/release.yml (stable releases)
- [ ] Create .github/workflows/beta.yml (beta releases)
- [ ] Create .github/workflows/daily.yml (daily builds)
- [ ] Create .github/workflows/docker.yml (Docker multi-arch)
- [ ] Add version bumping automation
- [ ] Add changelog generation
- [ ] Add binary artifact upload
- [ ] Add Docker image push to GHCR
- [ ] Add manifest annotations for multi-arch images

### Gitea Actions (if applicable)
- [ ] Create .gitea/workflows/ (mirror of GitHub workflows)

---

## PART 16: HEALTH & VERSIONING

### Health Endpoints
- [ ] Implement /healthz (basic health check)
- [ ] Implement /api/v1/healthz (detailed health with all services)
- [ ] Add database connectivity check
- [ ] Add Tor service check (if enabled)
- [ ] Add scheduler status check
- [ ] Return proper status codes (200 OK, 503 Service Unavailable)
- [ ] Add uptime, version, commit info to health response

### Version Endpoint
- [ ] Implement version information endpoint
- [ ] Include version, commit, build date, Go version
- [ ] Add platform/arch information

---

## PART 17: WEB FRONTEND

### Frontend Infrastructure
- [ ] Create src/server/templates/ for HTML templates
- [ ] Implement project-wide theme system (light/dark/auto)
- [ ] Set dark theme as default
- [ ] Add theme toggle UI component
- [ ] Embed templates using embed.FS
- [ ] Create base layout template
- [ ] Add CSS files (NO inline CSS)
- [ ] Implement responsive design (mobile-first)

### Theme Application
- [ ] Apply theme to main web UI
- [ ] Apply theme to admin panel
- [ ] Apply theme to Swagger UI
- [ ] Apply theme to GraphQL Playground
- [ ] Apply theme to ReadTheDocs (via mkdocs.yml)
- [ ] Test theme persistence (localStorage)
- [ ] Test theme auto-detection (prefers-color-scheme)

### Accessibility
- [ ] Ensure WCAG 2.1 AA compliance
- [ ] Test with screen readers
- [ ] Add ARIA labels
- [ ] Ensure keyboard navigation works
- [ ] Test color contrast in both themes

---

## PART 18: SERVER CONFIGURATION

### Server Package
- [ ] Implement src/server/server.go (HTTP server setup)
- [ ] Add chi router configuration
- [ ] Implement middleware stack
- [ ] Add CORS support
- [ ] Add compression middleware
- [ ] Implement request ID middleware
- [ ] Add access logging middleware
- [ ] Add security headers middleware

---

## PART 19: ADMIN PANEL

### Admin UI Implementation
- [x] Some admin models exist
- [ ] Create admin panel templates
- [ ] Implement /admin route (session auth)
- [ ] Implement /api/v1/admin/ (bearer token auth)
- [ ] Add first-run setup wizard
- [ ] Create settings pages for all config sections
  - [ ] Server settings (address, port, mode)
  - [ ] User management settings
  - [ ] Weather service settings
  - [ ] Email/SMTP settings
  - [ ] Scheduler settings
  - [ ] SSL/TLS settings
  - [ ] Security settings
  - [ ] Backup settings
  - [ ] Tor settings
  - [ ] Custom domain settings (if applicable)
- [ ] Add tooltips/help text for all settings
- [ ] Implement live reload (apply changes without restart)
- [ ] Add validation with clear error messages
- [ ] Show defaults and current values

### Admin Authentication
- [x] Admin model exists
- [ ] Implement admin user creation (first-run)
- [ ] Add admin login/logout
- [ ] Add session management
- [ ] Add bearer token generation for API access
- [ ] Implement password reset
- [ ] Add 2FA support (TOTP)
- [ ] Add Passkey/WebAuthn support
- [ ] Generate recovery keys on MFA enable

---

## PART 20: API STRUCTURE

### REST API
- [ ] Implement /api/v1/weather (current weather)
- [ ] Implement /api/v1/weather/forecast (weather forecast)
- [ ] Implement /api/v1/weather/location (weather by location)
- [ ] Implement /api/v1/weather/search (search weather data)
- [ ] Add API versioning support
- [ ] Implement standard error responses
- [ ] Add pagination for list endpoints
- [ ] Add filtering and sorting
- [ ] Implement rate limiting per endpoint

### Swagger/OpenAPI
- [ ] Create src/swagger/swagger.go
- [ ] Generate OpenAPI 3.0 spec
- [ ] Serve Swagger UI at /openapi
- [ ] Serve OpenAPI JSON at /openapi.json (NO YAML)
- [ ] Apply project theme to Swagger UI
- [ ] Document all REST endpoints
- [ ] Add request/response examples
- [ ] Include authentication documentation

### GraphQL API
- [ ] Create src/graphql/graphql.go
- [ ] Implement GraphQL schema for weather data
- [ ] Add queries (weather, forecast, locations)
- [ ] Add mutations (if applicable)
- [ ] Sync GraphQL with REST API (same data)
- [ ] Serve GraphQL Playground at /graphql
- [ ] Apply project theme to GraphQL Playground
- [ ] Add authentication to GraphQL

---

## PART 21: SSL/TLS & LET'S ENCRYPT

### SSL Package
- [ ] Create src/ssl/ package
- [ ] Implement Let's Encrypt support (HTTP-01, TLS-ALPN-01, DNS-01)
- [ ] Add manual certificate support
- [ ] Implement automatic renewal via scheduler
- [ ] Add certificate validation
- [ ] Support multiple domains
- [ ] Add HSTS when SSL enabled
- [ ] Store certificates in config/security/

---

## PART 22: SECURITY & LOGGING

### Security Implementation
- [ ] Implement all security headers (CSP, X-Frame-Options, etc.)
- [ ] Add rate limiting middleware
- [ ] Implement IP-based blocking
- [ ] Add CSRF protection for forms
- [ ] Implement audit logging
- [ ] Add security event logging
- [ ] Add login attempt tracking
- [ ] Implement account lockout after failed attempts

### Logging System
- [ ] Implement structured logging
- [ ] Create access.log (HTTP requests)
- [ ] Create server.log (application events)
- [ ] Create error.log (errors and warnings)
- [ ] Create audit.log (security events, JSON only)
- [ ] Create security.log (auth, permissions)
- [ ] Add log rotation support
- [ ] Implement configurable log levels
- [ ] Add log format configuration (JSON, text)

---

## PART 23: USER MANAGEMENT

### User System
- [x] User model exists
- [x] Session model exists
- [ ] Implement user registration (open by default)
- [ ] Add user login/logout
- [ ] Implement password hashing (Argon2)
- [ ] Add legacy bcrypt verification
- [ ] Implement email verification
- [ ] Add password reset flow
- [ ] Implement user profile management
- [ ] Add user roles and permissions

### Multi-User Features
- [ ] Enable multi-user mode by default (weather.multiuser_enabled: true)
- [ ] Enable open registration by default (weather.open_registration: true)
- [ ] Add user management UI in admin panel
- [ ] Implement user search and filtering
- [ ] Add user suspension/deletion
- [ ] Implement user API tokens

### Authentication Methods
- [x] TOTP model may exist
- [ ] Implement TOTP 2FA
- [ ] Implement Passkey/WebAuthn
- [ ] Add LDAP/Active Directory support
- [ ] Add OAuth2 providers (optional)
- [ ] Add OpenID Connect support (optional)
- [ ] Implement session management with configurable timeout

---

## PART 24: DATABASE & CLUSTER

### Database Implementation
- [ ] Create src/database/ package
- [ ] Implement SQLite support (modernc.org/sqlite)
- [ ] Add PostgreSQL support (github.com/jackc/pgx/v5)
- [ ] Add MySQL support (github.com/go-sql-driver/mysql)
- [ ] Implement auto schema creation (CREATE TABLE IF NOT EXISTS)
- [ ] Add automatic migrations on startup
- [ ] Implement connection pooling
- [ ] Add database health checks

### Cluster Support
- [ ] Implement cluster mode configuration
- [ ] Add node discovery
- [ ] Implement config synchronization
- [ ] Add distributed task locking (via Redis/Valkey)
- [ ] Test multi-node deployment
- [ ] Add cluster health monitoring

---

## PART 25: BACKUP & RESTORE

### Backup System
- [ ] Implement automated daily backups
- [ ] Add AES-256-GCM encryption for backups
- [ ] Configure backup retention (max 4 by default)
- [ ] Store backups in data/backup/
- [ ] Add backup scheduling via scheduler
- [ ] Implement backup file rotation

### Restore Functionality
- [ ] Implement --maintenance backup (manual backup)
- [ ] Implement --maintenance restore (restore from backup)
- [ ] Add backup verification
- [ ] Add restore validation
- [ ] Support backup encryption password
- [ ] Test backup/restore end-to-end

---

## PART 26: EMAIL & NOTIFICATIONS

### Email System
- [ ] Implement SMTP configuration
- [ ] Add email template system
- [ ] Create customizable email templates
- [ ] Implement email verification emails
- [ ] Add password reset emails
- [ ] Add notification emails
- [ ] Test email delivery
- [ ] Handle missing SMTP gracefully (disable email features, don't error)

### Notification System
- [x] Notification model exists
- [ ] Implement in-app notifications
- [ ] Add WebUI notification display
- [ ] Implement notification preferences
- [ ] Add notification history
- [ ] Test notification delivery

---

## PART 27: SCHEDULER

### Scheduler Implementation
- [ ] Create src/scheduler/ package using github.com/robfig/cron/v3
- [ ] Implement daily backup task (02:00)
- [ ] Add SSL renewal task (03:00 daily)
- [ ] Add GeoIP update task (03:00 Sunday)
- [ ] Add session cleanup task (hourly)
- [ ] Implement weather data update task (configurable interval)
- [ ] Add cluster-aware task locking
- [ ] Implement scheduler health monitoring
- [ ] Add scheduler configuration in admin panel

---

## PART 28: GEOIP

### GeoIP Implementation
- [ ] Download ip-location-db on first run
- [ ] Store GeoIP database in config/security/
- [ ] Implement country-based blocking (deny_countries)
- [ ] Add GeoIP update via scheduler (weekly)
- [ ] Add country management in admin panel
- [ ] Implement IP lookup functionality
- [ ] Add GeoIP-based features to weather location queries

---

## PART 29: METRICS

### Metrics System
- [ ] Implement Prometheus metrics endpoint
- [ ] Add standard Go metrics (memory, goroutines, etc.)
- [ ] Add custom metrics (API requests, weather queries, etc.)
- [ ] Add database query metrics
- [ ] Implement request duration histograms
- [ ] Add counter metrics for all endpoints
- [ ] Test metrics collection and export

---

## PART 30: TOR HIDDEN SERVICE

### Tor Integration
- [ ] Detect tor binary installation
- [ ] Auto-enable Tor if binary found
- [ ] Configure dedicated tor process
- [ ] Generate .onion address
- [ ] Store Tor keys in data/tor/
- [ ] Display .onion address in admin panel
- [ ] Implement vanity address generation (optional)
- [ ] Add Tor health monitoring

---

## PART 31: ERROR HANDLING & CACHING

### Error Handling
- [ ] Implement standard error response format
- [ ] Add error code system
- [ ] Implement context-aware error messages (user vs admin vs log)
- [ ] Add panic recovery middleware
- [ ] Implement error logging with stack traces
- [ ] Add user-friendly error pages

### Caching
- [ ] Implement caching for weather data
- [ ] Add cache invalidation
- [ ] Support Redis/Valkey for distributed caching
- [ ] Add in-memory cache for single node
- [ ] Implement cache TTL configuration
- [ ] Add cache hit/miss metrics

---

## PART 32: I18N & A11Y

### Internationalization
- [ ] Create src/locales/ directory
- [ ] Add English (en-US) locale (default)
- [ ] Implement translation system
- [ ] Add language selector in UI
- [ ] Translate all user-facing strings
- [ ] Add locale detection from Accept-Language header

### Accessibility
- [ ] Ensure WCAG 2.1 AA compliance
- [ ] Add skip navigation links
- [ ] Ensure all images have alt text
- [ ] Add ARIA labels for interactive elements
- [ ] Test keyboard navigation
- [ ] Test with screen readers (NVDA, JAWS, VoiceOver)
- [ ] Verify color contrast in both themes
- [ ] Add focus indicators
- [ ] Ensure form validation is accessible

---

## PART 33: READTHEDOCS DOCUMENTATION

### Documentation Setup
- [x] docs/ directory exists
- [x] mkdocs.yml exists
- [ ] Verify .readthedocs.yaml configuration
- [ ] Add docs/requirements.txt with mkdocs dependencies
- [ ] Configure MkDocs Material theme
- [ ] Set dark theme as default in mkdocs.yml
- [ ] Add theme toggle (light/dark/auto)

### Documentation Pages
- [ ] Write docs/index.md (overview, features)
- [ ] Write docs/installation.md (Docker, binary, from source)
- [ ] Write docs/configuration.md (all config options)
- [ ] Write docs/api.md (REST, GraphQL, Swagger)
- [ ] Write docs/admin.md (admin panel guide)
- [ ] Write docs/cli.md (CLI client guide, if applicable)
- [ ] Write docs/development.md (contributing, building, testing)
- [ ] Add docs/stylesheets/dark.css (optional customization)
- [ ] Add docs/stylesheets/light.css (optional customization)

### ReadTheDocs Integration
- [ ] Set up ReadTheDocs project
- [ ] Configure webhook for auto-builds
- [ ] Verify documentation builds successfully
- [ ] Test all documentation links
- [ ] Add documentation badge to README.md

---

## PART 34: CLI CLIENT (OPTIONAL)

### CLI Client Decision
- [ ] Decide if weather API needs CLI client
  - Weather API with complex queries → YES (power users need CLI)
  - Simple web-only app → SKIP

### CLI Client Implementation (if YES)
- [ ] Create cmd/weather-cli/main.go
- [ ] Implement --server flag (API endpoint)
- [ ] Implement --token flag (API authentication)
- [ ] Implement --output flag (json, yaml, table)
- [ ] Implement --tui flag (Terminal UI mode)
- [ ] Add TUI with dark theme matching project theme
- [ ] Implement weather query commands
- [ ] Add forecast commands
- [ ] Add location search commands
- [ ] Create cli.yml configuration file (~/.config/weather/cli.yml)
- [ ] Build alongside server in Makefile
- [ ] Add CLI client documentation

---

## PART 35: CUSTOM DOMAINS (OPTIONAL)

### Custom Domains Decision
- [ ] Decide if weather API needs custom domains
  - Multi-tenant SaaS → YES
  - Single-instance API → SKIP
  - Weather data service → LIKELY SKIP

### Custom Domains Implementation (if YES)
- [x] Domain model exists
- [ ] Add domain verification system
- [ ] Implement DNS validation
- [ ] Add SSL certificate per domain
- [ ] Create domain management UI in admin panel
- [ ] Add domain-based routing
- [ ] Implement domain health checks

---

## PART 36: PROJECT-SPECIFIC SECTIONS

### Weather Service Implementation
- [ ] Create src/service/weather.go
- [ ] Implement current weather data retrieval
- [ ] Add weather forecast functionality (7 days default)
- [ ] Implement location-based weather queries
- [ ] Add weather search functionality
- [ ] Create weather.json data file (src/data/weather.json)
- [ ] Create locations.json database (src/data/locations.json)
- [ ] Implement automatic weather data updates (3600s interval)
- [ ] Add weather API integration (external service)
- [ ] Configure API keys in config

### Data Files
- [ ] Create src/data/ directory
- [ ] Add weather.json with sample data
- [ ] Add locations.json with location database
- [ ] Implement data loading on startup
- [ ] Add data validation

---

## FINAL CHECKPOINT: COMPLIANCE VERIFICATION

### Document Compliance
- [x] AI.md exists and is complete
- [ ] TODO.AI.md tracks all remaining tasks
- [x] README.md exists
- [ ] Verify README.md follows spec order (About, Features, Production, CLI, Config, API, Dev)
- [x] LICENSE.md exists with MIT license
- [ ] Verify embedded licenses section in LICENSE.md

### Build Compliance
- [ ] Verify CGO_ENABLED=0 for all builds
- [ ] Test builds on all 4 OSes (Linux, macOS, Windows, BSD)
- [ ] Test builds for both architectures (amd64, arm64)
- [ ] Verify single static binary with embedded assets
- [ ] Test no external runtime dependencies

### Configuration Compliance
- [ ] Verify server.yml format (not .yaml)
- [ ] Test WEATHER_* environment variables
- [ ] Verify CLI flags override env vars
- [ ] Verify env vars override file config
- [ ] Test ParseBool with all truthy/falsy values
- [ ] Verify all settings have sane defaults

### Frontend Compliance
- [ ] Verify project-wide theme system (light/dark/auto)
- [ ] Confirm dark theme is default
- [ ] Verify NO inline CSS
- [ ] Verify NO JavaScript alerts (use toast notifications)
- [ ] Test mobile-first responsive design
- [ ] Verify WCAG 2.1 AA compliance in both themes

### API Compliance
- [ ] REST API at /api/v1/
- [ ] Swagger UI at /openapi
- [ ] OpenAPI spec at /openapi.json (JSON only)
- [ ] GraphQL at /graphql
- [ ] Health check at /healthz
- [ ] All APIs in sync (same data)

### Admin Panel Compliance
- [ ] Web UI at /admin (session auth)
- [ ] API at /api/v1/admin/ (bearer token)
- [ ] First-run setup wizard
- [ ] All settings configurable via UI (100% coverage)
- [ ] Live reload without restart
- [ ] Tooltips on all settings

### Server CLI Compliance
- [ ] --help, --version (no privileges)
- [ ] --status (show running status)
- [ ] --config, --data, --log, --pid
- [ ] --service (install/start/stop/restart/uninstall)
- [ ] --maintenance (backup/restore/setup)
- [ ] --update (check/yes/branch)
- [ ] --daemon (Unix only)
- [ ] --debug (enable debug mode)

### Database Compliance
- [ ] SQLite for local/single-node (default)
- [ ] PostgreSQL/MySQL for cluster
- [ ] Self-creating schema
- [ ] Automatic migrations
- [ ] modernc.org/sqlite (NO mattn/go-sqlite3)

### Scheduler Compliance
- [ ] Built-in scheduler always running
- [ ] Backup: 02:00 daily
- [ ] SSL renewal: 03:00 daily
- [ ] GeoIP update: 03:00 Sunday
- [ ] Session cleanup: hourly
- [ ] Weather updates: configurable interval
- [ ] Cluster-aware task locking

### Backup Compliance
- [ ] Automatic daily backups
- [ ] AES-256-GCM encryption
- [ ] Max 4 backups retained
- [ ] Restore via --maintenance restore

### Email Compliance
- [ ] SMTP required for email
- [ ] No SMTP = disable email (no hidden errors)
- [ ] Customizable templates
- [ ] WebUI notifications always available

### SSL/TLS Compliance
- [ ] Let's Encrypt support (HTTP-01, TLS-ALPN-01, DNS-01)
- [ ] Manual certificate support
- [ ] Auto-renewal via scheduler
- [ ] HSTS when SSL enabled

### Security Compliance
- [ ] All security headers
- [ ] Rate limiting enabled
- [ ] Audit logging
- [ ] 2FA support (TOTP, WebAuthn)
- [ ] Session management
- [ ] GeoIP blocking

### Logging Compliance
- [ ] access.log, server.log, error.log, audit.log, security.log
- [ ] Configurable formats and rotation
- [ ] Audit log: JSON only, daily rotation
- [ ] Structured logging

### Build & Deploy Compliance
- [ ] Makefile: build, release, docker, test, dev, clean
- [ ] CI/CD: release, beta, daily, docker workflows
- [ ] 8 platform builds (4 OS × 2 arch)
- [ ] Docker: Alpine base, tini init, non-root user
- [ ] Multi-stage Dockerfile

### Documentation Compliance
- [ ] docs/ directory ONLY for ReadTheDocs
- [ ] mkdocs.yml in project root
- [ ] .readthedocs.yaml in project root
- [ ] docs/requirements.txt
- [ ] MkDocs Material theme with light/dark/auto
- [ ] Required pages: index, installation, configuration, api, admin, development
- [ ] ReadTheDocs URL configured
- [ ] Documentation badge in README.md

### Tor Compliance
- [ ] Auto-enabled when tor binary found
- [ ] Dedicated tor process
- [ ] .onion address management
- [ ] Vanity address generation

### Weather-Specific Compliance
- [ ] Multi-user enabled by default
- [ ] Open registration enabled by default
- [ ] Forecast days: 7 (configurable)
- [ ] Update interval: 3600s (configurable)
- [ ] Location search enabled
- [ ] Weather API endpoints functional
- [ ] GraphQL weather queries functional

---

## SUMMARY

**Total Tasks:** 400+

**Categories:**
- Core Infrastructure: ~80 tasks
- Web/Frontend/API: ~70 tasks
- Security/Auth/Users: ~60 tasks
- Database/Backup/Scheduler: ~45 tasks
- Docker/CI/CD: ~35 tasks
- Documentation: ~30 tasks
- Testing: ~20 tasks
- Weather-Specific: ~25 tasks
- Optional (CLI Client, Custom Domains): ~20 tasks
- Compliance Verification: ~50 tasks

---

## FIRST 10 PRIORITY TASKS

These are the critical path items to start with:

1. **Implement src/paths/paths.go** - OS-specific path detection (Linux, macOS, Windows, BSD)
2. **Complete src/config/config.go** - YAML parsing, ParseBool, env vars, CLI flags, live reload
3. **Implement src/mode/mode.go** - Production/development mode detection and behaviors
4. **Complete server CLI flags** - All --help, --version, --config, --data, --log, --pid, --status, --debug
5. **Implement PID file handling** - Stale detection, process verification, platform-specific
6. **Create weather service** (src/service/weather.go) - Core weather data functionality
7. **Implement REST API endpoints** - /api/v1/weather, forecast, location, search
8. **Complete admin panel** - First-run setup, all settings pages with live reload
9. **Implement database schema** - Auto-creation, migrations for weather data
10. **Create Dockerfile and entrypoint** - Multi-stage build, signal handling, Tor integration

---

## NOTES

- This is a COMPLETE implementation from scratch following AI.md specification
- Multi-user and open registration are ENABLED by default per spec
- All NON-NEGOTIABLE sections MUST be implemented exactly as specified
- Weather-specific features build on top of standard template
- Tor is auto-enabled when binary is installed (included in Docker image)
- Theme system is project-wide: dark default, light/auto options
- CLI client decision pending (likely YES for power users)
- Custom domains likely not needed for weather API (single-instance service)

---

**END OF TODO.AI.md**
