# Frontend Rules (PART 16, 17)

⚠️ **Server does the work. Client displays the result.** ⚠️

## CRITICAL - NEVER DO
- ❌ Client-side rendering (React, Vue, Angular)
- ❌ Require JavaScript for core functionality
- ❌ Client-side routing (SPA)
- ❌ Business logic in JavaScript
- ❌ Let long strings break mobile layout
- ❌ Desktop-first CSS (use mobile-first)
- ❌ Inline CSS or JavaScript
- ❌ JavaScript alerts (use toast notifications)
- ❌ Generic placeholder content in pages
- ❌ Stub templates or "coming soon" pages

## REQUIRED - ALWAYS DO
- ✅ Server-side rendering (Go templates)
- ✅ Progressive enhancement (works without JS)
- ✅ Mobile-first responsive CSS
- ✅ CSS `word-break: break-all` for long strings
- ✅ Full admin panel with ALL settings
- ✅ WCAG 2.1 AA accessibility
- ✅ Touch targets minimum 44x44px
- ✅ Content from IDEA.md for /server/about, /server/help

## LONG STRINGS (REQUIRED CSS)
```css
.long-string, .ip-address, .onion-address, .api-token, .hash {
  word-break: break-all;
  overflow-wrap: break-word;
  font-family: monospace;
}
```

## BREAKPOINTS (mobile-first)
| Target | CSS |
|--------|-----|
| Mobile (base) | No media query |
| Tablet+ | @media (min-width: 768px) |
| Desktop+ | @media (min-width: 1024px) |

## SERVER VS CLIENT
| Task | Where |
|------|-------|
| Data validation | SERVER |
| HTML rendering | SERVER |
| Business logic | SERVER |
| Theme toggle | Client JS (UX enhancement) |
| Copy to clipboard | Client JS (browser API) |

## ADMIN PANEL (PART 17)
- ALL settings editable via WebUI
- No SSH/CLI required for configuration
- Grouped logically with tooltips
- Real-time validation
- Live reload (no restart needed)

---
**Full details: AI.md PART 16, PART 17**
