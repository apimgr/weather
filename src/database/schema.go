package database

import (
	"database/sql"
	"fmt"
	"time"

	_ "modernc.org/sqlite"
)

const SchemaVersion = 1

// Schema contains all table creation SQL
var Schema = `
-- Schema version tracking
CREATE TABLE IF NOT EXISTS schema_version (
	version INTEGER PRIMARY KEY,
	applied_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

-- Users table
CREATE TABLE IF NOT EXISTS users (
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	email TEXT UNIQUE NOT NULL,
	password_hash TEXT NOT NULL,
	role TEXT DEFAULT 'user' CHECK(role IN ('admin', 'user')),
	created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
	updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_users_email ON users(email);
CREATE INDEX IF NOT EXISTS idx_users_role ON users(role);

-- API Tokens table
CREATE TABLE IF NOT EXISTS api_tokens (
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	user_id INTEGER NOT NULL,
	token TEXT UNIQUE NOT NULL,
	name TEXT,
	created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
	last_used_at DATETIME,
	FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_tokens_token ON api_tokens(token);
CREATE INDEX IF NOT EXISTS idx_tokens_user ON api_tokens(user_id);

-- Saved Locations table
CREATE TABLE IF NOT EXISTS saved_locations (
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	user_id INTEGER NOT NULL,
	name TEXT NOT NULL,
	latitude REAL NOT NULL,
	longitude REAL NOT NULL,
	timezone TEXT,
	alerts_enabled BOOLEAN DEFAULT 1,
	created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
	updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
	FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_locations_user ON saved_locations(user_id);

-- Weather Alerts table
CREATE TABLE IF NOT EXISTS weather_alerts (
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	location_id INTEGER NOT NULL,
	alert_type TEXT NOT NULL,
	severity TEXT NOT NULL CHECK(severity IN ('info', 'warning', 'severe', 'critical')),
	title TEXT NOT NULL,
	message TEXT NOT NULL,
	source TEXT,
	created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
	expires_at DATETIME,
	FOREIGN KEY (location_id) REFERENCES saved_locations(id) ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_alerts_location ON weather_alerts(location_id);
CREATE INDEX IF NOT EXISTS idx_alerts_expires ON weather_alerts(expires_at);

-- Server Settings table
CREATE TABLE IF NOT EXISTS settings (
	key TEXT PRIMARY KEY,
	value TEXT,
	type TEXT DEFAULT 'string',
	description TEXT,
	updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

-- Scheduled Tasks table
CREATE TABLE IF NOT EXISTS scheduled_tasks (
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	name TEXT NOT NULL,
	schedule TEXT NOT NULL,
	task_type TEXT NOT NULL,
	enabled BOOLEAN DEFAULT 1,
	last_run DATETIME,
	next_run DATETIME,
	last_result TEXT,
	created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
	updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_tasks_enabled ON scheduled_tasks(enabled);
CREATE INDEX IF NOT EXISTS idx_tasks_next_run ON scheduled_tasks(next_run);

-- Audit Log table (optional, disabled by default)
CREATE TABLE IF NOT EXISTS audit_log (
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	user_id INTEGER,
	action TEXT NOT NULL,
	resource TEXT,
	details TEXT,
	ip_address TEXT,
	user_agent TEXT,
	created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
	FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE SET NULL
);

CREATE INDEX IF NOT EXISTS idx_audit_user ON audit_log(user_id);
CREATE INDEX IF NOT EXISTS idx_audit_created ON audit_log(created_at);

-- Notifications table
CREATE TABLE IF NOT EXISTS notifications (
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	user_id INTEGER NOT NULL,
	type TEXT NOT NULL,
	title TEXT NOT NULL,
	message TEXT NOT NULL,
	link TEXT,
	read BOOLEAN DEFAULT 0,
	created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
	FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_notif_user ON notifications(user_id);
CREATE INDEX IF NOT EXISTS idx_notif_read ON notifications(read);
CREATE INDEX IF NOT EXISTS idx_notif_created ON notifications(created_at);

-- Sessions table
CREATE TABLE IF NOT EXISTS sessions (
	id TEXT PRIMARY KEY,
	user_id INTEGER NOT NULL,
	data TEXT,
	expires_at DATETIME NOT NULL,
	created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
	FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_sessions_user ON sessions(user_id);
CREATE INDEX IF NOT EXISTS idx_sessions_expires ON sessions(expires_at);

-- Rate Limiting table
CREATE TABLE IF NOT EXISTS rate_limits (
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	identifier TEXT NOT NULL,
	endpoint TEXT NOT NULL,
	count INTEGER DEFAULT 1,
	window_start DATETIME DEFAULT CURRENT_TIMESTAMP,
	UNIQUE(identifier, endpoint, window_start)
);

CREATE INDEX IF NOT EXISTS idx_ratelimit_identifier ON rate_limits(identifier, endpoint);
CREATE INDEX IF NOT EXISTS idx_ratelimit_window ON rate_limits(window_start);
`

