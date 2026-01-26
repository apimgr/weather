package mode

import (
	"os"
	"testing"
)

func TestSetAppMode(t *testing.T) {
	tests := []struct {
		input    string
		expected AppMode
	}{
		{"development", Development},
		{"dev", Development},
		{"production", Production},
		{"prod", Production},
		{"DEVELOPMENT", Development},
		{"PRODUCTION", Production},
		// Default to production
		{"anything", Production},
	}

	for _, tt := range tests {
		SetAppMode(tt.input)
		if Current() != tt.expected {
			t.Errorf("SetAppMode(%q): got %v, want %v", tt.input, Current(), tt.expected)
		}
	}
}

func TestIsAppModeDev(t *testing.T) {
	SetAppMode("development")
	if !IsAppModeDev() {
		t.Error("IsAppModeDev() should return true when mode is development")
	}
	if IsAppModeProd() {
		t.Error("IsAppModeProd() should return false when mode is development")
	}
}

func TestIsAppModeProd(t *testing.T) {
	SetAppMode("production")
	if !IsAppModeProd() {
		t.Error("IsAppModeProd() should return true when mode is production")
	}
	if IsAppModeDev() {
		t.Error("IsAppModeDev() should return false when mode is production")
	}
}

func TestSetDebugEnabled(t *testing.T) {
	SetDebugEnabled(true)
	if !IsDebugEnabled() {
		t.Error("IsDebugEnabled() should return true when debug is enabled")
	}

	SetDebugEnabled(false)
	if IsDebugEnabled() {
		t.Error("IsDebugEnabled() should return false when debug is disabled")
	}
}

func TestModeString(t *testing.T) {
	SetAppMode("production")
	SetDebugEnabled(false)
	if ModeString() != "production" {
		t.Errorf("ModeString() = %q, want %q", ModeString(), "production")
	}

	SetDebugEnabled(true)
	if ModeString() != "production [debugging]" {
		t.Errorf("ModeString() = %q, want %q", ModeString(), "production [debugging]")
	}

	SetAppMode("development")
	SetDebugEnabled(false)
	if ModeString() != "development" {
		t.Errorf("ModeString() = %q, want %q", ModeString(), "development")
	}

	SetDebugEnabled(true)
	if ModeString() != "development [debugging]" {
		t.Errorf("ModeString() = %q, want %q", ModeString(), "development [debugging]")
	}
}

func TestFromEnv(t *testing.T) {
	// Save original env vars
	origMode := os.Getenv("MODE")
	origDebug := os.Getenv("DEBUG")
	defer func() {
		os.Setenv("MODE", origMode)
		os.Setenv("DEBUG", origDebug)
	}()

	// Test MODE env var
	os.Setenv("MODE", "development")
	os.Setenv("DEBUG", "")
	FromEnv()
	if !IsAppModeDev() {
		t.Error("FromEnv() should set development mode from MODE env var")
	}

	// Test DEBUG env var
	os.Setenv("MODE", "production")
	os.Setenv("DEBUG", "true")
	FromEnv()
	if !IsDebugEnabled() {
		t.Error("FromEnv() should enable debug from DEBUG env var")
	}
}

func TestAppModeString(t *testing.T) {
	tests := []struct {
		mode AppMode
		want string
	}{
		{Production, "production"},
		{Development, "development"},
	}

	for _, tt := range tests {
		if got := tt.mode.String(); got != tt.want {
			t.Errorf("AppMode.String() = %q, want %q", got, tt.want)
		}
	}
}
