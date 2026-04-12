# Features Rules (PART 18-23)

⚠️ **These rules are NON-NEGOTIABLE. Violations are bugs.** ⚠️

## CRITICAL - NEVER DO

- Never use external schedulers (cron, systemd timers, etc.)
- Never embed GeoIP databases in binary -- always download on first run
- Never disable the built-in scheduler -- it is ALWAYS running
- Never skip email template defaults -- all templates must work without SMTP config
- Never send WebUI notifications via email when user is active
- Never hardcode backup encryption password -- never store it
- Never skip metrics -- Prometheus-compatible metrics are required

## CRITICAL - ALWAYS DO

- Use the built-in internal scheduler for ALL scheduled tasks
- Download GeoIP databases on first run; update via scheduler
- Expose Prometheus-compatible metrics endpoint
- Follow Prometheus naming conventions for all metrics
- Include all required scheduled tasks
- Support customizable email templates with sensible defaults
- Use WebUI notification system first; email only when user is offline

## Email (PART 18)

- ALL projects MUST have customizable email templates
- ALL templates MUST have sensible defaults (work without SMTP config)
- ALL account-related emails MUST include: sender identity, action description, security notice, next steps
- Do NOT hardcode template content -- make it configurable

## WebUI Notifications (PART 18)

- ALWAYS available regardless of SMTP configuration
- WebUI notifications used when user is active
- Email ONLY used when user is not active/online

## Scheduler (PART 19)

- ALWAYS RUNNING -- no enable/disable setting
- NEVER use external schedulers (cron, systemd, launchd, Task Scheduler, Kubernetes CronJob)
- Every project MUST include these scheduled tasks:
  - Security database updates
  - GeoIP database updates
  - Session cleanup
  - Log rotation
  - Backup (if configured)
  - Update checks

## GeoIP (PART 20)

- Use sapics/ip-location-db
- NEVER embed databases in binary
- Download on first run
- Update via scheduler

## Metrics (PART 21)

- ALL projects MUST have built-in Prometheus-compatible metrics
- Expose at /metrics endpoint
- Follow Prometheus naming conventions

### Required Metrics

| Category | Examples |
|----------|---------|
| Application | uptime, version info, build info |
| HTTP | request count, duration, status codes |
| Database | query count, duration, connection pool |
| Authentication | login attempts, failures, active sessions |

## Backup (PART 22)

- If backup.encryption.enabled is true, backup MUST be encrypted
- Password is NEVER stored -- admin must remember/provide it
- Restore requires re-authentication

## Updates (PART 23)

- All downloads MUST be verified against checksums from release
- Built-in update check via scheduler

## Reference

For complete details, see AI.md PART 18 (28324-29646), PART 19 (29647-30131), PART 20 (30132-30204), PART 21 (30205-31649), PART 22 (31650-32378), PART 23 (32379-32857)
