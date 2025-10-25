//go:build windows
// +build windows

package main

import (
	"log"
	"os"
	"syscall"

	"weather-go/src/database"
	"weather-go/src/models"
	"weather-go/src/utils"
)

func handlePlatformSignal(sig os.Signal, db *database.DB, appLogger *utils.Logger, dirPaths *utils.DirectoryPaths) bool {
	switch sig {
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
