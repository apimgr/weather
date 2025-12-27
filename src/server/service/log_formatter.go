package services

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

// LogFormat represents supported log format types
// TEMPLATE.md Part 25: Support 7 log formats
type LogFormat string

const (
	LogFormatApache   LogFormat = "apache"   // Apache Combined Log Format
	LogFormatNginx    LogFormat = "nginx"    // Nginx access log format
	LogFormatJSON     LogFormat = "json"     // JSON structured logs
	LogFormatFail2ban LogFormat = "fail2ban" // fail2ban-compatible format
	LogFormatSyslog   LogFormat = "syslog"   // RFC 5424 Syslog
	LogFormatCEF      LogFormat = "cef"      // Common Event Format (ArcSight)
	LogFormatText     LogFormat = "text"     // Custom text format
)

// LogEntry represents a single log entry with all fields
type LogEntry struct {
	Timestamp      time.Time
	RemoteAddr     string
	Method         string
	Path           string
	Protocol       string
	StatusCode     int
	BytesSent      int
	Referer        string
	UserAgent      string
	RequestTime    float64 // in seconds
	RequestID      string
	Username       string // For authenticated requests
	ErrorMessage   string // For error logs
	Facility       string // For syslog
	Severity       string // For syslog/CEF
	DeviceVendor   string // For CEF
	DeviceProduct  string // For CEF
	DeviceVersion  string // For CEF
	SignatureID    string // For CEF
	Name           string // For CEF
}

// LogFormatter handles different log output formats
type LogFormatter struct {
	format         LogFormat
	deviceVendor   string
	deviceProduct  string
	deviceVersion  string
}

// NewLogFormatter creates a new log formatter
func NewLogFormatter(format LogFormat) *LogFormatter {
	return &LogFormatter{
		format:         format,
		deviceVendor:   "apimgr",
		deviceProduct:  "weather",
		deviceVersion:  "1.0",
	}
}

// Format formats a log entry according to the configured format
func (f *LogFormatter) Format(entry *LogEntry) string {
	switch f.format {
	case LogFormatApache:
		return f.formatApache(entry)
	case LogFormatNginx:
		return f.formatNginx(entry)
	case LogFormatJSON:
		return f.formatJSON(entry)
	case LogFormatFail2ban:
		return f.formatFail2ban(entry)
	case LogFormatSyslog:
		return f.formatSyslog(entry)
	case LogFormatCEF:
		return f.formatCEF(entry)
	case LogFormatText:
		return f.formatText(entry)
	default:
		return f.formatApache(entry) // Default to Apache
	}
}

// formatApache formats in Apache Combined Log Format
// TEMPLATE.md Part 25: Apache Combined format
// Format: %h %l %u %t "%r" %>s %b "%{Referer}i" "%{User-agent}i"
func (f *LogFormatter) formatApache(entry *LogEntry) string {
	username := "-"
	if entry.Username != "" {
		username = entry.Username
	}

	referer := "-"
	if entry.Referer != "" {
		referer = entry.Referer
	}

	userAgent := "-"
	if entry.UserAgent != "" {
		userAgent = entry.UserAgent
	}

	timestamp := entry.Timestamp.Format("02/Jan/2006:15:04:05 -0700")

	return fmt.Sprintf("%s - %s [%s] \"%s %s %s\" %d %d \"%s\" \"%s\"",
		entry.RemoteAddr,
		username,
		timestamp,
		entry.Method,
		entry.Path,
		entry.Protocol,
		entry.StatusCode,
		entry.BytesSent,
		referer,
		userAgent,
	)
}

// formatNginx formats in Nginx access log format
// TEMPLATE.md Part 25: Nginx format
func (f *LogFormatter) formatNginx(entry *LogEntry) string {
	username := "-"
	if entry.Username != "" {
		username = entry.Username
	}

	referer := "-"
	if entry.Referer != "" {
		referer = entry.Referer
	}

	userAgent := "-"
	if entry.UserAgent != "" {
		userAgent = entry.UserAgent
	}

	timestamp := entry.Timestamp.Format("02/Jan/2006:15:04:05 -0700")

	return fmt.Sprintf("%s - %s [%s] \"%s %s %s\" %d %d \"%s\" \"%s\" %.3f",
		entry.RemoteAddr,
		username,
		timestamp,
		entry.Method,
		entry.Path,
		entry.Protocol,
		entry.StatusCode,
		entry.BytesSent,
		referer,
		userAgent,
		entry.RequestTime,
	)
}

