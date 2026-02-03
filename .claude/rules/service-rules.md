# Service Rules (PART 24, 25)

⚠️ **These rules are NON-NEGOTIABLE. Violations are bugs.** ⚠️

## CRITICAL - NEVER DO
- ❌ Require manual service file creation
- ❌ Skip privilege checks before service operations
- ❌ Assume root/admin on first run
- ❌ Use different service interface per OS

## REQUIRED - ALWAYS DO
- ✅ --service flag for service management
- ✅ Auto-detect privilege level
- ✅ Request elevation when needed
- ✅ Support all 4 OSes with native service managers
- ✅ PID file management
- ✅ Graceful shutdown on signals

## SERVICE FLAG
```
--service start      Start service
--service stop       Stop service
--service restart    Restart service
--service reload     Reload configuration
--service --install  Install as system service
--service --uninstall Remove system service
--service --disable  Disable service
--service --help     Show service help
```

## PRIVILEGE ESCALATION (PART 24)
1. Detect if running as root/admin
2. If not, request elevation
3. Linux: sudo/pkexec
4. macOS: osascript
5. Windows: UAC prompt

## NATIVE SERVICE MANAGERS (PART 25)
| OS | Service Manager | File Location |
|----|-----------------|---------------|
| Linux | systemd | /etc/systemd/system/weather.service |
| macOS | launchd | /Library/LaunchDaemons/com.apimgr.weather.plist |
| FreeBSD | rc.d | /usr/local/etc/rc.d/weather |
| Windows | SCM | Windows Service Manager |

## SIGNAL HANDLING
| Signal | Action |
|--------|--------|
| SIGTERM | Graceful shutdown |
| SIGINT | Graceful shutdown |
| SIGHUP | Reload configuration |
| SIGUSR1 | Reopen log files |

## PID FILE
- Created on start: {pid_dir}/weather.pid
- Removed on clean shutdown
- Used for status checks and signaling

---
**Full details: AI.md PART 24, PART 25**
