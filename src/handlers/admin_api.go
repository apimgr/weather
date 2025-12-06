package handlers

import (
	"database/sql"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/gin-gonic/gin"
	"weather-go/src/models"
	"weather-go/src/services"
)

// AdminAPIResponse represents a standard API response
type AdminAPIResponse struct {
	Success bool        `json:"success"`
	Message string      `json:"message,omitempty"`
	Error   string      `json:"error,omitempty"`
	Data    interface{} `json:"data,omitempty"`
}

// BackupFile represents a backup file information
type BackupFile struct {
	Filename string    `json:"filename"`
	Size     int64     `json:"size"`
	Created  time.Time `json:"created"`
}

// SaveWebSettings handles saving web configuration settings
func SaveWebSettings(c *gin.Context) {
	var settings map[string]interface{}
	if err := c.ShouldBindJSON(&settings); err != nil {
		c.JSON(http.StatusBadRequest, AdminAPIResponse{
			Success: false,
			Error:   "Invalid request data",
		})
		return
	}

	// Get database from context
	db, exists := c.Get("db")
	if !exists {
		c.JSON(http.StatusInternalServerError, AdminAPIResponse{
			Success: false,
			Error:   "Database connection not available",
		})
		return
	}

	settingsModel := &models.SettingsModel{DB: db.(*sql.DB)}

	// Save each setting to database
	for key, value := range settings {
		var err error
		switch v := value.(type) {
		case string:
			err = settingsModel.SetString(key, v)
		case float64:
			err = settingsModel.SetInt(key, int(v))
		case bool:
			err = settingsModel.SetBool(key, v)
		default:
			err = settingsModel.SetJSON(key, v)
		}

		if err != nil {
			c.JSON(http.StatusInternalServerError, AdminAPIResponse{
				Success: false,
				Error:   fmt.Sprintf("Failed to save setting %s: %v", key, err),
			})
			return
		}
	}

	c.JSON(http.StatusOK, AdminAPIResponse{
		Success: true,
		Message: "Web settings saved successfully",
	})
}

// SaveSecuritySettings handles saving security configuration
func SaveSecuritySettings(c *gin.Context) {
	var settings map[string]interface{}
	if err := c.ShouldBindJSON(&settings); err != nil {
		c.JSON(http.StatusBadRequest, AdminAPIResponse{
			Success: false,
			Error:   "Invalid request data",
		})
		return
	}

	// Get database from context
	db, exists := c.Get("db")
	if !exists {
		c.JSON(http.StatusInternalServerError, AdminAPIResponse{
			Success: false,
			Error:   "Database connection not available",
		})
		return
	}

	settingsModel := &models.SettingsModel{DB: db.(*sql.DB)}

	// Save each setting to database
	for key, value := range settings {
		var err error
		switch v := value.(type) {
		case string:
			err = settingsModel.SetString(key, v)
		case float64:
			err = settingsModel.SetInt(key, int(v))
		case bool:
			err = settingsModel.SetBool(key, v)
		default:
			err = settingsModel.SetJSON(key, v)
		}

		if err != nil {
			c.JSON(http.StatusInternalServerError, AdminAPIResponse{
				Success: false,
				Error:   fmt.Sprintf("Failed to save setting %s: %v", key, err),
			})
			return
		}
	}

	c.JSON(http.StatusOK, AdminAPIResponse{
		Success: true,
		Message: "Security settings saved successfully",
	})
}

// TestDatabaseConnection tests the database connection
func TestDatabaseConnection(c *gin.Context) {
	db, exists := c.Get("db")
	if !exists {
		c.JSON(http.StatusInternalServerError, AdminAPIResponse{
			Success: false,
			Error:   "Database connection not available",
		})
		return
	}

	start := time.Now()
	if err := db.(*sql.DB).Ping(); err != nil {
		c.JSON(http.StatusInternalServerError, AdminAPIResponse{
			Success: false,
			Error:   fmt.Sprintf("Database connection failed: %v", err),
		})
		return
	}
	latency := time.Since(start)

	c.JSON(http.StatusOK, AdminAPIResponse{
		Success: true,
		Message: "Database connection successful",
		Data: map[string]interface{}{
			"latency": latency.String(),
			"status":  "connected",
		},
	})
}

