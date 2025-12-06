package cli

import (
	"flag"
	"fmt"
	"os"
	"runtime"
)

// Version information (set via ldflags)
var (
	Version   = "dev"
	BuildDate = "unknown"
	GitCommit = "unknown"
)

// Command represents a CLI command
type Command struct {
	Name        string
	Description string
	Privileged  bool // Requires root/admin
	Handler     func(args []string) error
}

// CLI manages command-line interface
type CLI struct {
	commands map[string]*Command
	flags    *flag.FlagSet
}

// New creates a new CLI instance
func New() *CLI {
	return &CLI{
		commands: make(map[string]*Command),
		flags:    flag.NewFlagSet("weather", flag.ExitOnError),
	}
}

// RegisterCommand registers a new CLI command
func (c *CLI) RegisterCommand(cmd *Command) {
	c.commands[cmd.Name] = cmd
}

// Parse parses command line arguments
func (c *CLI) Parse(args []string) error {
	// Define standard flags
	var (
		showHelp      = c.flags.Bool("help", false, "Show this help message")
		showVersion   = c.flags.Bool("version", false, "Show version information")
		showStatus    = c.flags.Bool("status", false, "Show server status and health")
		healthcheck   = c.flags.Bool("healthcheck", false, "Run health check and exit (for Docker)")
		mode          = c.flags.String("mode", "", "Application mode: production or development")
		port          = c.flags.String("port", "", "Server port")
		address       = c.flags.String("address", "", "Listen address")
		dataDir       = c.flags.String("data", "", "Data directory")
		configDir     = c.flags.String("config", "", "Configuration directory")
		serviceCmd    = c.flags.String("service", "", "Service management: start, stop, restart, reload, --install, --uninstall")
		maintenanceCmd = c.flags.String("maintenance", "", "Maintenance: backup, restore, update, mode")
		updateCmd     = c.flags.String("update", "", "Update: check, yes, branch {stable|beta|daily}")
	)

	if err := c.flags.Parse(args); err != nil {
		return err
	}

	// Handle help flag
	if *showHelp {
		c.ShowHelp()
		os.Exit(0)
	}

	// Handle version flag
	if *showVersion {
		c.ShowVersion()
		os.Exit(0)
	}

	// Handle status flag - store for main.go to check
	if *showStatus {
		os.Setenv("CLI_STATUS_FLAG", "1")
		return nil
	}

	// Handle healthcheck flag - store for main.go to check
	if *healthcheck {
		os.Setenv("CLI_HEALTHCHECK_FLAG", "1")
		return nil
	}

	// Handle service command
	if *serviceCmd != "" {
		if cmd, ok := c.commands["service"]; ok {
			return cmd.Handler([]string{*serviceCmd})
		}
		return fmt.Errorf("service command not registered")
	}

	// Handle maintenance command
	if *maintenanceCmd != "" {
		if cmd, ok := c.commands["maintenance"]; ok {
			return cmd.Handler(c.flags.Args())
		}
		return fmt.Errorf("maintenance command not registered")
	}

	// Handle update command
	if *updateCmd != "" {
		if cmd, ok := c.commands["update"]; ok {
			return cmd.Handler(c.flags.Args())
		}
		return fmt.Errorf("update command not registered")
	}

	// Store flags for later access
	if *mode != "" {
		os.Setenv("MODE", *mode)
	}
	if *port != "" {
		os.Setenv("PORT", *port)
	}
	if *address != "" {
		os.Setenv("LISTEN", *address)
	}
	if *dataDir != "" {
		os.Setenv("DATA_DIR", *dataDir)
	}
	if *configDir != "" {
		os.Setenv("CONFIG_DIR", *configDir)
	}

	return nil
}

