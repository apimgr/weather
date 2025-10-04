package services

import (
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"sync"
)

// GeoIPData represents location data from GeoIP lookup
type GeoIPData struct {
	IP          string
	City        string
	Region      string
	Country     string
	CountryCode string
	Latitude    float64
	Longitude   float64
	Timezone    string
}

// GeoIPService handles IP-based geolocation
type GeoIPService struct {
	mu        sync.RWMutex
	enabled   bool
	dataPath  string
	dataURL   string
	cache     map[string]*GeoIPData
	cacheTime map[string]int64
}

// NewGeoIPService creates a new GeoIP service
func NewGeoIPService(dataURL, dataPath string) *GeoIPService {
	if dataURL == "" {
		dataURL = "https://github.com/P3TERX/GeoLite.mmdb/raw/download/GeoLite2-City.mmdb"
	}
	if dataPath == "" {
		dataPath = "/tmp/GeoLite2-City.mmdb"
	}

	gs := &GeoIPService{
		dataURL:   dataURL,
		dataPath:  dataPath,
		cache:     make(map[string]*GeoIPData),
		cacheTime: make(map[string]int64),
	}

	// Try to load database in background
	go gs.loadDatabase()

	return gs
}

// loadDatabase downloads and prepares the GeoIP database
func (gs *GeoIPService) loadDatabase() {
	fmt.Println("🌍 Loading GeoIP database...")

	// Check if database already exists and use cached version
	if _, err := os.Stat(gs.dataPath); err == nil {
		gs.mu.Lock()
		gs.enabled = true
		gs.mu.Unlock()
		fmt.Println("📂 GeoIP database loaded from cache")
		return
	}

	// Download database
	fmt.Printf("📥 Downloading GeoIP database from %s...\n", gs.dataURL)
	resp, err := http.Get(gs.dataURL)
	if err != nil {
		fmt.Printf("❌ Failed to download GeoIP database: %v\n", err)
		return
	}
	defer resp.Body.Close()

	// Create directory if needed
	dir := filepath.Dir(gs.dataPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		fmt.Printf("❌ Failed to create directory: %v\n", err)
		return
	}

	// Write to file
	out, err := os.Create(gs.dataPath)
	if err != nil {
		fmt.Printf("❌ Failed to create file: %v\n", err)
		return
	}
	defer out.Close()

	_, err = io.Copy(out, resp.Body)
	if err != nil {
		fmt.Printf("❌ Failed to write GeoIP database: %v\n", err)
		return
	}

	gs.mu.Lock()
	gs.enabled = true
	gs.mu.Unlock()

	fmt.Println("✅ GeoIP database downloaded successfully")
}

// LookupIP performs IP geolocation lookup
// Note: This is a simplified implementation using ip-api.com as fallback
// For production, you should use a proper MMDB reader library
func (gs *GeoIPService) LookupIP(ip string) (*GeoIPData, error) {
	gs.mu.RLock()
	enabled := gs.enabled
	gs.mu.RUnlock()

	if !enabled {
		return nil, fmt.Errorf("GeoIP database not loaded")
	}

	// Check cache first
	gs.mu.RLock()
	if cached, ok := gs.cache[ip]; ok {
		gs.mu.RUnlock()
		return cached, nil
	}
	gs.mu.RUnlock()

	// Parse IP to validate
	parsedIP := net.ParseIP(ip)
	if parsedIP == nil {
		return nil, fmt.Errorf("invalid IP address: %s", ip)
	}

	// Use ip-api.com as fallback (free, no key required)
	url := fmt.Sprintf("http://ip-api.com/json/%s", ip)
	resp, err := http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to lookup IP: %w", err)
	}
	defer resp.Body.Close()

	var result struct {
		Status      string  `json:"status"`
		Country     string  `json:"country"`
		CountryCode string  `json:"countryCode"`
		Region      string  `json:"regionName"`
		City        string  `json:"city"`
		Lat         float64 `json:"lat"`
		Lon         float64 `json:"lon"`
		Timezone    string  `json:"timezone"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to parse GeoIP response: %w", err)
	}

	if result.Status != "success" {
		return nil, fmt.Errorf("GeoIP lookup failed for IP: %s", ip)
	}

	geoData := &GeoIPData{
		IP:          ip,
		City:        result.City,
		Region:      result.Region,
		Country:     result.Country,
		CountryCode: result.CountryCode,
		Latitude:    result.Lat,
		Longitude:   result.Lon,
		Timezone:    result.Timezone,
	}

	// Cache result
	gs.mu.Lock()
	gs.cache[ip] = geoData
	gs.mu.Unlock()

	return geoData, nil
}

// IsEnabled returns whether GeoIP is enabled
func (gs *GeoIPService) IsEnabled() bool {
	gs.mu.RLock()
	defer gs.mu.RUnlock()
	return gs.enabled
}

// Reload reloads the GeoIP database
func (gs *GeoIPService) Reload() error {
	fmt.Println("🔄 Reloading GeoIP database...")

	gs.mu.Lock()
	gs.enabled = false
	gs.mu.Unlock()

	// Load database synchronously for reload
	gs.loadDatabase()

	gs.mu.RLock()
	enabled := gs.enabled
	gs.mu.RUnlock()

	if !enabled {
		return fmt.Errorf("GeoIP reload failed - database not enabled after reload")
	}

	return nil
}