// DefaultSettings are inserted on first setup
var DefaultSettings = map[string]string{
	"server.port":              "random", // Will be set to actual port on first run
	"server.address":           "0.0.0.0",
	"server.theme":             "dark",
	"auth.session_timeout":     "86400", // 24 hours
	"auth.require_email_verification": "false",
	"rate_limit.anonymous":     "120",
	"rate_limit.window":        "3600", // 1 hour
	"audit.enabled":            "false",
	"notifications.email":      "false",
	"notifications.webhook":    "false",
	"notifications.push":       "false",
	"weather.refresh_interval": "600", // 10 minutes
	"alerts.enabled":           "true",
	"alerts.check_interval":    "300", // 5 minutes
	"backup.enabled":           "true",
	"backup.interval":          "86400", // Daily
	"log.format":               "apache",
	"log.level":                "info",
}

// DB represents the database connection
type DB struct {
	*sql.DB
}

// InitDB initializes the database and creates tables
func InitDB(dbPath string) (*DB, error) {
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// Enable foreign keys
	if _, err := db.Exec("PRAGMA foreign_keys = ON"); err != nil {
		return nil, fmt.Errorf("failed to enable foreign keys: %w", err)
	}

	// Enable WAL mode for better concurrency
	if _, err := db.Exec("PRAGMA journal_mode = WAL"); err != nil {
		return nil, fmt.Errorf("failed to enable WAL mode: %w", err)
	}

	// Create schema
	if _, err := db.Exec(Schema); err != nil {
		return nil, fmt.Errorf("failed to create schema: %w", err)
	}

	// Check schema version
	var currentVersion int
	err = db.QueryRow("SELECT COALESCE(MAX(version), 0) FROM schema_version").Scan(&currentVersion)
	if err != nil {
		return nil, fmt.Errorf("failed to check schema version: %w", err)
	}

	// Insert schema version if new database
	if currentVersion == 0 {
		if _, err := db.Exec("INSERT INTO schema_version (version) VALUES (?)", SchemaVersion); err != nil {
			return nil, fmt.Errorf("failed to insert schema version: %w", err)
		}

		// Insert default settings
		for key, value := range DefaultSettings {
			_, err := db.Exec(`
				INSERT INTO settings (key, value) VALUES (?, ?)
				ON CONFLICT(key) DO NOTHING
			`, key, value)
			if err != nil {
				return nil, fmt.Errorf("failed to insert default setting %s: %w", key, err)
			}
		}
	}

	return &DB{db}, nil
}

// IsFirstRun checks if this is the first run (no users exist)
func (db *DB) IsFirstRun() (bool, error) {
	var count int
	err := db.QueryRow("SELECT COUNT(*) FROM users").Scan(&count)
	if err != nil {
		return false, err
	}
	return count == 0, nil
}

// GetSetting retrieves a setting value
func (db *DB) GetSetting(key string) (string, error) {
	var value string
	err := db.QueryRow("SELECT value FROM settings WHERE key = ?", key).Scan(&value)
	return value, err
}

// SetSetting updates or inserts a setting
func (db *DB) SetSetting(key, value string) error {
	_, err := db.Exec(`
		INSERT INTO settings (key, value, updated_at)
		VALUES (?, ?, ?)
		ON CONFLICT(key) DO UPDATE SET value = ?, updated_at = ?
	`, key, value, time.Now(), value, time.Now())
	return err
}

// CleanupExpiredSessions removes expired sessions
func (db *DB) CleanupExpiredSessions() error {
	_, err := db.Exec("DELETE FROM sessions WHERE expires_at < ?", time.Now())
	return err
}

// CleanupExpiredAlerts removes expired alerts
func (db *DB) CleanupExpiredAlerts() error {
	_, err := db.Exec("DELETE FROM weather_alerts WHERE expires_at IS NOT NULL AND expires_at < ?", time.Now())
	return err
}

// CleanupOldAuditLogs removes audit logs older than retention period
func (db *DB) CleanupOldAuditLogs(retentionDays int) error {
	cutoff := time.Now().AddDate(0, 0, -retentionDays)
	_, err := db.Exec("DELETE FROM audit_log WHERE created_at < ?", cutoff)
	return err
}

// CleanupRateLimits removes old rate limit entries
func (db *DB) CleanupRateLimits() error {
	// Remove entries older than 2 hours
	cutoff := time.Now().Add(-2 * time.Hour)
	_, err := db.Exec("DELETE FROM rate_limits WHERE window_start < ?", cutoff)
	return err
}
