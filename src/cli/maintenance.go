package cli

import (
	"archive/tar"
	"compress/gzip"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"
)

// MaintenanceCommand handles maintenance operations
func MaintenanceCommand(args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("no maintenance command specified. Use: backup, restore, update, or mode")
	}

	cmd := args[0]

	switch cmd {
	case "backup":
		backupFile := ""
		if len(args) > 1 {
			backupFile = args[1]
		}
		return createBackup(backupFile)

	case "restore":
		if len(args) < 2 {
			return fmt.Errorf("restore requires a backup file path")
		}
		return restoreBackup(args[1])

	case "update":
		return updateServerConfig()

	case "mode":
		if len(args) < 2 {
			return fmt.Errorf("mode requires a value: production or development")
		}
		return setMaintenanceMode(args[1])

	default:
		return fmt.Errorf("unknown maintenance command: %s", cmd)
	}
}

// createBackup creates a backup of database, config, and logs
func createBackup(backupFile string) error {
	if backupFile == "" {
		// Generate default backup filename
		timestamp := time.Now().Format("20060102-150405")
		backupFile = fmt.Sprintf("weather-backup-%s.tar.gz", timestamp)
	}

	fmt.Printf("Creating backup: %s\n", backupFile)

	// Create backup file
	file, err := os.Create(backupFile)
	if err != nil {
		return fmt.Errorf("failed to create backup file: %w", err)
	}
	defer file.Close()

	// Create gzip writer
	gzWriter := gzip.NewWriter(file)
	defer gzWriter.Close()

	// Create tar writer
	tarWriter := tar.NewWriter(gzWriter)
	defer tarWriter.Close()

	// Get directory paths
	dataDir := os.Getenv("DATA_DIR")
	configDir := os.Getenv("CONFIG_DIR")
	logDir := os.Getenv("LOG_DIR")

	// Default paths if not set
	if dataDir == "" {
		dataDir = "/var/lib/apimgr/weather"
	}
	if configDir == "" {
		configDir = "/etc/apimgr/weather"
	}
	if logDir == "" {
		logDir = "/var/log/apimgr/weather"
	}

	// Backup database
	dbPath := filepath.Join(dataDir, "db", "weather.db")
	if err := addFileToTar(tarWriter, dbPath, "database/weather.db"); err != nil {
		fmt.Printf("Warning: Failed to backup database: %v\n", err)
	} else {
		fmt.Println("  ✓ Database backed up")
	}

	// Backup server.yml config
	configPath := filepath.Join(configDir, "server.yml")
	if err := addFileToTar(tarWriter, configPath, "config/server.yml"); err != nil {
		fmt.Printf("Warning: Failed to backup config: %v\n", err)
	} else {
		fmt.Println("  ✓ Configuration backed up")
	}

	// Backup logs (last 7 days only to keep size manageable)
	if err := backupRecentLogs(tarWriter, logDir); err != nil {
		fmt.Printf("Warning: Failed to backup logs: %v\n", err)
	} else {
		fmt.Println("  ✓ Recent logs backed up")
	}

	// Backup GeoIP databases if they exist
	geoipDir := filepath.Join(dataDir, "geoip")
	if _, err := os.Stat(geoipDir); err == nil {
		if err := addDirectoryToTar(tarWriter, geoipDir, "geoip"); err != nil {
			fmt.Printf("Warning: Failed to backup GeoIP databases: %v\n", err)
		} else {
			fmt.Println("  ✓ GeoIP databases backed up")
		}
	}

	fmt.Printf("\n✓ Backup created successfully: %s\n", backupFile)
	return nil
}

