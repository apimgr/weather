package mode

import (
	"os"
	"testing"
)

func TestSet(t *testing.T) {
	tests := []struct {
		input    string
		expected Mode
	}{
		{"development", Development},
		{"dev", Development},
		{"production", Production},
		{"prod", Production},
		{"DEVELOPMENT", Development},
		{"PRODUCTION", Production},
		{"anything", Production}, // Default to production
	}

	for _, tt := range tests {
		Set(tt.input)
		if Current() != tt.expected {
			t.Errorf("Set(%q): got %v, want %v", tt.input, Current(), tt.expected)
		}
	}
}

func TestIsDevelopment(t *testing.T) {
	Set("development")
	if !IsDevelopment() {
		t.Error("IsDevelopment() should return true when mode is development")
	}
	if IsProduction() {
		t.Error("IsProduction() should return false when mode is development")
	}
}

func TestIsProduction(t *testing.T) {
	Set("production")
	if !IsProduction() {
		t.Error("IsProduction() should return true when mode is production")
	}
	if IsDevelopment() {
		t.Error("IsDevelopment() should return false when mode is production")
	}
}

func TestSetDebug(t *testing.T) {
	SetDebug(true)
	if !IsDebug() {
		t.Error("IsDebug() should return true when debug is enabled")
	}

	SetDebug(false)
	if IsDebug() {
		t.Error("IsDebug() should return false when debug is disabled")
	}
}

func TestModeString(t *testing.T) {
	Set("production")
	SetDebug(false)
	if ModeString() != "production" {
		t.Errorf("ModeString() = %q, want %q", ModeString(), "production")
	}

	SetDebug(true)
	if ModeString() != "production [debugging]" {
		t.Errorf("ModeString() = %q, want %q", ModeString(), "production [debugging]")
	}

	Set("development")
	SetDebug(false)
	if ModeString() != "development" {
		t.Errorf("ModeString() = %q, want %q", ModeString(), "development")
	}

	SetDebug(true)
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
	if !IsDevelopment() {
		t.Error("FromEnv() should set development mode from MODE env var")
	}

	// Test DEBUG env var
	os.Setenv("MODE", "production")
	os.Setenv("DEBUG", "true")
	FromEnv()
	if !IsDebug() {
		t.Error("FromEnv() should enable debug from DEBUG env var")
	}
}

func TestModeString_String(t *testing.T) {
	tests := []struct {
		mode Mode
		want string
	}{
		{Production, "production"},
		{Development, "development"},
	}

	for _, tt := range tests {
		if got := tt.mode.String(); got != tt.want {
			t.Errorf("Mode.String() = %q, want %q", got, tt.want)
		}
	}
}
