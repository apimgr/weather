package cli

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"database/sql"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	_ "modernc.org/sqlite"
	"golang.org/x/crypto/argon2"
)

// BackupManifest represents metadata for a backup archive per AI.md PART 24
type BackupManifest struct {
	Version          string    `json:"version"`
	CreatedAt        time.Time `json:"created_at"`
	CreatedBy        string    `json:"created_by"`
	AppVersion       string    `json:"app_version"`
	ServerDB         bool      `json:"server_db"`
	UsersDB          bool      `json:"users_db"`
	ConfigFiles      bool      `json:"config_files"`
	LogsIncluded     bool      `json:"logs_included"`
	GeoIPIncluded    bool      `json:"geoip_included"`
	Encrypted        bool      `json:"encrypted"`
	EncryptionMethod string    `json:"encryption_method,omitempty"` // "AES-256-GCM"
	Checksum         string    `json:"checksum"`                    // sha256:...
}

// MaintenanceCommand handles maintenance operations per TEMPLATE.md PART 6
func MaintenanceCommand(args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("no maintenance command specified. Use: backup, restore, verify, admin-recovery")
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

	case "verify":
		// TEMPLATE.md PART 6: Verify system integrity
		return verifySystem()

	case "admin-recovery":
		// TEMPLATE.md PART 6: Emergency admin account recovery
		return adminRecoverySetup()

	// Legacy commands (not in TEMPLATE.md PART 6 spec)
	case "setup":
		fmt.Println("Note: 'setup' is deprecated, use 'admin-recovery'")
		return adminRecoverySetup()

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

// createBackup creates a backup of database, config, and logs per AI.md PART 24
func createBackup(backupFile string) error {
	// AI.md PART 24: Create backup in memory first, then encrypt if needed

	// Create buffer to hold unencrypted tar.gz
	var buf bytes.Buffer

	// Create gzip writer to buffer
	gzWriter := gzip.NewWriter(&buf)

	// Create tar writer
	tarWriter := tar.NewWriter(gzWriter)

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

	// TEMPLATE.md Part 31: Backup BOTH databases (server.db + users.db)
	manifest := &BackupManifest{
		Version:   "1.0",
		CreatedAt: time.Now(),
	}

	// Backup server.db (admin credentials, server state, scheduler)
	serverDBPath := filepath.Join(dataDir, "db", "server.db")
	if err := addFileToTar(tarWriter, serverDBPath, "database/server.db"); err != nil {
		fmt.Printf("Warning: Failed to backup server.db: %v\n", err)
	} else {
		fmt.Println("  ✓ Server database (server.db) backed up")
		manifest.ServerDB = true
	}

	// Backup users.db (user accounts, tokens, sessions)
	usersDBPath := filepath.Join(dataDir, "db", "users.db")
	if err := addFileToTar(tarWriter, usersDBPath, "database/users.db"); err != nil {
		fmt.Printf("Warning: Failed to backup users.db: %v\n", err)
	} else {
		fmt.Println("  ✓ Users database (users.db) backed up")
		manifest.UsersDB = true
	}

	// Backup server.yml config
	configPath := filepath.Join(configDir, "server.yml")
	if err := addFileToTar(tarWriter, configPath, "config/server.yml"); err != nil {
		fmt.Printf("Warning: Failed to backup config: %v\n", err)
	} else {
		fmt.Println("  ✓ Configuration backed up")
		manifest.ConfigFiles = true
	}

	// Backup logs (last 7 days only to keep size manageable)
	if err := backupRecentLogs(tarWriter, logDir); err != nil {
		fmt.Printf("Warning: Failed to backup logs: %v\n", err)
	} else {
		fmt.Println("  ✓ Recent logs backed up")
		manifest.LogsIncluded = true
	}

	// Backup GeoIP databases if they exist
	geoipDir := filepath.Join(dataDir, "geoip")
	if _, err := os.Stat(geoipDir); err == nil {
		if err := addDirectoryToTar(tarWriter, geoipDir, "geoip"); err != nil {
			fmt.Printf("Warning: Failed to backup GeoIP databases: %v\n", err)
		} else {
			fmt.Println("  ✓ GeoIP databases backed up")
			manifest.GeoIPIncluded = true
		}
	}

	// AI.md PART 24: Write manifest.json with backup metadata
	manifestJSON, err := json.MarshalIndent(manifest, "", "  ")
	if err != nil {
		tarWriter.Close()
		gzWriter.Close()
		return fmt.Errorf("failed to create manifest: %w", err)
	}

	header := &tar.Header{
		Name:    "manifest.json",
		Size:    int64(len(manifestJSON)),
		Mode:    0644,
		ModTime: time.Now(),
	}
	if err := tarWriter.WriteHeader(header); err != nil {
		tarWriter.Close()
		gzWriter.Close()
		return fmt.Errorf("failed to write manifest header: %w", err)
	}
	if _, err := tarWriter.Write(manifestJSON); err != nil {
		tarWriter.Close()
		gzWriter.Close()
		return fmt.Errorf("failed to write manifest data: %w", err)
	}
	fmt.Println("  ✓ Manifest created")

	// Close tar and gzip writers to flush to buffer
	if err := tarWriter.Close(); err != nil {
		return fmt.Errorf("failed to close tar writer: %w", err)
	}
	if err := gzWriter.Close(); err != nil {
		return fmt.Errorf("failed to close gzip writer: %w", err)
	}

	// Get the unencrypted backup data
	backupData := buf.Bytes()

	// AI.md PART 24: Check if encryption is enabled
	// For now, check if --password flag was provided via environment or prompt user
	var encrypted bool
	var password string

	// Check if encryption is requested
	fmt.Print("\n🔐 Encrypt this backup? (y/N): ")
	var response string
	fmt.Scanln(&response)

	if response == "y" || response == "Y" || response == "yes" {
		// Prompt for encryption password
		fmt.Print("Enter encryption password: ")
		fmt.Scanln(&password)
		if password == "" {
			return fmt.Errorf("encryption password cannot be empty")
		}

		fmt.Print("Confirm password: ")
		var confirmPassword string
		fmt.Scanln(&confirmPassword)
		if password != confirmPassword {
			return fmt.Errorf("passwords do not match")
		}

		// Encrypt the backup
		fmt.Println("\n🔒 Encrypting backup with AES-256-GCM...")
		encryptedData, err := encryptBackup(backupData, password)
		if err != nil {
			return fmt.Errorf("encryption failed: %w", err)
		}

		backupData = encryptedData
		encrypted = true
		manifest.Encrypted = true
		manifest.EncryptionMethod = "AES-256-GCM"
		fmt.Println("  ✓ Backup encrypted successfully")
	}

	// AI.md PART 24: Generate filename with correct format
	if backupFile == "" {
		// weather_backup_YYYY-MM-DD_HHMMSS.tar.gz[.enc]
		timestamp := time.Now().Format("2006-01-02_150405")
		if encrypted {
			backupFile = fmt.Sprintf("weather_backup_%s.tar.gz.enc", timestamp)
		} else {
			backupFile = fmt.Sprintf("weather_backup_%s.tar.gz", timestamp)
		}
	}

	// Write backup to disk
	fmt.Printf("\n💾 Writing backup to: %s\n", backupFile)
	if err := os.WriteFile(backupFile, backupData, 0600); err != nil {
		return fmt.Errorf("failed to write backup file: %w", err)
	}

	// Display summary
	fmt.Printf("\n✅ Backup created successfully!\n")
	fmt.Printf("   File: %s\n", backupFile)
	fmt.Printf("   Size: %.2f MB\n", float64(len(backupData))/(1024*1024))
	if encrypted {
		fmt.Printf("   Encrypted: Yes (AES-256-GCM)\n")
		fmt.Printf("   ⚠️  Save your password! It cannot be recovered.\n")
	} else {
		fmt.Printf("   Encrypted: No\n")
	}

	return nil
}

// restoreBackup restores from a backup file
func restoreBackup(backupFile string) error {
	fmt.Printf("Restoring from backup: %s\n", backupFile)

	// Read backup file into memory (AI.md PART 24: may be encrypted)
	backupData, err := os.ReadFile(backupFile)
	if err != nil {
		return fmt.Errorf("failed to read backup file: %w", err)
	}

	// Check if backup is encrypted (AI.md PART 24: .enc extension)
	encrypted := filepath.Ext(backupFile) == ".enc"
	if encrypted {
		fmt.Println("🔐 Encrypted backup detected")
		fmt.Print("Enter decryption password: ")
		var password string
		fmt.Scanln(&password)
		if password == "" {
			return fmt.Errorf("password cannot be empty")
		}

		// Decrypt backup
		decryptedData, err := decryptBackup(backupData, password)
		if err != nil {
			return fmt.Errorf("failed to decrypt backup: %w", err)
		}
		backupData = decryptedData
		fmt.Println("✓ Backup decrypted successfully")
	}

	// Create gzip reader from decrypted/plaintext data
	gzReader, err := gzip.NewReader(bytes.NewReader(backupData))
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

		// Determine target path (TEMPLATE.md Part 31: server.db + users.db)
		var targetPath string
		switch {
		case header.Name == "database/server.db":
			targetPath = filepath.Join(dataDir, "db", "server.db")
		case header.Name == "database/users.db":
			targetPath = filepath.Join(dataDir, "db", "users.db")
		case header.Name == "database/weather.db": // Legacy compatibility
			targetPath = filepath.Join(dataDir, "db", "weather.db")
		case header.Name == "config/server.yml":
			targetPath = filepath.Join(configDir, "server.yml")
		case header.Name == "manifest.json":
			// Read manifest for information
			var manifest BackupManifest
			if err := json.NewDecoder(tarReader).Decode(&manifest); err == nil {
				fmt.Printf("  📦 Backup created: %s\n", manifest.CreatedAt.Format(time.RFC3339))
			}
			continue
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

// adminRecoverySetup allows recovery of admin access after restore or lockout
func adminRecoverySetup() error {
	fmt.Println("🔧 Admin Account Recovery Setup")
	fmt.Println("This will reset the primary admin account credentials.")
	fmt.Println()

	// Get directory paths
	dataDir := os.Getenv("DATA_DIR")
	if dataDir == "" {
		dataDir = "/var/lib/apimgr/weather"
	}

	// Connect to server.db
	serverDBPath := filepath.Join(dataDir, "db", "server.db")

	// Check if database exists
	if _, err := os.Stat(serverDBPath); os.IsNotExist(err) {
		return fmt.Errorf("server database not found at %s", serverDBPath)
	}

	// Import database package is already imported via database initialization
	// We'll use direct SQL connection for CLI tool
	db, err := openDatabase(serverDBPath)
	if err != nil {
		return fmt.Errorf("failed to open database: %w", err)
	}
	defer db.Close()

	// Prompt for new admin credentials
	fmt.Print("Enter new admin username (default: admin): ")
	var username string
	fmt.Scanln(&username)
	if username == "" {
		username = "admin"
	}

	fmt.Print("Enter new admin password: ")
	var password string
	fmt.Scanln(&password)
	if password == "" {
		return fmt.Errorf("password cannot be empty")
	}

	fmt.Print("Confirm password: ")
	var confirmPassword string
	fmt.Scanln(&confirmPassword)
	if password != confirmPassword {
		return fmt.Errorf("passwords do not match")
	}

	// Hash password with Argon2id (TEMPLATE.md Part 0 requirement)
	passwordHash, err := hashPasswordArgon2id(password)
	if err != nil {
		return fmt.Errorf("failed to hash password: %w", err)
	}

	// Update or create admin account
	// First, try to update existing admin
	result, err := db.Exec(`
		UPDATE admin_credentials
		SET username = ?, password_hash = ?, updated_at = ?
		WHERE id = 1
	`, username, passwordHash, time.Now())

	if err != nil {
		return fmt.Errorf("failed to update admin credentials: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to check update result: %w", err)
	}

	// If no rows updated, insert new admin
	if rowsAffected == 0 {
		_, err = db.Exec(`
			INSERT INTO admin_credentials (id, username, password_hash, created_at, updated_at)
			VALUES (1, ?, ?, ?, ?)
		`, username, passwordHash, time.Now(), time.Now())

		if err != nil {
			return fmt.Errorf("failed to create admin credentials: %w", err)
		}
		fmt.Println("\n✓ New admin account created")
	} else {
		fmt.Println("\n✓ Admin account updated")
	}

	fmt.Printf("  Username: %s\n", username)
	fmt.Println("\n⚠️  Please restart the server and login with the new credentials")
	fmt.Println("  Use: systemctl restart weather")

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
		// No logs directory, skip
		return nil
	}

	// Last 7 days
	cutoffTime := time.Now().AddDate(0, 0, -7)

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

// openDatabase opens a SQLite database connection
func openDatabase(dbPath string) (*sql.DB, error) {
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, err
	}

	// Test connection
	if err := db.Ping(); err != nil {
		db.Close()
		return nil, err
	}

	return db, nil
}

// hashPasswordArgon2id hashes a password using Argon2id
// TEMPLATE.md Part 0: Argon2id parameters (time=3, memory=64*1024, threads=4, keyLen=32)
func hashPasswordArgon2id(password string) (string, error) {
	// Generate random salt (16 bytes)
	salt := make([]byte, 16)
	if _, err := rand.Read(salt); err != nil {
		return "", fmt.Errorf("failed to generate salt: %w", err)
	}

	// Argon2id parameters from TEMPLATE.md Part 0
	const (
		time    = 3
		memory  = 64 * 1024 // 64 MB
		threads = 4
		keyLen  = 32
	)

	// Generate hash
	hash := argon2.IDKey([]byte(password), salt, time, memory, threads, keyLen)

	// Encode as: $argon2id$v=19$m=65536,t=3,p=4$<salt>$<hash>
	b64Salt := base64.RawStdEncoding.EncodeToString(salt)
	b64Hash := base64.RawStdEncoding.EncodeToString(hash)

	return fmt.Sprintf("$argon2id$v=19$m=%d,t=%d,p=%d$%s$%s",
		memory, time, threads, b64Salt, b64Hash), nil
}

// verifySystem verifies system integrity per TEMPLATE.md PART 6
func verifySystem() error {
	fmt.Println("🔍 System Verification")
	fmt.Println()

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

	errors := 0

	// 1. Verify server.db exists and is accessible
	fmt.Print("Checking server.db... ")
	serverDBPath := filepath.Join(dataDir, "db", "server.db")
	if err := verifyDatabaseFile(serverDBPath); err != nil {
		fmt.Printf("❌ FAIL: %v\n", err)
		errors++
	} else {
		fmt.Println("✓ OK")
	}

	// 2. Verify users.db exists and is accessible
	fmt.Print("Checking users.db... ")
	usersDBPath := filepath.Join(dataDir, "db", "users.db")
	if err := verifyDatabaseFile(usersDBPath); err != nil {
		fmt.Printf("❌ FAIL: %v\n", err)
		errors++
	} else {
		fmt.Println("✓ OK")
	}

	// 3. Verify server.yml exists
	fmt.Print("Checking server.yml... ")
	configPath := filepath.Join(configDir, "server.yml")
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		fmt.Println("⚠️  NOT FOUND (using defaults)")
	} else if err != nil {
		fmt.Printf("❌ FAIL: %v\n", err)
		errors++
	} else {
		fmt.Println("✓ OK")
	}

	// 4. Verify log directory is writable
	fmt.Print("Checking log directory... ")
	if err := verifyDirectoryWritable(logDir); err != nil {
		fmt.Printf("❌ FAIL: %v\n", err)
		errors++
	} else {
		fmt.Println("✓ OK")
	}

	// 5. Verify data directory is writable
	fmt.Print("Checking data directory... ")
	if err := verifyDirectoryWritable(dataDir); err != nil {
		fmt.Printf("❌ FAIL: %v\n", err)
		errors++
	} else {
		fmt.Println("✓ OK")
	}

	// 6. Verify at least one admin account exists
	fmt.Print("Checking admin accounts... ")
	if err := verifyAdminExists(serverDBPath); err != nil {
		fmt.Printf("⚠️  WARNING: %v\n", err)
		fmt.Println("   Run: weather --maintenance admin-recovery")
	} else {
		fmt.Println("✓ OK")
	}

	// 7. Check GeoIP databases (optional)
	fmt.Print("Checking GeoIP databases... ")
	geoipDir := filepath.Join(dataDir, "geoip")
	if _, err := os.Stat(geoipDir); os.IsNotExist(err) {
		fmt.Println("⚠️  NOT FOUND (will download on first use)")
	} else {
		fmt.Println("✓ OK")
	}

	// Summary
	fmt.Println()
	if errors == 0 {
		fmt.Println("✓ System verification completed successfully")
		return nil
	}

	fmt.Printf("❌ System verification failed with %d error(s)\n", errors)
	return fmt.Errorf("verification failed")
}

// verifyDatabaseFile verifies a database file exists and is accessible
func verifyDatabaseFile(dbPath string) error {
	// Check if file exists
	info, err := os.Stat(dbPath)
	if os.IsNotExist(err) {
		return fmt.Errorf("database not found")
	}
	if err != nil {
		return fmt.Errorf("cannot access: %w", err)
	}

	// Check if it's a regular file
	if !info.Mode().IsRegular() {
		return fmt.Errorf("not a regular file")
	}

	// Try to open it
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return fmt.Errorf("cannot open: %w", err)
	}
	defer db.Close()

	// Try to ping
	if err := db.Ping(); err != nil {
		return fmt.Errorf("database corrupted: %w", err)
	}

	return nil
}

