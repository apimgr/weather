package services

import (
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/oschwald/geoip2-golang"
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
	mu           sync.RWMutex
	enabled      bool
	ipv4Path     string
	ipv6Path     string
	ipv4URL      string
	ipv6URL      string
	cache        map[string]*GeoIPData
	cacheTime    map[string]int64
}

// NewGeoIPService creates a new GeoIP service
// Downloads DB-IP City databases (separate IPv4 and IPv6) on first run, updates weekly
func NewGeoIPService(configDir string) *GeoIPService {
	geoipDir := filepath.Join(configDir, "geoip")
	ipv4Path := filepath.Join(geoipDir, "dbip-city-ipv4.mmdb")
	ipv6Path := filepath.Join(geoipDir, "dbip-city-ipv6.mmdb")

	gs := &GeoIPService{
		ipv4URL:   "https://cdn.jsdelivr.net/npm/@ip-location-db/dbip-city-mmdb/dbip-city-ipv4.mmdb",
		ipv6URL:   "https://cdn.jsdelivr.net/npm/@ip-location-db/dbip-city-mmdb/dbip-city-ipv6.mmdb",
		ipv4Path:  ipv4Path,
		ipv6Path:  ipv6Path,
		cache:     make(map[string]*GeoIPData),
		cacheTime: make(map[string]int64),
	}

	// Try to load databases in background
	go gs.loadDatabases()

	return gs
}

// loadDatabases downloads and prepares both IPv4 and IPv6 GeoIP databases
func (gs *GeoIPService) loadDatabases() {
	fmt.Println("🌍 Loading GeoIP databases...")

	// Create directory if needed
	dir := filepath.Dir(gs.ipv4Path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		fmt.Printf("❌ Failed to create directory: %v\n", err)
		return
	}

	// Check if both databases exist
	ipv4Exists := false
	ipv6Exists := false

	if _, err := os.Stat(gs.ipv4Path); err == nil {
		ipv4Exists = true
	}
	if _, err := os.Stat(gs.ipv6Path); err == nil {
		ipv6Exists = true
	}

	if ipv4Exists && ipv6Exists {
		gs.mu.Lock()
		gs.enabled = true
		gs.mu.Unlock()
		fmt.Println("📂 GeoIP databases loaded from cache (IPv4 + IPv6)")
		return
	}

	// Download IPv4 database
	if !ipv4Exists {
		fmt.Printf("📥 Downloading IPv4 GeoIP database...\n")
		if err := gs.downloadDatabase(gs.ipv4URL, gs.ipv4Path); err != nil {
			fmt.Printf("❌ Failed to download IPv4 database: %v\n", err)
			return
		}
		fmt.Println("✅ IPv4 GeoIP database downloaded")
	}

	// Download IPv6 database
	if !ipv6Exists {
		fmt.Printf("📥 Downloading IPv6 GeoIP database...\n")
		if err := gs.downloadDatabase(gs.ipv6URL, gs.ipv6Path); err != nil {
			fmt.Printf("❌ Failed to download IPv6 database: %v\n", err)
			return
		}
		fmt.Println("✅ IPv6 GeoIP database downloaded")
	}

	gs.mu.Lock()
	gs.enabled = true
	gs.mu.Unlock()

	fmt.Println("✅ GeoIP databases ready (IPv4 + IPv6 from DB-IP)")
}

// downloadDatabase downloads a database file
func (gs *GeoIPService) downloadDatabase(url, path string) error {
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("download returned status %d", resp.StatusCode)
	}

	out, err := os.Create(path)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, resp.Body)
	return err
}

