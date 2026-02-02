//go:build windows
// +build windows

package main

// setDirPermissions sets directory permissions on Windows
// Per AI.md PART 33 line 45075-45086: Windows directories inherit ACLs from %APPDATA%
// No explicit ACL modification needed for user directories
func setDirPermissions(dir string) error {
	// Windows: ACLs are inherited from parent by default in user directories
	// %APPDATA% and %LOCALAPPDATA% already have user-only access
	return nil
}

// setFilePermissions sets file permissions on Windows
// Per AI.md PART 33 line 45088-45093: Windows files inherit ACLs from directory
func setFilePermissions(path string) error {
	// Windows: files inherit ACLs from directory
	// For sensitive files, we could set explicit ACLs, but for user
	// directories this is not needed as they inherit from %APPDATA%
	return nil
}
