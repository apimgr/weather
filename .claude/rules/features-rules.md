# Features Rules (PART 18-23)

⚠️ **These rules are NON-NEGOTIABLE. Violations are bugs.** ⚠️

## CRITICAL - NEVER DO
- ❌ Use external cron/scheduler (built-in only)
- ❌ Skip backup encryption for sensitive data
- ❌ Store GeoIP database in repo (download at runtime)
- ❌ Hardcode SMTP settings (support multiple providers)
- ❌ Skip audit logging for admin actions

## CRITICAL - ALWAYS DO
- ✅ Built-in scheduler for ALL background tasks (PART 19)
- ✅ SMTP with auto-detection and 40+ provider presets (PART 18)
- ✅ GeoIP with embedded database and auto-updates (PART 20)
- ✅ Prometheus metrics at /metrics (internal only) (PART 21)
- ✅ Automated backup with encryption support (PART 22)
- ✅ Self-update capability with rollback (PART 23)

## SCHEDULER (PART 19)
| Task Type | Examples |
|-----------|----------|
| Cleanup | Sessions, tokens, rate limits, audit logs |
| Maintenance | Log rotation, backups, health checks |
| Updates | GeoIP database, blocklists, CVE database |
| External | Weather alerts, SSL renewal |

## EMAIL (PART 18)
- SMTP auto-detection from MX records
- 40+ provider presets (Gmail, Outlook, etc.)
- TLS modes: auto, starttls, tls, none
- Notification queue with retry and dead letter

## GEOIP (PART 20)
- Embedded database (sapics/ip-location-db)
- Monthly auto-updates via scheduler
- IP → Country, City, Timezone
- Integration with rate limiting and blocking

## METRICS (PART 21)
- Prometheus format at /metrics
- INTERNAL only (never expose publicly)
- Include: requests, latency, cache stats, DB connections
- Go runtime stats, custom application metrics

## BACKUP (PART 22)
| Type | Frequency | Retention |
|------|-----------|-----------|
| Daily | 24h | 7 days |
| Hourly | 1h | 24 hours |
| Manual | On-demand | Forever |

- Format: .tar.gz with optional encryption
- Include: database, config, certificates
- Cluster-aware: coordinated backups

## UPDATE (PART 23)
```
--update check         # Check for updates
--update yes           # Install update
--update branch stable # Switch branch
--update branch beta
--update branch daily
```

- Automatic rollback on failure
- Preserve configuration
- Verify checksums

---
For complete details, see AI.md PART 18-23
