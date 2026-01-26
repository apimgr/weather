# Service Rules

@AI.md PART 24, 25: Privilege Escalation, Service Management

## Privilege Escalation (PART 24)
- Run as non-root user
- Privileged ports via capabilities (Linux)
- Escalate only when needed
- Drop privileges after binding

## Service Management (PART 25)
Supported init systems:
| OS | Init System | File Location |
|----|-------------|---------------|
| Linux | systemd | `/etc/systemd/system/{project}.service` |
| Linux | runit | `/etc/sv/{project}/run` |
| Linux | rc.d | `/etc/rc.d/{project}` |
| macOS | launchd | `/Library/LaunchDaemons/{org}.{project}.plist` |
| Windows | SCM | Windows Service |

## Service User
- User: `{project}`
- Group: `{project}`
- Created automatically if missing
