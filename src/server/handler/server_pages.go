package handler

import (
	"fmt"
	"net/http"
	"time"

	"github.com/apimgr/weather/src/config"
	"github.com/apimgr/weather/src/database"
	"github.com/apimgr/weather/src/server/model"
	"github.com/apimgr/weather/src/server/service"
	"github.com/apimgr/weather/src/utils"

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
				"Version":     "1.0.0",
				"BuildDate":   "",
				"Mode":        cfg.Server.Mode,
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
				"BuildDate": "",
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

		// Try to send email via SMTP if configured (AI.md PART 26)
		smtpService := GetSMTPService(c)
		if smtpService != nil {
			// SMTP available - send email
			emailBody := fmt.Sprintf(`Contact Form Submission

From: %s <%s>
Subject: %s

Message:
%s

---
IP: %s
User Agent: %s
Time: %s`, form.Name, form.Email, form.Subject, form.Message, c.ClientIP(), c.Request.UserAgent(), time.Now().Format("2006-01-02 15:04:05"))

			// Get admin email from config
			cfg, _ := config.LoadConfig()
			adminEmail := "admin@localhost"
			if cfg != nil && cfg.Server.Admin.Email != "" {
				adminEmail = cfg.Server.Admin.Email
			}

			err := smtpService.SendEmail(adminEmail, fmt.Sprintf("Contact: %s", form.Subject), emailBody)
			if err != nil {
				// Email failed - save to database as fallback
				if err := saveContactToDB(c, form.Name, form.Email, form.Subject, form.Message); err != nil {
					c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": "Failed to send message"})
					return
				}
				c.JSON(http.StatusOK, gin.H{"success": true, "message": "Your message has been saved. We'll respond as soon as possible."})
				return
			}
			c.JSON(http.StatusOK, gin.H{"success": true, "message": "Thank you for contacting us. We'll get back to you soon."})
		} else {
			// No SMTP - save to database (AI.md PART 26)
			if err := saveContactToDB(c, form.Name, form.Email, form.Subject, form.Message); err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": "Failed to save message"})
				return
			}
			c.JSON(http.StatusOK, gin.H{"success": true, "message": "Your message has been saved. We'll respond as soon as possible."})
		}
	}
}

// saveContactToDB saves contact form submission to database when SMTP unavailable
// Per AI.md PART 26: Graceful degradation when SMTP not configured
func saveContactToDB(c *gin.Context, name, email, subject, message string) error {
	dbInterface, exists := c.Get("db")
	if !exists {
		return fmt.Errorf("database not available")
	}
	db, ok := dbInterface.(*database.DB)
	if !ok || db == nil {
		return fmt.Errorf("database not available")
	}

	// Create contact_submissions table if not exists
	_, err := db.DB.Exec(`
		CREATE TABLE IF NOT EXISTS contact_submissions (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			name TEXT NOT NULL,
			email TEXT NOT NULL,
			subject TEXT NOT NULL,
			message TEXT NOT NULL,
			ip_address TEXT,
			user_agent TEXT,
			status TEXT NOT NULL DEFAULT 'pending',
			created_at INTEGER NOT NULL DEFAULT (strftime('%s', 'now'))
		)
	`)
	if err != nil {
		return fmt.Errorf("failed to create table: %w", err)
	}

	// Insert contact submission
	_, err = db.DB.Exec(`
		INSERT INTO contact_submissions (name, email, subject, message, ip_address, user_agent)
		VALUES (?, ?, ?, ?, ?, ?)
	`, name, email, subject, message, c.ClientIP(), c.Request.UserAgent())

	return err
}

// GetSMTPService returns the SMTP service from context if available
func GetSMTPService(c *gin.Context) *service.SMTPService {
	if smtp, exists := c.Get("smtp"); exists {
		if smtpService, ok := smtp.(*service.SMTPService); ok {
			return smtpService
		}
	}
	return nil
}
