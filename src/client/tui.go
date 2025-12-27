package client

import (
	"fmt"
)

// runTUI launches the interactive TUI mode
// TODO: Implement full TUI with bubbletea library and Dracula theme
func runTUI(config *Config) error {
	fmt.Println("╔═════════════════════════════════════════════════╗")
	fmt.Println("║      Weather CLI - Interactive Mode (TUI)      ║")
	fmt.Println("╚═════════════════════════════════════════════════╝")
	fmt.Println()
	fmt.Println("TUI Mode Coming Soon!")
	fmt.Println()
	fmt.Println("The full Terminal User Interface (TUI) mode with:")
	fmt.Println("  • Interactive weather browsing")
	fmt.Println("  • Real-time updates")
	fmt.Println("  • Keyboard navigation")
	fmt.Println("  • Dracula theme matching the web interface")
	fmt.Println()
	fmt.Println("Will be implemented using Bubble Tea framework.")
	fmt.Println()
	fmt.Println("For now, use the standard CLI commands:")
	fmt.Println("  weather-cli current --location \"Your City\"")
	fmt.Println("  weather-cli forecast --zip 10001")
	fmt.Println("  weather-cli alerts --lat 40.7128 --lon -74.0060")
	fmt.Println()
	fmt.Println("Server: " + config.Server)
	fmt.Println()

	return nil
}
