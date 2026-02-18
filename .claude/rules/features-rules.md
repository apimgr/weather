# Features Rules (PART 18-23)

⚠️ **These rules are NON-NEGOTIABLE. Violations are bugs.** ⚠️

## CRITICAL - NEVER DO
- ❌ Use external cron → use internal scheduler
- ❌ Skip email verification
- ❌ Store backups unencrypted
- ❌ Skip GeoIP for geolocation

## REQUIRED - ALWAYS DO
- ✅ Built-in scheduler for all background tasks
- ✅ SMTP email with 40+ provider presets
- ✅ GeoIP database with monthly updates
- ✅ Prometheus metrics on /metrics
- ✅ Encrypted backups with compression
- ✅ Self-update capability

## EMAIL (PART 18)
- SMTP auto-detection from domain MX records
- 40+ provider presets (Gmail, Outlook, etc.)
- Template-based emails
- Queue with retry logic

## SCHEDULER (PART 19)
- Internal cron-like scheduler
- NO external cron/systemd timers
- Persistent task history
- Admin panel management

## GEOIP (PART 20)
- Embedded ip-location-db (sapics)
- Monthly database updates
- Country blocking support
- IP geolocation API

## METRICS (PART 21)
- Prometheus format on /metrics
- INTERNAL only (not public)
- Bearer token authentication
- All app metrics exposed

## BACKUP & RESTORE (PART 22)
- Automated daily/hourly backups
- AES-256-GCM encryption
- Gzip compression
- Config + database + uploads
- Cluster-aware (coordinator backup)

## UPDATE (PART 23)
- `--update check` - check for updates
- `--update yes` - download and apply
- `--update branch stable|beta|daily`
- Semantic versioning
- Rollback support

---
**Full details: AI.md PART 18, PART 19, PART 20, PART 21, PART 22, PART 23**
