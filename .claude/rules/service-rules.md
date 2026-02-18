# Service Rules (PART 24, 25)

⚠️ **These rules are NON-NEGOTIABLE. Violations are bugs.** ⚠️

## CRITICAL - NEVER DO
- ❌ Require manual escalation for every start
- ❌ Skip privilege dropping after port binding
- ❌ Run as root for entire lifetime

## REQUIRED - ALWAYS DO
- ✅ `--service install` for one-time setup
- ✅ Privilege drop after binding privileged ports
- ✅ Platform-specific service management
- ✅ Auto-create system user if needed

## PRIVILEGE ESCALATION
| Mode | How Started | Privilege |
|------|-------------|-----------|
| Service | `sudo weather --service install` | Root → drop to user |
| User | `weather` | User-level only |

## SERVICE COMMANDS
```bash
--service install     # Install as system service
--service uninstall   # Remove system service
--service start       # Start the service
--service stop        # Stop the service
--service restart     # Restart the service
--service reload      # Reload configuration
--service --help      # Show service help
```

## UNIX PRIVILEGE DROP
1. Start as root (via service manager)
2. Create system user `weather` if needed
3. Create directories, set ownership
4. Bind privileged ports (80, 443)
5. **DROP PRIVILEGES** to `weather` user
6. Run application as non-root

## WINDOWS
- Uses Virtual Service Account (NT SERVICE\weather)
- No privilege drop needed (already minimal)
- Service installed via SCM

## SUPPORTED SERVICE MANAGERS
| Platform | Init System |
|----------|-------------|
| Linux | systemd, runit, rc.d |
| macOS | launchd |
| FreeBSD | rc.d |
| Windows | SCM (Service Control Manager) |

---
**Full details: AI.md PART 24, PART 25**
