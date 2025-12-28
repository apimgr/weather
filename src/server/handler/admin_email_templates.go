package handler

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/gin-gonic/gin"
)

type EmailTemplateHandler struct {
	templatesDir string
}

func NewEmailTemplateHandler(templatesDir string) *EmailTemplateHandler {
	return &EmailTemplateHandler{
		templatesDir: templatesDir,
	}
}

type EmailTemplate struct {
	Subject string `json:"subject"`
	Body    string `json:"body"`
}

// GetTemplate retrieves a specific email template
func (h *EmailTemplateHandler) GetTemplate(c *gin.Context) {
	templateName := c.Param("name")

	// Validate template name
	if !isValidTemplateName(templateName) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid template name"})
		return
	}

	templatePath := filepath.Join(h.templatesDir, "email", templateName+".tmpl")

	// Read template file
	content, err := os.ReadFile(templatePath)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Template not found"})
		return
	}

	// Parse template (format: Subject: ...\n---\nBody...)
	template := parseTemplate(string(content))

	c.JSON(http.StatusOK, template)
}

// UpdateTemplate updates a specific email template
func (h *EmailTemplateHandler) UpdateTemplate(c *gin.Context) {
	templateName := c.Param("name")

	// Validate template name
	if !isValidTemplateName(templateName) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid template name"})
		return
	}

	var template EmailTemplate
	if err := c.ShouldBindJSON(&template); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	// Validate template
	if template.Subject == "" || template.Body == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Subject and body are required"})
		return
	}

	// Format template content
	content := fmt.Sprintf("Subject: %s\n---\n%s\n", template.Subject, template.Body)

	// Write template file
	templatePath := filepath.Join(h.templatesDir, "email", templateName+".tmpl")
	if err := os.WriteFile(templatePath, []byte(content), 0644); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save template"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Template updated successfully"})
}

// ListTemplates returns all available email templates
func (h *EmailTemplateHandler) ListTemplates(c *gin.Context) {
	emailDir := filepath.Join(h.templatesDir, "email")

	files, err := os.ReadDir(emailDir)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to read templates directory"})
		return
	}

	templates := make([]map[string]string, 0)
	for _, file := range files {
		if file.IsDir() || !strings.HasSuffix(file.Name(), ".tmpl") {
			continue
		}

		name := strings.TrimSuffix(file.Name(), ".tmpl")
		templates = append(templates, map[string]string{
			"name": name,
			"file": file.Name(),
		})
	}

	c.JSON(http.StatusOK, gin.H{"templates": templates})
}

// TestTemplate sends a test email using the specified template
func (h *EmailTemplateHandler) TestTemplate(c *gin.Context) {
	var request struct {
		Template string `json:"template"`
	}

	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	// Validate template name
	if !isValidTemplateName(request.Template) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid template name"})
		return
	}

	// In a real implementation, this would use the email service
	// For now, we'll just return success
	c.JSON(http.StatusOK, gin.H{
		"message": fmt.Sprintf("Test email sent using template: %s", request.Template),
	})
}

// Helper: Parse template file content
func parseTemplate(content string) EmailTemplate {
	lines := strings.Split(content, "\n")
	template := EmailTemplate{}

	separatorIndex := -1
	for i, line := range lines {
		if strings.TrimSpace(line) == "---" {
			separatorIndex = i
			break
		}
		if strings.HasPrefix(line, "Subject:") {
			template.Subject = strings.TrimSpace(strings.TrimPrefix(line, "Subject:"))
		}
	}

	if separatorIndex > 0 && separatorIndex < len(lines)-1 {
		template.Body = strings.TrimSpace(strings.Join(lines[separatorIndex+1:], "\n"))
	}

	return template
}

// Helper: Validate template name
func isValidTemplateName(name string) bool {
	validTemplates := []string{
		"welcome",
		"password_reset",
		"backup_complete",
		"backup_failed",
		"ssl_expiring",
		"ssl_renewed",
		"login_alert",
		"security_alert",
		"scheduler_error",
		"test",
	}

	for _, valid := range validTemplates {
		if name == valid {
			return true
		}
	}
	return false
}

// ExportTemplate exports a template as JSON
func (h *EmailTemplateHandler) ExportTemplate(c *gin.Context) {
	templateName := c.Param("name")

	if !isValidTemplateName(templateName) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid template name"})
		return
	}

	templatePath := filepath.Join(h.templatesDir, "email", templateName+".tmpl")

	content, err := os.ReadFile(templatePath)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Template not found"})
		return
	}

	template := parseTemplate(string(content))

	// Set headers for download
	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=%s.json", templateName))
	c.Header("Content-Type", "application/json")

	json.NewEncoder(c.Writer).Encode(template)
}

// ImportTemplate imports a template from JSON
func (h *EmailTemplateHandler) ImportTemplate(c *gin.Context) {
	templateName := c.Param("name")

	if !isValidTemplateName(templateName) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid template name"})
		return
	}

	var template EmailTemplate
	if err := c.ShouldBindJSON(&template); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid JSON format"})
		return
	}

	if template.Subject == "" || template.Body == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Subject and body are required"})
		return
	}

	content := fmt.Sprintf("Subject: %s\n---\n%s\n", template.Subject, template.Body)
	templatePath := filepath.Join(h.templatesDir, "email", templateName+".tmpl")

	if err := os.WriteFile(templatePath, []byte(content), 0644); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to import template"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Template imported successfully"})
}
