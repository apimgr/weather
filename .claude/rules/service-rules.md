# Service & Deployment Rules (PART 24, 25)

⚠️ **These rules are NON-NEGOTIABLE. Violations are bugs.** ⚠️

## CRITICAL - NEVER DO

- Never use commonly reserved UIDs/GIDs even if they appear available
- Never hardcode UIDs/GIDs -- find an unused ID dynamically
- Never skip service manager support -- all service managers are required

## CRITICAL - ALWAYS DO

- Support ALL service managers (systemd, OpenRC, SysV, launchd, Windows Services)
- Use the same UID and GID value for the service user
- Find an unused UID/GID where both are available AND not reserved by well-known services
- Support privilege escalation/dropping (start as root if needed for ports, drop after bind)

## Service User Rules

| Rule | Detail |
|------|--------|
| UID == GID | MUST be the same value |
| Unused ID | Find one not reserved by well-known services |
| Never hardcode | Detect unused ID dynamically during install |

## Reserved UIDs/GIDs to Avoid

NEVER use these even if they appear available:
- Common service accounts (www-data, nobody, daemon, etc.)
- IDs used by well-known services across distributions
- Detect dynamically -- do not hardcode any specific ID

## Service Manager Support (ALL Required)

| Platform | Service Manager |
|----------|----------------|
| Linux (systemd) | systemd unit file |
| Linux (OpenRC) | OpenRC init script |
| Linux (SysV) | SysV init script |
| macOS | launchd plist |
| Windows | Windows Service |
| FreeBSD | rc.d script |

## Privilege Escalation Logic

Follow PART 5 "Smart Escalation Logic" for complete escalation flow:
- If binding to privileged port (< 1024): may need to start as root
- After binding: drop privileges to service user
- If no privilege needed: never run as root

## Let's Encrypt (REQUIRED)

- ALL projects MUST have built-in Let's Encrypt support
- Auto-renew certificates
- Configurable via admin panel

## Tor Hidden Service

- ALWAYS enabled if Tor binary is found
- No enable/disable toggle
- Application starts its own dedicated Tor process
- Tor dirs: {config_dir}/tor/, {data_dir}/tor/, {log_dir}/tor.log

## Reference

For complete details, see AI.md PART 24 (32858-33755), PART 25 (33756-33939)
