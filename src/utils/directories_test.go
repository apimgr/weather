package utils

import (
	"os"
	"path/filepath"
)

// GetTestDirectoryPaths returns directory paths for testing (uses temp directory)
func GetTestDirectoryPaths() (*DirectoryPaths, error) {
	tempBase := filepath.Join(os.TempDir(), "weather-test")

	paths := &DirectoryPaths{
		Config: filepath.Join(tempBase, "config"),
		Data:   filepath.Join(tempBase, "data"),
		Log:    filepath.Join(tempBase, "logs"),
		Cache:  filepath.Join(tempBase, "cache"),
	}

	return paths, nil
}

// CleanupTestDirectories removes test directories
func CleanupTestDirectories(paths *DirectoryPaths) error {
	// Remove the parent temp directory
	tempBase := filepath.Dir(paths.Config)
	return os.RemoveAll(tempBase)
}
