package utils

import (
	"fmt"
	"os"

	"github.com/gin-gonic/gin"
)

// GetHostInfo detects the host information from request headers
func GetHostInfo(c *gin.Context) *HostInfo {
	// Detect protocol from headers or TLS
	protocol := c.GetHeader("X-Forwarded-Proto")
	if protocol == "" {
		if c.Request.TLS != nil {
			protocol = "https"
		} else {
			protocol = "http"
		}
	}

	// Detect hostname from headers
	hostname := c.GetHeader("X-Forwarded-Host")
	if hostname == "" {
		hostname = c.Request.Host
	}

	// Build full host URL
	fullHost := fmt.Sprintf("%s://%s", protocol, hostname)

	return &HostInfo{
		Protocol:   protocol,
		Hostname:   hostname,
		Port:       os.Getenv("PORT"),
		FullHost:   fullHost,
		ExampleURL: fmt.Sprintf("%s/London", fullHost),
	}
}

// GetClientIP extracts the real client IP from reverse proxy headers
func GetClientIP(c *gin.Context) string {
	// Check Cloudflare header first
	if ip := c.GetHeader("CF-Connecting-IP"); ip != "" {
		return ip
	}

	// Check X-Real-IP
	if ip := c.GetHeader("X-Real-IP"); ip != "" {
		return ip
	}

	// Check X-Forwarded-For (use first IP)
	if ip := c.GetHeader("X-Forwarded-For"); ip != "" {
		// X-Forwarded-For can contain multiple IPs: "client, proxy1, proxy2"
		// We want the first one (the actual client)
		for i, char := range ip {
			if char == ',' {
				return ip[:i]
			}
		}
		return ip
	}

	// Check True-Client-IP
	if ip := c.GetHeader("True-Client-IP"); ip != "" {
		return ip
	}

	// Fallback to Gin's ClientIP
	return c.ClientIP()
}

// IsLocalhost checks if an IP is localhost or private
func IsLocalhost(ip string) bool {
	localIPs := []string{
		"127.0.0.1",
		"::1",
		"localhost",
		"172.17.0.1", // Docker bridge
		"172.18.0.1",
		"172.19.0.1",
	}

	for _, localIP := range localIPs {
		if ip == localIP {
			return true
		}
	}

	// Check for private IP ranges
	if len(ip) > 3 {
		if ip[:4] == "10." || ip[:8] == "192.168." {
			return true
		}
	}

	return false
}

// IsBrowser detects if the request is from a web browser
func IsBrowser(c *gin.Context) bool {
	userAgent := c.GetHeader("User-Agent")
	if userAgent == "" {
		return false
	}

	// Simple browser detection
	ua := userAgent
	return contains(ua, "Mozilla") &&
		(contains(ua, "Chrome") || contains(ua, "Firefox") || contains(ua, "Safari") || contains(ua, "Edge"))
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && findSubstring(s, substr)
}

func findSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
