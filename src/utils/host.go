package utils

import (
	"fmt"
	"net"
	"os"
	"os/exec"
	"strings"

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

	// Detect hostname - prioritize DOMAIN env variable
	hostname := os.Getenv("DOMAIN")
	if hostname == "" {
		hostname = c.GetHeader("X-Forwarded-Host")
	}
	if hostname == "" {
		hostname = c.Request.Host
	}

	// Filter out invalid hostnames that should never be shown to users
	hostname = sanitizeHostname(hostname)

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

// sanitizeHostname filters out invalid hostnames and returns the most relevant one
func sanitizeHostname(hostname string) string {
	// Remove port from hostname for checking
	host := hostname
	if strings.Contains(hostname, ":") {
		host, _, _ = net.SplitHostPort(hostname)
	}

	// List of invalid hostnames that should never be shown
	invalidHosts := []string{
		"localhost",
		"127.0.0.1",
		"0.0.0.0",
		"::1",
		"[::1]",
		"[::]",
	}

	// Check if hostname is invalid
	for _, invalid := range invalidHosts {
		if host == invalid {
			// Try to get actual server IP or FQDN
			return getRealHostname(hostname)
		}
	}

	// Check if hostname looks like a container ID (Docker assigns random container names/IDs)
	if len(host) == 12 || len(host) == 64 {
		// Could be a Docker container ID, try to get real hostname
		return getRealHostname(hostname)
	}

	return hostname
}

// getRealHostname attempts to get the real hostname or IP
func getRealHostname(fallback string) string {
	// Try DOMAIN env variable first
	if domain := os.Getenv("DOMAIN"); domain != "" {
		return domain
	}

	// Try HOSTNAME env variable
	if hostname := os.Getenv("HOSTNAME"); hostname != "" && !isInvalidHost(hostname) {
		// Add port back if fallback had one
		if strings.Contains(fallback, ":") {
			_, port, _ := net.SplitHostPort(fallback)
			if port != "" && port != "80" && port != "443" {
				return hostname + ":" + port
			}
		}
		return hostname
	}

	// Try to get system hostname
	if cmd := exec.Command("hostname", "-I"); cmd.Err == nil {
		if output, err := cmd.Output(); err == nil {
			ips := strings.Fields(strings.TrimSpace(string(output)))
			if len(ips) > 0 && !isInvalidHost(ips[0]) {
				// Use first non-loopback IP
				if strings.Contains(fallback, ":") {
					_, port, _ := net.SplitHostPort(fallback)
					if port != "" && port != "80" && port != "443" {
						return ips[0] + ":" + port
					}
				}
				return ips[0]
			}
		}
	}

	// Try to get FQDN
	if cmd := exec.Command("hostname", "-f"); cmd.Err == nil {
		if output, err := cmd.Output(); err == nil {
			fqdn := strings.TrimSpace(string(output))
			if fqdn != "" && !isInvalidHost(fqdn) {
				if strings.Contains(fallback, ":") {
					_, port, _ := net.SplitHostPort(fallback)
					if port != "" && port != "80" && port != "443" {
						return fqdn + ":" + port
					}
				}
				return fqdn
			}
		}
	}

	// Last resort: return fallback
	return fallback
}

// isInvalidHost checks if hostname should not be shown to users
func isInvalidHost(host string) bool {
	// Remove port if present
	if strings.Contains(host, ":") {
		host, _, _ = net.SplitHostPort(host)
	}

	invalidHosts := []string{
		"localhost",
		"127.0.0.1",
		"0.0.0.0",
		"::1",
		"",
	}

	for _, invalid := range invalidHosts {
		if host == invalid {
			return true
		}
	}

	// Check for Docker container IDs (12 or 64 char hex strings)
	if len(host) == 12 || len(host) == 64 {
		// Check if it's all hex characters
		isHex := true
		for _, c := range host {
			if !((c >= '0' && c <= '9') || (c >= 'a' && c <= 'f')) {
				isHex = false
				break
			}
		}
		if isHex {
			return true
		}
	}

	return false
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