// LookupIP performs IP geolocation lookup using DB-IP City databases
func (gs *GeoIPService) LookupIP(ip string) (*GeoIPData, error) {
	// Check if databases are loaded
	gs.mu.RLock()
	enabled := gs.enabled
	gs.mu.RUnlock()

	if !enabled {
		return nil, fmt.Errorf("GeoIP databases not loaded yet")
	}

	// Check cache first
	gs.mu.RLock()
	if cached, ok := gs.cache[ip]; ok {
		gs.mu.RUnlock()
		return cached, nil
	}
	gs.mu.RUnlock()

	// Parse IP address
	parsedIP := net.ParseIP(ip)
	if parsedIP == nil {
		return nil, fmt.Errorf("invalid IP address: %s", ip)
	}

	// Check if local/private IP
	if parsedIP.IsLoopback() || parsedIP.IsPrivate() {
		return nil, fmt.Errorf("cannot geolocate private/local IP: %s", ip)
	}

	// Determine which database to use (IPv4 or IPv6)
	dbPath := gs.ipv4Path
	if strings.Contains(ip, ":") {
		// IPv6 address
		dbPath = gs.ipv6Path
	}

	// Open and query appropriate database
	reader, err := geoip2.Open(dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open GeoIP database: %w", err)
	}
	defer reader.Close()

	// Lookup city data
	record, err := reader.City(parsedIP)
	if err != nil {
		return nil, fmt.Errorf("IP lookup failed: %w", err)
	}

	// Extract city name (prefer English)
	city := ""
	if len(record.City.Names) > 0 {
		if name, ok := record.City.Names["en"]; ok {
			city = name
		} else {
			// Fallback to first available name
			for _, name := range record.City.Names {
				city = name
				break
			}
		}
	}

	// Extract region/state (prefer English)
	region := ""
	if len(record.Subdivisions) > 0 {
		if name, ok := record.Subdivisions[0].Names["en"]; ok {
			region = name
		}
	}

	// Extract country (prefer English)
	country := ""
	countryCode := ""
	if len(record.Country.Names) > 0 {
		if name, ok := record.Country.Names["en"]; ok {
			country = name
		}
		countryCode = record.Country.IsoCode
	}

	// Extract timezone
	timezone := record.Location.TimeZone
	if timezone == "" {
		timezone = "UTC"
	}

	geoData := &GeoIPData{
		IP:          ip,
		City:        city,
		Region:      region,
		Country:     country,
		CountryCode: countryCode,
		Latitude:    record.Location.Latitude,
		Longitude:   record.Location.Longitude,
		Timezone:    timezone,
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

// UpdateDatabase downloads the latest GeoIP databases (both IPv4 and IPv6)
func (gs *GeoIPService) UpdateDatabase() error {
	fmt.Println("📥 Updating GeoIP databases...")

	// Update IPv4 database
	if err := gs.updateSingleDatabase(gs.ipv4URL, gs.ipv4Path, "IPv4"); err != nil {
		return fmt.Errorf("failed to update IPv4 database: %w", err)
	}

	// Update IPv6 database
	if err := gs.updateSingleDatabase(gs.ipv6URL, gs.ipv6Path, "IPv6"); err != nil {
		return fmt.Errorf("failed to update IPv6 database: %w", err)
	}

	// Clear cache to force fresh lookups
	gs.mu.Lock()
	gs.cache = make(map[string]*GeoIPData)
	gs.cacheTime = make(map[string]int64)
	gs.mu.Unlock()

	fmt.Println("✅ GeoIP databases updated successfully (IPv4 + IPv6)")
	return nil
}

// updateSingleDatabase updates a single database file
func (gs *GeoIPService) updateSingleDatabase(url, path, name string) error {
	fmt.Printf("  📥 Downloading %s database...\n", name)

	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("download returned status %d", resp.StatusCode)
	}

	// Create temporary file
	tempPath := path + ".tmp"
	out, err := os.Create(tempPath)
	if err != nil {
		return err
	}
	defer out.Close()

	// Write to temp file
	_, err = io.Copy(out, resp.Body)
	if err != nil {
		os.Remove(tempPath)
		return err
	}
	out.Close()

	// Verify database is valid
	testReader, err := geoip2.Open(tempPath)
	if err != nil {
		os.Remove(tempPath)
		return fmt.Errorf("downloaded database is invalid: %w", err)
	}
	testReader.Close()

	// Replace old with new
	if err := os.Rename(tempPath, path); err != nil {
		os.Remove(tempPath)
		return err
	}

	fmt.Printf("  ✅ %s database updated\n", name)
	return nil
}

// Reload reloads the GeoIP databases
func (gs *GeoIPService) Reload() error {
	fmt.Println("🔄 Reloading GeoIP databases...")

	gs.mu.Lock()
	gs.enabled = false
	gs.mu.Unlock()

	// Load databases synchronously for reload
	gs.loadDatabases()

	gs.mu.RLock()
	enabled := gs.enabled
	gs.mu.RUnlock()

	if !enabled {
		return fmt.Errorf("GeoIP reload failed - databases not enabled after reload")
	}

	fmt.Println("✅ GeoIP databases reloaded")
	return nil
}
