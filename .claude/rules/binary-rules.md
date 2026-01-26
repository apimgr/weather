# Binary Rules

@AI.md PART 7, 8, 33: Binary Requirements, Server CLI, Client/Agent

## Binary Requirements (PART 7)
- CGO_ENABLED=0 always
- Single static binary with embedded assets
- Support 4 OS x 2 arch = 8 binaries
- Binary naming: `{project}-{os}-{arch}`

## Server CLI (PART 8)
Required flags:
- `--help`, `-h` - Show help
- `--version`, `-v` - Show version
- `--config` - Config file path
- `--address` - Bind address
- `--debug` - Enable debug mode

## Version Output Format
```
{binaryname} v{version}
Built: {timestamp}
Go: {goversion}
OS/Arch: {os}/{arch}
```

## Client Binary (PART 33)
- Name: `{project}-cli`
- REQUIRED for all projects
- Supports CLI, TUI, GUI modes
- Auto-detects mode from terminal
