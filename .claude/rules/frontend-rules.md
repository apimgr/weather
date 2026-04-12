# Frontend Rules (PART 16, 17)

⚠️ **These rules are NON-NEGOTIABLE. Violations are bugs.** ⚠️

## CRITICAL - NEVER DO

- Never use client-side rendering (React/Vue/Angular) -- server-side Go templates only
- Never require JavaScript for core functionality
- Never let long strings overflow containers or break mobile layouts
- Never show blank space for empty states -- always show proper empty state UI
- Never link to /admin from any public pages
- Never use default JavaScript UI elements -- always custom styled components
- Never put admin links in public nav
- Never create .env, .env.example, or .env.sample files
- Never render unvalidated custom verification meta tags
- Never use generic content placeholders -- content must come from IDEA.md

## CRITICAL - ALWAYS DO

- Use Go html/template for ALL frontend HTML
- Embed all templates and static assets in the binary
- Make all frontend work without JavaScript (progressive enhancement)
- Use mobile-first CSS (base styles for mobile, media queries for larger screens)
- Implement light AND dark themes (both must pass WCAG AA 4.5:1 contrast)
- Apply URL normalization middleware FIRST in chain
- Show CSRF protection on ALL forms
- Include cookie consent banner (always enabled -- we use cookies)
- Serve sitemap.xml at /sitemap.xml (exclude admin, auth, API)
- Use word-break CSS to prevent long string overflow

## Template Structure

```
src/templates/
  shared/
    head.tmpl        # meta, CSS links (REQUIRED)
    scripts.tmpl     # JS includes before </body> (REQUIRED)
    error.tmpl       # Error pages (REQUIRED, uses site theme)
  public/
    header.tmpl      # Public header (REQUIRED)
    nav.tmpl         # Public nav (REQUIRED)
    footer.tmpl      # Public footer (REQUIRED)
  admin/
    header.tmpl      # Admin header (REQUIRED)
    sidebar.tmpl     # Admin sidebar (REQUIRED)
    footer.tmpl      # Admin footer (REQUIRED)
```

Every page template MUST include header, nav, and footer partials.

## Theme System

- Light AND dark themes MUST be easy to read (no color conflicts)
- Both themes MUST pass WCAG AA contrast (4.5:1 minimum)
- Theme switching MUST work without page reload
- Syntax highlighting MUST adapt to theme
- Focus indicators MUST be visible in both themes
- Keyboard navigation MUST work identically in both themes
- Screen readers MUST work correctly in both themes

## URL Normalization (Middleware -- First in Chain)

- No double slashes
- No trailing slashes (redirect to canonical)
- Lowercase paths

## Public Nav Rules

Public nav NEVER contains:
- Links to /admin
- Server-internal paths
- Admin-only features

## Admin Panel Requirements

- MUST require authentication -- NEVER bypass in tests
- All internal systems MUST use the configured admin path (default: admin)
- Scheduler section MUST be visible in admin panel
- Admin panel handles ALL settings

## Admin Auth Features (REQUIRED)

| Feature | Applies To |
|---------|-----------|
| Argon2id password hashing | All Server Admins |
| TOTP 2FA support | All Server Admins |
| Passkey/WebAuthn support | All Server Admins |
| Recovery keys (when MFA enabled) | All Server Admins |
| Session timeout | All Server Admins |
| API token security | All Server Admins |
| Audit logging | All Server Admins |

## Cookie Consent

- ALWAYS enabled (we use cookies for sessions and preferences)
- Fixed bottom banner shown until user responds
- When "Decline": no analytics, no tracking, no 3rd-party scripts loaded

## Sitemap Rules

| Route Type | Include? |
|------------|---------|
| Public pages | YES |
| Admin pages (/admin/*) | NEVER |
| Auth pages (/auth/*) | NEVER |
| API endpoints (/api/*) | NEVER |

## Content Source

ALL page content (home, about, privacy, terms) MUST come from IDEA.md -- NEVER use generic placeholders.

## Setup Wizard (First Run)

App works perfectly with sane defaults before setup. Setup wizard is optional. Server is fully functional immediately on first run.

## Reference

For complete details, see AI.md PART 16 (20118-26286), PART 17 (26287-28323)
