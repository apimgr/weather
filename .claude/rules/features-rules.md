# Features Rules

@AI.md PART 18-23: Email, Scheduler, GeoIP, Metrics, Backup, Update

## Email & Notifications (PART 18)
- SMTP auto-detection
- Template-based emails
- Notification preferences per user

## Scheduler (PART 19)
- Internal scheduler (NO external cron)
- Built-in tasks: ssl.renewal, geoip.update, session.cleanup
- cluster.heartbeat every 30s (cluster mode)
- Catch-up window for missed tasks

## GeoIP (PART 20)
- ip-location-db for country detection
- Weekly automatic updates
- Country blocking support

## Metrics (PART 21)
- Prometheus format at `/metrics`
- INTERNAL only (not public)
- Request counts, latency, DB connections

## Backup & Restore (PART 22)
- Automatic daily backups
- Encrypted backups supported
- Restore via admin panel or CLI

## Update Command (PART 23)
- `--update` flag for self-update
- Verify binary signatures
- Rollback on failure
