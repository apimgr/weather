package handlers

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/apimgr/weather/src/server/model"
	"github.com/apimgr/weather/src/server/service"
)

// TorAdminHandler handles Tor administration endpoints
type TorAdminHandler struct {
	torService      *services.TorService
	settingsModel   *models.SettingsModel
	vanityGenerator *services.VanityGenerator
	keyManager      *services.TorKeyManager
}

// NewTorAdminHandler creates a new Tor admin handler
func NewTorAdminHandler(torService *services.TorService, settingsModel *models.SettingsModel, dataDir string) *TorAdminHandler {
	return &TorAdminHandler{
		torService:      torService,
		settingsModel:   settingsModel,
		vanityGenerator: services.NewVanityGenerator(),
		keyManager:      services.NewTorKeyManager(dataDir + "/tor"),
	}
}

// GetStatus returns Tor service status
// GET /api/v1/admin/server/tor/status
func (h *TorAdminHandler) GetStatus(c *gin.Context) {
	status := h.torService.GetStatus()

	c.JSON(http.StatusOK, gin.H{
		"status": status,
	})
}

// GetHealth returns Tor service health status
// GET /api/v1/admin/server/tor/health
func (h *TorAdminHandler) GetHealth(c *gin.Context) {
	health := h.torService.GetHealthStatus()

	c.JSON(http.StatusOK, gin.H{
		"health": health,
	})
}

// Enable enables the Tor service
// POST /api/v1/admin/server/tor/enable
func (h *TorAdminHandler) Enable(c *gin.Context) {
	// Get HTTP port from settings or context
	// Default, should be retrieved from actual server config
	httpPort := 8080

	// Update setting
	if err := h.settingsModel.SetBool("tor.enabled", true); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": gin.H{
				"code":    "DATABASE_ERROR",
				"message": fmt.Sprintf("Failed to update settings: %v", err),
			},
		})
		return
	}

	// Start Tor service
	if err := h.torService.Start(httpPort); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": gin.H{
				"code":    "TOR_START_FAILED",
				"message": fmt.Sprintf("Failed to start Tor: %v", err),
			},
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Tor service enabled and started",
		"status":  h.torService.GetStatus(),
	})
}

// Disable disables the Tor service
// POST /api/v1/admin/server/tor/disable
func (h *TorAdminHandler) Disable(c *gin.Context) {
	// Update setting
	if err := h.settingsModel.SetBool("tor.enabled", false); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": gin.H{
				"code":    "DATABASE_ERROR",
				"message": fmt.Sprintf("Failed to update settings: %v", err),
			},
		})
		return
	}

	// Stop Tor service
	if err := h.torService.Stop(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": gin.H{
				"code":    "TOR_STOP_FAILED",
				"message": fmt.Sprintf("Failed to stop Tor: %v", err),
			},
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Tor service disabled and stopped",
	})
}

// Regenerate regenerates the .onion address
// POST /api/v1/admin/server/tor/regenerate
func (h *TorAdminHandler) Regenerate(c *gin.Context) {
	// Should be retrieved from actual server config
	httpPort := 8080

	if err := h.torService.RegenerateAddress(httpPort); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": gin.H{
				"code":    "REGENERATE_FAILED",
				"message": fmt.Sprintf("Failed to regenerate address: %v", err),
			},
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Tor address regenerated successfully",
		"address": h.torService.GetOnionAddress(),
	})
}