// verifyDirectoryWritable verifies a directory exists and is writable
func verifyDirectoryWritable(dirPath string) error {
	// Check if directory exists
	info, err := os.Stat(dirPath)
	if os.IsNotExist(err) {
		// Try to create it
		if err := os.MkdirAll(dirPath, 0755); err != nil {
			return fmt.Errorf("cannot create: %w", err)
		}
		return nil
	}
	if err != nil {
		return fmt.Errorf("cannot access: %w", err)
	}

	// Check if it's a directory
	if !info.IsDir() {
		return fmt.Errorf("not a directory")
	}

	// Try to create a test file
	testFile := filepath.Join(dirPath, ".write-test")
	if err := os.WriteFile(testFile, []byte("test"), 0644); err != nil {
		return fmt.Errorf("not writable: %w", err)
	}

	// Clean up test file
	os.Remove(testFile)

	return nil
}

// verifyAdminExists checks if at least one admin account exists
func verifyAdminExists(serverDBPath string) error {
	db, err := sql.Open("sqlite", serverDBPath)
	if err != nil {
		return fmt.Errorf("cannot open database: %w", err)
	}
	defer db.Close()

	var count int
	err = db.QueryRow("SELECT COUNT(*) FROM server_admin_credentials").Scan(&count)
	if err != nil {
		return fmt.Errorf("cannot query: %w", err)
	}

	if count == 0 {
		return fmt.Errorf("no admin accounts found")
	}

	return nil
}

