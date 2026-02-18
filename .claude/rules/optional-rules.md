# Optional Rules (PART 34, 35, 36)

⚠️ **OPTIONAL until implemented. Once implemented, NON-NEGOTIABLE.** ⚠️

## CRITICAL - NEVER DO (if not using)
- ❌ Include user registration code if not using PART 34
- ❌ Include org code if not using PART 35
- ❌ Include custom domain code if not using PART 36
- ❌ Add disabled feature toggles (users.enabled: false)

## WHEN TO INCLUDE

### PART 34: Multi-User
| Include If | Skip If |
|------------|---------|
| Regular user accounts needed | Admin-only APIs |
| User registration required | Simple data APIs |
| User profiles needed | Read-only services |

### PART 35: Organizations (requires PART 34)
| Include If | Skip If |
|------------|---------|
| Team collaboration | Single-user apps |
| B2B features | Consumer apps |
| Shared workspaces | Personal tools |

### PART 36: Custom Domains
| Include If | Skip If |
|------------|---------|
| Users want branded URLs | Simple APIs |
| Content under user domains | Internal tools |

## IF IMPLEMENTING PART 34
- ✅ Separate users table (not admins)
- ✅ Registration modes: public, private, disabled
- ✅ Username blocklist validation
- ✅ Profile fields: visibility, avatar, bio, website
- ✅ 2FA support: TOTP, Passkeys
- ✅ User preferences table
- ✅ User API tokens (usr_ prefix)

## IF IMPLEMENTING PART 35
- ✅ Organizations table
- ✅ Organization membership
- ✅ Org API tokens (org_ prefix)
- ✅ Org-level settings

## IF IMPLEMENTING PART 36
- ✅ Custom domains table
- ✅ Domain verification (DNS TXT)
- ✅ SSL cert provisioning
- ✅ Domain routing

## CLEAN CODEBASE RULE
If NOT using an optional PART:
- Zero code references
- No conditionals checking for feature
- No config toggles
- No database tables
- Code written as if feature never existed

---
**Full details: AI.md PART 34, PART 35, PART 36**
