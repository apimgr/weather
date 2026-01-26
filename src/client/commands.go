package main

import (
	"flag"
	"fmt"
)

// handleCurrentCommand handles the current weather command
func handleCurrentCommand(config *CLIConfig, args []string) error {
	flagSet := flag.NewFlagSet("current", flag.ContinueOnError)
	lat := flagSet.Float64("lat", 0, "Latitude")
	lon := flagSet.Float64("lon", 0, "Longitude")
	zip := flagSet.String("zip", "", "ZIP code")
	location := flagSet.String("location", "", "Location name")

	if err := flagSet.Parse(args); err != nil {
		return NewUsageError(err.Error())
	}

	// Build query parameters
	apiPath := config.GetAPIPath()
	var path string
	if *zip != "" {
		path = fmt.Sprintf("%s/weather?zip=%s", apiPath, *zip)
	} else if *location != "" {
		path = fmt.Sprintf("%s/weather?location=%s", apiPath, *location)
	} else if *lat != 0 && *lon != 0 {
		path = fmt.Sprintf("%s/weather?lat=%f&lon=%f", apiPath, *lat, *lon)
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
func handleForecastCommand(config *CLIConfig, args []string) error {
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
	apiPath := config.GetAPIPath()
	var path string
	if *zip != "" {
		path = fmt.Sprintf("%s/weather/forecast?zip=%s&days=%d", apiPath, *zip, *days)
	} else if *location != "" {
		path = fmt.Sprintf("%s/weather/forecast?location=%s&days=%d", apiPath, *location, *days)
	} else if *lat != 0 && *lon != 0 {
		path = fmt.Sprintf("%s/weather/forecast?lat=%f&lon=%f&days=%d", apiPath, *lat, *lon, *days)
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
func handleAlertsCommand(config *CLIConfig, args []string) error {
	flagSet := flag.NewFlagSet("alerts", flag.ContinueOnError)
	lat := flagSet.Float64("lat", 0, "Latitude")
	lon := flagSet.Float64("lon", 0, "Longitude")
	zip := flagSet.String("zip", "", "ZIP code")
	location := flagSet.String("location", "", "Location name")

	if err := flagSet.Parse(args); err != nil {
		return NewUsageError(err.Error())
	}

	// Build query parameters
	apiPath := config.GetAPIPath()
	var path string
	if *zip != "" {
		path = fmt.Sprintf("%s/weather/alerts?zip=%s", apiPath, *zip)
	} else if *location != "" {
		path = fmt.Sprintf("%s/weather/alerts?location=%s", apiPath, *location)
	} else if *lat != 0 && *lon != 0 {
		path = fmt.Sprintf("%s/weather/alerts?lat=%f&lon=%f", apiPath, *lat, *lon)
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
func handleMoonCommand(config *CLIConfig, args []string) error {
	flagSet := flag.NewFlagSet("moon", flag.ContinueOnError)
	date := flagSet.String("date", "", "Date (YYYY-MM-DD, default: today)")

	if err := flagSet.Parse(args); err != nil {
		return NewUsageError(err.Error())
	}

	// Build query parameters
	apiPath := config.GetAPIPath()
	path := fmt.Sprintf("%s/weather/moon", apiPath)
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
func handleHistoryCommand(config *CLIConfig, args []string) error {
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
	apiPath := config.GetAPIPath()
	var path string
	if *zip != "" {
		path = fmt.Sprintf("%s/weather/history?zip=%s&date=%s", apiPath, *zip, *date)
	} else if *location != "" {
		path = fmt.Sprintf("%s/weather/history?location=%s&date=%s", apiPath, *location, *date)
	} else if *lat != 0 && *lon != 0 {
		path = fmt.Sprintf("%s/weather/history?lat=%f&lon=%f&date=%s", apiPath, *lat, *lon, *date)
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
