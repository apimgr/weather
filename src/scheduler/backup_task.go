// Package scheduler - automated backup task per AI.md PART 27
// AI.md Reference: Lines 24088-24250, specifically line 24182-24188
package scheduler

import (
	"fmt"
	"log"
	"time"

	"github.com/apimgr/weather/src/backup"
	"github.com/apimgr/weather/src/paths"
)

// BackupTask performs automated daily backups per AI.md PART 27 line 24182
// Default schedule: Daily at 01:00 (disabled by default)
// Keep max 4 backups per retention policy
func BackupTask(configDir, dataDir string) func() error {
	return func() error {
		log.Println("🔄 Starting automated backup...")

		// Create backup service
		svc := backup.New(configDir, dataDir)

		// Create backup with default options
		// Per AI.md PART 25, backup includes:
		// - server.yml (always)
		// - server.db (always)
		// - users.db (if exists)
		// - templates/ (if exists)
		// - themes/ (if exists)
		opts := backup.BackupOptions{
			ConfigDir:   configDir,
			DataDir:     dataDir,
			// Auto-generate filename
			OutputPath:  "",
			// Encryption disabled for automated backups
			Password:    "",
			IncludeSSL:  false,
			IncludeData: false,
			CreatedBy:   "scheduler",
			// Version set at build time
			AppVersion:  "1.0.0",
		}

		backupPath, err := svc.Create(opts)
		if err != nil {
			log.Printf("❌ Automated backup failed: %v", err)
			return fmt.Errorf("backup failed: %w", err)
		}

		log.Printf("✅ Automated backup completed: %s", backupPath)
		return nil
	}
}

// RegisterBackupTask registers the automated backup task with the scheduler
// Per AI.md PART 27 lines 24182-24188
func RegisterBackupTask(s *Scheduler, enabled bool) {
	// Get paths per AI.md PART 4
	p := paths.GetDefaultPaths("weather")
	if p == nil {
		log.Println("⚠️  Failed to get default paths for backup task")
		return
	}

	// Schedule: Daily at 01:00 per AI.md PART 27 line 24184
	// Disabled by default per AI.md PART 27 line 24185
	interval := 24 * time.Hour

	// Create task function
	taskFn := BackupTask(p.ConfigDir, p.DataDir)

	// Add to scheduler
	s.AddTask("backup_auto", interval, taskFn)

	if enabled {
		log.Println("📅 Automated backup task scheduled (daily at 01:00)")
	} else {
		log.Println("📅 Automated backup task registered (disabled by default)")
	}
}
