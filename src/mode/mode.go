package mode

import (
	"os"
	"runtime"
	"strings"

	"github.com/apimgr/weather/src/config"
)

var (
	currentMode  = Production
	debugEnabled = false
)

// Mode represents the application execution mode
type Mode int

const (
	// Production mode (default) - strict security, minimal logging
	Production Mode = iota
	// Development mode - relaxed security, verbose logging
	Development
)

// String returns the string representation of the mode
func (m Mode) String() string {
	switch m {
	case Development:
		return "development"
	default:
		return "production"
	}
}

// Set sets the application mode
func Set(m string) {
	switch strings.ToLower(m) {
	case "dev", "development":
		currentMode = Development
	default:
		currentMode = Production
	}
	updateProfilingSettings()
}

// SetDebug enables or disables debug mode
func SetDebug(enabled bool) {
	debugEnabled = enabled
	updateProfilingSettings()
}

// updateProfilingSettings enables/disables profiling based on debug flag
func updateProfilingSettings() {
	if debugEnabled {
		// Enable profiling when debug is on
		runtime.SetBlockProfileRate(1)
		runtime.SetMutexProfileFraction(1)
	} else {
		// Disable profiling when debug is off
		runtime.SetBlockProfileRate(0)
		runtime.SetMutexProfileFraction(0)
	}
}

// Current returns the current mode
func Current() Mode {
	return currentMode
}

// IsDevelopment returns true if in development mode
func IsDevelopment() bool {
	return currentMode == Development
}

// IsProduction returns true if in production mode
func IsProduction() bool {
	return currentMode == Production
}

// IsDebug returns true if debug mode is enabled (--debug or DEBUG=true)
func IsDebug() bool {
	return debugEnabled
}

// ModeString returns mode string with debug suffix if enabled
func ModeString() string {
	s := currentMode.String()
	if debugEnabled {
		s += " [debugging]"
	}
	return s
}

// FromEnv sets mode and debug from environment variables
func FromEnv() {
	if m := os.Getenv("MODE"); m != "" {
		Set(m)
	}
	if config.IsTruthy(os.Getenv("DEBUG")) {
		SetDebug(true)
	}
}
