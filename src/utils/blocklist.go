package utils

import (
	"strings"
)

// UsernameBlocklist contains reserved usernames that cannot be registered
// This list prevents users from registering common system usernames, admin terms,
// and potentially confusing identifiers
var UsernameBlocklist = []string{
	// System/Admin terms
	"admin", "administrator", "root", "superuser", "sudo", "sysadmin",
	"system", "moderator", "mod", "owner", "staff", "support",
	"help", "helpdesk", "webmaster", "postmaster", "hostmaster",
	"admins", "administrators",

	// Service accounts
	"noreply", "no-reply", "mailer-daemon", "daemon", "nobody",
	"service", "api", "bot", "system-bot", "automated",

	// Common usernames
	"user", "guest", "anonymous", "anon", "public", "default",
	"test", "testing", "demo", "example", "sample", "tests",

	// Security-related
	"security", "abuse", "spam", "phishing", "fraud", "scam",
	"hacker", "exploit", "malware", "virus",

	// Application-specific
	"weather", "weatherapp", "app", "application", "server",
	"api-server", "backend", "frontend", "database", "db",

	// Reserved routes/endpoints
	"login", "logout", "register", "signup", "signin", "signout",
	"auth", "authentication", "session", "sessions", "token", "tokens",
	"user", "users", "account", "accounts", "profile", "profiles",
	"settings", "preferences", "config", "configuration",
	"dashboard", "panel", "admin-panel", "control-panel",
	"api", "v1", "v2", "v3", "version", "versions",
	"docs", "documentation", "api-docs", "swagger",
	"static", "assets", "public", "files", "uploads",
	"download", "downloads", "export", "exports",
	"health", "healthz", "healthcheck", "status", "ping",
	"metrics", "stats", "statistics", "analytics",
	"webhook", "webhooks", "callback", "callbacks",
	"notification", "notifications", "alert", "alerts",
	"message", "messages", "email", "emails", "mail",
	"search", "find", "query", "queries",
	"feed", "feeds", "rss", "atom",
	"blog", "blogs", "post", "posts", "article", "articles",
	"page", "pages", "site", "sites",
	"tag", "tags", "category", "categories",
	"group", "groups", "team", "teams", "organization", "organizations",
	"company", "companies", "business", "businesses",

	// Social/Communication
	"follow", "following", "follower", "followers",
	"friend", "friends", "contact", "contacts",
	"chat", "message", "dm", "inbox", "outbox",

	// Actions/Verbs
	"create", "new", "add", "edit", "update", "delete", "remove",
	"show", "view", "list", "index", "home", "main",
	"about", "info", "information", "faq", "faqs",
	"terms", "tos", "privacy", "legal", "dmca",
	"cookie", "cookies", "gdpr", "ccpa",

	// Common brands/services (prevent impersonation)
	"google", "facebook", "twitter", "github", "microsoft",
	"apple", "amazon", "netflix", "youtube", "instagram",
	"linkedin", "reddit", "discord", "slack", "zoom",

	// Weather-specific
	"forecast", "forecasts", "temperature", "humidity",
	"wind", "rain", "snow", "storm", "hurricane", "earthquake",
	"moon", "sun", "cloud", "clouds", "sky",
	"location", "locations", "saved", "favorite", "favorites",

	// Offensive/Inappropriate (basic list)
	"fuck", "shit", "ass", "bitch", "damn", "hell",
	"sex", "porn", "xxx", "adult", "nsfw",
	"kill", "death", "die", "suicide", "murder",
	"nazi", "hitler", "terrorist", "terrorism",
	"drug", "drugs", "cocaine", "heroin", "meth",

	// Special characters/confusing
	"null", "nil", "undefined", "void", "none",
	"true", "false", "yes", "no", "ok", "error",
	"success", "fail", "failure", "warning",
	"placeholder", "tmp", "temp", "temporary",

	// Single characters (too generic/confusing)
	"a", "b", "c", "d", "e", "f", "g", "h", "i", "j",
	"k", "l", "m", "n", "o", "p", "q", "r", "s", "t",
	"u", "v", "w", "x", "y", "z",
	"0", "1", "2", "3", "4", "5", "6", "7", "8", "9",

	// Common short patterns
	"00", "01", "10", "11", "99",

	// Common variations
	"admin1", "admin2", "admin123", "test1", "test123",
	"user1", "user123", "guest1", "guest123",

	// Automated/Bot accounts
	"robot", "crawler", "spider", "scraper", "agent",
	"script", "automation", "auto",

	// Special cases
	"me", "myself", "self", "you", "your", "yours",
	"everyone", "all", "here", "channel",
	"system-admin", "sys-admin", "site-admin",
	"super-admin", "super-user", "power-user",
}

// IsUsernameBlocked checks if a username (email local part) is on the blocklist
// Returns true if the username is blocked, false otherwise
func IsUsernameBlocked(email string) bool {
	// Extract username from email (everything before @)
	parts := strings.Split(email, "@")
	if len(parts) < 2 || parts[0] == "" {
		return true // Invalid email format or empty username
	}

	username := strings.ToLower(strings.TrimSpace(parts[0]))

	// Empty username check
	if username == "" {
		return true
	}

	// Check exact matches
	for _, blocked := range UsernameBlocklist {
		if username == strings.ToLower(blocked) {
			return true
		}
	}

	// Check for variations with numbers/special chars at the end
	// e.g., "admin123", "admin_1", "admin-test"
	// But NOT "administrator" or "testing" (where blocked word is a substring)
	for _, blocked := range UsernameBlocklist {
		blockedLower := strings.ToLower(blocked)

		// Only check if username starts with the blocked term
		if !strings.HasPrefix(username, blockedLower) {
			continue
		}

		// Get what comes after the blocked term
		suffix := username[len(blockedLower):]

		// If there's a suffix, only block if it's "simple" (numbers/punctuation only)
		// This allows "administrator" and "testing" but blocks "admin123" and "test_1"
		if suffix != "" && !isSimpleSuffix(suffix) {
			continue // Has complex suffix, allow it
		}

		// If we get here, it's either exact match or blocked term + simple suffix
		if suffix == "" || isSimpleSuffix(suffix) {
			return true
		}
	}

	return false
}

// isSimpleSuffix checks if a string contains only numbers, hyphens, and underscores
func isSimpleSuffix(s string) bool {
	if s == "" {
		return true
	}

	for _, c := range s {
		if !((c >= '0' && c <= '9') || c == '-' || c == '_' || c == '.') {
			return false
		}
	}
	return true
}

// GetBlocklistSize returns the number of blocked usernames
func GetBlocklistSize() int {
	return len(UsernameBlocklist)
}

// IsBlocklistPublic always returns true - this blocklist is intentionally public
// to allow users to understand why their chosen username was rejected
func IsBlocklistPublic() bool {
	return true
}
