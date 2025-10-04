package utils

import "testing"

func TestIsUsernameBlocked(t *testing.T) {
	tests := []struct {
		email    string
		blocked  bool
		testName string
	}{
		// Exact matches
		{"admin@example.com", true, "admin exact match"},
		{"root@example.com", true, "root exact match"},
		{"test@example.com", true, "test exact match"},

		// Variations with numbers
		{"admin123@example.com", true, "admin with numbers (exact blocklist match)"},
		{"admin_1@example.com", true, "admin with underscore"},
		{"test-123@example.com", true, "test with hyphen"},
		{"test1@example.com", true, "test1 (exact blocklist match)"},

		// Should NOT be blocked (legitimate usernames)
		{"administrator123@example.com", true, "administrator with suffix (blocked)"},
		{"johnadmin@example.com", false, "admin in middle"},
		{"testing@example.com", true, "testing is blocked"},
		{"user.name@example.com", false, "legitimate user"},
		{"alice@example.com", false, "normal name"},
		{"bob.smith@example.com", false, "normal name with dot"},

		// Case insensitive
		{"ADMIN@example.com", true, "uppercase admin"},
		{"RoOt@example.com", true, "mixed case root"},

		// Single characters
		{"a@example.com", true, "single letter"},
		{"z@example.com", true, "single letter z"},

		// Numbers
		{"1@example.com", true, "single digit"},
		{"123@example.com", true, "starts with blocked single digit"},
		{"00@example.com", true, "00 reserved"},
		{"john1234@example.com", false, "legitimate username with numbers"},

		// Edge cases
		{"@example.com", true, "empty username"},
		{"support@example.com", true, "support reserved"},
		{"noreply@example.com", true, "noreply reserved"},
	}

	for _, tt := range tests {
		t.Run(tt.testName, func(t *testing.T) {
			result := IsUsernameBlocked(tt.email)
			if result != tt.blocked {
				t.Errorf("IsUsernameBlocked(%q) = %v, want %v", tt.email, result, tt.blocked)
			}
		})
	}
}

func TestGetBlocklistSize(t *testing.T) {
	size := GetBlocklistSize()
	if size == 0 {
		t.Error("Blocklist size should not be zero")
	}
	if size != len(UsernameBlocklist) {
		t.Errorf("GetBlocklistSize() = %d, want %d", size, len(UsernameBlocklist))
	}
}

func TestIsBlocklistPublic(t *testing.T) {
	if !IsBlocklistPublic() {
		t.Error("Blocklist should be public")
	}
}
