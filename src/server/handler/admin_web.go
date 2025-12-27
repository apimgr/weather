package handlers

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/apimgr/weather/src/config"
	"github.com/apimgr/weather/src/database"
)

// AdminWebHandler handles web settings administration
type AdminWebHandler struct {
	db *database.DB
}

// NewAdminWebHandler creates a new admin web handler
func NewAdminWebHandler(db *database.DB) *AdminWebHandler {
	return &AdminWebHandler{db: db}
}

// ShowWebSettings renders the web settings admin page
// GET /admin/server/web
func (h *AdminWebHandler) ShowWebSettings(c *gin.Context) {
	cfg := config.GetGlobalConfig()

	robotsTxt := ""
	securityTxt := ""
	if cfg != nil {
		robotsTxt = cfg.Web.RobotsTxt
		securityTxt = cfg.Web.SecurityTxt
	}

	// Get app URL for template variable replacement
	scheme := "http"
	if c.Request.TLS != nil {
		scheme = "https"
	}
	appURL := scheme + "://" + c.Request.Host

	c.HTML(http.StatusOK, "admin_web.tmpl", gin.H{
		"Title":       "Web Settings",
		"RobotsTxt":   robotsTxt,
		"SecurityTxt": securityTxt,
		"AppURL":      appURL,
		"User":        c.MustGet("user"),
	})
}

// GetRobotsTxt retrieves robots.txt content
// GET /api/v1/admin/server/web/robots
func (h *AdminWebHandler) GetRobotsTxt(c *gin.Context) {
	cfg := config.GetGlobalConfig()
	robotsTxt := ""

	if cfg != nil {
		robotsTxt = cfg.Web.RobotsTxt
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"content": robotsTxt,
	})
}

// UpdateRobotsTxt updates robots.txt content
// PATCH /api/v1/admin/server/web/robots
func (h *AdminWebHandler) UpdateRobotsTxt(c *gin.Context) {
	var req struct {
		Content string `json:"content" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": gin.H{
				"code":    "INVALID_REQUEST",
				"message": "Invalid request body",
			},
		})
		return
	}

	// Update server.yml file (config watcher will auto-reload)
	if err := config.UpdateWebRobotsTxt(req.Content); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": gin.H{
				"code":    "UPDATE_FAILED",
				"message": "Failed to update robots.txt in server.yml",
			},
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "robots.txt updated successfully (will auto-reload)",
	})
}

// GetSecurityTxt retrieves security.txt content
// GET /api/v1/admin/server/web/security
func (h *AdminWebHandler) GetSecurityTxt(c *gin.Context) {
	cfg := config.GetGlobalConfig()
	securityTxt := ""

	if cfg != nil {
		securityTxt = cfg.Web.SecurityTxt
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"content": securityTxt,
	})
}

// UpdateSecurityTxt updates security.txt content
// PATCH /api/v1/admin/server/web/security
func (h *AdminWebHandler) UpdateSecurityTxt(c *gin.Context) {
	var req struct {
		Content string `json:"content" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": gin.H{
				"code":    "INVALID_REQUEST",
				"message": "Invalid request body",
			},
		})
		return
	}

	// Update server.yml file (config watcher will auto-reload)
	if err := config.UpdateWebSecurityTxt(req.Content); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": gin.H{
				"code":    "UPDATE_FAILED",
				"message": "Failed to update security.txt in server.yml",
			},
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "security.txt updated successfully (will auto-reload)",
	})
}

// ServeRobotsTxt serves robots.txt file
// GET /robots.txt
func (h *AdminWebHandler) ServeRobotsTxt(c *gin.Context) {
	cfg := config.GetGlobalConfig()
	robotsTxt := `User-agent: *
Allow: /`

	if cfg != nil && cfg.Web.RobotsTxt != "" {
		robotsTxt = cfg.Web.RobotsTxt
	}

	// Replace template variables
	scheme := "http"
	if c.Request.TLS != nil {
		scheme = "https"
	}
	appURL := scheme + "://" + c.Request.Host

	robotsTxt = strings.ReplaceAll(robotsTxt, "{app_url}", appURL)

	c.Header("Content-Type", "text/plain; charset=utf-8")
	// Cache for 24 hours
	c.Header("Cache-Control", "public, max-age=86400")
	c.String(http.StatusOK, robotsTxt)
}

// ServeSecurityTxt serves security.txt file
// GET /.well-known/security.txt
func (h *AdminWebHandler) ServeSecurityTxt(c *gin.Context) {
	cfg := config.GetGlobalConfig()
	securityTxt := ""

	if cfg != nil && cfg.Web.SecurityTxt != "" {
		securityTxt = cfg.Web.SecurityTxt
	}

	if securityTxt == "" {
		// Generate default security.txt if not configured
		securityTxt = `Contact: mailto:security@example.com
Expires: 2026-12-31T23:59:59.000Z
Preferred-Languages: en`
	}

	// Replace template variables
	scheme := "http"
	if c.Request.TLS != nil {
		scheme = "https"
	}
	appURL := scheme + "://" + c.Request.Host

	securityTxt = strings.ReplaceAll(securityTxt, "{app_url}", appURL)

	c.Header("Content-Type", "text/plain; charset=utf-8")
	// Cache for 24 hours
	c.Header("Cache-Control", "public, max-age=86400")
	c.String(http.StatusOK, securityTxt)
}
