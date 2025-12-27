package models

import (
	"database/sql"
	"fmt"
	"strings"
	"time"
)

// Domain represents a custom domain configuration
// TEMPLATE.md PART 34: Custom domain support
type Domain struct {
	ID           int64     `json:"id"`
	Domain       string    `json:"domain"`
	UserID       *int64    `json:"user_id,omitempty"`   // Null for server default
	IsVerified   bool      `json:"is_verified"`
	IsActive     bool      `json:"is_active"`
	SSLEnabled   bool      `json:"ssl_enabled"`
	SSLCertPath  string    `json:"ssl_cert_path,omitempty"`
	SSLKeyPath   string    `json:"ssl_key_path,omitempty"`
	RedirectWWW  bool      `json:"redirect_www"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
	VerifiedAt   *time.Time `json:"verified_at,omitempty"`
}

// DomainModel handles domain-related database operations
type DomainModel struct {
	DB *sql.DB
}

// Create creates a new custom domain
func (m *DomainModel) Create(domain string, userID *int64) (*Domain, error) {
	// Normalize domain (lowercase, remove protocol)
	domain = normalizeDomain(domain)

	// Validate domain format
	if err := validateDomain(domain); err != nil {
		return nil, err
	}

	// Check if domain already exists
	var exists int
	err := m.DB.QueryRow(`
		SELECT COUNT(*) FROM custom_domains WHERE domain = ?
	`, domain).Scan(&exists)

	if err != nil {
		return nil, fmt.Errorf("failed to check domain existence: %w", err)
	}

	if exists > 0 {
		return nil, fmt.Errorf("domain already registered")
	}

	// Insert domain
	result, err := m.DB.Exec(`
		INSERT INTO custom_domains (domain, user_id, is_verified, is_active, ssl_enabled, redirect_www, created_at, updated_at)
		VALUES (?, ?, 0, 0, 0, 1, ?, ?)
	`, domain, userID, time.Now(), time.Now())

	if err != nil {
		return nil, fmt.Errorf("failed to create domain: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return nil, fmt.Errorf("failed to get domain ID: %w", err)
	}

	return m.GetByID(id)
}

// GetByID retrieves a domain by ID
func (m *DomainModel) GetByID(id int64) (*Domain, error) {
	domain := &Domain{}

	err := m.DB.QueryRow(`
		SELECT id, domain, user_id, is_verified, is_active, ssl_enabled,
		       COALESCE(ssl_cert_path, ''), COALESCE(ssl_key_path, ''),
		       redirect_www, created_at, updated_at, verified_at
		FROM custom_domains
		WHERE id = ?
	`, id).Scan(
		&domain.ID, &domain.Domain, &domain.UserID, &domain.IsVerified,
		&domain.IsActive, &domain.SSLEnabled, &domain.SSLCertPath,
		&domain.SSLKeyPath, &domain.RedirectWWW, &domain.CreatedAt,
		&domain.UpdatedAt, &domain.VerifiedAt,
	)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("domain not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get domain: %w", err)
	}

	return domain, nil
}

// GetByDomain retrieves a domain by domain name
func (m *DomainModel) GetByDomain(domainName string) (*Domain, error) {
	domain := &Domain{}

	domainName = normalizeDomain(domainName)

	err := m.DB.QueryRow(`
		SELECT id, domain, user_id, is_verified, is_active, ssl_enabled,
		       COALESCE(ssl_cert_path, ''), COALESCE(ssl_key_path, ''),
		       redirect_www, created_at, updated_at, verified_at
		FROM custom_domains
		WHERE domain = ?
	`, domainName).Scan(
		&domain.ID, &domain.Domain, &domain.UserID, &domain.IsVerified,
		&domain.IsActive, &domain.SSLEnabled, &domain.SSLCertPath,
		&domain.SSLKeyPath, &domain.RedirectWWW, &domain.CreatedAt,
		&domain.UpdatedAt, &domain.VerifiedAt,
	)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("domain not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get domain: %w", err)
	}

	return domain, nil
}

// List retrieves all domains (admin) or user's domains
func (m *DomainModel) List(userID *int64) ([]*Domain, error) {
	var rows *sql.Rows
	var err error

	if userID == nil {
		// Admin: get all domains
		rows, err = m.DB.Query(`
			SELECT id, domain, user_id, is_verified, is_active, ssl_enabled,
			       COALESCE(ssl_cert_path, ''), COALESCE(ssl_key_path, ''),
			       redirect_www, created_at, updated_at, verified_at
			FROM custom_domains
			ORDER BY created_at DESC
		`)
	} else {
		// User: get only their domains
		rows, err = m.DB.Query(`
			SELECT id, domain, user_id, is_verified, is_active, ssl_enabled,
			       COALESCE(ssl_cert_path, ''), COALESCE(ssl_key_path, ''),
			       redirect_www, created_at, updated_at, verified_at
			FROM custom_domains
			WHERE user_id = ?
			ORDER BY created_at DESC
		`, userID)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to list domains: %w", err)
	}
	defer rows.Close()

	domains := []*Domain{}
	for rows.Next() {
		domain := &Domain{}
		err := rows.Scan(
			&domain.ID, &domain.Domain, &domain.UserID, &domain.IsVerified,
			&domain.IsActive, &domain.SSLEnabled, &domain.SSLCertPath,
			&domain.SSLKeyPath, &domain.RedirectWWW, &domain.CreatedAt,
			&domain.UpdatedAt, &domain.VerifiedAt,
		)
		if err != nil {
			continue
		}
		domains = append(domains, domain)
	}

	return domains, nil
}

// Verify marks a domain as verified
func (m *DomainModel) Verify(id int64) error {
	now := time.Now()
	_, err := m.DB.Exec(`
		UPDATE custom_domains
		SET is_verified = 1, verified_at = ?, updated_at = ?
		WHERE id = ?
	`, now, now, id)

	return err
}

// Activate activates a domain (makes it live)
func (m *DomainModel) Activate(id int64) error {
	_, err := m.DB.Exec(`
		UPDATE custom_domains
		SET is_active = 1, updated_at = ?
		WHERE id = ? AND is_verified = 1
	`, time.Now(), id)

	if err != nil {
		return err
	}

	// Check if any rows were affected
	// If not, domain might not be verified
	return nil
}

// Deactivate deactivates a domain
func (m *DomainModel) Deactivate(id int64) error {
	_, err := m.DB.Exec(`
		UPDATE custom_domains
		SET is_active = 0, updated_at = ?
		WHERE id = ?
	`, time.Now(), id)

	return err
}

// UpdateSSL updates SSL configuration for a domain
func (m *DomainModel) UpdateSSL(id int64, certPath, keyPath string) error {
	_, err := m.DB.Exec(`
		UPDATE custom_domains
		SET ssl_enabled = 1, ssl_cert_path = ?, ssl_key_path = ?, updated_at = ?
		WHERE id = ?
	`, certPath, keyPath, time.Now(), id)

	return err
}

// Delete deletes a domain
func (m *DomainModel) Delete(id int64) error {
	_, err := m.DB.Exec(`DELETE FROM custom_domains WHERE id = ?`, id)
	return err
}

// Helper functions

// normalizeDomain normalizes a domain name
func normalizeDomain(domain string) string {
	// Remove http:// or https://
	domain = strings.TrimPrefix(domain, "http://")
	domain = strings.TrimPrefix(domain, "https://")

	// Remove trailing slash
	domain = strings.TrimSuffix(domain, "/")

	// Remove www. prefix if present (we'll handle this separately)
	// domain = strings.TrimPrefix(domain, "www.")

	// Lowercase
	domain = strings.ToLower(domain)

	return domain
}

// validateDomain validates domain format
func validateDomain(domain string) error {
	if domain == "" {
		return fmt.Errorf("domain cannot be empty")
	}

	// Check length
	if len(domain) > 253 {
		return fmt.Errorf("domain too long (max 253 characters)")
	}

	// Basic validation: must contain at least one dot
	if !strings.Contains(domain, ".") {
		return fmt.Errorf("invalid domain format (must be like example.com)")
	}

	// Check for invalid characters
	for _, char := range domain {
		if !((char >= 'a' && char <= 'z') || (char >= '0' && char <= '9') || char == '.' || char == '-') {
			return fmt.Errorf("invalid character in domain: %c", char)
		}
	}

	// Ensure doesn't start or end with hyphen or dot
	if strings.HasPrefix(domain, "-") || strings.HasSuffix(domain, "-") ||
		strings.HasPrefix(domain, ".") || strings.HasSuffix(domain, ".") {
		return fmt.Errorf("domain cannot start or end with hyphen or dot")
	}

	return nil
}

// GetVerificationToken generates a verification token for a domain
func (m *DomainModel) GetVerificationToken(id int64) (string, error) {
	var domain string
	err := m.DB.QueryRow(`
		SELECT domain FROM custom_domains WHERE id = ?
	`, id).Scan(&domain)

	if err != nil {
		return "", err
	}

	// Simple token: weather-verify-{domain-hash}
	// In production, use crypto/rand for secure token
	token := fmt.Sprintf("weather-verify-%d", id)

	return token, nil
}
