package services

import (
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"path/filepath"
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

// GeoIPService handles IP-based geolocation using sapics/ip-location-db
type GeoIPService struct {
	mu          sync.RWMutex
	enabled     bool
	cityIPv4Path string
	cityIPv6Path string
	countryPath  string
	asnPath      string
	cache        map[string]*GeoIPData
	cacheTime    map[string]int64
}

// Database URLs from sapics/ip-location-db (SPEC Section 16)
const (
	cityIPv4URL  = "https://cdn.jsdelivr.net/npm/@ip-location-db/geolite2-city-mmdb/geolite2-city-ipv4.mmdb"
	cityIPv6URL  = "https://cdn.jsdelivr.net/npm/@ip-location-db/geolite2-city-mmdb/geolite2-city-ipv6.mmdb"
	countryURL   = "https://cdn.jsdelivr.net/npm/@ip-location-db/geo-whois-asn-country-mmdb/geo-whois-asn-country.mmdb"
	asnURL       = "https://cdn.jsdelivr.net/npm/@ip-location-db/asn-mmdb/asn.mmdb"
)

// NewGeoIPService creates a new GeoIP service
// Downloads 4 databases from sapics/ip-location-db on first run, updates weekly
func NewGeoIPService(configDir string) *GeoIPService {
	geoipDir := filepath.Join(configDir, "geoip")

	gs := &GeoIPService{
		cityIPv4Path: filepath.Join(geoipDir, "geolite2-city-ipv4.mmdb"),
		cityIPv6Path: filepath.Join(geoipDir, "geolite2-city-ipv6.mmdb"),
		countryPath:  filepath.Join(geoipDir, "geo-whois-asn-country.mmdb"),
		asnPath:      filepath.Join(geoipDir, "asn.mmdb"),
		cache:        make(map[string]*GeoIPData),
		cacheTime:    make(map[string]int64),
	}

	// Try to load databases in background
	go gs.loadDatabases()

	return gs
}

// loadDatabases downloads and prepares all 4 GeoIP databases from sapics/ip-location-db
func (gs *GeoIPService) loadDatabases() {
	fmt.Println("🌍 Loading GeoIP databases (sapics/ip-location-db)...")

	// Create directory if needed
	dir := filepath.Dir(gs.cityIPv4Path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		fmt.Printf("❌ Failed to create directory: %v\n", err)
		return
	}

	// Check if all 4 databases exist
	databases := map[string]string{
		"City IPv4":  gs.cityIPv4Path,
		"City IPv6":  gs.cityIPv6Path,
		"Country":    gs.countryPath,
		"ASN":        gs.asnPath,
	}

	allExist := true
	for name, path := range databases {
		if _, err := os.Stat(path); err != nil {
			allExist = false
			fmt.Printf("  📥 %s database not found, will download\n", name)
		}
	}

	if allExist {
		gs.mu.Lock()
		gs.enabled = true
		gs.mu.Unlock()
		fmt.Println("📂 GeoIP databases loaded from cache (4 databases)")
		return
	}

	// Download all 4 databases
	downloads := map[string]string{
		gs.cityIPv4Path: cityIPv4URL,
		gs.cityIPv6Path: cityIPv6URL,
		gs.countryPath:  countryURL,
		gs.asnPath:      asnURL,
	}

	for path, url := range downloads {
		if _, err := os.Stat(path); err == nil {
			continue // Already exists
		}

		filename := filepath.Base(path)
		fmt.Printf("📥 Downloading %s...\n", filename)

		if err := gs.downloadDatabase(url, path); err != nil {
			fmt.Printf("❌ Failed to download %s: %v\n", filename, err)
			return
		}
		fmt.Printf("✅ %s downloaded\n", filename)
	}

	gs.mu.Lock()
	gs.enabled = true
	gs.mu.Unlock()

	fmt.Println("✅ All 4 GeoIP databases ready (sapics/ip-location-db: ~103MB total)")
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

	// Determine which city database to use (IPv4 or IPv6)
	cityDBPath := gs.cityIPv4Path
	if parsedIP.To4() == nil {
		// IPv6 address
		cityDBPath = gs.cityIPv6Path
	}

	// Try city database first
	cityReader, err := geoip2.Open(cityDBPath)
	if err != nil {
		// Fallback to country database
		return gs.lookupCountry(parsedIP, ip)
	}
	defer cityReader.Close()

	// Lookup city data
	record, err := cityReader.City(parsedIP)
	if err != nil {
		// Fallback to country database
		cityReader.Close()
		return gs.lookupCountry(parsedIP, ip)
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

// lookupCountry performs country-level lookup (fallback when city lookup fails)
func (gs *GeoIPService) lookupCountry(parsedIP net.IP, ipStr string) (*GeoIPData, error) {
	// Open country database (combined IPv4/IPv6)
	countryReader, err := geoip2.Open(gs.countryPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open country database: %w", err)
	}
	defer countryReader.Close()

	// Lookup country data
	record, err := countryReader.Country(parsedIP)
	if err != nil {
		return nil, fmt.Errorf("country lookup failed: %w", err)
	}

	// Extract country info
	country := ""
	countryCode := ""
	if len(record.Country.Names) > 0 {
		if name, ok := record.Country.Names["en"]; ok {
			country = name
		}
		countryCode = record.Country.IsoCode
	}

	geoData := &GeoIPData{
		IP:          ipStr,
		City:        "", // Country-level only, no city
		Region:      "",
		Country:     country,
		CountryCode: countryCode,
		Latitude:    0, // Country database doesn't have coordinates
		Longitude:   0,
		Timezone:    "",
	}

	// Cache result
	gs.mu.Lock()
	gs.cache[ipStr] = geoData
	gs.mu.Unlock()

	return geoData, nil
}

// IsEnabled returns whether GeoIP is enabled
func (gs *GeoIPService) IsEnabled() bool {
	gs.mu.RLock()
	defer gs.mu.RUnlock()
	return gs.enabled
}

// UpdateDatabase downloads the latest GeoIP databases (all 4 from sapics)
func (gs *GeoIPService) UpdateDatabase() error {
	fmt.Println("📥 Updating GeoIP databases (sapics/ip-location-db)...")

	// Update all 4 databases
	updates := map[string]string{
		gs.cityIPv4Path: cityIPv4URL,
		gs.cityIPv6Path: cityIPv6URL,
		gs.countryPath:  countryURL,
		gs.asnPath:      asnURL,
	}

	for path, url := range updates {
		filename := filepath.Base(path)
		if err := gs.updateSingleDatabase(url, path, filename); err != nil {
			return fmt.Errorf("failed to update %s: %w", filename, err)
		}
	}

	// Clear cache to force fresh lookups
	gs.mu.Lock()
	gs.cache = make(map[string]*GeoIPData)
	gs.cacheTime = make(map[string]int64)
	gs.mu.Unlock()

	fmt.Println("✅ All 4 GeoIP databases updated successfully")
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
