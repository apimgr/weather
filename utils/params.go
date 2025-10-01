package utils

import (
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

	// Format parameters (0-4)
	if format := c.Query("format"); format != "" {
		switch format {
		case "0":
			params.Format = 0
		case "1":
			params.Format = 1
		case "2":
			params.Format = 2
		case "3":
			params.Format = 3
		case "4":
			params.Format = 4
		}
	}

	// Unit parameters
	if c.Query("u") != "" {
		params.Units = "imperial" // USCS/Imperial
	} else if c.Query("m") != "" {
		params.Units = "metric" // Metric
	} else if c.Query("M") != "" {
		params.Units = "metric" // SI (metric with m/s for wind)
	} else if units := c.Query("units"); units != "" {
		params.Units = units
	}

	// Display options
	if c.Query("0") != "" {
		params.Days = 0 // Current only
	} else if c.Query("1") != "" {
		params.Days = 1 // Current + today
	} else if c.Query("2") != "" {
		params.Days = 2 // Current + today + tomorrow
	}

	// Style options
	if c.Query("F") != "" {
		params.NoFooter = true // Hide footer
	}
	if c.Query("n") != "" {
		params.Narrow = true // Narrow format
	}
	if c.Query("q") != "" {
		params.Quiet = true // Quiet mode
	}
	if c.Query("Q") != "" {
		params.SuperQuiet = true // Super quiet mode
	}
	if c.Query("T") != "" {
		params.NoColors = true // No terminal colors
	}
	if c.Query("A") != "" {
		params.ForceANSI = true // Force ANSI/terminal output
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
