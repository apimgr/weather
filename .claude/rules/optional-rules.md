# Optional Rules

@AI.md PART 32, 34, 35, 36: Tor, Multi-User, Organizations, Custom Domains

## Tor Hidden Service (PART 32)
- Binary controls Tor process
- Auto-generate .onion address
- Optional, enabled via config

## Multi-User (PART 34) - OPTIONAL
- Regular User accounts (separate from Server Admins)
- User registration, profiles, preferences
- When implemented: becomes NON-NEGOTIABLE
- If NOT using: NO user tables, NO /auth/register

## Organizations (PART 35) - OPTIONAL
- Requires PART 34 (Multi-User)
- Team collaboration features
- Org-level settings and billing
- If NOT using: NO org tables, NO /orgs/* routes

## Custom Domains (PART 36) - OPTIONAL
- User/org branded domains
- SSL certificate management
- DNS verification
- If NOT using: NO custom_domains table

## Rule
Once implemented, optional PART becomes NON-NEGOTIABLE.
If NOT using, feature must be COMPLETELY absent from code.
