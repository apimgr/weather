package handler

import (
	"database/sql"
	"net/http"
	"strconv"

	"github.com/apimgr/weather/src/server/model"
	"github.com/apimgr/weather/src/utils"

	"github.com/gin-gonic/gin"
)

// DomainHandlers contains domain-related HTTP handlers
// TEMPLATE.md PART 34: Custom domain support API
type DomainHandlers struct {
	ServerDB *sql.DB
	Logger   *utils.Logger
}

// NewDomainHandlers creates a new DomainHandlers instance
func NewDomainHandlers(serverDB *sql.DB, logger *utils.Logger) *DomainHandlers {
	return &DomainHandlers{
		ServerDB: serverDB,
		Logger:   logger,
	}
}

// ListDomains returns all domains (admin) or user's domains
// GET /{api_version}/admin/domains
// GET /{api_version}/users/domains
func (h *DomainHandlers) ListDomains(c *gin.Context) {
	user, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	u := user.(*models.User)
	domainModel := &models.DomainModel{DB: h.ServerDB}

	var domains []*models.Domain
	var err error

	// Admin can see all domains, users see only their own
	if u.Role == "admin" {
		domains, err = domainModel.List(nil)
	} else {
		domains, err = domainModel.List(&u.ID)
	}

	if err != nil {
		h.Logger.Error("Failed to list domains: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to list domains"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"domains": domains,
		"count":   len(domains),
	})
}

// GetDomain returns a single domain by ID
// GET /{api_version}/admin/domains/:id
func (h *DomainHandlers) GetDomain(c *gin.Context) {
	idStr := c.Param("id")

	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid domain ID"})
		return
	}

	domainModel := &models.DomainModel{DB: h.ServerDB}
	domain, err := domainModel.GetByID(id)
	if err != nil {
		h.Logger.Error("Failed to get domain: %v", err)
		c.JSON(http.StatusNotFound, gin.H{"error": "Domain not found"})
		return
	}

	// Check authorization - users can only see their own domains
	user, _ := c.Get("user")
	u := user.(*models.User)
	if u.Role != "admin" && (domain.UserID == nil || *domain.UserID != u.ID) {
		c.JSON(http.StatusForbidden, gin.H{"error": "Forbidden"})
		return
	}

	c.JSON(http.StatusOK, domain)
}

