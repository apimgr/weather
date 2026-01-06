package client

// Version information (set by main via ldflags)
var (
	Version   = "dev"
	GitCommit = "unknown"
	BuildDate = "unknown"
)

// UserAgent returns the User-Agent string for API requests
// ALWAYS uses original project name even if binary renamed
func UserAgent() string {
	return "weather-cli/" + Version
}
