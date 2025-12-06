package handlers

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	"weather-go/src/services"
	"weather-go/src/utils"
)

// HurricaneHandler handles hurricane tracking requests
type HurricaneHandler struct {
	hurricaneService *services.HurricaneService
}

// NewHurricaneHandler creates a new hurricane handler
func NewHurricaneHandler(hurricaneService *services.HurricaneService) *HurricaneHandler {
	return &HurricaneHandler{
		hurricaneService: hurricaneService,
	}
}

// HandleHurricaneRequest handles hurricane tracking page requests
func (h *HurricaneHandler) HandleHurricaneRequest(c *gin.Context) {
	// Check if user wants JSON
	accept := c.GetHeader("Accept")
	wantsJSON := strings.Contains(accept, "application/json")

	// Get active storms
	data, err := h.hurricaneService.GetActiveStorms()
	if err != nil {
		if wantsJSON {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch hurricane data"})
		} else {
			c.String(http.StatusInternalServerError, "Failed to fetch hurricane data: %v", err)
		}
		return
	}

	// Return JSON if requested
	if wantsJSON {
		c.JSON(http.StatusOK, data)
		return
	}

	// Check user agent to determine if browser or console
	isBrowser := utils.IsBrowser(c)

	if isBrowser {
		// Render HTML template
		hostInfo := utils.GetHostInfo(c)
		c.HTML(http.StatusOK, "hurricane.tmpl", gin.H{
			"Title":    "Active Hurricanes & Tropical Storms",
			"Storms":   data.ActiveStorms,
			"Count":    len(data.ActiveStorms),
			"HostInfo": hostInfo,
		})
	} else {
		// Render console output
		output := h.renderConsoleOutput(data)
		c.String(http.StatusOK, output)
	}
}

// HandleHurricaneAPI handles JSON API requests for hurricane data
func (h *HurricaneHandler) HandleHurricaneAPI(c *gin.Context) {
	data, err := h.hurricaneService.GetActiveStorms()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch hurricane data"})
		return
	}

	c.JSON(http.StatusOK, data)
}

// renderConsoleOutput renders hurricane data for console/terminal
func (h *HurricaneHandler) renderConsoleOutput(data *services.HurricaneData) string {
	if len(data.ActiveStorms) == 0 {
		return "🌊 No active tropical storms or hurricanes at this time.\n\n"
	}

	output := "🌀 Active Tropical Storms & Hurricanes\n"
	output += "═══════════════════════════════════════\n\n"

	for _, storm := range data.ActiveStorms {
		icon := h.hurricaneService.GetStormIcon(storm.Classification, storm.WindSpeed)
		category := h.hurricaneService.GetStormCategory(storm.WindSpeed)

		output += icon + " " + storm.Name + "\n"
		output += "   Category: " + category + "\n"
		output += "   Wind Speed: " + formatInt(storm.WindSpeed) + " mph\n"
		output += "   Pressure: " + formatInt(storm.Pressure) + " mb\n"
		output += "   Location: " + formatFloat(storm.Latitude) + ", " + formatFloat(storm.Longitude) + "\n"

		if storm.MovementSpeed > 0 {
			output += "   Movement: " + storm.MovementDir + " at " + formatInt(storm.MovementSpeed) + " mph\n"
		}

		output += "   Last Update: " + storm.LastUpdate + "\n"

		if storm.PublicAdvisory != "" {
			output += "   Advisory: " + storm.PublicAdvisory + "\n"
		}

		output += "\n"
	}

	output += "Data from NOAA National Hurricane Center\n"
	output += "Updates every 10 minutes\n\n"

	return output
}

// Helper functions
func formatInt(val int) string {
	if val == 0 {
		return "N/A"
	}
	return formatIntToStr(val)
}

func formatFloat(val float64) string {
	if val == 0 {
		return "N/A"
	}
	return formatFloatToStr(val)
}

func formatIntToStr(val int) string {
	// Simple int to string
	if val < 0 {
		return "-" + formatIntToStr(-val)
	}
	if val < 10 {
		return string(rune('0' + val))
	}
	return formatIntToStr(val/10) + string(rune('0'+val%10))
}

func formatFloatToStr(val float64) string {
	// Simple float to string with 2 decimals
	intPart := int(val)
	fracPart := int((val - float64(intPart)) * 100)
	if fracPart < 0 {
		fracPart = -fracPart
	}
	result := formatIntToStr(intPart) + "."
	if fracPart < 10 {
		result += "0"
	}
	result += formatIntToStr(fracPart)
	return result
}
