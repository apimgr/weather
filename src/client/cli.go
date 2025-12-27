package client

import (
	"flag"
	"fmt"
	"os"
	"strings"
)

// Execute is the main entry point for the CLI
func Execute() error {
	// Global flags
	flagSet := flag.NewFlagSet("weather-cli", flag.ContinueOnError)
	flagSet.Usage = func() {
		printUsage()
	}

	// Define global flags
	serverFlag := flagSet.String("server", "", "Server URL (overrides config)")
	tokenFlag := flagSet.String("token", "", "API token (overrides config)")
	outputFlag := flagSet.String("output", "", "Output format: json, table, plain (overrides config)")
	configFlag := flagSet.String("config", "", "Config file path (default: ~/.config/weather/cli.yml)")
	tuiFlag := flagSet.Bool("tui", false, "Launch TUI mode")
	noColorFlag := flagSet.Bool("no-color", false, "Disable colored output")
	timeoutFlag := flagSet.Int("timeout", 0, "Request timeout in seconds (overrides config)")
	versionFlag := flagSet.Bool("version", false, "Show version information")
	helpFlag := flagSet.Bool("help", false, "Show help")

	// Parse flags
	if err := flagSet.Parse(os.Args[1:]); err != nil {
		return NewUsageError(err.Error())
	}

	// Show version
	if *versionFlag {
		return printVersion()
	}

	// Show help
	if *helpFlag || (flagSet.NArg() == 0 && !*tuiFlag) {
		printUsage()
		return nil
	}

	// Load config
	config, err := LoadConfig()
	if err != nil {
		return err
	}

	// Override config with flags
	if *serverFlag != "" {
		config.Server = *serverFlag
	}
	if *tokenFlag != "" {
		config.Token = *tokenFlag
	}
	if *outputFlag != "" {
		config.Output = *outputFlag
	}
	if *noColorFlag {
		config.NoColor = true
	}
	if *timeoutFlag > 0 {
		config.Timeout = *timeoutFlag
	}

	// TUI mode
	if *tuiFlag {
		return runTUI(config)
	}

	// Get command
	args := flagSet.Args()
	if len(args) == 0 {
		return NewUsageError("no command specified")
	}

	command := args[0]
	commandArgs := args[1:]

	// Route to appropriate command handler
	switch command {
	case "config":
		return handleConfigCommand(commandArgs)
	case "version":
		return printVersion()
	case "current":
		return handleCurrentCommand(config, commandArgs)
	case "forecast":
		return handleForecastCommand(config, commandArgs)
	case "alerts":
		return handleAlertsCommand(config, commandArgs)
	case "moon":
		return handleMoonCommand(config, commandArgs)
	case "history":
		return handleHistoryCommand(config, commandArgs)
	case "tui":
		return runTUI(config)
	default:
		return NewUsageError(fmt.Sprintf("unknown command: %s", command))
	}
}

// printUsage prints the usage information
func printUsage() {
	fmt.Println("Weather CLI - Command-line interface for Weather Service")
	fmt.Println()
	fmt.Println("Usage:")
	fmt.Println("  weather-cli [flags] <command> [args]")
	fmt.Println()
	fmt.Println("Commands:")
	fmt.Println("  config       Manage configuration (init, show, get, set)")
	fmt.Println("  version      Show version information")
	fmt.Println("  current      Get current weather")
	fmt.Println("  forecast     Get weather forecast")
	fmt.Println("  alerts       Get weather alerts")
	fmt.Println("  moon         Get moon phase information")
	fmt.Println("  history      Get historical weather data")
	fmt.Println("  tui          Launch interactive TUI mode")
	fmt.Println()
	fmt.Println("Global Flags:")
	fmt.Println("  --server <url>       Server URL (default: http://localhost:64948)")
	fmt.Println("  --token <token>      API authentication token")
	fmt.Println("  --output <format>    Output format: json, table, plain (default: table)")
	fmt.Println("  --config <path>      Config file path")
	fmt.Println("  --tui                Launch TUI mode")
	fmt.Println("  --no-color           Disable colored output")
	fmt.Println("  --timeout <seconds>  Request timeout (default: 30)")
	fmt.Println("  --version            Show version information")
	fmt.Println("  --help               Show this help message")
	fmt.Println()
	fmt.Println("Examples:")
	fmt.Println("  weather-cli config init")
	fmt.Println("  weather-cli config set server http://localhost:64948")
	fmt.Println("  weather-cli config set token YOUR_API_TOKEN")
	fmt.Println("  weather-cli current --lat 40.7128 --lon -74.0060")
	fmt.Println("  weather-cli forecast --zip 10001")
	fmt.Println("  weather-cli alerts --location \"New York, NY\"")
	fmt.Println("  weather-cli moon")
	fmt.Println("  weather-cli tui")
	fmt.Println()
	fmt.Println("Configuration:")
	fmt.Println("  Config file location: ~/.config/weather/cli.yml")
	fmt.Println("  Initialize config: weather-cli config init")
	fmt.Println()
}

// printVersion prints version information
func printVersion() error {
	fmt.Printf("weather-cli version %s\n", Version)
	fmt.Printf("Git commit: %s\n", GitCommit)
	fmt.Printf("Build date: %s\n", BuildDate)
	return nil
}

// handleConfigCommand handles config subcommands
func handleConfigCommand(args []string) error {
	if len(args) == 0 {
		return NewUsageError("config command requires a subcommand (init, show, get, set)")
	}

	subcommand := args[0]

	switch subcommand {
	case "init":
		return InitConfig()

	case "show":
		config, err := LoadConfig()
		if err != nil {
			return err
		}
		data, err := os.ReadFile(mustConfigPath())
		if err == nil {
			fmt.Println(string(data))
		} else {
			// If file doesn't exist, show default config
			fmt.Printf("server: %s\n", config.Server)
			fmt.Printf("token: %s\n", config.Token)
			fmt.Printf("output: %s\n", config.Output)
			fmt.Printf("timeout: %d\n", config.Timeout)
			fmt.Printf("no_color: %t\n", config.NoColor)
		}
		return nil

	case "get":
		if len(args) < 2 {
			return NewUsageError("config get requires a key")
		}
		value, err := GetConfigValue(args[1])
		if err != nil {
			return err
		}
		fmt.Println(value)
		return nil

	case "set":
		if len(args) < 3 {
			return NewUsageError("config set requires a key and value")
		}
		if err := SetConfigValue(args[1], strings.Join(args[2:], " ")); err != nil {
			return err
		}
		fmt.Printf("Configuration updated: %s = %s\n", args[1], strings.Join(args[2:], " "))
		return nil

	default:
		return NewUsageError(fmt.Sprintf("unknown config subcommand: %s", subcommand))
	}
}

// mustConfigPath returns the config path or panics
func mustConfigPath() string {
	path, err := ConfigPath()
	if err != nil {
		panic(err)
	}
	return path
}
