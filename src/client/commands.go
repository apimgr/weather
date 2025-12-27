package client

import (
	"flag"
	"fmt"
)

// handleCurrentCommand handles the current weather command
func handleCurrentCommand(config *Config, args []string) error {
	flagSet := flag.NewFlagSet("current", flag.ContinueOnError)
	lat := flagSet.Float64("lat", 0, "Latitude")
	lon := flagSet.Float64("lon", 0, "Longitude")
	zip := flagSet.String("zip", "", "ZIP code")
	location := flagSet.String("location", "", "Location name")

	if err := flagSet.Parse(args); err != nil {
		return NewUsageError(err.Error())
	}

	// Build query parameters
	var path string
	if *zip != "" {
		path = fmt.Sprintf("/api/v1/weather?zip=%s", *zip)
	} else if *location != "" {
		path = fmt.Sprintf("/api/v1/weather?location=%s", *location)
	} else if *lat != 0 && *lon != 0 {
		path = fmt.Sprintf("/api/v1/weather?lat=%f&lon=%f", *lat, *lon)
	} else {
		return NewUsageError("must specify --lat/--lon, --zip, or --location")
	}

	// Make API request
	client := NewHTTPClient(config)
	var result map[string]interface{}
	if err := client.GetJSON(path, &result); err != nil {
		return err
	}

	// Format and print output
	formatter := NewFormatter(config.Output, config.NoColor)
	fmt.Println(formatter.FormatWeatherCurrent(result))

	return nil
}

// handleForecastCommand handles the forecast command
func handleForecastCommand(config *Config, args []string) error {
	flagSet := flag.NewFlagSet("forecast", flag.ContinueOnError)
	lat := flagSet.Float64("lat", 0, "Latitude")
	lon := flagSet.Float64("lon", 0, "Longitude")
	zip := flagSet.String("zip", "", "ZIP code")
	location := flagSet.String("location", "", "Location name")
	days := flagSet.Int("days", 7, "Number of days (default: 7)")

	if err := flagSet.Parse(args); err != nil {
		return NewUsageError(err.Error())
	}

	// Build query parameters
	var path string
	if *zip != "" {
		path = fmt.Sprintf("/api/v1/weather/forecast?zip=%s&days=%d", *zip, *days)
	} else if *location != "" {
		path = fmt.Sprintf("/api/v1/weather/forecast?location=%s&days=%d", *location, *days)
	} else if *lat != 0 && *lon != 0 {
		path = fmt.Sprintf("/api/v1/weather/forecast?lat=%f&lon=%f&days=%d", *lat, *lon, *days)
	} else {
		return NewUsageError("must specify --lat/--lon, --zip, or --location")
	}

	// Make API request
	client := NewHTTPClient(config)
	var result map[string]interface{}
	if err := client.GetJSON(path, &result); err != nil {
		return err
	}

	// Format and print output
	formatter := NewFormatter(config.Output, config.NoColor)
	fmt.Println(formatter.FormatForecast(result))

	return nil
}

// handleAlertsCommand handles the alerts command
func handleAlertsCommand(config *Config, args []string) error {
	flagSet := flag.NewFlagSet("alerts", flag.ContinueOnError)
	lat := flagSet.Float64("lat", 0, "Latitude")
	lon := flagSet.Float64("lon", 0, "Longitude")
	zip := flagSet.String("zip", "", "ZIP code")
	location := flagSet.String("location", "", "Location name")

	if err := flagSet.Parse(args); err != nil {
		return NewUsageError(err.Error())
	}

	// Build query parameters
	var path string
	if *zip != "" {
		path = fmt.Sprintf("/api/v1/weather/alerts?zip=%s", *zip)
	} else if *location != "" {
		path = fmt.Sprintf("/api/v1/weather/alerts?location=%s", *location)
	} else if *lat != 0 && *lon != 0 {
		path = fmt.Sprintf("/api/v1/weather/alerts?lat=%f&lon=%f", *lat, *lon)
	} else {
		return NewUsageError("must specify --lat/--lon, --zip, or --location")
	}

	// Make API request
	client := NewHTTPClient(config)
	var result map[string]interface{}
	if err := client.GetJSON(path, &result); err != nil {
		return err
	}

	// Format and print output
	formatter := NewFormatter(config.Output, config.NoColor)
	fmt.Println(formatter.FormatAlerts(result))

	return nil
}

// handleMoonCommand handles the moon phase command
func handleMoonCommand(config *Config, args []string) error {
	flagSet := flag.NewFlagSet("moon", flag.ContinueOnError)
	date := flagSet.String("date", "", "Date (YYYY-MM-DD, default: today)")

	if err := flagSet.Parse(args); err != nil {
		return NewUsageError(err.Error())
	}

	// Build query parameters
	path := "/api/v1/weather/moon"
	if *date != "" {
		path = fmt.Sprintf("%s?date=%s", path, *date)
	}

	// Make API request
	client := NewHTTPClient(config)
	var result map[string]interface{}
	if err := client.GetJSON(path, &result); err != nil {
		return err
	}

	// Format and print output
	formatter := NewFormatter(config.Output, config.NoColor)
	fmt.Println(formatter.FormatMoon(result))

	return nil
}

// handleHistoryCommand handles the historical weather command
func handleHistoryCommand(config *Config, args []string) error {
	flagSet := flag.NewFlagSet("history", flag.ContinueOnError)
	lat := flagSet.Float64("lat", 0, "Latitude")
	lon := flagSet.Float64("lon", 0, "Longitude")
	zip := flagSet.String("zip", "", "ZIP code")
	location := flagSet.String("location", "", "Location name")
	date := flagSet.String("date", "", "Date (YYYY-MM-DD)")

	if err := flagSet.Parse(args); err != nil {
		return NewUsageError(err.Error())
	}

	if *date == "" {
		return NewUsageError("--date is required for history command")
	}

	// Build query parameters
	var path string
	if *zip != "" {
		path = fmt.Sprintf("/api/v1/weather/history?zip=%s&date=%s", *zip, *date)
	} else if *location != "" {
		path = fmt.Sprintf("/api/v1/weather/history?location=%s&date=%s", *location, *date)
	} else if *lat != 0 && *lon != 0 {
		path = fmt.Sprintf("/api/v1/weather/history?lat=%f&lon=%f&date=%s", *lat, *lon, *date)
	} else {
		return NewUsageError("must specify --lat/--lon, --zip, or --location")
	}

	// Make API request
	client := NewHTTPClient(config)
	var result map[string]interface{}
	if err := client.GetJSON(path, &result); err != nil {
		return err
	}

	// Format and print output (reuse current weather formatter for historical data)
	formatter := NewFormatter(config.Output, config.NoColor)
	fmt.Println(formatter.FormatWeatherCurrent(result))

	return nil
}
