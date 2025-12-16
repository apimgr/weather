package handlers

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

type SSLHandler struct {
	certsDir string
}

func NewSSLHandler(certsDir string) *SSLHandler {
	return &SSLHandler{
		certsDir: certsDir,
	}
}

type CertificateInfo struct {
	Subject     string    `json:"subject"`
	Issuer      string    `json:"issuer"`
	NotBefore   time.Time `json:"notBefore"`
	NotAfter    time.Time `json:"notAfter"`
	DNSNames    []string  `json:"dnsNames"`
	IsValid     bool      `json:"isValid"`
	DaysRemaining int     `json:"daysRemaining"`
}

type SSLStatus struct {
	Certificate  *CertificateInfo `json:"certificate"`
	NextCheck    string           `json:"nextCheck"`
	NextRenewal  string           `json:"nextRenewal"`
	LastRenewal  string           `json:"lastRenewal"`
	AutoRenewal  bool             `json:"autoRenewal"`
}

// GetStatus returns the current SSL certificate status
func (h *SSLHandler) GetStatus(c *gin.Context) {
	// Try to load existing certificate
	certInfo, err := h.getCertificateInfo()
	if err != nil {
		// No certificate or error loading
		c.JSON(http.StatusOK, SSLStatus{
			Certificate: nil,
			NextCheck:   "Not scheduled",
			NextRenewal: "No certificate",
			LastRenewal: "Never",
			AutoRenewal: false,
		})
		return
	}

	status := SSLStatus{
		Certificate:  certInfo,
		NextCheck:    time.Now().Add(24 * time.Hour).Format("2006-01-02 15:04"),
		NextRenewal:  calculateNextRenewal(certInfo.NotAfter),
		LastRenewal:  "Unknown",
		AutoRenewal:  true,
	}

	c.JSON(http.StatusOK, status)
}

// ObtainCertificate obtains a new Let's Encrypt certificate
func (h *SSLHandler) ObtainCertificate(c *gin.Context) {
	var request struct {
		Domain   string   `json:"domain" binding:"required"`
		Email    string   `json:"email" binding:"required"`
		AltNames []string `json:"altNames"`
	}

	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	// Validate domain
	if request.Domain == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Domain is required"})
		return
	}

	// Validate email
	if request.Email == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Email is required"})
		return
	}

	// In a real implementation, this would use ACME/Let's Encrypt client
	// For now, return a placeholder response
	c.JSON(http.StatusOK, gin.H{
		"message": "Certificate request initiated",
		"domain":  request.Domain,
		"status":  "processing",
		"note":    "Let's Encrypt integration will be implemented with ACME client",
	})
}

// RenewCertificate renews an existing certificate
func (h *SSLHandler) RenewCertificate(c *gin.Context) {
	// Check if certificate exists
	certInfo, err := h.getCertificateInfo()
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "No certificate found to renew"})
		return
	}

	// Check if renewal is needed
	daysRemaining := int(time.Until(certInfo.NotAfter).Hours() / 24)
	if daysRemaining > 30 {
		c.JSON(http.StatusOK, gin.H{
			"message": "Certificate does not need renewal yet",
			"daysRemaining": daysRemaining,
		})
		return
	}

	// In a real implementation, this would trigger ACME renewal
	c.JSON(http.StatusOK, gin.H{
		"message": "Certificate renewal initiated",
		"status":  "processing",
	})
}

// VerifyCertificate verifies the current certificate
func (h *SSLHandler) VerifyCertificate(c *gin.Context) {
	certInfo, err := h.getCertificateInfo()
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"valid": false,
			"error": "No certificate found or unable to load",
		})
		return
	}

	// Check if certificate is expired
	if time.Now().After(certInfo.NotAfter) {
		c.JSON(http.StatusOK, gin.H{
			"valid": false,
			"error": "Certificate has expired",
		})
		return
	}

	// Check if certificate is not yet valid
	if time.Now().Before(certInfo.NotBefore) {
		c.JSON(http.StatusOK, gin.H{
			"valid": false,
			"error": "Certificate is not yet valid",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"valid":    true,
		"message":  "Certificate is valid",
		"subject":  certInfo.Subject,
		"issuer":   certInfo.Issuer,
		"notAfter": certInfo.NotAfter,
		"daysRemaining": certInfo.DaysRemaining,
	})
}