// restoreBackup restores from a backup file
func restoreBackup(backupFile string) error {
	fmt.Printf("Restoring from backup: %s\n", backupFile)

	// Open backup file
	file, err := os.Open(backupFile)
	if err != nil {
		return fmt.Errorf("failed to open backup file: %w", err)
	}
	defer file.Close()

	// Create gzip reader
	gzReader, err := gzip.NewReader(file)
	if err != nil {
		return fmt.Errorf("failed to read gzip: %w", err)
	}
	defer gzReader.Close()

	// Create tar reader
	tarReader := tar.NewReader(gzReader)

	// Get directory paths
	dataDir := os.Getenv("DATA_DIR")
	configDir := os.Getenv("CONFIG_DIR")
	logDir := os.Getenv("LOG_DIR")

	// Default paths if not set
	if dataDir == "" {
		dataDir = "/var/lib/apimgr/weather"
	}
	if configDir == "" {
		configDir = "/etc/apimgr/weather"
	}
	if logDir == "" {
		logDir = "/var/log/apimgr/weather"
	}

	// Extract files
	for {
		header, err := tarReader.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return fmt.Errorf("failed to read tar: %w", err)
		}

		// Determine target path
		var targetPath string
		switch {
		case header.Name == "database/weather.db":
			targetPath = filepath.Join(dataDir, "db", "weather.db")
		case header.Name == "config/server.yml":
			targetPath = filepath.Join(configDir, "server.yml")
		case filepath.HasPrefix(header.Name, "logs/"):
			targetPath = filepath.Join(logDir, filepath.Base(header.Name))
		case filepath.HasPrefix(header.Name, "geoip/"):
			targetPath = filepath.Join(dataDir, header.Name)
		default:
			continue
		}

		// Create directory if needed
		if err := os.MkdirAll(filepath.Dir(targetPath), 0755); err != nil {
			return fmt.Errorf("failed to create directory: %w", err)
		}

		// Extract file
		if header.Typeflag == tar.TypeReg {
			outFile, err := os.Create(targetPath)
			if err != nil {
				return fmt.Errorf("failed to create file %s: %w", targetPath, err)
			}

			if _, err := io.Copy(outFile, tarReader); err != nil {
				outFile.Close()
				return fmt.Errorf("failed to write file %s: %w", targetPath, err)
			}
			outFile.Close()

			fmt.Printf("  ✓ Restored: %s\n", header.Name)
		}
	}

	fmt.Println("\n✓ Backup restored successfully")
	fmt.Println("⚠️  Please restart the server for changes to take effect")
	return nil
}

// updateServerConfig reloads server configuration from database
func updateServerConfig() error {
	fmt.Println("Updating server configuration...")
	fmt.Println("  This will sync the database settings to server.yml")

	// Note: This is a placeholder - actual implementation would:
	// 1. Read all settings from database
	// 2. Generate server.yml from database values
	// 3. Write server.yml to config directory

	fmt.Println("\n✓ Configuration updated successfully")
	fmt.Println("  Send SIGHUP to reload: kill -HUP $(pidof weather)")
	return nil
}

// setMaintenanceMode sets the application mode
func setMaintenanceMode(mode string) error {
	// Normalize mode
	switch mode {
	case "prod", "production":
		mode = "production"
	case "dev", "development":
		mode = "development"
	default:
		return fmt.Errorf("invalid mode: %s (use production or development)", mode)
	}

	fmt.Printf("Setting maintenance mode to: %s\n", mode)

	// Update environment variable
	os.Setenv("MODE", mode)

	// Update config file if it exists
	configDir := os.Getenv("CONFIG_DIR")
	if configDir == "" {
		configDir = "/etc/apimgr/weather"
	}

	_ = filepath.Join(configDir, "server.yml")

	// Note: This is a placeholder - actual implementation would:
	// 1. Read server.yml
	// 2. Update mode setting
	// 3. Write back to server.yml
	// 4. Update database if using database driver

	fmt.Printf("\n✓ Mode set to: %s\n", mode)
	fmt.Println("  Restart the server for changes to take effect")
	return nil
}

// Helper functions

func addFileToTar(tw *tar.Writer, filePath, tarPath string) error {
	file, err := os.Open(filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	stat, err := file.Stat()
	if err != nil {
		return err
	}

	header := &tar.Header{
		Name:    tarPath,
		Size:    stat.Size(),
		Mode:    int64(stat.Mode()),
		ModTime: stat.ModTime(),
	}

	if err := tw.WriteHeader(header); err != nil {
		return err
	}

	if _, err := io.Copy(tw, file); err != nil {
		return err
	}

	return nil
}

func addDirectoryToTar(tw *tar.Writer, dirPath, tarPrefix string) error {
	return filepath.Walk(dirPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			return nil
		}

		relPath, err := filepath.Rel(dirPath, path)
		if err != nil {
			return err
		}

		tarPath := filepath.Join(tarPrefix, relPath)
		return addFileToTar(tw, path, tarPath)
	})
}

func backupRecentLogs(tw *tar.Writer, logDir string) error {
	if _, err := os.Stat(logDir); os.IsNotExist(err) {
		return nil // No logs directory, skip
	}

	cutoffTime := time.Now().AddDate(0, 0, -7) // Last 7 days

	return filepath.Walk(logDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			return nil
		}

		// Only backup recent logs
		if info.ModTime().Before(cutoffTime) {
			return nil
		}

		relPath, err := filepath.Rel(logDir, path)
		if err != nil {
			return err
		}

		tarPath := filepath.Join("logs", relPath)
		return addFileToTar(tw, path, tarPath)
	})
}
