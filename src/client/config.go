package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

// CLIConfig represents the CLI client configuration
// Per AI.md PART 33: CLI configuration with cluster support
type CLIConfig struct {
	// Server URL (deprecated, use ServerPrimary)
	Server string `yaml:"server"`
	// Primary server URL
	// Per AI.md PART 33: server.primary in cli.yml
	ServerPrimary string `yaml:"server_primary,omitempty"`
	// Cluster node URLs for failover
	// Per AI.md PART 33: server.cluster in cli.yml
	ServerCluster []string `yaml:"server_cluster,omitempty"`
	// API token
	Token string `yaml:"token"`
	// Output format (json, table, plain)
	Output string `yaml:"output"`
	// Request timeout in seconds
	Timeout int `yaml:"timeout"`
	// Disable colored output
	NoColor bool `yaml:"no_color"`
	// User/org context (--user flag value)
	User string `yaml:"user,omitempty"`
	// API base path (e.g., /api/v1)
	// Per AI.md PART 14: configurable API version
	APIBasePath string `yaml:"api_base_path,omitempty"`
}

// GetAPIPath returns the API base path (default: /api/v1)
func (c *CLIConfig) GetAPIPath() string {
	if c.APIBasePath == "" {
		return "/api/v1"
	}
	return c.APIBasePath
}

// GetPrimaryServer returns the primary server URL
// Per AI.md PART 33: Priority order for server resolution
func (c *CLIConfig) GetPrimaryServer() string {
	// 1. ServerPrimary (new field)
	if c.ServerPrimary != "" {
		return c.ServerPrimary
	}
	// 2. Server (legacy/backwards compatibility)
	if c.Server != "" {
		return c.Server
	}
	// 3. Default
	return "http://localhost:64948"
}

// GetAllServers returns all available servers (primary + cluster) for failover
// Per AI.md PART 33: CLI cluster failover support
func (c *CLIConfig) GetAllServers() []string {
	servers := []string{}

	// Add primary server first
	primary := c.GetPrimaryServer()
	if primary != "" {
		servers = append(servers, primary)
	}

	// Add cluster nodes
	for _, node := range c.ServerCluster {
		if node != "" && node != primary {
			servers = append(servers, node)
		}
	}

	return servers
}

// Hardcoded project name for internal use (paths, env vars)
// Per AI.md PART 36: Internal uses hardcoded project name
const projectName = "weather"

// DefaultConfig returns the default configuration
func DefaultConfig() *CLIConfig {
	return &CLIConfig{
		Server:  "http://localhost:64948",
		Token:   "",
		Output:  "table",
		Timeout: 30,
		NoColor: false,
		User:    "",
	}
}

// ConfigDir returns the config directory path
// Per AI.md PART 36: ~/.config/apimgr/weather/
func ConfigDir() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("failed to get home directory: %w", err)
	}

	return filepath.Join(home, ".config", "apimgr", projectName), nil
}

// ConfigPath returns the path to the config file
// Per AI.md PART 36: ~/.config/apimgr/weather/cli.yml
func ConfigPath() (string, error) {
	configDir, err := ConfigDir()
	if err != nil {
		return "", err
	}

	return filepath.Join(configDir, "cli.yml"), nil
}

// TokenPath returns the path to the token file
// Per AI.md PART 36: ~/.config/apimgr/weather/token
func TokenPath() (string, error) {
	configDir, err := ConfigDir()
	if err != nil {
		return "", err
	}

	return filepath.Join(configDir, "token"), nil
}

// LoadConfig loads the configuration from the default config file
func LoadConfig() (*CLIConfig, error) {
	return LoadConfigFromPath("")
}

// LoadConfigFromPath loads configuration from a specific path
// If path is empty, uses the default config path
func LoadConfigFromPath(path string) (*CLIConfig, error) {
	var configPath string
	var err error

	if path != "" {
		configPath = path
	} else {
		configPath, err = ConfigPath()
		if err != nil {
			return nil, NewConfigError(fmt.Sprintf("failed to get config path: %v", err))
		}
	}

	// Start with defaults
	config := DefaultConfig()

	// If config exists, load it
	if _, err := os.Stat(configPath); err == nil {
		data, err := os.ReadFile(configPath)
		if err != nil {
			return nil, NewConfigError(fmt.Sprintf("failed to read config: %v", err))
		}

		if err := yaml.Unmarshal(data, config); err != nil {
			return nil, NewConfigError(fmt.Sprintf("failed to parse config: %v", err))
		}
	}

	return config, nil
}

// GetToken returns the API token using priority order per AI.md PART 36
// Priority: 1. explicit flag, 2. token-file flag, 3. env var, 4. config, 5. token file
func GetToken(flagToken, flagTokenFile string, config *CLIConfig) string {
	// 1. Explicit --token flag
	if flagToken != "" {
		return flagToken
	}

	// 2. --token-file flag
	if flagTokenFile != "" {
		if data, err := os.ReadFile(flagTokenFile); err == nil {
			return strings.TrimSpace(string(data))
		}
	}

	// 3. Environment variable: WEATHER_TOKEN
	// Per AI.md PART 36: os.Getenv(strings.ToUpper(projectName) + "_TOKEN")
	envKey := strings.ToUpper(projectName) + "_TOKEN"
	if token := os.Getenv(envKey); token != "" {
		return token
	}

	// 4. CLIConfig file token
	if config.Token != "" {
		return config.Token
	}

	// 5. Default token file
	tokenPath, err := TokenPath()
	if err == nil {
		if data, err := os.ReadFile(tokenPath); err == nil {
			return strings.TrimSpace(string(data))
		}
	}

	return "" // No token (anonymous access if allowed)
}

// SaveConfig saves the configuration to the config file
func SaveConfig(config *CLIConfig) error {
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

// SaveToken saves the token to the token file
// Per AI.md PART 36: weather-cli login saves to ~/.config/apimgr/weather/token
func SaveToken(token string) error {
	tokenPath, err := TokenPath()
	if err != nil {
		return NewConfigError(fmt.Sprintf("failed to get token path: %v", err))
	}

	// Create config directory if it doesn't exist
	configDir := filepath.Dir(tokenPath)
	if err := os.MkdirAll(configDir, 0700); err != nil {
		return NewConfigError(fmt.Sprintf("failed to create config directory: %v", err))
	}

	// Write token with restrictive permissions
	if err := os.WriteFile(tokenPath, []byte(token+"\n"), 0600); err != nil {
		return NewConfigError(fmt.Sprintf("failed to write token: %v", err))
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
	case "user":
		return config.User, nil
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
		switch strings.ToLower(value) {
		case "true", "1", "yes":
			config.NoColor = true
		case "false", "0", "no":
			config.NoColor = false
		default:
			return NewConfigError("no_color must be true or false")
		}
	case "user":
		config.User = value
	default:
		return NewConfigError(fmt.Sprintf("unknown config key: %s", key))
	}

	return SaveConfig(config)
}
