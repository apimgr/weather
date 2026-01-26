package main

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"golang.org/x/term"
)

// Execute is the main entry point for the CLI
// Per AI.md PART 36: Automatic mode detection, no --tui flag
func Execute() error {
	// Get actual binary name for display
	// Per AI.md PART 36: Display uses filepath.Base(os.Args[0])
	binaryName := filepath.Base(os.Args[0])

	// Check for exit-immediately flags first (before any parsing)
	// Per AI.md PART 36: -h, --help, -v, --version exit immediately (never TUI)
	for _, arg := range os.Args[1:] {
		switch arg {
		case "-h", "--help":
			printUsage(binaryName)
			return nil
		case "-v", "--version":
			return printVersion(binaryName)
		}
	}

	// Global flags
	flagSet := flag.NewFlagSet(binaryName, flag.ContinueOnError)
	flagSet.Usage = func() {
		printUsage(binaryName)
	}

	// Define global flags
	// Per AI.md PART 34: Required flags
	serverFlag := flagSet.String("server", "", "Server URL (overrides config)")
	tokenFlag := flagSet.String("token", "", "API token (overrides config)")
	tokenFileFlag := flagSet.String("token-file", "", "Read token from file")
	userFlag := flagSet.String("user", "", "User or org context (@user, +org, or auto-detect)")
	outputFlag := flagSet.String("output", "", "Output format: json, table, plain (overrides config)")
	configFlag := flagSet.String("config", "", "CLIConfig file path (default: ~/.config/apimgr/weather/cli.yml)")
	noColorFlag := flagSet.Bool("no-color", false, "Disable colored output")
	timeoutFlag := flagSet.Int("timeout", 0, "Request timeout in seconds (overrides config)")
	debugFlag := flagSet.Bool("debug", false, "Enable debug output")
	tuiFlag := flagSet.Bool("tui", false, "Launch TUI mode")

	// Parse flags
	if err := flagSet.Parse(os.Args[1:]); err != nil {
		return NewUsageError(err.Error())
	}

	// Load config (from specified path or default)
	config, err := LoadConfigFromPath(*configFlag)
	if err != nil {
		return err
	}

	// Get token using priority order (flag, file, env, config, token file)
	// Per AI.md PART 36: Token sources priority
	config.Token = GetToken(*tokenFlag, *tokenFileFlag, config)

	// Override config with flags
	if *serverFlag != "" {
		config.Server = *serverFlag
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
	if *userFlag != "" {
		config.User = *userFlag
	}

	// Detect mode
	// Per AI.md PART 34: --tui flag forces TUI mode, otherwise auto-detect
	mode := detectMode(flagSet.Args(), *tuiFlag)

	if *debugFlag {
		fmt.Fprintf(os.Stderr, "[DEBUG] Mode: %s\n", mode)
		fmt.Fprintf(os.Stderr, "[DEBUG] Server: %s\n", config.Server)
		fmt.Fprintf(os.Stderr, "[DEBUG] Token: %s\n", maskToken(config.Token))
	}

	// Handle based on mode
	switch mode {
	case "tui":
		return runTUI(config)
	case "plain":
		// Force plain output for non-TTY
		config.NoColor = true
		fallthrough
	case "cli":
		return handleCommand(config, flagSet.Args(), binaryName)
	default:
		return NewUsageError("failed to detect mode")
	}
}

// detectMode determines if we should use TUI or CLI mode
// Per AI.md PART 34: --tui flag forces TUI, otherwise smart detection
func detectMode(args []string, tuiFlag bool) string {
	// Not a terminal = plain output (regardless of --tui flag)
	if !term.IsTerminal(int(os.Stdout.Fd())) {
		return "plain"
	}

	// --tui flag explicitly forces TUI mode
	// Per AI.md PART 34 line 27805: --tui flag launches TUI mode
	if tuiFlag {
		return "tui"
	}

	// Config-only flags don't prevent TUI
	// Per AI.md PART 34: --config, --server, --token, --debug still allow TUI
	configFlags := map[string]bool{
		"--config": true, "--server": true, "--token": true, "--debug": true,
		"--token-file": true, "--user": true, "--no-color": true, "--timeout": true,
		"--output": true, "--tui": true,
	}

	// Check if any non-config args provided
	for _, arg := range args {
		if !strings.HasPrefix(arg, "-") {
			// Command or argument provided = CLI mode
			return "cli"
		}
		flag := strings.Split(arg, "=")[0]
		if !configFlags[flag] {
			// Unknown flag = CLI mode
			return "cli"
		}
	}

	// No args or config-only args = TUI mode
	// Per AI.md PART 34: Interactive terminal + no command = TUI
	return "tui"
}

// handleCommand routes to the appropriate command handler
func handleCommand(config *CLIConfig, args []string, binaryName string) error {
	if len(args) == 0 {
		printUsage(binaryName)
		return nil
	}

	command := args[0]
	commandArgs := args[1:]

	// Route to appropriate command handler
	switch command {
	case "config":
		return handleConfigCommand(commandArgs, binaryName)
	case "version":
		return printVersion(binaryName)
	case "tui":
		// Per AI.md PART 34 line 27884: tui command launches TUI mode
		return runTUI(config)
	case "login":
		return handleLoginCommand(config, commandArgs)
	case "logout":
		return handleLogoutCommand()
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
	case "earthquakes":
		return handleEarthquakesCommand(config, commandArgs)
	case "hurricanes":
		return handleHurricanesCommand(config, commandArgs)
	default:
		return NewUsageError(fmt.Sprintf("unknown command: %s", command))
	}
}

// printUsage prints the usage information
// Per AI.md PART 36: Shows actual binary name
func printUsage(binaryName string) {
	fmt.Printf("%s - Command-line interface for Weather Service\n", binaryName)
	fmt.Println()
	fmt.Println("Usage:")
	fmt.Printf("  %s [flags] <command> [args]\n", binaryName)
	fmt.Println()
	fmt.Println("Commands:")
	fmt.Println("  config       Manage configuration (init, show, get, set)")
	fmt.Println("  version      Show version information")
	fmt.Println("  tui          Launch interactive TUI")
	fmt.Println("  login        Authenticate and save token")
	fmt.Println("  logout       Remove saved token")
	fmt.Println("  current      Get current weather")
	fmt.Println("  forecast     Get weather forecast")
	fmt.Println("  alerts       Get weather alerts")
	fmt.Println("  moon         Get moon phase information")
	fmt.Println("  history      Get historical weather data")
	fmt.Println("  earthquakes  Get earthquake data")
	fmt.Println("  hurricanes   Get hurricane data")
	fmt.Println()
	fmt.Println("Global Flags:")
	fmt.Println("  --server <url>        Server URL (default: http://localhost:64948)")
	fmt.Println("  --token <token>       API authentication token")
	fmt.Println("  --token-file <path>   Read token from file")
	fmt.Println("  --user <name>         User or org context (@user, +org, or name)")
	fmt.Println("  --output <format>     Output format: json, table, plain (default: table)")
	fmt.Println("  --config <path>       CLIConfig file path")
	fmt.Println("  --no-color            Disable colored output")
	fmt.Println("  --timeout <seconds>   Request timeout (default: 30)")
	fmt.Println("  --debug               Enable debug output")
	fmt.Println("  --tui                 Launch TUI mode")
	fmt.Println("  --version, -v         Show version information")
	fmt.Println("  --help, -h            Show this help message")
	fmt.Println()
	fmt.Println("Token Sources (priority order):")
	fmt.Println("  1. --token flag")
	fmt.Println("  2. --token-file flag")
	fmt.Println("  3. WEATHER_TOKEN environment variable")
	fmt.Println("  4. CLIConfig file (~/.config/apimgr/weather/cli.yml)")
	fmt.Println("  5. Token file (~/.config/apimgr/weather/token)")
	fmt.Println()
	fmt.Println("User Context (--user flag):")
	fmt.Println("  @username    Explicit user context")
	fmt.Println("  +orgname     Explicit org context")
	fmt.Println("  name         Auto-detect user or org")
	fmt.Println()
	fmt.Println("Examples:")
	fmt.Printf("  %s config init\n", binaryName)
	fmt.Printf("  %s login\n", binaryName)
	fmt.Printf("  %s current --lat 40.7128 --lon -74.0060\n", binaryName)
	fmt.Printf("  %s forecast --location \"New York, NY\"\n", binaryName)
	fmt.Printf("  %s --user @alice current --zip 10001\n", binaryName)
	fmt.Println()
	fmt.Println("Configuration:")
	fmt.Println("  CLIConfig file: ~/.config/apimgr/weather/cli.yml")
	fmt.Println("  Token file:  ~/.config/apimgr/weather/token")
	fmt.Println()
	fmt.Println("TUI Mode:")
	fmt.Printf("  Run '%s' without arguments to launch interactive TUI\n", binaryName)
	fmt.Println()
}

// printVersion prints version information per AI.md PART 16 line 13778
// Per AI.md PART 34: Shows actual binary name, same format as server
// Format per spec:
//
//	{binaryName} v{version}
//	Built: {timestamp}
//	Go: {go_version}
//	OS/Arch: {os}/{arch}
func printVersion(binaryName string) error {
	fmt.Printf("%s v%s\n", binaryName, Version)
	fmt.Printf("Built: %s\n", BuildDate)
	fmt.Printf("Go: %s\n", runtime.Version())
	fmt.Printf("OS/Arch: %s/%s\n", runtime.GOOS, runtime.GOARCH)
	return nil
}

// handleConfigCommand handles config subcommands
func handleConfigCommand(args []string, binaryName string) error {
	if len(args) == 0 {
		return NewUsageError("config command requires a subcommand (init, show, get, set)")
	}

	subcommand := args[0]

	switch subcommand {
	case "init":
		return InitConfig()

	case "show":
		configPath, err := ConfigPath()
		if err != nil {
			return err
		}
		data, err := os.ReadFile(configPath)
		if err == nil {
			fmt.Println(string(data))
		} else {
			// If file doesn't exist, show default config
			config := DefaultConfig()
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

	case "path":
		configPath, err := ConfigPath()
		if err != nil {
			return err
		}
		fmt.Println(configPath)
		return nil

	default:
		return NewUsageError(fmt.Sprintf("unknown config subcommand: %s", subcommand))
	}
}

// handleLoginCommand handles the login command
// Per AI.md PART 36: weather-cli login saves token
func handleLoginCommand(config *CLIConfig, args []string) error {
	fmt.Printf("Server: %s\n", config.Server)
	fmt.Println()

	// Get username/email
	fmt.Print("Username or email: ")
	reader := bufio.NewReader(os.Stdin)
	identifier, err := reader.ReadString('\n')
	if err != nil {
		return NewAPIError(fmt.Sprintf("failed to read input: %v", err))
	}
	identifier = strings.TrimSpace(identifier)

	// Get password (hidden)
	fmt.Print("Password: ")
	password, err := term.ReadPassword(int(os.Stdin.Fd()))
	if err != nil {
		return NewAPIError(fmt.Sprintf("failed to read password: %v", err))
	}
	fmt.Println()

	// Make login request
	client := NewHTTPClient(config)
	loginReq := map[string]string{
		"identifier": identifier,
		"password":   string(password),
	}

	var result map[string]interface{}
	if err := client.PostJSON(config.GetAPIPath()+"/auth/login", loginReq, &result); err != nil {
		return err
	}

	// Extract token from response
	token, ok := result["token"].(string)
	if !ok {
		return NewAPIError("login successful but no token in response")
	}

	// Save token
	if err := SaveToken(token); err != nil {
		return err
	}

	tokenPath, _ := TokenPath()
	fmt.Printf("\nLogin successful! Token saved to %s\n", tokenPath)
	return nil
}

// handleLogoutCommand handles the logout command
func handleLogoutCommand() error {
	tokenPath, err := TokenPath()
	if err != nil {
		return err
	}

	// Remove token file
	if err := os.Remove(tokenPath); err != nil {
		if os.IsNotExist(err) {
			fmt.Println("No token file found. Already logged out.")
			return nil
		}
		return NewConfigError(fmt.Sprintf("failed to remove token: %v", err))
	}

	fmt.Println("Logged out. Token removed.")
	return nil
}

// handleEarthquakesCommand handles the earthquakes command
func handleEarthquakesCommand(config *CLIConfig, args []string) error {
	flagSet := flag.NewFlagSet("earthquakes", flag.ContinueOnError)
	minMagnitude := flagSet.Float64("min-mag", 0, "Minimum magnitude")
	limit := flagSet.Int("limit", 10, "Number of results")

	if err := flagSet.Parse(args); err != nil {
		return NewUsageError(err.Error())
	}

	path := fmt.Sprintf("%s/earthquakes?min_magnitude=%f&limit=%d", config.GetAPIPath(), *minMagnitude, *limit)

	client := NewHTTPClient(config)
	var result map[string]interface{}
	if err := client.GetJSON(path, &result); err != nil {
		return err
	}

	formatter := NewFormatter(config.Output, config.NoColor)
	fmt.Println(formatter.FormatJSON(result))

	return nil
}

// handleHurricanesCommand handles the hurricanes command
func handleHurricanesCommand(config *CLIConfig, args []string) error {
	flagSet := flag.NewFlagSet("hurricanes", flag.ContinueOnError)
	active := flagSet.Bool("active", true, "Show only active storms")

	if err := flagSet.Parse(args); err != nil {
		return NewUsageError(err.Error())
	}

	path := config.GetAPIPath() + "/hurricanes"
	if *active {
		path += "?active=true"
	}

	client := NewHTTPClient(config)
	var result map[string]interface{}
	if err := client.GetJSON(path, &result); err != nil {
		return err
	}

	formatter := NewFormatter(config.Output, config.NoColor)
	fmt.Println(formatter.FormatJSON(result))

	return nil
}

// maskToken masks a token for debug output
func maskToken(token string) string {
	if token == "" {
		return "(empty)"
	}
	if len(token) < 8 {
		return "***"
	}
	return token[:4] + "..." + token[len(token)-4:]
}
