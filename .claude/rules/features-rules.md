# Features Rules (PART 18-23)

⚠️ **These rules are NON-NEGOTIABLE. Violations are bugs.** ⚠️

## CRITICAL - NEVER DO
- ❌ Use external cron/Task Scheduler (built-in scheduler only)
- ❌ Skip email validation before sending
- ❌ Store GeoIP database in repo (download at runtime)
- ❌ Expose Prometheus metrics without auth option
- ❌ Create backups without encryption option
- ❌ Skip update verification (checksum required)

## REQUIRED - ALWAYS DO
- ✅ Built-in scheduler for all background tasks (PART 19)
- ✅ Email via SMTP with provider presets (PART 18)
- ✅ GeoIP with auto-update (sapics/ip-location-db) (PART 20)
- ✅ Prometheus metrics at /metrics (PART 21)
- ✅ Automated backup/restore with encryption (PART 22)
- ✅ Self-update with checksum verification (PART 23)

## SCHEDULER (PART 19)
Internal cron-like scheduler using github.com/robfig/cron/v3

Common tasks:
- Log rotation (daily)
- Session cleanup (15min)
- Token cleanup (15min)
- Backup (daily/hourly)
- GeoIP update (weekly)
- SSL renewal check (daily)
- Health self-check (5min)

## EMAIL (PART 18)
- SMTP with 40+ provider presets
- Queue with retry logic
- Dead letter handling
- Template support

## GEOIP (PART 20)
- sapics/ip-location-db (free, MIT licensed)
- Monthly automatic updates via scheduler
- Country-level blocking support

## METRICS (PART 21)
- Prometheus format at /metrics
- Standard metrics: requests, latency, errors
- Custom metrics per project

## BACKUP/RESTORE (PART 22)
- Automated daily/hourly backups
- Encryption option (AES-256-GCM)
- Restore via admin panel or CLI
- Configurable retention

## UPDATE (PART 23)
- --update check (check for updates)
- --update yes (install update)
- --update branch {stable|beta|daily}
- Checksum verification required
- Rollback on failure

---
**Full details: AI.md PART 18, 19, 20, 21, 22, 23**