// ShowHelp displays help information
func (c *CLI) ShowHelp() {
	fmt.Println("Weather Service - Production-grade weather API server")
	fmt.Println()
	fmt.Printf("Version: %s\n", Version)
	fmt.Printf("Build Date: %s\n", BuildDate)
	fmt.Printf("Platform: %s/%s\n\n", runtime.GOOS, runtime.GOARCH)

	fmt.Println("USAGE:")
	fmt.Println("  weather [flags]")
	fmt.Println()

	fmt.Println("SERVER FLAGS:")
	fmt.Println("  --help                          Show this help message")
	fmt.Println("  --version                       Show version information")
	fmt.Println("  --mode {production|development} Set application mode (default: production)")
	fmt.Println("  --port {port}                   Server port (default: auto-assign 64000-64999)")
	fmt.Println("  --address {listen}              Listen address (default: auto-detect :: or 0.0.0.0)")
	fmt.Println("  --data {datadir}                Data directory")
	fmt.Println("  --config {etcdir}               Configuration directory")
	fmt.Println()

	fmt.Println("INFORMATION:")
	fmt.Println("  --status                        Show server status and health")
	fmt.Println("  --healthcheck                   Run health check and exit (for Docker)")
	fmt.Println()

	fmt.Println("SERVICE MANAGEMENT:")
	fmt.Println("  --service start                 Start the service")
	fmt.Println("  --service stop                  Stop the service")
	fmt.Println("  --service restart               Restart the service")
	fmt.Println("  --service reload                Reload configuration")
	fmt.Println("  --service --install             Install as system service")
	fmt.Println("  --service --uninstall           Uninstall system service")
	fmt.Println("  --service --disable             Disable system service")
	fmt.Println("  --service --help                Show service management help")
	fmt.Println()

	fmt.Println("MAINTENANCE:")
	fmt.Println("  --maintenance backup [file]     Backup database and configuration")
	fmt.Println("  --maintenance restore [file]    Restore from backup")
	fmt.Println("  --maintenance update            Update server configuration")
	fmt.Println("  --maintenance mode {prod|dev}   Set maintenance mode")
	fmt.Println()

	fmt.Println("UPDATE:")
	fmt.Println("  --update check                  Check for updates")
	fmt.Println("  --update yes                    Update to latest stable")
	fmt.Println("  --update branch stable          Update from stable branch")
	fmt.Println("  --update branch beta            Update from beta branch")
	fmt.Println("  --update branch daily           Update from daily build")
	fmt.Println()

	fmt.Println("ENVIRONMENT VARIABLES:")
	fmt.Println("  MODE              Application mode: production (default) or development")
	fmt.Println("  DEBUG             Enable debug mode (NEVER in production!)")
	fmt.Println("  PORT              Server port")
	fmt.Println("  LISTEN            Listen address")
	fmt.Println("  DATA_DIR          Data directory")
	fmt.Println("  CONFIG_DIR        Configuration directory")
	fmt.Println("  LOG_DIR           Log directory")
	fmt.Println("  DATABASE_DRIVER   Database driver (file, sqlite, postgres, etc.)")
	fmt.Println("  DATABASE_URL      Database connection string")
	fmt.Println()

	fmt.Println("EXAMPLES:")
	fmt.Println("  weather                                  # Start server (production mode)")
	fmt.Println("  weather --mode development               # Start in development mode")
	fmt.Println("  weather --port 8080                      # Start on port 8080")
	fmt.Println("  weather --service --install              # Install as system service")
	fmt.Println("  weather --maintenance backup backup.tar  # Create backup")
	fmt.Println("  weather --update check                   # Check for updates")
	fmt.Println()

	fmt.Println("DOCUMENTATION:")
	fmt.Println("  GitHub:  https://github.com/apimgr/weather")
	fmt.Println("  Docs:    https://weather.apimgr.us")
	fmt.Println()
}

// ShowVersion displays version information
func (c *CLI) ShowVersion() {
	fmt.Printf("Weather Service v%s\n", Version)
	fmt.Printf("Build Date: %s\n", BuildDate)
	fmt.Printf("Git Commit: %s\n", GitCommit)
	fmt.Printf("Go Version: %s\n", runtime.Version())
	fmt.Printf("Platform: %s/%s\n", runtime.GOOS, runtime.GOARCH)
}

// GetFlag returns a flag value
func (c *CLI) GetFlag(name string) interface{} {
	return c.flags.Lookup(name)
}

// IsFlagSet checks if a flag was set
func (c *CLI) IsFlagSet(name string) bool {
	found := false
	c.flags.Visit(func(f *flag.Flag) {
		if f.Name == name {
			found = true
		}
	})
	return found
}
