# Service Rules (PART 24, 25)

⚠️ **These rules are NON-NEGOTIABLE. Violations are bugs.** ⚠️

## CRITICAL - NEVER DO
- ❌ Require manual service file creation
- ❌ Assume user has root/admin access
- ❌ Skip privilege detection before escalation
- ❌ Run as root after binding privileged ports (Unix)
- ❌ Prompt for escalation when user cannot escalate

## CRITICAL - ALWAYS DO
- ✅ Auto-generate service files on `--service --install`
- ✅ Detect privilege level (root vs user)
- ✅ Check if user CAN escalate before prompting
- ✅ Drop privileges after binding ports <1024 (Unix)
- ✅ Support both system and user service installations
- ✅ Support all service managers per platform

## SERVICE MANAGEMENT
```bash
weather --service --install    # Install service
weather --service --uninstall  # Remove service
weather --service --disable    # Disable autostart
weather --service start        # Start service
weather --service stop         # Stop service
weather --service restart      # Restart service
weather --service reload       # Reload config
weather --service status       # Show status (read-only)
```

## PLATFORM SERVICE MANAGERS
| Platform | System Service | User Service |
|----------|---------------|--------------|
| Linux | systemd | systemd --user |
| macOS | launchd (LaunchDaemons) | launchd (LaunchAgents) |
| FreeBSD | rc.d | user crontab |
| Windows | SCM (Services) | Task Scheduler |

## PRIVILEGE ESCALATION FLOW
```
User runs: weather --service --install
    │
    ├─► Is elevated (root/admin)?
    │   └─► YES: Install system service
    │
    ├─► Can escalate (sudo/UAC)?
    │   └─► YES: Ask user, re-exec elevated
    │
    └─► Cannot escalate?
        └─► Install user service OR show clear error
```

## UNIX PRIVILEGE DROP
1. Start as root (service manager)
2. Create `weather` user if needed
3. Bind privileged ports (80, 443)
4. **DROP to `weather` user**
5. Initialize application
6. Serve requests (as user)

## WINDOWS VIRTUAL SERVICE ACCOUNT
- Auto-created: `NT SERVICE\weather`
- No manual user creation needed
- Minimal privileges by design
- No privilege dropping (already minimal)

## OPERATIONS REQUIRING AUTH
| Operation | Authorization |
|-----------|---------------|
| `--maintenance setup` | First-run OR setup token OR root |
| `--maintenance restore` | Admin creds OR root OR empty DB |
| `--maintenance mode` | Admin creds OR root |

---
For complete details, see AI.md PART 24, 25
