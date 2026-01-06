package client

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDefaultConfig(t *testing.T) {
	config := DefaultConfig()

	if config.Server != "http://localhost:64948" {
		t.Errorf("Expected default server to be http://localhost:64948, got %s", config.Server)
	}

	if config.Token != "" {
		t.Errorf("Expected default token to be empty, got %s", config.Token)
	}

	if config.Output != "table" {
		t.Errorf("Expected default output to be table, got %s", config.Output)
	}

	if config.Timeout != 30 {
		t.Errorf("Expected default timeout to be 30, got %d", config.Timeout)
	}

	if config.NoColor {
		t.Error("Expected default NoColor to be false")
	}
}

func TestConfigPath(t *testing.T) {
	path, err := ConfigPath()
	if err != nil {
		t.Fatalf("ConfigPath() failed: %v", err)
	}

	if !filepath.IsAbs(path) {
		t.Errorf("Expected absolute path, got %s", path)
	}

	if filepath.Base(path) != "cli.yml" {
		t.Errorf("Expected filename to be cli.yml, got %s", filepath.Base(path))
	}
}

func TestLoadConfigNonExistent(t *testing.T) {
	// Set a custom config path that doesn't exist
	origHome := os.Getenv("HOME")
	defer os.Setenv("HOME", origHome)

	tempDir := t.TempDir()
	os.Setenv("HOME", tempDir)

	config, err := LoadConfig()
	if err != nil {
		t.Fatalf("LoadConfig() failed: %v", err)
	}

	// Should return default config
	if config.Server != "http://localhost:64948" {
		t.Errorf("Expected default server, got %s", config.Server)
	}
}

func TestSaveAndLoadConfig(t *testing.T) {
	// Set up temporary directory for config
	origHome := os.Getenv("HOME")
	defer os.Setenv("HOME", origHome)

	tempDir := t.TempDir()
	os.Setenv("HOME", tempDir)

	// Create config
	config := &Config{
		Server:  "http://test.example.com",
		Token:   "test-token",
		Output:  "json",
		Timeout: 60,
		NoColor: true,
	}

	// Save config
	if err := SaveConfig(config); err != nil {
		t.Fatalf("SaveConfig() failed: %v", err)
	}

	// Load config
	loaded, err := LoadConfig()
	if err != nil {
		t.Fatalf("LoadConfig() failed: %v", err)
	}

	// Verify all fields
	if loaded.Server != config.Server {
		t.Errorf("Expected server %s, got %s", config.Server, loaded.Server)
	}
	if loaded.Token != config.Token {
		t.Errorf("Expected token %s, got %s", config.Token, loaded.Token)
	}
	if loaded.Output != config.Output {
		t.Errorf("Expected output %s, got %s", config.Output, loaded.Output)
	}
	if loaded.Timeout != config.Timeout {
		t.Errorf("Expected timeout %d, got %d", config.Timeout, loaded.Timeout)
	}
	if loaded.NoColor != config.NoColor {
		t.Errorf("Expected NoColor %t, got %t", config.NoColor, loaded.NoColor)
	}
}

func TestInitConfig(t *testing.T) {
	// Set up temporary directory for config
	origHome := os.Getenv("HOME")
	defer os.Setenv("HOME", origHome)

	tempDir := t.TempDir()
	os.Setenv("HOME", tempDir)

	// Initialize config
	if err := InitConfig(); err != nil {
		t.Fatalf("InitConfig() failed: %v", err)
	}

	// Verify config file exists
	configPath, _ := ConfigPath()
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		t.Error("Config file was not created")
	}

	// Try to init again (should fail)
	err := InitConfig()
	if err == nil {
		t.Error("Expected error when initializing existing config")
	}
	if exitErr, ok := err.(*ExitError); ok {
		if exitErr.Code != ExitConfigError {
			t.Errorf("Expected ExitConfigError, got exit code %d", exitErr.Code)
		}
	} else {
		t.Error("Expected ExitError type")
	}
}

func TestGetConfigValue(t *testing.T) {
	// Set up temporary directory for config
	origHome := os.Getenv("HOME")
	defer os.Setenv("HOME", origHome)

	tempDir := t.TempDir()
	os.Setenv("HOME", tempDir)

	// Save test config
	config := &Config{
		Server:  "http://test.example.com",
		Token:   "test-token",
		Output:  "json",
		Timeout: 45,
		NoColor: true,
	}
	SaveConfig(config)

	tests := []struct {
		key      string
		expected string
	}{
		{"server", "http://test.example.com"},
		{"token", "test-token"},
		{"output", "json"},
		{"timeout", "45"},
		{"no_color", "true"},
	}

	for _, tt := range tests {
		t.Run(tt.key, func(t *testing.T) {
			value, err := GetConfigValue(tt.key)
			if err != nil {
				t.Fatalf("GetConfigValue(%s) failed: %v", tt.key, err)
			}
			if value != tt.expected {
				t.Errorf("Expected %s, got %s", tt.expected, value)
			}
		})
	}

	// Test invalid key
	_, err := GetConfigValue("invalid_key")
	if err == nil {
		t.Error("Expected error for invalid key")
	}
}

func TestSetConfigValue(t *testing.T) {
	// Set up temporary directory for config
	origHome := os.Getenv("HOME")
	defer os.Setenv("HOME", origHome)

	tempDir := t.TempDir()
	os.Setenv("HOME", tempDir)

	// Initialize config
	InitConfig()

	tests := []struct {
		key   string
		value string
	}{
		{"server", "http://new.example.com"},
		{"token", "new-token"},
		{"output", "json"},
		{"timeout", "60"},
		{"no_color", "true"},
	}

	for _, tt := range tests {
		t.Run(tt.key, func(t *testing.T) {
			if err := SetConfigValue(tt.key, tt.value); err != nil {
				t.Fatalf("SetConfigValue(%s, %s) failed: %v", tt.key, tt.value, err)
			}

			// Verify the value was set
			value, err := GetConfigValue(tt.key)
			if err != nil {
				t.Fatalf("GetConfigValue(%s) failed: %v", tt.key, err)
			}
			if value != tt.value {
				t.Errorf("Expected %s, got %s", tt.value, value)
			}
		})
	}
}

func TestSetConfigValueValidation(t *testing.T) {
	// Set up temporary directory for config
	origHome := os.Getenv("HOME")
	defer os.Setenv("HOME", origHome)

	tempDir := t.TempDir()
	os.Setenv("HOME", tempDir)

	// Initialize config
	InitConfig()

	// Test invalid output format
	err := SetConfigValue("output", "invalid")
	if err == nil {
		t.Error("Expected error for invalid output format")
	}

	// Test invalid timeout
	err = SetConfigValue("timeout", "abc")
	if err == nil {
		t.Error("Expected error for invalid timeout")
	}

	// Test timeout out of range
	err = SetConfigValue("timeout", "500")
	if err == nil {
		t.Error("Expected error for timeout out of range")
	}

	// Test invalid no_color
	err = SetConfigValue("no_color", "maybe")
	if err == nil {
		t.Error("Expected error for invalid no_color value")
	}

	// Test invalid key
	err = SetConfigValue("invalid_key", "value")
	if err == nil {
		t.Error("Expected error for invalid key")
	}
}