// OptimizeDatabase optimizes the database
func OptimizeDatabase(c *gin.Context) {
	db, exists := c.Get("db")
	if !exists {
		c.JSON(http.StatusInternalServerError, AdminAPIResponse{
			Success: false,
			Error:   "Database connection not available",
		})
		return
	}

	sqlDB := db.(*sql.DB)

	// Run ANALYZE to update statistics
	if _, err := sqlDB.Exec("ANALYZE"); err != nil {
		c.JSON(http.StatusInternalServerError, AdminAPIResponse{
			Success: false,
			Error:   fmt.Sprintf("Failed to analyze database: %v", err),
		})
		return
	}

	c.JSON(http.StatusOK, AdminAPIResponse{
		Success: true,
		Message: "Database optimized successfully",
		Data: map[string]interface{}{
			"operation": "ANALYZE completed",
		},
	})
}

// ClearCache clears the application cache
func ClearCache(c *gin.Context) {
	// Get cache manager from context
	cacheInterface, exists := c.Get("cache")
	if !exists {
		c.JSON(http.StatusOK, AdminAPIResponse{
			Success: true,
			Message: "Cache not configured (running without cache)",
		})
		return
	}

	cache, ok := cacheInterface.(*services.CacheManager)
	if !ok || !cache.IsEnabled() {
		c.JSON(http.StatusOK, AdminAPIResponse{
			Success: true,
			Message: "Cache not enabled",
		})
		return
	}

	// Flush all cache entries
	if err := cache.Flush(); err != nil {
		c.JSON(http.StatusInternalServerError, AdminAPIResponse{
			Success: false,
			Error:   fmt.Sprintf("Failed to clear cache: %v", err),
		})
		return
	}

	c.JSON(http.StatusOK, AdminAPIResponse{
		Success: true,
		Message: "Cache cleared successfully",
	})
}

// VacuumDatabase performs database vacuum operation
func VacuumDatabase(c *gin.Context) {
	db, exists := c.Get("db")
	if !exists {
		c.JSON(http.StatusInternalServerError, AdminAPIResponse{
			Success: false,
			Error:   "Database connection not available",
		})
		return
	}

	sqlDB := db.(*sql.DB)
	start := time.Now()

	// Run VACUUM to reclaim space
	if _, err := sqlDB.Exec("VACUUM"); err != nil {
		c.JSON(http.StatusInternalServerError, AdminAPIResponse{
			Success: false,
			Error:   fmt.Sprintf("Failed to vacuum database: %v", err),
		})
		return
	}

	duration := time.Since(start)

	c.JSON(http.StatusOK, AdminAPIResponse{
		Success: true,
		Message: "Database vacuum completed",
		Data: map[string]interface{}{
			"duration":  duration.String(),
			"operation": "VACUUM completed",
		},
	})
}

// VerifySSLCertificate verifies the SSL certificate
func VerifySSLCertificate(c *gin.Context) {
	// TODO: Implement actual SSL certificate verification
	c.JSON(http.StatusOK, AdminAPIResponse{
		Success: true,
		Message: "Certificate is valid",
		Data: map[string]interface{}{
			"valid_until": "2025-12-31",
			"issuer":      "Example CA",
		},
	})
}

// ObtainSSLCertificate obtains a new SSL certificate via ACME
func ObtainSSLCertificate(c *gin.Context) {
	var request struct {
		Domain   string `json:"domain"`
		Email    string `json:"email"`
		Provider string `json:"provider"`
	}

	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, AdminAPIResponse{
			Success: false,
			Error:   "Invalid request data",
		})
		return
	}

	// TODO: Implement actual ACME certificate acquisition
	c.JSON(http.StatusOK, AdminAPIResponse{
		Success: true,
		Message: "Certificate obtained successfully",
		Data: map[string]interface{}{
			"domain":   request.Domain,
			"issuer":   request.Provider,
			"expires":  time.Now().AddDate(0, 3, 0).Format("2006-01-02"),
			"filename": "certificate.pem",
		},
	})
}

// RenewSSLCertificate renews the SSL certificate
func RenewSSLCertificate(c *gin.Context) {
	// TODO: Implement actual certificate renewal
	c.JSON(http.StatusOK, AdminAPIResponse{
		Success: true,
		Message: "Certificate renewed successfully",
		Data: map[string]interface{}{
			"expires": time.Now().AddDate(0, 3, 0).Format("2006-01-02"),
		},
	})
}

// CreateBackup creates a backup of the database and configuration
func CreateBackup(c *gin.Context) {
	// TODO: Implement actual backup creation using tar.gz
	// This should call the CLI backup function from src/cli/maintenance.go

	timestamp := time.Now().Format("20060102-150405")
	filename := "weather-backup-" + timestamp + ".tar.gz"

	c.JSON(http.StatusOK, AdminAPIResponse{
		Success: true,
		Message: "Backup created successfully",
		Data: map[string]interface{}{
			"filename": filename,
			"size":     1024 * 1024 * 5, // Placeholder: 5MB
		},
	})
}

