package utils

import (
	"fmt"
	"regexp"
	"strings"
)

// Username validation rules
const (
	MinUsernameLength = 3
	MaxUsernameLength = 20
)

var usernameRegex = regexp.MustCompile(`^[a-zA-Z0-9_]+$`)

// ValidateUsername validates a username according to the rules:
// - Length between 3 and 20 characters
// - Only alphanumeric characters and underscores
// - Not on the blocklist
func ValidateUsername(username string) error {
	// Trim spaces
	username = strings.TrimSpace(username)

	// Check length
	if len(username) < MinUsernameLength {
		return fmt.Errorf("username must be at least %d characters long", MinUsernameLength)
	}

	if len(username) > MaxUsernameLength {
		return fmt.Errorf("username must be no more than %d characters long", MaxUsernameLength)
	}

	// Check format (alphanumeric + underscore only)
	if !usernameRegex.MatchString(username) {
		return fmt.Errorf("username can only contain letters, numbers, and underscores")
	}

	// Check against blocklist
	// Create a fake email to use the blocklist checker
	fakeEmail := username + "@example.com"
	if IsUsernameBlocked(fakeEmail) {
		return fmt.Errorf("this username is reserved and cannot be used")
	}

	return nil
}

// NormalizeUsername converts username to lowercase and trims spaces
func NormalizeUsername(username string) string {
	return strings.ToLower(strings.TrimSpace(username))
}

// ValidatePhone validates a phone number
// Basic validation - can be enhanced with libphonenumber later
func ValidatePhone(phone string) error {
	if phone == "" {
		return nil // Phone is optional
	}

	// Remove common formatting characters
	cleaned := strings.Map(func(r rune) rune {
		if r >= '0' && r <= '9' || r == '+' {
			return r
		}
		return -1
	}, phone)

	// Check length (basic validation)
	if len(cleaned) < 10 || len(cleaned) > 15 {
		return fmt.Errorf("phone number must be between 10 and 15 digits")
	}

	return nil
}

// NormalizePhone removes formatting from phone number
func NormalizePhone(phone string) string {
	if phone == "" {
		return ""
	}

	// Keep only digits and leading +
	cleaned := strings.Map(func(r rune) rune {
		if r >= '0' && r <= '9' || r == '+' {
			return r
		}
		return -1
	}, phone)

	return cleaned
}
