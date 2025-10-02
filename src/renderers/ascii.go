package renderers

import (
	"fmt"
	"math"
	"strings"
	"time"

	"weather-go/src/utils"
)

// ASCIIRenderer handles full wttr.in-style ASCII art weather display
type ASCIIRenderer struct {
	width    int
	noColors bool
}

// NewASCIIRenderer creates a new ASCII renderer
func NewASCIIRenderer() *ASCIIRenderer {
	return &ASCIIRenderer{
		width: 120,
	}
}

// RenderFull renders the full wttr.in style weather display
func (r *ASCIIRenderer) RenderFull(weather *utils.WeatherData, params utils.RenderParams) string {
	// Set noColors flag for this render
	r.noColors = params.NoColors

	var lines []string

	// Header (skip if quiet mode)
	if !params.Quiet {
		lines = append(lines, r.renderHeader(weather.Location))
		lines = append(lines, "")
	}

	// Current weather with ASCII art
	lines = append(lines, r.renderCurrentWeatherArt(weather.Current, params)...)

	// Add forecast based on format
	if params.Days > 0 {
		lines = append(lines, "")
		lines = append(lines, "") // Extra spacing before forecast table

		// Render forecast table with specified number of days
		forecastDays := weather.Forecast
		if len(forecastDays) > params.Days {
			forecastDays = forecastDays[:params.Days]
		}
		lines = append(lines, r.renderForecastTable(forecastDays, params)...)
	} else if params.Days == -1 {
		// Default: show full forecast
		lines = append(lines, "")
		lines = append(lines, "") // Extra spacing before forecast table
		lines = append(lines, r.renderForecastTable(weather.Forecast, params)...)
	}

	if !params.NoFooter {
		lines = append(lines, "")
		lines = append(lines, r.renderFooter(weather.Location))
	}

	// Add two newlines at the end for proper terminal spacing
	return strings.Join(lines, "\n") + "\n\n"
}

// renderHeader renders the weather report header
func (r *ASCIIRenderer) renderHeader(location utils.LocationData) string {
	fullLocation := r.getFullLocationName(location)
	header := fmt.Sprintf("Weather report: %s", r.capitalizeLocation(fullLocation))
	return r.colorize(header, "#f1fa8c", true) // Yellow, bold
}

// capitalizeLocation capitalizes first letter of each location part
func (r *ASCIIRenderer) capitalizeLocation(location string) string {
	// Split by comma to get individual parts
	parts := strings.Split(location, ",")
	for i, part := range parts {
		part = strings.TrimSpace(part)

		// If it's a 2-letter country/state code, capitalize entirely
		if len(part) == 2 && strings.ToUpper(part) == part {
			parts[i] = strings.ToUpper(part)
		} else {
			// Capitalize first letter of each word
			words := strings.Fields(part)
			for j, word := range words {
				if len(word) > 0 {
					// Check if it's a short abbreviation (2-3 letters all caps)
					if len(word) <= 3 && strings.ToUpper(word) == word {
						words[j] = strings.ToUpper(word)
					} else {
						words[j] = strings.ToUpper(string(word[0])) + strings.ToLower(word[1:])
					}
				}
			}
			parts[i] = strings.Join(words, " ")
		}
	}
	return strings.Join(parts, ", ")
}

// getFullLocationName gets the full location display name with coordinates
func (r *ASCIIRenderer) getFullLocationName(location utils.LocationData) string {
	var baseName string

	// Use the pre-built full name from geocoding API if available
	if location.FullName != "" {
		baseName = location.FullName
	} else if location.ShortName != "" {
		baseName = location.ShortName
	} else {
		// Fallback: build basic name
		baseName = fmt.Sprintf("%s, %s", location.Name, location.Country)
	}

	// Append coordinates like wttr.in does
	return fmt.Sprintf("%s [%.7f,%.7f]", baseName, location.Latitude, location.Longitude)
}

