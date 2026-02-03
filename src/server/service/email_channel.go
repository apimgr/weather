package service

import (
	"fmt"
)

// EmailChannel implements NotificationChannel for email delivery
type EmailChannel struct {
	smtp    *SMTPService
	enabled bool
}

// NewEmailChannel creates a new email channel
func NewEmailChannel(smtp *SMTPService) *EmailChannel {
	return &EmailChannel{
		smtp:    smtp,
		enabled: smtp != nil && smtp.IsEnabled(),
	}
}

// GetType returns the channel type
func (e *EmailChannel) GetType() string {
	return "email"
}

// GetName returns the channel display name
func (e *EmailChannel) GetName() string {
	return "Email (SMTP)"
}

// IsEnabled returns whether the channel is enabled
func (e *EmailChannel) IsEnabled() bool {
	return e.enabled && e.smtp != nil && e.smtp.IsEnabled()
}

// Send sends an email notification
func (e *EmailChannel) Send(recipient, subject, body string, metadata map[string]interface{}) error {
	if !e.IsEnabled() {
		return fmt.Errorf("email channel not enabled: SMTP not configured")
	}

	// Use the SMTP service to send email
	err := e.smtp.SendEmail(recipient, subject, body)
	if err != nil {
		return fmt.Errorf("failed to send email: %w", err)
	}

	return nil
}

// Test sends a test email to verify configuration
func (e *EmailChannel) Test(recipient string) error {
	if e.smtp == nil {
		return fmt.Errorf("SMTP service not initialized")
	}
	return e.smtp.SendTestEmail(recipient)
}

// ValidateConfig validates the channel configuration
func (e *EmailChannel) ValidateConfig(config map[string]interface{}) error {
	// Required fields for SMTP
	required := []string{"host", "port", "from_address"}
	for _, field := range required {
		if _, ok := config[field]; !ok {
			return fmt.Errorf("missing required field: %s", field)
		}
	}

	// Validate port is a number
	if port, ok := config["port"]; ok {
		switch port.(type) {
		case int, int64, float64:
			// OK
		default:
			return fmt.Errorf("port must be a number")
		}
	}

	return nil
}

// Refresh refreshes the channel state (re-checks SMTP configuration)
func (e *EmailChannel) Refresh() {
	if e.smtp != nil {
		e.enabled = e.smtp.IsEnabled()
	}
}