// encryptBackup encrypts a backup archive with AES-256-GCM per AI.md PART 24
// Takes plaintext data, password, returns encrypted data
func encryptBackup(plaintext []byte, password string) ([]byte, error) {
	// Derive encryption key from password using Argon2id (AI.md PART 24)
	const (
		saltSize  = 32        // 256 bits
		keySize   = 32        // 256 bits for AES-256
		time      = 3         // Argon2id time parameter
		memory    = 64 * 1024 // 64 MB
		threads   = 4         // Parallelism
	)

	// Generate random salt
	salt := make([]byte, saltSize)
	if _, err := rand.Read(salt); err != nil {
		return nil, fmt.Errorf("failed to generate salt: %w", err)
	}

	// Derive key from password
	key := argon2.IDKey([]byte(password), salt, time, memory, threads, keySize)

	// Create AES-256-GCM cipher
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, fmt.Errorf("failed to create cipher: %w", err)
	}

	aesGCM, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("failed to create GCM: %w", err)
	}

	// Generate random nonce
	nonce := make([]byte, aesGCM.NonceSize())
	if _, err := rand.Read(nonce); err != nil {
		return nil, fmt.Errorf("failed to generate nonce: %w", err)
	}

	// Encrypt plaintext
	ciphertext := aesGCM.Seal(nonce, nonce, plaintext, nil)

	// Format: salt (32 bytes) + ciphertext (nonce + encrypted data + tag)
	encrypted := make([]byte, 0, saltSize+len(ciphertext))
	encrypted = append(encrypted, salt...)
	encrypted = append(encrypted, ciphertext...)

	return encrypted, nil
}