// CreateDomain creates a new custom domain
// POST /{api_version}/admin/domains
// POST /{api_version}/users/domains
func (h *DomainHandlers) CreateDomain(c *gin.Context) {
	var req struct {
		Domain string `json:"domain"`
		UserID *int64 `json:"user_id,omitempty"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	if req.Domain == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Domain is required"})
		return
	}

	user, _ := c.Get("user")
	u := user.(*models.User)

	// Non-admin users can only create domains for themselves
	if u.Role != "admin" {
		req.UserID = &u.ID
	}

	domainModel := &models.DomainModel{DB: h.ServerDB}
	domain, err := domainModel.Create(req.Domain, req.UserID)
	if err != nil {
		h.Logger.Error("Failed to create domain: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	h.Logger.Info("Domain created: %s (ID: %d)", domain.Domain, domain.ID)
	c.JSON(http.StatusCreated, domain)
}

// GetVerificationToken returns the verification token for a domain
// GET /{api_version}/admin/domains/:id/verification
func (h *DomainHandlers) GetVerificationToken(c *gin.Context) {
	idStr := c.Param("id")

	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid domain ID"})
		return
	}

	domainModel := &models.DomainModel{DB: h.ServerDB}

	// Check if domain exists and user has access
	domain, err := domainModel.GetByID(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Domain not found"})
		return
	}

	user, _ := c.Get("user")
	u := user.(*models.User)
	if u.Role != "admin" && (domain.UserID == nil || *domain.UserID != u.ID) {
		c.JSON(http.StatusForbidden, gin.H{"error": "Forbidden"})
		return
	}

	token, err := domainModel.GetVerificationToken(id)
	if err != nil {
		h.Logger.Error("Failed to get verification token: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get verification token"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"token":        token,
		"instructions": "Add this token as a TXT record at _weather-verify." + domain.Domain,
		"example":      "_weather-verify." + domain.Domain + " TXT \"" + token + "\"",
	})
}

// VerifyDomain marks a domain as verified after DNS verification
// PUT /{api_version}/admin/domains/:id/verify
func (h *DomainHandlers) VerifyDomain(c *gin.Context) {
	idStr := c.Param("id")

	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid domain ID"})
		return
	}

	domainModel := &models.DomainModel{DB: h.ServerDB}

	// Check if domain exists and user has access
	domain, err := domainModel.GetByID(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Domain not found"})
		return
	}

	user, _ := c.Get("user")
	u := user.(*models.User)
	if u.Role != "admin" && (domain.UserID == nil || *domain.UserID != u.ID) {
		c.JSON(http.StatusForbidden, gin.H{"error": "Forbidden"})
		return
	}

	// In production, verify DNS TXT record here
	// For now, we'll trust the admin/user verification

	err = domainModel.Verify(id)
	if err != nil {
		h.Logger.Error("Failed to verify domain: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to verify domain"})
		return
	}

	h.Logger.Info("Domain verified: %s (ID: %d)", domain.Domain, domain.ID)
	c.JSON(http.StatusOK, gin.H{
		"message": "Domain verified successfully",
		"domain":  domain.Domain,
	})
}

// ActivateDomain activates a domain (makes it live)
// PUT /{api_version}/admin/domains/:id/activate
func (h *DomainHandlers) ActivateDomain(c *gin.Context) {
	idStr := c.Param("id")

	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid domain ID"})
		return
	}

	domainModel := &models.DomainModel{DB: h.ServerDB}

	// Check if domain exists and user has access
	domain, err := domainModel.GetByID(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Domain not found"})
		return
	}

	user, _ := c.Get("user")
	u := user.(*models.User)
	if u.Role != "admin" && (domain.UserID == nil || *domain.UserID != u.ID) {
		c.JSON(http.StatusForbidden, gin.H{"error": "Forbidden"})
		return
	}

	if !domain.IsVerified {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Domain must be verified before activation"})
		return
	}

	err = domainModel.Activate(id)
	if err != nil {
		h.Logger.Error("Failed to activate domain: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to activate domain"})
		return
	}

	h.Logger.Info("Domain activated: %s (ID: %d)", domain.Domain, domain.ID)
	c.JSON(http.StatusOK, gin.H{
		"message": "Domain activated successfully",
		"domain":  domain.Domain,
	})
}

// DeactivateDomain deactivates a domain
// PUT /{api_version}/admin/domains/:id/deactivate
func (h *DomainHandlers) DeactivateDomain(c *gin.Context) {
	idStr := c.Param("id")

	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid domain ID"})
		return
	}

	domainModel := &models.DomainModel{DB: h.ServerDB}

	// Check if domain exists and user has access
	domain, err := domainModel.GetByID(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Domain not found"})
		return
	}

	user, _ := c.Get("user")
	u := user.(*models.User)
	if u.Role != "admin" && (domain.UserID == nil || *domain.UserID != u.ID) {
		c.JSON(http.StatusForbidden, gin.H{"error": "Forbidden"})
		return
	}

	err = domainModel.Deactivate(id)
	if err != nil {
		h.Logger.Error("Failed to deactivate domain: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to deactivate domain"})
		return
	}

	h.Logger.Info("Domain deactivated: %s (ID: %d)", domain.Domain, domain.ID)
	c.JSON(http.StatusOK, gin.H{
		"message": "Domain deactivated successfully",
		"domain":  domain.Domain,
	})
}

// UpdateSSL updates SSL configuration for a domain
// PUT /{api_version}/admin/domains/:id/ssl
func (h *DomainHandlers) UpdateSSL(c *gin.Context) {
	idStr := c.Param("id")

	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid domain ID"})
		return
	}

	var req struct {
		CertPath string `json:"cert_path"`
		KeyPath  string `json:"key_path"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	if req.CertPath == "" || req.KeyPath == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Both cert_path and key_path are required"})
		return
	}

	domainModel := &models.DomainModel{DB: h.ServerDB}

	// Check if domain exists and user has access
	domain, err := domainModel.GetByID(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Domain not found"})
		return
	}

	user, _ := c.Get("user")
	u := user.(*models.User)
	if u.Role != "admin" && (domain.UserID == nil || *domain.UserID != u.ID) {
		c.JSON(http.StatusForbidden, gin.H{"error": "Forbidden"})
		return
	}

	err = domainModel.UpdateSSL(id, req.CertPath, req.KeyPath)
	if err != nil {
		h.Logger.Error("Failed to update SSL: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update SSL configuration"})
		return
	}

	h.Logger.Info("SSL updated for domain: %s (ID: %d)", domain.Domain, domain.ID)
	c.JSON(http.StatusOK, gin.H{
		"message": "SSL configuration updated successfully",
		"domain":  domain.Domain,
	})
}

// DeleteDomain deletes a custom domain
// DELETE /{api_version}/admin/domains/:id
func (h *DomainHandlers) DeleteDomain(c *gin.Context) {
	idStr := c.Param("id")

	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid domain ID"})
		return
	}

	domainModel := &models.DomainModel{DB: h.ServerDB}

	// Check if domain exists and user has access
	domain, err := domainModel.GetByID(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Domain not found"})
		return
	}

	user, _ := c.Get("user")
	u := user.(*models.User)
	if u.Role != "admin" && (domain.UserID == nil || *domain.UserID != u.ID) {
		c.JSON(http.StatusForbidden, gin.H{"error": "Forbidden"})
		return
	}

	err = domainModel.Delete(id)
	if err != nil {
		h.Logger.Error("Failed to delete domain: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete domain"})
		return
	}

	h.Logger.Info("Domain deleted: %s (ID: %d)", domain.Domain, domain.ID)
	c.JSON(http.StatusOK, gin.H{
		"message": "Domain deleted successfully",
	})
}
