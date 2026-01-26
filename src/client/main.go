//go:build !test
// +build !test

// Package main provides the CLI client entry point
// AI.md: All source code in src/ directory
package main

import (
	"fmt"
	"os"
)

func main() {
	// Execute CLI
	if err := Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)

		// Exit with appropriate code based on error type
		if exitErr, ok := err.(*ExitError); ok {
			os.Exit(exitErr.Code)
		}
		os.Exit(1)
	}
}
