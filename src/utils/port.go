package utils

import (
	"database/sql"
	"fmt"
	"math/rand"
	"net"
	"time"
)

const (
	// Port range for random selection (64000-64999)
	MinPort = 64000
	MaxPort = 64999
)

// PortManager handles port selection and persistence
type PortManager struct {
	db *sql.DB
}

// NewPortManager creates a new port manager
func NewPortManager(db *sql.DB) *PortManager {
	return &PortManager{db: db}
}

// IsPortAvailable checks if a port is available for binding
func IsPortAvailable(port int) bool {
	// Try to bind to the port
	addr := fmt.Sprintf(":%d", port)
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		return false
	}
	listener.Close()
	return true
}

// GetRandomAvailablePort finds a random available port in the 64000-64999 range
func GetRandomAvailablePort() (int, error) {
	// Initialize random seed
	rand.Seed(time.Now().UnixNano())

	// Try up to 100 times to find an available port
	for i := 0; i < 100; i++ {
		port := rand.Intn(MaxPort-MinPort+1) + MinPort
		if IsPortAvailable(port) {
			return port, nil
		}
	}

	return 0, fmt.Errorf("could not find available port in range %d-%d after 100 attempts", MinPort, MaxPort)
}

// GetOrAssignPort retrieves the saved port from database or assigns a new one
func (pm *PortManager) GetOrAssignPort(portType string) (int, error) {
	// Try to get saved port from database
	var savedPort int
	err := pm.db.QueryRow("SELECT value FROM settings WHERE key = ?", "server.port."+portType).Scan(&savedPort)

	if err == nil && savedPort > 0 {
		// Found saved port, check if it's still available
		if IsPortAvailable(savedPort) {
			return savedPort, nil
		}
		// Port is no longer available, fall through to assign new one
	}

	// No saved port or port not available, assign a new one
	newPort, err := GetRandomAvailablePort()
	if err != nil {
		return 0, err
	}

	// Save to database
	_, err = pm.db.Exec(`
		INSERT INTO settings (key, value, type, description, updated_at)
		VALUES (?, ?, 'integer', ?, ?)
		ON CONFLICT(key) DO UPDATE SET value = ?, updated_at = ?
	`, "server.port."+portType, fmt.Sprintf("%d", newPort),
		fmt.Sprintf("Auto-assigned %s port", portType),
		time.Now(), fmt.Sprintf("%d", newPort), time.Now())

	if err != nil {
		return 0, fmt.Errorf("failed to save port to database: %w", err)
	}

	return newPort, nil
}

// SavePort saves a port to the database
func (pm *PortManager) SavePort(portType string, port int) error {
	_, err := pm.db.Exec(`
		INSERT INTO settings (key, value, type, description, updated_at)
		VALUES (?, ?, 'integer', ?, ?)
		ON CONFLICT(key) DO UPDATE SET value = ?, updated_at = ?
	`, "server.port."+portType, fmt.Sprintf("%d", port),
		fmt.Sprintf("%s port", portType),
		time.Now(), fmt.Sprintf("%d", port), time.Now())

	return err
}

// GetSavedPort retrieves a saved port from the database
func (pm *PortManager) GetSavedPort(portType string) (int, error) {
	var savedPort int
	err := pm.db.QueryRow("SELECT value FROM settings WHERE key = ?", "server.port."+portType).Scan(&savedPort)
	if err != nil {
		return 0, err
	}
	return savedPort, nil
}