// renderCurrentWeatherArt renders the current weather with ASCII art
func (r *ASCIIRenderer) renderCurrentWeatherArt(current utils.CurrentData, params utils.RenderParams) []string {
	temp := int(math.Round(current.Temperature))
	feelsLike := int(math.Round(current.FeelsLike))
	tempUnit := getTemperatureUnit(params.Units)
	speedUnit := getSpeedUnit(params.Units)
	precipUnit := getPrecipitationUnit(params.Units)
	windSpeed := int(math.Round(current.WindSpeed))
	description := current.Condition
	precipitation := fmt.Sprintf("%.1f", current.Precipitation)

	art := r.getColoredWeatherArt(current.WeatherCode, params.NoColors)
	windDir := getWindArrow(current.WindDirection)

	// Determine if it's day or night based on the icon (if available)
	_ = !strings.Contains(current.Icon, "🌙") // isDay variable for future use

	// Create current weather display with Dracula colors
	lines := []string{
		fmt.Sprintf("%s     %s", art[0], r.colorize(description, "#8be9fd", false)),
		fmt.Sprintf("%s     %s", art[1], r.colorize(fmt.Sprintf("+%d(%d) %s", temp, feelsLike, tempUnit), "#f1fa8c", false)),
		fmt.Sprintf("%s     %s", art[2], r.colorize(fmt.Sprintf("%s %d %s", windDir, windSpeed, speedUnit), "#50fa7b", false)),
		fmt.Sprintf("%s     %s", art[3], r.colorize(fmt.Sprintf("%d %s", getPressure(current.Pressure, params.Units), getPressureUnit(params.Units)), "#bd93f9", false)),
		fmt.Sprintf("%s     %s", art[4], r.colorize(fmt.Sprintf("%s %s", precipitation, precipUnit), "#ff79c6", false)),
	}

	return lines
}

// renderForecastTable renders the forecast table
func (r *ASCIIRenderer) renderForecastTable(forecast []utils.ForecastData, params utils.RenderParams) []string {
	var lines []string

	days := forecast
	if len(days) > 3 {
		days = days[:3]
	}

	if len(days) == 0 {
		return []string{"No forecast data available"}
	}

	// Get first day for header
	firstDate, _ := time.Parse("2006-01-02", days[0].Date)
	dayHeader := firstDate.Format("Mon 2 Jan")

	// Calculate dynamic spacing based on date length for perfect alignment
	_ = 123                // baseTableWidth: Base table without day header
	dayBoxPadding := (120 - 13) / 2

	// Top day header box - mathematically centered
	lines = append(lines, strings.Repeat(" ", dayBoxPadding)+r.colorize("┌─────────────┐", "#bd93f9", false)+strings.Repeat(" ", 120-dayBoxPadding-13))

	// Date header row with box drawing
	// Simplified: just center the day header without complex spacing
	lines = append(lines,
		r.colorize("┌──────────────────────────────┬───────────────────────┤", "#bd93f9", false)+
			" "+r.colorize(dayHeader, "#ff79c6", false)+" "+
			r.colorize("├───────────────────────┬──────────────────────────────┐", "#bd93f9", false))

	// Time period headers
	lines = append(lines,
		r.colorize("│", "#bd93f9", false)+r.colorize("            Morning           ", "#ffb86c", false)+
			r.colorize("│", "#bd93f9", false)+r.colorize("             Noon      ", "#ffb86c", false)+
			r.colorize("└──────┬──────┘", "#bd93f9", false)+r.colorize("     Evening           ", "#ffb86c", false)+
			r.colorize("│", "#bd93f9", false)+r.colorize("             Night            ", "#ffb86c", false)+
			r.colorize("│", "#bd93f9", false))

	// Data separator
	lines = append(lines, r.colorize("├──────────────────────────────┼──────────────────────────────┼──────────────────────────────┼──────────────────────────────┤", "#bd93f9", false))

	// Add each day's forecast
	for i, day := range days {
		periods := r.generateDayPeriods(day, params)

		// Each period gets 5 lines of ASCII art + data with colors
		for lineIndex := 0; lineIndex < 5; lineIndex++ {
			mornLine := padToWidth(periods.morning[lineIndex], 30)
			noonLine := padToWidth(periods.noon[lineIndex], 30)
			evenLine := padToWidth(periods.evening[lineIndex], 30)
			nightLine := padToWidth(periods.night[lineIndex], 30)

			line := r.colorize("│", "#bd93f9", false) + mornLine +
				r.colorize("│", "#bd93f9", false) + noonLine +
				r.colorize("│", "#bd93f9", false) + evenLine +
				r.colorize("│", "#bd93f9", false) + nightLine +
				r.colorize("│", "#bd93f9", false)
			lines = append(lines, line)
		}

		// Add separator between days (except last)
		if i < len(days)-1 {
			lines = append(lines, r.colorize("├──────────────────────────────┼──────────────────────────────┼──────────────────────────────┼──────────────────────────────┤", "#bd93f9", false))
			lines = append(lines, "") // Add blank line between days for better readability
		}
	}

	lines = append(lines, r.colorize("└──────────────────────────────┴──────────────────────────────┴──────────────────────────────┴──────────────────────────────┘", "#bd93f9", false))

	return lines
}

