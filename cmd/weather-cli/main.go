package main

import (
	"fmt"
	"os"

	"github.com/apimgr/weather/src/client"
)

var (
	// Version info (set via ldflags during build)
	Version   = "dev"
	GitCommit = "unknown"
	BuildDate = "unknown"
)

func main() {
	// Set version info for the client
	client.Version = Version
	client.GitCommit = GitCommit
	client.BuildDate = BuildDate

	// Execute CLI
	if err := client.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)

		// Exit with appropriate code based on error type
		if exitErr, ok := err.(*client.ExitError); ok {
			os.Exit(exitErr.Code)
		}
		os.Exit(1)
	}
}
