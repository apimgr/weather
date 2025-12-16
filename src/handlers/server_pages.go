package handlers

import (
	"log"
	"net/http"
	"weather-go/src/config"
	"weather-go/src/database"
	"weather-go/src/models"
	"weather-go/src/utils"

	"github.com/gin-gonic/gin"
)

// ShowAboutPage renders the about page
func ShowAboutPage(db *database.DB, cfg *config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		user, _ := c.Get("user")

		// Get server configuration
		settingsModel := &models.SettingsModel{DB: db.DB}

		// Get Tor configuration if available
		torEnabled := settingsModel.GetBool("tor.enabled", false)
		onionAddress := settingsModel.GetString("tor.onion_address", "")

		c.HTML(http.StatusOK, "about.tmpl", gin.H{
			"user": user,
			"page": "about",
			"server": gin.H{
				"Title":       cfg.Server.Branding.Title,
				"Description": cfg.Server.Branding.Description,
				"Version":     cfg.Version,
				"BuildDate":   cfg.BuildDate,
				"Mode":        cfg.Mode,
				"GitOrg":      "apimgr",
				"GitRepo":     "weather",
				"Tor": gin.H{
					"Enabled":      torEnabled,
					"OnionAddress": onionAddress,
				},
			},
			"HostInfo": utils.GetHostInfo(c),
		})
	}
}

// ShowPrivacyPage renders the privacy policy page
func ShowPrivacyPage(db *database.DB, cfg *config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		user, _ := c.Get("user")

		c.HTML(http.StatusOK, "privacy.tmpl", gin.H{
			"user": user,
			"page": "privacy",
			"server": gin.H{
				"Title":     cfg.Server.Branding.Title,
				"BuildDate": cfg.BuildDate,
			},
			"HostInfo": utils.GetHostInfo(c),
		})
	}
}

// ShowContactPage renders the contact form page
func ShowContactPage(db *database.DB, cfg *config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		user, _ := c.Get("user")

		c.HTML(http.StatusOK, "contact.tmpl", gin.H{
			"user": user,
			"page": "contact",
			"server": gin.H{
				"Title":   cfg.Server.Branding.Title,
				"GitOrg":  "apimgr",
				"GitRepo": "weather",
			},
			"HostInfo": utils.GetHostInfo(c),
		})
	}
}

// ShowHelpPage renders the help & documentation page
func ShowHelpPage(db *database.DB, cfg *config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		user, _ := c.Get("user")

		c.HTML(http.StatusOK, "help.tmpl", gin.H{
			"user": user,
			"page": "help",
			"server": gin.H{
				"Title":   cfg.Server.Branding.Title,
				"GitOrg":  "apimgr",
				"GitRepo": "weather",
			},
			"HostInfo": utils.GetHostInfo(c),
		})
	}
}

// HandleContactFormSubmission handles the contact form POST request (API endpoint)
func HandleContactFormSubmission(db *database.DB, cfg *config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		var form struct {
			Name    string `json:"name" binding:"required"`
			Email   string `json:"email" binding:"required,email"`
			Subject string `json:"subject" binding:"required"`
			Message string `json:"message" binding:"required"`
		}

		if err := c.ShouldBindJSON(&form); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid form data"})
			return
		}

		// Log the contact form submission
		// In a production environment, this would send an email or create a support ticket
		// For now, we'll just log it
		log.Printf("📧 Contact form submission: Name=%s, Email=%s, Subject=%s", form.Name, form.Email, form.Subject)

		// TODO: Implement email sending via SMTP when configured
		// TODO: Or create a support ticket in the database

		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"message": "Thank you for contacting us. We'll get back to you soon.",
		})
	}
}
