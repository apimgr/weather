# Optional Rules (PART 34, 35, 36)

⚠️ **When implemented, these become NON-NEGOTIABLE.** ⚠️

## STATUS FOR THIS PROJECT
- **PART 34 (Multi-User)**: IMPLEMENTED ✅
- **PART 35 (Organizations)**: Not implemented
- **PART 36 (Custom Domains)**: Not implemented

---

## PART 34: MULTI-USER (IMPLEMENTED)

### CRITICAL - NEVER DO
- ❌ Confuse Server Admin with Regular User (different tables)
- ❌ Allow usernames with special characters except `-` and `_`
- ❌ Allow usernames starting with numbers
- ❌ Allow reserved usernames (admin, api, www, etc.)
- ❌ Skip email verification for public registration

### CRITICAL - ALWAYS DO
- ✅ Separate tables: `usr_admins` vs `usr_users`
- ✅ Username format: lowercase, 3-30 chars, starts with letter
- ✅ Registration modes: public, private (invite), disabled
- ✅ Email verification required for public registration
- ✅ 2FA support: TOTP, Passkeys, recovery keys
- ✅ Profile fields: visibility, avatar, bio, website
- ✅ User preferences table for settings

### ACCOUNT TYPES
| Type | Table | Purpose |
|------|-------|---------|
| Server Admin | `usr_admins` | App administration |
| Primary Admin | `usr_admins` | First admin (cannot delete) |
| Regular User | `usr_users` | End-user accounts |

### USER ROUTES
| Web Route | API Route | Purpose |
|-----------|-----------|---------|
| `/users/settings` | `GET/PATCH /api/v1/users/settings` | Account settings |
| `/users/tokens` | `GET/POST/DELETE /api/v1/users/tokens` | API tokens |
| - | `GET /api/v1/public/users/:username` | Public profile |

### API TOKEN PREFIXES
| Owner | Prefix | Example |
|-------|--------|---------|
| Admin | `adm_` | `adm_abc123...` |
| User | `usr_` | `usr_xyz789...` |
| Org | `org_` | `org_def456...` |

---

## PART 35: ORGANIZATIONS (NOT IMPLEMENTED)

If implemented, requires:
- Org membership and roles
- Org-level API tokens
- Org admin vs member permissions
- Requires PART 34

---

## PART 36: CUSTOM DOMAINS (NOT IMPLEMENTED)

If implemented, requires:
- User/org branded domains
- DNS verification (TXT record)
- SSL certificates per domain
- Requires PART 34 (and optionally 35)

---

## UNUSED FEATURES RULE

**If a PART is NOT implemented, its code must NOT exist:**

- No references to the feature
- No `if enabled` conditionals
- No stub functions or empty tables
- No config options for the feature
- No hidden UI elements

**The code should be written as if the feature never existed.**

---
For complete details, see AI.md PART 34, 35, 36
