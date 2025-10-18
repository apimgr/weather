// Package paths provides OS-specific directory path detection
package paths

import (
	"os"
	"path/filepath"
	"runtime"
)

// Paths holds OS-specific directory paths
type Paths struct {
	DataDir   string
	ConfigDir string
	LogDir    string
	CacheDir  string
	TempDir   string
}

// GetDefaultPaths returns default directory paths for the current OS
func GetDefaultPaths(appName string) *Paths {
	switch runtime.GOOS {
	case "linux":
		return getLinuxPaths(appName)
	case "darwin":
		return getDarwinPaths(appName)
	case "windows":
		return getWindowsPaths(appName)
	case "freebsd", "openbsd", "netbsd":
		return getBSDPaths(appName)
	default:
		return getLinuxPaths(appName) // Fallback to Linux paths
	}
}

// getLinuxPaths returns Linux-specific paths (XDG Base Directory spec)
func getLinuxPaths(appName string) *Paths {
	homeDir, _ := os.UserHomeDir()

	// Check if running as root
	if os.Geteuid() == 0 {
		return &Paths{
			DataDir:   filepath.Join("/var/lib", appName),
			ConfigDir: filepath.Join("/etc", appName),
			LogDir:    filepath.Join("/var/log", appName),
			CacheDir:  filepath.Join("/var/cache", appName),
			TempDir:   filepath.Join("/tmp", appName),
		}
	}

	// User-level paths (XDG Base Directory)
	dataHome := os.Getenv("XDG_DATA_HOME")
	if dataHome == "" {
		dataHome = filepath.Join(homeDir, ".local", "share")
	}

	configHome := os.Getenv("XDG_CONFIG_HOME")
	if configHome == "" {
		configHome = filepath.Join(homeDir, ".config")
	}

	cacheHome := os.Getenv("XDG_CACHE_HOME")
	if cacheHome == "" {
		cacheHome = filepath.Join(homeDir, ".cache")
	}

	return &Paths{
		DataDir:   filepath.Join(dataHome, appName),
		ConfigDir: filepath.Join(configHome, appName),
		LogDir:    filepath.Join(homeDir, ".local", "state", appName),
		CacheDir:  filepath.Join(cacheHome, appName),
		TempDir:   filepath.Join(os.TempDir(), appName),
	}
}

// getDarwinPaths returns macOS-specific paths
func getDarwinPaths(appName string) *Paths {
	homeDir, _ := os.UserHomeDir()

	return &Paths{
		DataDir:   filepath.Join(homeDir, "Library", "Application Support", appName),
		ConfigDir: filepath.Join(homeDir, "Library", "Preferences", appName),
		LogDir:    filepath.Join(homeDir, "Library", "Logs", appName),
		CacheDir:  filepath.Join(homeDir, "Library", "Caches", appName),
		TempDir:   filepath.Join(os.TempDir(), appName),
	}
}

// getWindowsPaths returns Windows-specific paths
func getWindowsPaths(appName string) *Paths {
	appData := os.Getenv("APPDATA")
	if appData == "" {
		appData = filepath.Join(os.Getenv("USERPROFILE"), "AppData", "Roaming")
	}

	localAppData := os.Getenv("LOCALAPPDATA")
	if localAppData == "" {
		localAppData = filepath.Join(os.Getenv("USERPROFILE"), "AppData", "Local")
	}

	programData := os.Getenv("PROGRAMDATA")
	if programData == "" {
		programData = "C:\\ProgramData"
	}

	return &Paths{
		DataDir:   filepath.Join(localAppData, appName),
		ConfigDir: filepath.Join(appData, appName),
		LogDir:    filepath.Join(localAppData, appName, "Logs"),
		CacheDir:  filepath.Join(localAppData, appName, "Cache"),
		TempDir:   filepath.Join(os.TempDir(), appName),
	}
}

// getBSDPaths returns BSD-specific paths (similar to Linux)
func getBSDPaths(appName string) *Paths {
	homeDir, _ := os.UserHomeDir()

	// Check if running as root
	if os.Geteuid() == 0 {
		return &Paths{
			DataDir:   filepath.Join("/var/db", appName),
			ConfigDir: filepath.Join("/usr/local/etc", appName),
			LogDir:    filepath.Join("/var/log", appName),
			CacheDir:  filepath.Join("/var/cache", appName),
			TempDir:   filepath.Join("/tmp", appName),
		}
	}

	// User-level paths
	return &Paths{
		DataDir:   filepath.Join(homeDir, ".local", "share", appName),
		ConfigDir: filepath.Join(homeDir, ".config", appName),
		LogDir:    filepath.Join(homeDir, ".local", "state", appName),
		CacheDir:  filepath.Join(homeDir, ".cache", appName),
		TempDir:   filepath.Join(os.TempDir(), appName),
	}
}

// EnsureDirectories creates all directories if they don't exist
func (p *Paths) EnsureDirectories() error {
	dirs := []string{
		p.DataDir,
		p.ConfigDir,
		p.LogDir,
		p.CacheDir,
		p.TempDir,
	}

	for _, dir := range dirs {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return err
		}
	}

	return nil
}

// Override allows overriding paths via environment variables
func (p *Paths) Override() {
	if dataDir := os.Getenv("DATA_DIR"); dataDir != "" {
		p.DataDir = dataDir
	}
	if configDir := os.Getenv("CONFIG_DIR"); configDir != "" {
		p.ConfigDir = configDir
	}
	if logDir := os.Getenv("LOG_DIR"); logDir != "" {
		p.LogDir = logDir
	}
	if cacheDir := os.Getenv("CACHE_DIR"); cacheDir != "" {
		p.CacheDir = cacheDir
	}
	if tempDir := os.Getenv("TEMP_DIR"); tempDir != "" {
		p.TempDir = tempDir
	}
}