// generateDayPeriods generates forecast data for different periods of the day
func (r *ASCIIRenderer) generateDayPeriods(day utils.ForecastData, params utils.RenderParams) struct {
	morning []string
	noon    []string
	evening []string
	night   []string
} {
	tempUnit := getTemperatureUnit(params.Units)
	speedUnit := getSpeedUnit(params.Units)
	precipUnit := getPrecipitationUnit(params.Units)

	// Generate temperatures for different periods
	tempLow := int(math.Round(day.TempMin))
	tempHigh := int(math.Round(day.TempMax))
	tempMorn := int(math.Round(float64(tempLow) + float64(tempHigh-tempLow)*0.3))
	tempNoon := tempHigh
	tempEven := int(math.Round(float64(tempLow) + float64(tempHigh-tempLow)*0.7))
	tempNight := tempLow

	windSpeed := int(math.Round(day.WindSpeed))
	windDir := getWindArrow(day.WindDirection)
	precipitation := fmt.Sprintf("%.1f", day.Precipitation)
	precipProb := 0 // Default if not available

	description := day.Condition
	// Don't truncate - let centerInWidth handle it

	// Create properly aligned content - each line exactly fits the column
	createPeriodLines := func(temp, feels int, windMod float64) []string {
		tempLine := fmt.Sprintf("+%d(%d) %s", temp, feels, tempUnit)
		windLine := fmt.Sprintf("%s %d %s", windDir, int(math.Max(1, float64(windSpeed)*windMod)), speedUnit)
		precipLine := fmt.Sprintf("%s %s | %d%%", precipitation, precipUnit, precipProb)

		return []string{
			centerInWidth(r.colorize(description, "#8be9fd", false), 30),
			centerInWidth(r.colorize(tempLine, "#f1fa8c", false), 30),
			centerInWidth(r.colorize(windLine, "#50fa7b", false), 30),
			centerInWidth(r.colorize("6 mi", "#bd93f9", false), 30),
			centerInWidth(r.colorize(precipLine, "#ff79c6", false), 30),
		}
	}

	return struct {
		morning []string
		noon    []string
		evening []string
		night   []string
	}{
		morning: createPeriodLines(tempMorn, tempMorn+2, 0.8),
		noon:    createPeriodLines(tempNoon, tempNoon+3, 1.0),
		evening: createPeriodLines(tempEven, tempEven+2, 1.2),
		night:   createPeriodLines(tempNight, tempNight+1, 0.6),
	}
}

// getColoredWeatherArt returns colored weather ASCII art
func (r *ASCIIRenderer) getColoredWeatherArt(weatherCode int, noColors bool) []string {
	baseArt := r.getWeatherArt(weatherCode, true) // isDay is determined from icon
	artColor := r.getWeatherColor(weatherCode, true)

	if noColors {
		return baseArt
	}

	colored := make([]string, len(baseArt))
	for i, line := range baseArt {
		colored[i] = r.colorize(line, artColor, false)
	}
	return colored
}

// getWeatherColor returns the Dracula theme color for a weather condition
func (r *ASCIIRenderer) getWeatherColor(weatherCode int, isDay bool) string {
	// Dracula theme colors for different weather conditions
	if weatherCode == 0 {
		if isDay {
			return "#f1fa8c" // Clear: yellow
		}
		return "#bd93f9" // Clear night: purple
	}
	if weatherCode <= 3 {
		return "#6272a4" // Cloudy: comment gray
	}
	if weatherCode >= 45 && weatherCode <= 48 {
		return "#f8f8f2" // Fog: foreground
	}
	if weatherCode >= 51 && weatherCode <= 67 {
		return "#8be9fd" // Rain: cyan
	}
	if weatherCode >= 71 && weatherCode <= 86 {
		return "#f8f8f2" // Snow: white
	}
	if weatherCode >= 95 {
		return "#ff5555" // Storms: red
	}

	return "#f8f8f2" // Default: foreground
}

// getWeatherArt returns ASCII art for a weather condition
func (r *ASCIIRenderer) getWeatherArt(weatherCode int, isDay bool) []string {
	// Clear sky
	if weatherCode == 0 {
		if isDay {
			return []string{
				"    \\   /    ",
				"     .-.     ",
				"  ― (   ) ―  ",
				"     `-'     ",
				"    /   \\    ",
			}
		}
		return []string{
			"             ",
			"     .--.    ",
			"  .-(    ).  ",
			" (___.__)_)  ",
			"             ",
		}
	}

	// Mainly clear
	if weatherCode == 1 {
		return []string{
			"   \\  /      ",
			" _ /\"\".-.    ",
			"   \\_(   ).  ",
			"   /(_(__)   ",
			"             ",
		}
	}

	// Partly cloudy / Overcast
	if weatherCode <= 3 {
		return []string{
			"             ",
			"     .--.    ",
			"  .-(    ).  ",
			" (___.__)_)  ",
			"             ",
		}
	}

	// Rain
	if weatherCode >= 61 && weatherCode <= 67 {
		return []string{
			"     .-.     ",
			"    (   ).   ",
			"   (___(__)  ",
			"    ‚'‚'‚'‚' ",
			"    ‚'‚'‚'‚' ",
		}
	}

	// Default to cloudy
	return []string{
		"             ",
		"     .--.    ",
		"  .-(    ).  ",
		" (___.__)_)  ",
		"             ",
	}
}

