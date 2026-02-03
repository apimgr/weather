# Optional Rules (PART 34, 35, 36)

⚠️ **OPTIONAL sections become NON-NEGOTIABLE when implemented.** ⚠️

## PART 34: MULTI-USER (Implemented for Weather)

### When to Include
- Apps with Regular User accounts
- Multi-tenant platforms
- User-specific data storage
- Social features, profiles

### If NOT Using Multi-User
These must NOT appear in code:
- users table
- User registration
- User preferences table
- User API tokens
- /auth/register routes
- User profiles

### If Using Multi-User (Weather uses this)
- Separate from Server Admin (different DB tables)
- Registration modes: public, private, disabled
- Profile fields: username, avatar, bio, website
- 2FA support: TOTP, Passkeys
- API tokens per user
- Privacy settings (profile visibility)

## PART 35: ORGANIZATIONS (Not Used)

### When to Include
- Team collaboration features
- B2B platforms
- Workspace-based apps
- Requires PART 34 (Multi-User)

### If NOT Using Organizations
These must NOT appear:
- organizations table
- Org membership
- Org ownership
- Org API tokens
- /orgs/* routes

## PART 36: CUSTOM DOMAINS (Not Used)

### When to Include
- Linktree-style apps
- Blog platforms
- White-label services
- Users want branded URLs

### If NOT Using Custom Domains
These must NOT appear:
- custom_domains table
- Domain verification
- User/org domain settings
- SSL for custom domains

## CLIENT-SIDE PREFERENCES (No PART 34 Required)
These work without user accounts:
- localStorage: Theme, language, UI prefs
- Cookies: Session prefs, consent flags

---
**Full details: AI.md PART 34, PART 35, PART 36**
