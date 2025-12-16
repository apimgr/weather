package config

import (
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// Config represents the application configuration
type Config struct {
	Version   string `yaml:"version"`
	BuildDate string `yaml:"build_date"`
	Mode      string `yaml:"mode"` // development, production

	Server ServerConfig `yaml:"server"`
}

// ServerConfig represents server-specific configuration
type ServerConfig struct {
	Branding BrandingConfig `yaml:"branding"`
}

// BrandingConfig represents branding configuration
type BrandingConfig struct {
	Title       string `yaml:"title"`
	Description string `yaml:"description"`
}

// LoadConfig loads configuration from server.yml
func LoadConfig() (*Config, error) {
	// Default config
	cfg := &Config{
		Version:   "1.0.0",
		BuildDate: "2025-12-11",
		Mode:      "production",
		Server: ServerConfig{
			Branding: BrandingConfig{
				Title:       "Weather Service",
				Description: "Professional weather tracking and forecasting service",
			},
		},
	}

	// Try to load from server.yml in project root
	configPath := findConfigFile()
	if configPath == "" {
		// No config file found, use defaults
		return cfg, nil
	}

	// Read config file
	data, err := os.ReadFile(configPath)
	if err != nil {
		return cfg, err
	}

	// Parse YAML
	if err := yaml.Unmarshal(data, cfg); err != nil {
		return cfg, err
	}

	return cfg, nil
}

// findConfigFile searches for server.yml in common locations
func findConfigFile() string {
	// Get current working directory
	cwd, err := os.Getwd()
	if err != nil {
		return ""
	}

	// Search paths (in order of priority)
	searchPaths := []string{
		filepath.Join(cwd, "server.yml"),
		filepath.Join(cwd, "../server.yml"),
		filepath.Join(cwd, "../../server.yml"),
		"/etc/weather/server.yml",
		"/opt/weather/server.yml",
	}

	for _, path := range searchPaths {
		if _, err := os.Stat(path); err == nil {
			return path
		}
	}

	return ""
}