// GenerateVanity starts vanity address generation
// POST /api/v1/admin/server/tor/vanity/generate
func (h *TorAdminHandler) GenerateVanity(c *gin.Context) {
	var req struct {
		Prefix string `json:"prefix" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": gin.H{
				"code":    "INVALID_REQUEST",
				"message": "Missing or invalid prefix",
			},
		})
		return
	}

	if err := h.vanityGenerator.Start(req.Prefix); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": gin.H{
				"code":    "GENERATION_FAILED",
				"message": fmt.Sprintf("Failed to start generation: %v", err),
			},
		})
		return
	}

	// Start monitoring for completion in background
	go h.monitorVanityGeneration()

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": fmt.Sprintf("Started generating vanity address with prefix: %s", req.Prefix),
		"status":  h.vanityGenerator.GetStatus(),
	})
}

// monitorVanityGeneration watches for completion and sends notification
func (h *TorAdminHandler) monitorVanityGeneration() {
	notifyCh := h.vanityGenerator.GetNotificationChannel()
	address := <-notifyCh

	// TODO: Send notification to user via notification system
	fmt.Printf("🎉 Vanity address generated: %s\n", address)
}

// GetVanityStatus returns vanity generation status
// GET /api/v1/admin/server/tor/vanity/status
func (h *TorAdminHandler) GetVanityStatus(c *gin.Context) {
	status := h.vanityGenerator.GetStatus()

	if status == nil {
		c.JSON(http.StatusOK, gin.H{
			"running": false,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"running":        status.Running,
		"prefix":         status.Prefix,
		"start_time":     status.StartTime,
		"attempts":       status.Attempts,
		"estimated_time": status.EstimatedTime,
		"address":        status.Address,
	})
}

// CancelVanity cancels vanity generation
// POST /api/v1/admin/server/tor/vanity/cancel
func (h *TorAdminHandler) CancelVanity(c *gin.Context) {
	if err := h.vanityGenerator.Cancel(); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": gin.H{
				"code":    "CANCEL_FAILED",
				"message": fmt.Sprintf("Failed to cancel: %v", err),
			},
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Vanity generation cancelled",
	})
}

// ApplyVanity applies the generated vanity keys
// POST /api/v1/admin/server/tor/vanity/apply
func (h *TorAdminHandler) ApplyVanity(c *gin.Context) {
	// Get generated keys
	publicKey, privateKey, err := h.vanityGenerator.GetKeys()
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": gin.H{
				"code":    "NO_KEYS",
				"message": fmt.Sprintf("No keys available: %v", err),
			},
		})
		return
	}

	// Import keys
	if err := h.keyManager.ImportKeys(publicKey, privateKey); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": gin.H{
				"code":    "IMPORT_FAILED",
				"message": fmt.Sprintf("Failed to import keys: %v", err),
			},
		})
		return
	}

	// Restart Tor with new keys
	httpPort := 8080
	if err := h.torService.RegenerateAddress(httpPort); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": gin.H{
				"code":    "RESTART_FAILED",
				"message": fmt.Sprintf("Failed to restart Tor: %v", err),
			},
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Vanity address applied successfully",
		"address": h.torService.GetOnionAddress(),
	})
}

// ImportKeys imports external Tor keys
// POST /api/v1/admin/server/tor/keys/import
func (h *TorAdminHandler) ImportKeys(c *gin.Context) {
	file, err := c.FormFile("key_file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": gin.H{
				"code":    "NO_FILE",
				"message": "No key file provided",
			},
		})
		return
	}

	// Save uploaded file temporarily
	tempPath := "/tmp/tor_key_upload"
	if err := c.SaveUploadedFile(file, tempPath); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": gin.H{
				"code":    "SAVE_FAILED",
				"message": fmt.Sprintf("Failed to save file: %v", err),
			},
		})
		return
	}

	// Import from file
	if err := h.keyManager.ImportFromFile(tempPath); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": gin.H{
				"code":    "IMPORT_FAILED",
				"message": fmt.Sprintf("Failed to import keys: %v", err),
			},
		})
		return
	}

	// Restart Tor with new keys
	httpPort := 8080
	if err := h.torService.RegenerateAddress(httpPort); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": gin.H{
				"code":    "RESTART_FAILED",
				"message": fmt.Sprintf("Failed to restart Tor: %v", err),
			},
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Keys imported and Tor restarted successfully",
		"address": h.torService.GetOnionAddress(),
	})
}

// ExportKeys exports current Tor keys
// GET /api/v1/admin/server/tor/keys/export
func (h *TorAdminHandler) ExportKeys(c *gin.Context) {
	publicKey, privateKey, err := h.keyManager.ExportKeys()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": gin.H{
				"code":    "EXPORT_FAILED",
				"message": fmt.Sprintf("Failed to export keys: %v", err),
			},
		})
		return
	}

	// Return private key file for download
	c.Header("Content-Disposition", "attachment; filename=hs_ed25519_secret_key")
	c.Header("Content-Type", "application/octet-stream")
	c.Header("Content-Length", strconv.Itoa(len(privateKey)))

	// Write key in Tor format (32-byte header + 32-byte key)
	header := []byte("== ed25519v1-secret: type0 ==")
	padding := make([]byte, 32-len(header))
	c.Writer.Write(header)
	c.Writer.Write(padding)
	c.Writer.Write(privateKey)

	// Public key can be derived from private
	_ = publicKey
}
