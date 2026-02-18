# Frontend Rules (PART 16, 17)

⚠️ **These rules are NON-NEGOTIABLE. Violations are bugs.** ⚠️

## CRITICAL - NEVER DO
- ❌ Client-side rendering (React, Vue, Angular, etc.)
- ❌ Require JavaScript for core functionality
- ❌ Client-side routing (SPA)
- ❌ Business logic in JavaScript
- ❌ Let long strings break mobile layout
- ❌ Desktop-first CSS (use mobile-first)
- ❌ Inline CSS or JavaScript
- ❌ JavaScript alerts (use toast notifications)
- ❌ Generic placeholder content in /server/about or /server/help pages
- ❌ "Your application name here" or "Feature 1, Feature 2" text
- ❌ Stub templates or "coming soon" pages
- ❌ Empty handlers or placeholder routes

## CRITICAL - ALWAYS DO
- ✅ Server-side rendering (Go templates)
- ✅ Progressive enhancement (works without JS)
- ✅ Mobile-first responsive CSS
- ✅ CSS `word-break: break-all` for long strings (IPv6, .onion, tokens)
- ✅ Full admin panel with ALL settings
- ✅ WCAG 2.1 AA accessibility
- ✅ Touch targets minimum 44x44px
- ✅ /server/about content from IDEA.md (name, tagline, description, features)
- ✅ /server/help content from IDEA.md (real endpoints, real examples)
- ✅ All pages fully functional - no "coming soon" or placeholder pages
- ✅ All routes implemented - no 501 Not Implemented responses

## PAGE CONTENT SOURCING
| Page | Content Source |
|------|----------------|
| /server/about | IDEA.md → name, tagline, description, features, links |
| /server/help | IDEA.md → real endpoints, real curl examples, real FAQ |
| /server/privacy | Config → `server.privacy.*` settings |
| /server/terms | Config → customizable, default template |
| /server/contact | Config → `server.contact` settings |

## SERVER VS CLIENT
| Task | Where | Why |
|------|-------|-----|
| Data validation | SERVER | Server is authoritative |
| HTML rendering | SERVER | Works without JS |
| Business logic | SERVER | Security, consistency |
| Formatting | SERVER | Consistent output |
| Theme toggle | Client JS | Instant UX feedback |
| Copy to clipboard | Client JS | Browser API required |
| Form feedback | Client JS | UX enhancement only |

## LONG STRINGS (REQUIRED CSS)
```css
.long-string, .ip-address, .onion-address, .api-token, .hash {
  word-break: break-all;
  overflow-wrap: break-word;
  font-family: monospace;
}
```

Apply to: IPv6, Tor .onion, API tokens, hashes, UUIDs, Base64

## BREAKPOINTS (mobile-first)
| Target | CSS |
|--------|-----|
| Mobile (base) | No media query |
| Tablet+ | `@media (min-width: 768px)` |
| Desktop+ | `@media (min-width: 1024px)` |

## ADMIN PANEL (PART 17)
- Full WebUI for ALL server settings
- No settings only available via config file
- Real-time validation and feedback
- Audit logging for all admin actions
- MFA support (TOTP, Passkeys)

---
For complete details, see AI.md PART 16, 17
