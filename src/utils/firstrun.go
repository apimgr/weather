package utils

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"time"

	"gopkg.in/yaml.v3"
)

// FirstRunConfig represents first-run auto-detected configuration
type FirstRunConfig struct {
	SMTPHost      string
	SMTPPort      int
	HTTPPort      int
	SetupToken    string
	IsFirstRun    bool
	IsDockerized  bool
	TorEnabled    bool
	OnionAddress  string
}

// DetectFirstRun checks if this is the first time the server is running
func DetectFirstRun(dataDir string) bool {
	serverDBPath := filepath.Join(dataDir, "server.db")
	_, err := os.Stat(serverDBPath)
	return os.IsNotExist(err)
}

// GenerateSetupToken generates a cryptographically secure one-time setup token
func GenerateSetupToken() (string, error) {
	// 128 bits
	bytes := make([]byte, 16)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}

// AutoDetectSMTP attempts to auto-detect available SMTP servers
// Tries common locations: localhost:25, docker bridge, etc.
func AutoDetectSMTP() (string, int) {
	// List of SMTP servers to try, in order of preference
	smtpServers := []struct {
		Host string
		Port int
	}{
		{"localhost", 25},
		{"localhost", 587},
		{"127.0.0.1", 25},
		// Docker bridge
		{"172.17.0.1", 25},
		// Docker Desktop
		{"host.docker.internal", 25},
		// VirtualBox NAT
		{"10.0.2.2", 25},
	}

	for _, server := range smtpServers {
		addr := fmt.Sprintf("%s:%d", server.Host, server.Port)
		conn, err := net.DialTimeout("tcp", addr, 2*time.Second)
		if err == nil {
			conn.Close()
			return server.Host, server.Port
		}
	}

	// Default fallback
	return "localhost", 25
}

// SelectRandomPort selects a random port in the 64000-64999 range
func SelectRandomPort() int {
	// Try random ports until we find an available one
	for attempts := 0; attempts < 100; attempts++ {
		port := MinPort + (int(time.Now().UnixNano()) % (MaxPort - MinPort + 1))
		if IsPortAvailable(port) {
			return port
		}
	}
	// Fallback to 64948 if we can't find a random one
	return 64948
}

// IsDockerized detects if running inside a Docker container
func IsDockerized() bool {
	// Check for .dockerenv file
	if _, err := os.Stat("/.dockerenv"); err == nil {
		return true
	}

	// Check cgroup for docker
	data, err := os.ReadFile("/proc/1/cgroup")
	if err == nil {
		content := string(data)
		if len(content) > 0 && (containsString(content, "docker") || containsString(content, "containerd")) {
			return true
		}
	}

	return false
}

// CreateDefaultServerYML creates server.yml with auto-detected settings
func CreateDefaultServerYML(configPath string, smtpHost string, smtpPort int) error {
	// Ensure config directory exists
	configDir := filepath.Dir(configPath)
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	// Default configuration
	config := map[string]interface{}{
		"mode": "production",
		"server": map[string]interface{}{
			"branding": map[string]interface{}{
				"title":       "Weather Service",
				"description": "Professional weather tracking and forecasting service with real-time updates, severe weather alerts, earthquake monitoring, and moon phase tracking",
				"tagline":     "Your reliable weather companion",
				"logo_url":    "",
				"favicon_url": "",
			},
			"seo": map[string]interface{}{
				"keywords":      "weather, forecast, alerts, tracking, severe weather, earthquakes, moon phases",
				"author":        "Weather Service",
				"canonical_url": "",
				"og_image":      "",
			},
			"theme": map[string]interface{}{
				"primary_color":   "#3b82f6",
				"secondary_color": "#8b5cf6",
				"dark_mode":       false,
			},
			"email": map[string]interface{}{
				"host":      smtpHost,
				"port":      smtpPort,
				"username":  "",
				"password":  "",
				"from":      "noreply@localhost",
				"from_name": "Weather Service",
				"use_tls":   false,
			},
			"notifications": map[string]interface{}{
				"enabled":         true,
				"email_enabled":   true,
				"webhook_enabled": false,
			},
			"rate_limit": map[string]interface{}{
				"enabled":          true,
				"requests_per_min": 60,
				"burst_size":       10,
			},
			"web": map[string]interface{}{
				"robots_txt":   "User-agent: *\nAllow: /\nDisallow: /admin\nDisallow: /api/v1/admin",
				"security_txt": "Contact: mailto:security@example.com\nExpires: 2026-12-31T23:59:59.000Z\nPreferred-Languages: en",
			},
			"tor": map[string]interface{}{
				"enabled":    false,
				"onion_addr": "",
			},
			"features": map[string]interface{}{
				"earthquakes":    true,
				"hurricanes":     true,
				"moon_phases":    true,
				"severe_weather": true,
				"audit_log":      true,
			},
		},
	}

	// Marshal to YAML
	data, err := yaml.Marshal(config)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	// Write to file
	header := "# Weather Service Configuration\n"
	header += "# Auto-generated on first run: " + time.Now().Format(time.RFC3339) + "\n"
	header += "# SMTP auto-detected: " + smtpHost + ":" + fmt.Sprintf("%d", smtpPort) + "\n\n"

	fullContent := header + string(data)
	if err := os.WriteFile(configPath, []byte(fullContent), 0644); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	return nil
}

// Helper function
func containsString(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > len(substr) &&
		(s[:len(substr)] == substr || s[len(s)-len(substr):] == substr ||
		 findSubstring(s, substr)))
}

func findSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