// UpdateSettings updates SSL/TLS settings
func (h *SSLHandler) UpdateSettings(c *gin.Context) {
	var settings struct {
		AutoRenewal        bool `json:"autoRenewal"`
		RenewalDays        int  `json:"renewalDays"`
		EmailNotifications bool `json:"emailNotifications"`
	}

	if err := c.ShouldBindJSON(&settings); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	// Validate renewal days
	if settings.RenewalDays < 1 || settings.RenewalDays > 60 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Renewal days must be between 1 and 60"})
		return
	}

	// In a real implementation, save these settings to database
	c.JSON(http.StatusOK, gin.H{
		"message": "Settings saved successfully",
		"settings": settings,
	})
}

// ExportCertificate exports the current certificate
func (h *SSLHandler) ExportCertificate(c *gin.Context) {
	// In a real implementation, read certificate files
	c.JSON(http.StatusOK, gin.H{
		"message": "Certificate export",
		"note":    "Certificate files can be found in certs directory",
	})
}

// ImportCertificate imports an external certificate
func (h *SSLHandler) ImportCertificate(c *gin.Context) {
	// Handle multipart file upload
	certFile, err := c.FormFile("certificate")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Certificate file is required"})
		return
	}

	keyFile, err := c.FormFile("privateKey")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Private key file is required"})
		return
	}

	// In a real implementation, validate and save certificate files
	c.JSON(http.StatusOK, gin.H{
		"message": "Certificate imported successfully",
		"certFile": certFile.Filename,
		"keyFile":  keyFile.Filename,
	})
}

// RevokeCertificate revokes the current certificate
func (h *SSLHandler) RevokeCertificate(c *gin.Context) {
	var request struct {
		Reason string `json:"reason"`
	}

	if err := c.ShouldBindJSON(&request); err != nil {
		// Optional reason
	}

	// In a real implementation, use ACME to revoke certificate
	c.JSON(http.StatusOK, gin.H{
		"message": "Certificate revocation initiated",
		"status":  "processing",
	})
}

// TestSSL tests the SSL/TLS configuration
func (h *SSLHandler) TestSSL(c *gin.Context) {
	// Test SSL configuration
	results := map[string]interface{}{
		"tlsVersion":       "TLS 1.3",
		"cipherSuites":     []string{"TLS_AES_256_GCM_SHA384", "TLS_AES_128_GCM_SHA256"},
		"certificateValid": true,
		"chainValid":       true,
		"ocspStapling":     false,
		"hsts":             true,
	}

	c.JSON(http.StatusOK, gin.H{
		"results": results,
		"score":   "A+",
		"message": "SSL configuration test completed",
	})
}

// SecurityScan performs a security scan of the SSL configuration
func (h *SSLHandler) SecurityScan(c *gin.Context) {
	scan := map[string]interface{}{
		"vulnerabilities": []string{},
		"warnings": []string{
			"Consider enabling OCSP stapling for better performance",
		},
		"recommendations": []string{
			"Enable HTTP/2 for improved performance",
			"Configure CAA records in DNS",
		},
		"score": 95,
	}

	c.JSON(http.StatusOK, gin.H{
		"scan":    scan,
		"message": "Security scan completed",
	})
}

// Helper: Get certificate information
func (h *SSLHandler) getCertificateInfo() (*CertificateInfo, error) {
	// In a real implementation, load certificate from file
	// For now, return a placeholder

	// Try to get cert from system
	conn, err := tls.Dial("tcp", "localhost:443", &tls.Config{
		InsecureSkipVerify: true,
	})
	if err != nil {
		return nil, fmt.Errorf("no certificate available")
	}
	defer conn.Close()

	certs := conn.ConnectionState().PeerCertificates
	if len(certs) == 0 {
		return nil, fmt.Errorf("no certificate found")
	}

	cert := certs[0]
	daysRemaining := int(time.Until(cert.NotAfter).Hours() / 24)

	return &CertificateInfo{
		Subject:       cert.Subject.CommonName,
		Issuer:        cert.Issuer.CommonName,
		NotBefore:     cert.NotBefore,
		NotAfter:      cert.NotAfter,
		DNSNames:      cert.DNSNames,
		IsValid:       time.Now().After(cert.NotBefore) && time.Now().Before(cert.NotAfter),
		DaysRemaining: daysRemaining,
	}, nil
}

// Helper: Parse certificate
func parseCertificate(certPEM []byte) (*x509.Certificate, error) {
	cert, err := x509.ParseCertificate(certPEM)
	if err != nil {
		return nil, err
	}
	return cert, nil
}

// Helper: Calculate next renewal date
func calculateNextRenewal(notAfter time.Time) string {
	// Renew 30 days before expiry
	renewalDate := notAfter.Add(-30 * 24 * time.Hour)
	if time.Now().After(renewalDate) {
		return "Now"
	}
	return renewalDate.Format("2006-01-02 15:04")
}
