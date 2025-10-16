package models

import (
	"database/sql"
	"encoding/json"
	"fmt"
)

// Setting represents a configuration setting
type Setting struct {
	Key         string `json:"key"`
	Value       string `json:"value"`
	Type        string `json:"type"`        // string, number, boolean, json
	Description string `json:"description"` // Human-readable description
}

// SettingsModel handles settings operations
type SettingsModel struct {
	DB *sql.DB
}

// Get retrieves a setting by key
func (m *SettingsModel) Get(key string) (*Setting, error) {
	setting := &Setting{}
	err := m.DB.QueryRow(
		"SELECT key, value, type FROM settings WHERE key = ?",
		key,
	).Scan(&setting.Key, &setting.Value, &setting.Type)

	if err != nil {
		return nil, err
	}

	return setting, nil
}

// GetString retrieves a setting value as string
func (m *SettingsModel) GetString(key, defaultValue string) string {
	setting, err := m.Get(key)
	if err != nil {
		return defaultValue
	}
	return setting.Value
}

// GetInt retrieves a setting value as int
func (m *SettingsModel) GetInt(key string, defaultValue int) int {
	setting, err := m.Get(key)
	if err != nil {
		return defaultValue
	}

	var value int
	if _, err := fmt.Sscanf(setting.Value, "%d", &value); err != nil {
		return defaultValue
	}
	return value
}

// GetBool retrieves a setting value as bool
func (m *SettingsModel) GetBool(key string, defaultValue bool) bool {
	setting, err := m.Get(key)
	if err != nil {
		return defaultValue
	}

	return setting.Value == "true" || setting.Value == "1"
}

// GetJSON retrieves a setting value as JSON
func (m *SettingsModel) GetJSON(key string, dest interface{}) error {
	setting, err := m.Get(key)
	if err != nil {
		return err
	}

	return json.Unmarshal([]byte(setting.Value), dest)
}

// Set creates or updates a setting
func (m *SettingsModel) Set(key, value, settingType string) error {
	_, err := m.DB.Exec(`
		INSERT INTO settings (key, value, type)
		VALUES (?, ?, ?)
		ON CONFLICT(key) DO UPDATE SET value = ?, type = ?
	`, key, value, settingType, value, settingType)

	return err
}

// SetWithDescription sets a setting with description
func (m *SettingsModel) SetWithDescription(key, value, settingType, description string) error {
	_, err := m.DB.Exec(`
		INSERT INTO settings (key, value, type, description)
		VALUES (?, ?, ?, ?)
		ON CONFLICT(key) DO UPDATE SET value = ?, type = ?, description = ?
	`, key, value, settingType, description, value, settingType, description)

	return err
}

// SetString sets a string setting
func (m *SettingsModel) SetString(key, value string) error {
	return m.Set(key, value, "string")
}

// SetInt sets an integer setting
func (m *SettingsModel) SetInt(key string, value int) error {
	return m.Set(key, fmt.Sprintf("%d", value), "number")
}

// SetBool sets a boolean setting
func (m *SettingsModel) SetBool(key string, value bool) error {
	stringValue := "false"
	if value {
		stringValue = "true"
	}
	return m.Set(key, stringValue, "boolean")
}

// SetJSON sets a JSON setting
func (m *SettingsModel) SetJSON(key string, value interface{}) error {
	jsonBytes, err := json.Marshal(value)
	if err != nil {
		return err
	}
	return m.Set(key, string(jsonBytes), "json")
}

// Delete removes a setting
func (m *SettingsModel) Delete(key string) error {
	_, err := m.DB.Exec("DELETE FROM settings WHERE key = ?", key)
	return err
}

// List returns all settings
func (m *SettingsModel) List() ([]*Setting, error) {
	rows, err := m.DB.Query("SELECT key, value, type FROM settings ORDER BY key")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var settings []*Setting
	for rows.Next() {
		setting := &Setting{}
		if err := rows.Scan(&setting.Key, &setting.Value, &setting.Type); err != nil {
			return nil, err
		}
		settings = append(settings, setting)
	}

	return settings, rows.Err()
}