// renderFooter renders the footer with attribution
func (r *ASCIIRenderer) renderFooter(location utils.LocationData) string {
	footer := r.colorize("Weather • Free weather data from Open-Meteo.com", "#6272a4", false)
	return footer
}

// Helper functions

// colorize applies ANSI color codes to text (method on renderer to access noColors)
func (r *ASCIIRenderer) colorize(text, hexColor string, bold bool) string {
	return colorizeWithFlag(text, hexColor, bold, r.noColors)
}

// colorizeWithFlag is a package-level function for rendering without renderer context
func colorizeWithFlag(text, hexColor string, bold bool, noColors bool) string {
	// If colors are disabled, return plain text
	if noColors {
		return text
	}

	// Convert hex to RGB
	red, green, blue := hexToRGB(hexColor)

	// ANSI escape code for 24-bit color
	colorCode := fmt.Sprintf("\033[38;2;%d;%d;%dm", red, green, blue)
	resetCode := "\033[0m"

	if bold {
		colorCode = "\033[1m" + colorCode
	}

	return colorCode + text + resetCode
}

// colorize is the original package function for backward compatibility
func colorize(text, hexColor string, bold bool) string {
	return colorizeWithFlag(text, hexColor, bold, false)
}

// hexToRGB converts hex color to RGB
func hexToRGB(hex string) (int, int, int) {
	// Remove # if present
	hex = strings.TrimPrefix(hex, "#")

	var r, g, b int
	fmt.Sscanf(hex, "%02x%02x%02x", &r, &g, &b)
	return r, g, b
}

// stripAnsiCodes removes ANSI color codes from text
func stripAnsiCodes(text string) string {
	// Simple regex-like replacement for ANSI codes
	result := ""
	inEscape := false
	for _, ch := range text {
		if ch == '\033' {
			inEscape = true
			continue
		}
		if inEscape {
			if ch == 'm' {
				inEscape = false
			}
			continue
		}
		result += string(ch)
	}
	return result
}

// padToWidth pads text to a specific width (accounting for ANSI codes)
func padToWidth(text string, width int) string {
	cleanText := stripAnsiCodes(text)
	actualLength := len(cleanText)
	padding := width - actualLength
	if padding < 0 {
		padding = 0
	}
	return text + strings.Repeat(" ", padding)
}

// centerInWidth centers text within a specific width (accounting for ANSI codes)
func centerInWidth(text string, width int) string {
	cleanText := stripAnsiCodes(text)
	textLength := len(cleanText)

	// If text is too long, truncate it
	if textLength > width {
		// Find the color codes in the original text
		hasColors := strings.Contains(text, "\x1b[")
		if hasColors {
			// Extract color prefix and reset code
			colorStart := text[:strings.Index(text, "m")+1]
			resetCode := "\x1b[0m"
			truncated := cleanText[:width]
			return colorStart + truncated + resetCode
		}
		return cleanText[:width]
	}

	totalPadding := width - textLength
	leftPadding := totalPadding / 2
	rightPadding := totalPadding - leftPadding

	return strings.Repeat(" ", leftPadding) + text + strings.Repeat(" ", rightPadding)
}

// getWindArrow returns wind direction arrow
func getWindArrow(degrees int) string {
	arrows := []string{"↓", "↙", "←", "↖", "↑", "↗", "→", "↘"}
	index := int(math.Round(float64(degrees)/45.0)) % 8
	return arrows[index]
}

// Unit conversion helpers

func getTemperatureUnit(units string) string {
	if units == "metric" || units == "M" {
		return "°C"
	}
	return "°F"
}

func getSpeedUnit(units string) string {
	if units == "M" {
		return "m/s"
	}
	if units == "metric" {
		return "km/h"
	}
	return "mph"
}

func getPrecipitationUnit(units string) string {
	if units == "metric" || units == "M" {
		return "mm"
	}
	return "in"
}

func getPressureUnit(units string) string {
	if units == "imperial" {
		return "inHg"
	}
	return "hPa"
}

func getPressure(pressure float64, units string) int {
	if units == "imperial" {
		return int(math.Round(pressure * 0.02953))
	}
	return int(math.Round(pressure))
}
