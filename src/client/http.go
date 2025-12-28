package client

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// HTTPClient wraps the HTTP client with config
type HTTPClient struct {
	Config     *Config
	HTTPClient *http.Client
}

// NewHTTPClient creates a new HTTP client
func NewHTTPClient(config *Config) *HTTPClient {
	return &HTTPClient{
		Config: config,
		HTTPClient: &http.Client{
			Timeout: time.Duration(config.Timeout) * time.Second,
		},
	}
}

// Get performs a GET request
func (c *HTTPClient) Get(path string) (*http.Response, error) {
	url := c.Config.Server + path

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, NewConnectionError(fmt.Sprintf("failed to create request: %v", err))
	}

	// Set headers
	req.Header.Set("User-Agent", UserAgent())
	if c.Config.Token != "" {
		req.Header.Set("Authorization", "Bearer "+c.Config.Token)
	}

	// Perform request
	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, NewConnectionError(fmt.Sprintf("failed to connect to server: %v", err))
	}

	// Check for errors
	if resp.StatusCode == 401 || resp.StatusCode == 403 {
		resp.Body.Close()
		return nil, NewAuthError("authentication failed - check your token")
	}

	if resp.StatusCode == 404 {
		resp.Body.Close()
		return nil, NewNotFoundError("resource not found")
	}

	if resp.StatusCode >= 400 {
		defer resp.Body.Close()
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("server error (%d): %s", resp.StatusCode, string(body))
	}

	return resp, nil
}

// GetJSON performs a GET request and decodes JSON response
func (c *HTTPClient) GetJSON(path string, result interface{}) error {
	resp, err := c.Get(path)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if err := json.NewDecoder(resp.Body).Decode(result); err != nil {
		return fmt.Errorf("failed to decode response: %w", err)
	}

	return nil
}

// CheckServerVersion checks if the server version is compatible
func (c *HTTPClient) CheckServerVersion() error {
	var versionResp struct {
		Version string `json:"version"`
	}

	if err := c.GetJSON("/api/v1/version", &versionResp); err != nil {
		// If version endpoint doesn't exist, skip check
		if _, ok := err.(*ExitError); ok && err.(*ExitError).Code == ExitNotFound {
			return nil
		}
		return err
	}

	// Basic version compatibility check
	// Version comparison can be added in future if needed

	return nil
}