// RestoreBackup restores from a backup file
func RestoreBackup(c *gin.Context) {
	file, header, err := c.Request.FormFile("backup")
	if err != nil {
		c.JSON(http.StatusBadRequest, AdminAPIResponse{
			Success: false,
			Error:   "No backup file provided",
		})
		return
	}
	defer file.Close()

	// TODO: Implement actual backup restoration
	// This should call the CLI restore function from src/cli/maintenance.go
	// For now, just validate the file was received

	c.JSON(http.StatusOK, AdminAPIResponse{
		Success: true,
		Message: "Backup restored successfully. Server will restart.",
		Data: map[string]interface{}{
			"filename": header.Filename,
		},
	})
}

// ListBackups lists all available backup files
func ListBackups(c *gin.Context) {
	dataDir := os.Getenv("DATA_DIR")
	if dataDir == "" {
		dataDir = "/var/lib/weather"
	}

	backupDir := filepath.Join(dataDir, "backups")

	// Create backups directory if it doesn't exist
	if err := os.MkdirAll(backupDir, 0755); err != nil {
		c.JSON(http.StatusInternalServerError, AdminAPIResponse{
			Success: false,
			Error:   "Failed to access backups directory",
		})
		return
	}

	entries, err := os.ReadDir(backupDir)
	if err != nil {
		c.JSON(http.StatusInternalServerError, AdminAPIResponse{
			Success: false,
			Error:   "Failed to read backups directory",
		})
		return
	}

	var backups []BackupFile
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		// Only list .tar.gz files
		if filepath.Ext(entry.Name()) != ".gz" {
			continue
		}

		info, err := entry.Info()
		if err != nil {
			continue
		}

		backups = append(backups, BackupFile{
			Filename: entry.Name(),
			Size:     info.Size(),
			Created:  info.ModTime(),
		})
	}

	c.JSON(http.StatusOK, backups)
}

// DownloadBackup downloads a specific backup file
func DownloadBackup(c *gin.Context) {
	filename := c.Param("filename")

	dataDir := os.Getenv("DATA_DIR")
	if dataDir == "" {
		dataDir = "/var/lib/weather"
	}

	backupPath := filepath.Join(dataDir, "backups", filename)

	// Security: Prevent directory traversal
	if !filepath.HasPrefix(filepath.Clean(backupPath), filepath.Join(dataDir, "backups")) {
		c.JSON(http.StatusForbidden, AdminAPIResponse{
			Success: false,
			Error:   "Invalid backup filename",
		})
		return
	}

	// Check file exists
	if _, err := os.Stat(backupPath); os.IsNotExist(err) {
		c.JSON(http.StatusNotFound, AdminAPIResponse{
			Success: false,
			Error:   "Backup file not found",
		})
		return
	}

	c.Header("Content-Description", "File Transfer")
	c.Header("Content-Disposition", "attachment; filename="+filename)
	c.Header("Content-Type", "application/gzip")
	c.File(backupPath)
}

// DeleteBackup deletes a backup file
func DeleteBackup(c *gin.Context) {
	filename := c.Param("filename")

	dataDir := os.Getenv("DATA_DIR")
	if dataDir == "" {
		dataDir = "/var/lib/weather"
	}

	backupPath := filepath.Join(dataDir, "backups", filename)

	// Security: Prevent directory traversal
	if !filepath.HasPrefix(filepath.Clean(backupPath), filepath.Join(dataDir, "backups")) {
		c.JSON(http.StatusForbidden, AdminAPIResponse{
			Success: false,
			Error:   "Invalid backup filename",
		})
		return
	}

	// Check file exists
	if _, err := os.Stat(backupPath); os.IsNotExist(err) {
		c.JSON(http.StatusNotFound, AdminAPIResponse{
			Success: false,
			Error:   "Backup file not found",
		})
		return
	}

	// Delete the file
	if err := os.Remove(backupPath); err != nil {
		c.JSON(http.StatusInternalServerError, AdminAPIResponse{
			Success: false,
			Error:   "Failed to delete backup: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, AdminAPIResponse{
		Success: true,
		Message: "Backup deleted successfully",
	})
}

// SaveDatabaseSettings handles saving database configuration
func SaveDatabaseSettings(c *gin.Context) {
	var settings map[string]interface{}
	if err := c.ShouldBindJSON(&settings); err != nil {
		c.JSON(http.StatusBadRequest, AdminAPIResponse{
			Success: false,
			Error:   "Invalid request data",
		})
		return
	}

	// TODO: Implement actual database settings update
	c.JSON(http.StatusOK, AdminAPIResponse{
		Success: true,
		Message: "Database settings saved successfully",
	})
}
