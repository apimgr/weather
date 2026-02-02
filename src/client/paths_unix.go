//go:build !windows
// +build !windows

package main

import "os"

// setDirPermissions sets directory permissions on Unix systems
// Per AI.md PART 33 line 45061-45068: Unix directories use 0700 (user-only access)
func setDirPermissions(dir string) error {
	return os.Chmod(dir, 0700)
}

// setFilePermissions sets file permissions on Unix systems
// Per AI.md PART 33 line 45070-45073: Unix files use 0600 (user-only read/write)
func setFilePermissions(path string) error {
	return os.Chmod(path, 0600)
}
