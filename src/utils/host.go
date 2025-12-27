package utils

import (
	"fmt"
	"net"
	"os"
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

	// Get hostname from request (reverse proxy headers first)
	hostname := GetHostFromRequest(c)

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

// GetHostFromRequest resolves hostname from request (TEMPLATE.md compliant)
// Priority: X-Forwarded-Host > X-Real-Host > X-Original-Host > GetFQDN()
func GetHostFromRequest(c *gin.Context) string {
	// 1. Reverse proxy headers (highest priority)
	for _, header := range []string{"X-Forwarded-Host", "X-Real-Host", "X-Original-Host"} {
		if host := c.GetHeader(header); host != "" {
			// Strip port if present
			if h, _, err := net.SplitHostPort(host); err == nil {
				return h
			}
			return host
		}
	}

	// 2. Fall back to static FQDN resolution
	return GetFQDN()
}

// GetFQDN resolves the fully qualified domain name (TEMPLATE.md compliant)
// Priority: DOMAIN env > os.Hostname() > HOSTNAME env > IPv6 > IPv4
func GetFQDN() string {
	// 1. DOMAIN env var (explicit user override)
	if domain := os.Getenv("DOMAIN"); domain != "" {
		return domain
	}

	// 2. os.Hostname() - cross-platform
	if hostname, err := os.Hostname(); err == nil && hostname != "" {
		if !isLoopback(hostname) {
			return hostname
		}
	}

	// 3. HOSTNAME env var
	if hostname := os.Getenv("HOSTNAME"); hostname != "" {
		if !isLoopback(hostname) {
			return hostname
		}
	}

	// 4. Global IPv6 (preferred for modern networks)
	if ipv6 := getGlobalIPv6(); ipv6 != "" {
		return ipv6
	}

	// 5. Global IPv4
	if ipv4 := getGlobalIPv4(); ipv4 != "" {
		return ipv4
	}

	// Last resort
	return "localhost"
}

// isLoopback checks if host is a loopback address
func isLoopback(host string) bool {
	lower := strings.ToLower(host)
	if lower == "localhost" {
		return true
	}
	if ip := net.ParseIP(host); ip != nil {
		return ip.IsLoopback()
	}
	return false
}

// getGlobalIPv6 returns the first global unicast IPv6 address
func getGlobalIPv6() string {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return ""
	}
	for _, addr := range addrs {
		if ipnet, ok := addr.(*net.IPNet); ok {
			if ipnet.IP.To4() == nil && ipnet.IP.IsGlobalUnicast() {
				return ipnet.IP.String()
			}
		}
	}
	return ""
}

// getGlobalIPv4 returns the first global unicast IPv4 address
func getGlobalIPv4() string {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return ""
	}
	for _, addr := range addrs {
		if ipnet, ok := addr.(*net.IPNet); ok {
			if ip4 := ipnet.IP.To4(); ip4 != nil && ipnet.IP.IsGlobalUnicast() {
				return ip4.String()
			}
		}
	}
	return ""
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
		// Docker bridge
		"172.17.0.1",
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
