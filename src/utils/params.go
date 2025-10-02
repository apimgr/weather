package utils

import (
	"strings"

	"github.com/gin-gonic/gin"
)

// ParseWttrParams parses wttr.in-compatible query parameters
func ParseWttrParams(c *gin.Context) *RenderParams {
	params := &RenderParams{
		Format:   0,
		Units:    "auto",
		Language: "en",
		Days:     2, // Default: current + today + tomorrow
	}

	// Check for wttr.in style combined parameters (e.g., ?TFm or ?qn)
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
				case '0':
					params.Days = 0
				case '1':
					params.Days = 1
				case '2':
					params.Days = 2
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

	// Unit parameters - check if parameter exists (wttr.in style)
	if _, exists := c.GetQuery("u"); exists {
		params.Units = "imperial" // USCS/Imperial
	} else if _, exists := c.GetQuery("m"); exists {
		params.Units = "metric" // Metric
	} else if _, exists := c.GetQuery("M"); exists {
		params.Units = "metric" // SI (metric with m/s for wind)
	} else if units := c.Query("units"); units != "" {
		params.Units = units
	}

	// Display options - check if parameter exists (wttr.in style)
	if _, exists := c.GetQuery("0"); exists {
		params.Days = 0 // Current only
	} else if _, exists := c.GetQuery("1"); exists {
		params.Days = 1 // Current + today
	} else if _, exists := c.GetQuery("2"); exists {
		params.Days = 2 // Current + today + tomorrow
	}

	// Style options - check if parameter exists (wttr.in style)
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

	return params
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
