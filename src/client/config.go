package client

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// Config represents the CLI client configuration
type Config struct {
	// Server URL
	Server  string `yaml:"server"`
	// API token
	Token   string `yaml:"token"`
	// Output format (json, table, plain)
	Output  string `yaml:"output"`
	// Request timeout in seconds
	Timeout int    `yaml:"timeout"`
	NoColor bool   `yaml:"no_color"`// Disable colored output
}

// DefaultConfig returns the default configuration
func DefaultConfig() *Config {
	return &Config{
		Server:  "http://localhost:64948",
		Token:   "",
		Output:  "table",
		Timeout: 30,
		NoColor: false,
	}
}

// ConfigPath returns the path to the config file
// Default: ~/.config/weather/cli.yml
func ConfigPath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("failed to get home directory: %w", err)
	}

	configDir := filepath.Join(home, ".config", "weather")
	return filepath.Join(configDir, "cli.yml"), nil
}

// LoadConfig loads the configuration from the config file
func LoadConfig() (*Config, error) {
	configPath, err := ConfigPath()
	if err != nil {
		return nil, NewConfigError(fmt.Sprintf("failed to get config path: %v", err))
	}

	// If config doesn't exist, return default
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		return DefaultConfig(), nil
	}

	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, NewConfigError(fmt.Sprintf("failed to read config: %v", err))
	}

	var config Config
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, NewConfigError(fmt.Sprintf("failed to parse config: %v", err))
	}

	return &config, nil
}

// SaveConfig saves the configuration to the config file
func SaveConfig(config *Config) error {
	configPath, err := ConfigPath()
	if err != nil {
		return NewConfigError(fmt.Sprintf("failed to get config path: %v", err))
	}

	// Create config directory if it doesn't exist
	configDir := filepath.Dir(configPath)
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return NewConfigError(fmt.Sprintf("failed to create config directory: %v", err))
	}

	data, err := yaml.Marshal(config)
	if err != nil {
		return NewConfigError(fmt.Sprintf("failed to marshal config: %v", err))
	}

	if err := os.WriteFile(configPath, data, 0600); err != nil {
		return NewConfigError(fmt.Sprintf("failed to write config: %v", err))
	}

	return nil
}

// InitConfig initializes a new configuration file
func InitConfig() error {
	configPath, err := ConfigPath()
	if err != nil {
		return NewConfigError(fmt.Sprintf("failed to get config path: %v", err))
	}

	// Check if config already exists
	if _, err := os.Stat(configPath); err == nil {
		return NewConfigError("config file already exists")
	}

	// Create default config
	config := DefaultConfig()
	if err := SaveConfig(config); err != nil {
		return err
	}

	fmt.Printf("Configuration file created at: %s\n", configPath)
	return nil
}

// GetConfigValue returns a specific configuration value
func GetConfigValue(key string) (string, error) {
	config, err := LoadConfig()
	if err != nil {
		return "", err
	}

	switch key {
	case "server":
		return config.Server, nil
	case "token":
		return config.Token, nil
	case "output":
		return config.Output, nil
	case "timeout":
		return fmt.Sprintf("%d", config.Timeout), nil
	case "no_color":
		return fmt.Sprintf("%t", config.NoColor), nil
	default:
		return "", NewConfigError(fmt.Sprintf("unknown config key: %s", key))
	}
}

// SetConfigValue sets a specific configuration value
func SetConfigValue(key, value string) error {
	config, err := LoadConfig()
	if err != nil {
		return err
	}

	switch key {
	case "server":
		config.Server = value
	case "token":
		config.Token = value
	case "output":
		if value != "json" && value != "table" && value != "plain" {
			return NewConfigError("output must be json, table, or plain")
		}
		config.Output = value
	case "timeout":
		var timeout int
		if _, err := fmt.Sscanf(value, "%d", &timeout); err != nil {
			return NewConfigError("timeout must be a number")
		}
		if timeout < 1 || timeout > 300 {
			return NewConfigError("timeout must be between 1 and 300 seconds")
		}
		config.Timeout = timeout
	case "no_color":
		var noColor bool
		if _, err := fmt.Sscanf(value, "%t", &noColor); err != nil {
			return NewConfigError("no_color must be true or false")
		}
		config.NoColor = noColor
	default:
		return NewConfigError(fmt.Sprintf("unknown config key: %s", key))
	}

	return SaveConfig(config)
}