// formatJSON formats as JSON structured log
// TEMPLATE.md Part 25: JSON format
func (f *LogFormatter) formatJSON(entry *LogEntry) string {
	logData := map[string]interface{}{
		"timestamp":    entry.Timestamp.Format(time.RFC3339Nano),
		"remote_addr":  entry.RemoteAddr,
		"method":       entry.Method,
		"path":         entry.Path,
		"protocol":     entry.Protocol,
		"status_code":  entry.StatusCode,
		"bytes_sent":   entry.BytesSent,
		"request_time": entry.RequestTime,
		"request_id":   entry.RequestID,
	}

	if entry.Username != "" {
		logData["username"] = entry.Username
	}

	if entry.Referer != "" {
		logData["referer"] = entry.Referer
	}

	if entry.UserAgent != "" {
		logData["user_agent"] = entry.UserAgent
	}

	if entry.ErrorMessage != "" {
		logData["error"] = entry.ErrorMessage
	}

	jsonBytes, _ := json.Marshal(logData)
	return string(jsonBytes)
}

// formatFail2ban formats for fail2ban compatibility
// TEMPLATE.md Part 25: fail2ban-compatible format
func (f *LogFormatter) formatFail2ban(entry *LogEntry) string {
	timestamp := entry.Timestamp.Format("2006-01-02 15:04:05")

	// fail2ban looks for patterns like:
	// "Failed login from <IP>" or "Authentication failure from <IP>"
	action := "access"
	if entry.StatusCode >= 400 {
		action = "failed"
	}
	if entry.StatusCode == 401 || entry.StatusCode == 403 {
		action = "auth_failed"
	}

	return fmt.Sprintf("[%s] %s: %s %s %s from %s - status=%d",
		timestamp,
		action,
		entry.Method,
		entry.Path,
		entry.Protocol,
		entry.RemoteAddr,
		entry.StatusCode,
	)
}

// formatSyslog formats according to RFC 5424 Syslog
// TEMPLATE.md Part 25: RFC 5424 Syslog format
func (f *LogFormatter) formatSyslog(entry *LogEntry) string {
	// RFC 5424 format: <PRI>VERSION TIMESTAMP HOSTNAME APP-NAME PROCID MSGID STRUCTURED-DATA MSG

	// Calculate priority: facility * 8 + severity
	facility := 16 // local0
	severity := 6  // informational
	if entry.StatusCode >= 500 {
		severity = 3 // error
	} else if entry.StatusCode >= 400 {
		severity = 4 // warning
	}
	priority := facility*8 + severity

	timestamp := entry.Timestamp.Format(time.RFC3339)
	hostname := "-"
	appName := "weather"
	procID := "-"
	msgID := entry.RequestID
	if msgID == "" {
		msgID = "-"
	}

	// Structured data
	structuredData := fmt.Sprintf("[request@48577 method=\"%s\" path=\"%s\" status=\"%d\" ip=\"%s\"]",
		entry.Method,
		entry.Path,
		entry.StatusCode,
		entry.RemoteAddr,
	)

	msg := fmt.Sprintf("%s %s %s %d bytes %.3fs",
		entry.Method,
		entry.Path,
		entry.Protocol,
		entry.BytesSent,
		entry.RequestTime,
	)

	return fmt.Sprintf("<%d>1 %s %s %s %s %s %s %s",
		priority,
		timestamp,
		hostname,
		appName,
		procID,
		msgID,
		structuredData,
		msg,
	)
}

