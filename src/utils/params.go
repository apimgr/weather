package utils

import (
	"fmt"
	"strings"

	"github.com/gin-gonic/gin"
)

const (
	// MaxForecastDays is the maximum number of forecast days available
	MaxForecastDays = 16
)

// ParseQueryParams parses query parameters for weather display formatting
func ParseQueryParams(c *gin.Context) *RenderParams {
	params := &RenderParams{
		Format:   0,
		Units:    "auto",
		Language: "en",
		Days:     3,    // Default: show 3 days (adaptive based on terminal width)
		Width:    0,    // 0 = auto-detect based on content
	}

	// Check for combined parameter flags (e.g., ?TFm or ?qn)
	// Look for any query key that might be a combination of flags
	for key := range c.Request.URL.Query() {
		if len(key) > 1 && c.Query(key) == "" {
			// This might be a combined flag parameter
			for _, char := range key {
				switch char {
				case 'T':
					params.NoColors = true
				case 'F':
					params.NoFooter = true
				case 'm':
					params.Units = "metric"
				case 'u':
					params.Units = "imperial"
				case 'M':
					params.Units = "metric"
				case 'q':
					params.Quiet = true
				case 'Q':
					params.SuperQuiet = true
				case 'n':
					params.Narrow = true
				case 'A':
					params.ForceANSI = true
				}
			}
		}
	}

	// Format parameters (0-4)
	if format := c.Query("format"); format != "" {
		switch format {
		case "0":
			params.Format = 0
		case "1":
			params.Format = 1
			params.NoColors = true // Format 1-4 always output plain text only
		case "2":
			params.Format = 2
			params.NoColors = true // Format 1-4 always output plain text only
		case "3":
			params.Format = 3
			params.NoColors = true // Format 1-4 always output plain text only
		case "4":
			params.Format = 4
			params.NoColors = true // Format 1-4 always output plain text only
		}
	}

	// Unit parameters - check if parameter exists
	if _, exists := c.GetQuery("u"); exists {
		params.Units = "imperial" // USCS/Imperial
	} else if _, exists := c.GetQuery("m"); exists {
		params.Units = "metric" // Metric
	} else if _, exists := c.GetQuery("M"); exists {
		params.Units = "metric" // SI (metric with m/s for wind)
	} else if units := c.Query("units"); units != "" {
		params.Units = units
	}

	// Days parameter with max capping
	if days := c.Query("days"); days != "" {
		requestedDays := parseIntSafe(days)
		if requestedDays < 0 {
			requestedDays = 0 // Negative values become 0 (current only)
		}
		if requestedDays > MaxForecastDays {
			requestedDays = MaxForecastDays // Cap at max available days
		}
		params.Days = requestedDays
	}

	// Style options - check if parameter exists
	if _, exists := c.GetQuery("F"); exists {
		params.NoFooter = true // Hide footer
	}
	if _, exists := c.GetQuery("n"); exists {
		params.Narrow = true // Narrow format
	}
	if _, exists := c.GetQuery("q"); exists {
		params.Quiet = true // Quiet mode
	}
	if _, exists := c.GetQuery("Q"); exists {
		params.SuperQuiet = true // Super quiet mode
	}
	if _, exists := c.GetQuery("T"); exists {
		params.NoColors = true // No terminal colors
	}
	if _, exists := c.GetQuery("A"); exists {
		params.ForceANSI = true // Force ANSI/terminal output
	}

	// Check Accept header for text/plain (also disable colors)
	accept := c.GetHeader("Accept")
	if strings.Contains(accept, "text/plain") {
		params.NoColors = true
	}

	// Language
	if lang := c.Query("lang"); lang != "" {
		params.Language = lang
	}

	// Terminal width detection (for adaptive layout)
	// Priority: query param > User-Agent header > default
	if width := c.Query("width"); width != "" {
		// Manual width specification
		if w := parseIntSafe(width); w > 0 {
			params.Width = w
		}
	} else if width := c.Query("cols"); width != "" {
		// Alternative: cols parameter
		if w := parseIntSafe(width); w > 0 {
			params.Width = w
		}
	} else {
		// Try to parse from User-Agent (some terminal tools include columns info)
		// Example: curl/<version> (COLUMNS=120)
		userAgent := c.GetHeader("User-Agent")
		if strings.Contains(userAgent, "COLUMNS=") {
			parts := strings.Split(userAgent, "COLUMNS=")
			if len(parts) > 1 {
				colPart := strings.Split(parts[1], ")")[0]
				colPart = strings.Split(colPart, " ")[0]
				if w := parseIntSafe(colPart); w > 0 {
					params.Width = w
				}
			}
		}
	}

	return params
}

// parseIntSafe safely parses an integer, returning 0 on error
func parseIntSafe(s string) int {
	var i int
	_, _ = fmt.Sscanf(s, "%d", &i)
	return i
}

// GetUnits returns the appropriate units based on parameters and location
func GetUnits(params *RenderParams, countryCode string) string {
	// If units explicitly set, use them
	if params.Units != "auto" {
		return params.Units
	}

	// Auto-detect based on country
	// US uses imperial, rest of world uses metric
	if countryCode == "US" {
		return "imperial"
	}

	return "metric"
}