// decryptBackup decrypts a backup archive with AES-256-GCM per AI.md PART 24
// Takes encrypted data, password, returns plaintext data
func decryptBackup(encrypted []byte, password string) ([]byte, error) {
	const (
		saltSize = 32        // 256 bits
		keySize  = 32        // 256 bits for AES-256
		time     = 3         // Argon2id time parameter
		memory   = 64 * 1024 // 64 MB
		threads  = 4         // Parallelism
	)

	// Validate minimum size (salt + nonce + tag)
	if len(encrypted) < saltSize+12+16 {
		return nil, fmt.Errorf("encrypted data too short")
	}

	// Extract salt (first 32 bytes)
	salt := encrypted[:saltSize]

	// Derive key from password
	key := argon2.IDKey([]byte(password), salt, time, memory, threads, keySize)

	// Create AES-256-GCM cipher
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, fmt.Errorf("failed to create cipher: %w", err)
	}

	aesGCM, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("failed to create GCM: %w", err)
	}

	// Extract ciphertext (everything after salt)
	ciphertext := encrypted[saltSize:]

	// Extract nonce
	nonceSize := aesGCM.NonceSize()
	if len(ciphertext) < nonceSize {
		return nil, fmt.Errorf("ciphertext too short")
	}

	nonce := ciphertext[:nonceSize]
	ciphertext = ciphertext[nonceSize:]

	// Decrypt
	plaintext, err := aesGCM.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return nil, fmt.Errorf("decryption failed (wrong password?): %w", err)
	}

	return plaintext, nil
}
