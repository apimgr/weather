//go:build !windows
// +build !windows

package main

import (
	"log"
	"os"
	"syscall"

	"github.com/gin-gonic/gin"
	"weather-go/src/database"
	"weather-go/src/models"
	"weather-go/src/utils"
)

func handlePlatformSignal(sig os.Signal, db *database.DB, appLogger *utils.Logger, dirPaths *utils.DirectoryPaths) bool {
	switch sig {
	case syscall.SIGUSR1:
		log.Println("📝 Received SIGUSR1, reopening log files...")
		if err := appLogger.RotateLogs(); err != nil {
			log.Printf("⚠️  Failed to rotate logs: %v", err)
		} else {
			log.Println("✅ Log files reopened")
		}
		return false

	case syscall.SIGUSR2:
		log.Println("🔧 Received SIGUSR2, toggling debug mode...")
		if gin.Mode() == gin.DebugMode {
			gin.SetMode(gin.ReleaseMode)
			log.Println("✅ Debug mode: OFF (release mode)")
		} else {
			gin.SetMode(gin.DebugMode)
			log.Println("✅ Debug mode: ON (debug mode)")
		}
		return false

	case syscall.SIGHUP:
		log.Println("🔄 Received SIGHUP, reloading configuration...")
		settingsModel := &models.SettingsModel{DB: db.DB}
		if err := settingsModel.InitializeDefaults(utils.GetBackupPath(dirPaths)); err != nil {
			log.Printf("⚠️  Failed to reload settings: %v", err)
		} else {
			log.Println("✅ Configuration reloaded")
		}
		return false
	}
	return false
}