// ListByPrefix returns all settings with a specific prefix
func (m *SettingsModel) ListByPrefix(prefix string) ([]*Setting, error) {
	rows, err := m.DB.Query(
		"SELECT key, value, type FROM settings WHERE key LIKE ? ORDER BY key",
		prefix+"%",
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var settings []*Setting
	for rows.Next() {
		setting := &Setting{}
		if err := rows.Scan(&setting.Key, &setting.Value, &setting.Type); err != nil {
			return nil, err
		}
		settings = append(settings, setting)
	}

	return settings, rows.Err()
}

// InitializeDefaults sets default values for all settings
// backupPath is optional - if empty, uses "/data/backups" as default
func (m *SettingsModel) InitializeDefaults(backupPath ...string) error {
	defaultBackupPath := "/data/backups"
	if len(backupPath) > 0 && backupPath[0] != "" {
		defaultBackupPath = backupPath[0]
	}

	defaults := map[string]Setting{
		// Server settings
		"server.title":          {Value: "Weather Service", Type: "string", Description: "Application display name shown in the header and page titles"},
		"server.tagline":        {Value: "Your personal weather dashboard", Type: "string", Description: "Short subtitle or slogan displayed under the title"},
		"server.description":    {Value: "A comprehensive platform for weather forecasts, moon phases, earthquakes, and hurricane tracking.", Type: "string", Description: "Full description shown on about page and in meta tags"},
		"server.address":        {Value: "0.0.0.0", Type: "string", Description: "Server listen address (0.0.0.0 for all interfaces)"},
		"server.http_port":      {Value: "0", Type: "number", Description: "HTTP port number (requires restart to apply)"},
		"server.https_port":     {Value: "0", Type: "number", Description: "HTTPS port number (0 to disable, requires restart)"},
		"server.https_enabled":  {Value: "false", Type: "boolean", Description: "Enable HTTPS/TLS connections"},
		"server.timezone":       {Value: "America/New_York", Type: "string", Description: "Default server timezone for date/time display"},
		"server.date_format":    {Value: "US", Type: "string", Description: "Date format: US (Jan 1, 2024), EU (1 Jan 2024), or ISO (2024-01-01)"},
		"server.time_format":    {Value: "12-hour", Type: "string", Description: "Time format: 12-hour (3:30 PM) or 24-hour (15:30)"},

		// Security settings
		"security.session_timeout":     {Value: "2592000", Type: "number", Description: "Session timeout in seconds (default: 2592000 = 30 days)"},
		"security.max_login_attempts":  {Value: "5", Type: "number", Description: "Maximum failed login attempts before account lockout"},
		"security.lockout_duration":    {Value: "30", Type: "number", Description: "Account lockout duration in minutes after max login attempts"},
		"security.password_min_length": {Value: "8", Type: "number", Description: "Minimum required password length for user accounts"},

		// Features
		"features.registration_enabled": {Value: "true", Type: "boolean", Description: "Allow new users to register accounts"},
		"features.api_enabled":          {Value: "true", Type: "boolean", Description: "Enable JSON API endpoints"},
		"features.weather_alerts":       {Value: "true", Type: "boolean", Description: "Enable weather alert notifications"},

		// Backup settings
		"backup.enabled":       {Value: "true", Type: "boolean", Description: "Enable automatic backup system"},
		"backup.interval":      {Value: "6", Type: "number", Description: "Backup interval in hours"},
		"backup.retention":     {Value: "30", Type: "number", Description: "Number of days to retain backups"},
		"backup.location":      {Value: defaultBackupPath, Type: "string", Description: "Directory path for storing backups"},

		// Logging settings
		"logging.level":         {Value: "info", Type: "string", Description: "Log level: debug, info, warn, error"},
		"logging.format":        {Value: "apache", Type: "string", Description: "Log format: apache, json, or plain"},
		"logging.access_log":    {Value: "true", Type: "boolean", Description: "Enable HTTP access logging"},
		"logging.error_log":     {Value: "true", Type: "boolean", Description: "Enable error logging"},
		"logging.audit_log":     {Value: "true", Type: "boolean", Description: "Enable audit logging for security events"},
		"logging.rotation_days": {Value: "30", Type: "number", Description: "Number of days to retain log files before rotation"},

		// SMTP settings
		"smtp.enabled":      {Value: "false", Type: "boolean", Description: "Enable email notifications via SMTP"},
		"smtp.host":         {Value: "", Type: "string", Description: "SMTP server hostname or IP address"},
		"smtp.port":         {Value: "587", Type: "number", Description: "SMTP server port (usually 587 for TLS, 465 for SSL)"},
		"smtp.username":     {Value: "", Type: "string", Description: "SMTP authentication username"},
		"smtp.password":     {Value: "", Type: "string", Description: "SMTP authentication password"},
		"smtp.from_address": {Value: "", Type: "string", Description: "Email address to send notifications from"},
		"smtp.from_name":    {Value: "", Type: "string", Description: "Display name for sent emails"},
		"smtp.use_tls":      {Value: "true", Type: "boolean", Description: "Use TLS/STARTTLS for secure SMTP connections"},
	}

	for key, setting := range defaults {
		// Only set if doesn't exist
		_, err := m.Get(key)
		if err == sql.ErrNoRows {
			if err := m.SetWithDescription(key, setting.Value, setting.Type, setting.Description); err != nil {
				return fmt.Errorf("failed to set default %s: %w", key, err)
			}
		}
	}

	return nil
}