// formatCEF formats as Common Event Format (ArcSight)
// TEMPLATE.md Part 25: CEF format
func (f *LogFormatter) formatCEF(entry *LogEntry) string {
	// CEF format:
	// CEF:Version|Device Vendor|Device Product|Device Version|Signature ID|Name|Severity|Extension

	version := "0"
	signatureID := fmt.Sprintf("HTTP_%d", entry.StatusCode)
	name := fmt.Sprintf("HTTP %s", entry.Method)

	severity := "3" // Medium
	if entry.StatusCode >= 500 {
		severity = "8" // High
	} else if entry.StatusCode >= 400 {
		severity = "5" // Medium-High
	} else if entry.StatusCode >= 300 {
		severity = "2" // Low
	} else if entry.StatusCode >= 200 {
		severity = "1" // Very Low
	}

	// Extension fields (key=value pairs)
	extensions := []string{
		fmt.Sprintf("rt=%d", entry.Timestamp.Unix()*1000), // milliseconds
		fmt.Sprintf("src=%s", entry.RemoteAddr),
		fmt.Sprintf("request=%s %s", entry.Method, entry.Path),
		fmt.Sprintf("requestMethod=%s", entry.Method),
		fmt.Sprintf("requestUrl=%s", entry.Path),
		fmt.Sprintf("cs1Label=Protocol"),
		fmt.Sprintf("cs1=%s", entry.Protocol),
		fmt.Sprintf("cs2Label=StatusCode"),
		fmt.Sprintf("cs2=%d", entry.StatusCode),
		fmt.Sprintf("cn1Label=BytesSent"),
		fmt.Sprintf("cn1=%d", entry.BytesSent),
		fmt.Sprintf("cn2Label=RequestTime"),
		fmt.Sprintf("cn2=%.3f", entry.RequestTime),
	}

	if entry.UserAgent != "" {
		extensions = append(extensions, fmt.Sprintf("requestClientApplication=%s", escapeCEF(entry.UserAgent)))
	}

	if entry.Username != "" {
		extensions = append(extensions, fmt.Sprintf("suser=%s", entry.Username))
	}

	extensionStr := strings.Join(extensions, " ")

	return fmt.Sprintf("CEF:%s|%s|%s|%s|%s|%s|%s|%s",
		version,
		f.deviceVendor,
		f.deviceProduct,
		f.deviceVersion,
		signatureID,
		name,
		severity,
		extensionStr,
	)
}

// formatText formats as custom text format
// TEMPLATE.md Part 25: Custom text format
func (f *LogFormatter) formatText(entry *LogEntry) string {
	timestamp := entry.Timestamp.Format("2006-01-02 15:04:05.000")

	status := fmt.Sprintf("%d", entry.StatusCode)
	if entry.StatusCode >= 500 {
		status = fmt.Sprintf("ERROR %d", entry.StatusCode)
	} else if entry.StatusCode >= 400 {
		status = fmt.Sprintf("WARN %d", entry.StatusCode)
	} else if entry.StatusCode >= 300 {
		status = fmt.Sprintf("REDIR %d", entry.StatusCode)
	} else {
		status = fmt.Sprintf("OK %d", entry.StatusCode)
	}

	parts := []string{
		fmt.Sprintf("[%s]", timestamp),
		fmt.Sprintf("[%s]", entry.RemoteAddr),
		fmt.Sprintf("[%s]", status),
		fmt.Sprintf("%s %s", entry.Method, entry.Path),
		fmt.Sprintf("%.0fms", entry.RequestTime*1000),
		fmt.Sprintf("%dB", entry.BytesSent),
	}

	if entry.RequestID != "" {
		parts = append(parts, fmt.Sprintf("id=%s", entry.RequestID))
	}

	if entry.Username != "" {
		parts = append(parts, fmt.Sprintf("user=%s", entry.Username))
	}

	return strings.Join(parts, " ")
}

// escapeCSF escapes special characters for CEF format
func escapeCEF(s string) string {
	s = strings.ReplaceAll(s, "\\", "\\\\")
	s = strings.ReplaceAll(s, "|", "\\|")
	s = strings.ReplaceAll(s, "=", "\\=")
	s = strings.ReplaceAll(s, "\n", "\\n")
	s = strings.ReplaceAll(s, "\r", "\\r")
	return s
}

// ExtractLogEntry extracts log entry data from Gin context
func ExtractLogEntry(c *gin.Context, startTime time.Time, bytesWritten int) *LogEntry {
	entry := &LogEntry{
		Timestamp:   startTime,
		RemoteAddr:  c.ClientIP(),
		Method:      c.Request.Method,
		Path:        c.Request.URL.Path,
		Protocol:    c.Request.Proto,
		StatusCode:  c.Writer.Status(),
		BytesSent:   bytesWritten,
		Referer:     c.Request.Referer(),
		UserAgent:   c.Request.UserAgent(),
		RequestTime: time.Since(startTime).Seconds(),
	}

	// Extract request ID if available
	if requestID, exists := c.Get("request_id"); exists {
		entry.RequestID = requestID.(string)
	}

	// Extract username if authenticated
	if username, exists := c.Get("username"); exists {
		entry.Username = username.(string)
	}

	return entry
}
