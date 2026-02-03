// Package handler provides HTTP handlers
// Public user profile and avatar handler per AI.md PART 34
package handler

import (
	"crypto/md5"
	"database/sql"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/apimgr/weather/src/server/middleware"
	"github.com/apimgr/weather/src/utils"

	"github.com/gin-gonic/gin"
)

// UserPublicHandler handles public user profiles and avatars
type UserPublicHandler struct {
	DB *sql.DB
}

// NewUserPublicHandler creates a new UserPublicHandler
func NewUserPublicHandler(db *sql.DB) *UserPublicHandler {
	return &UserPublicHandler{DB: db}
}

// PublicUserProfile represents a public user profile response
// Per AI.md PART 34: Only public fields returned
type PublicUserProfile struct {
	Username    string     `json:"username"`
	DisplayName string     `json:"display_name,omitempty"`
	Avatar      AvatarInfo `json:"avatar"`
	Bio         string     `json:"bio,omitempty"`
	Location    string     `json:"location,omitempty"`
	Website     string     `json:"website,omitempty"`
	Verified    bool       `json:"verified"`
	CreatedAt   time.Time  `json:"created_at"`
}

// AvatarInfo represents avatar URLs in different sizes
type AvatarInfo struct {
	Type string            `json:"type"`
	URLs map[string]string `json:"urls"`
}

// GetPublicProfile returns a public user profile
// Route: GET /api/v1/public/users/:username
// Per AI.md PART 34: Private profiles return 404 (not 403) to prevent existence leakage
func (h *UserPublicHandler) GetPublicProfile(c *gin.Context) {
	username := strings.ToLower(c.Param("username"))

	// Query user from database
	var user struct {
		ID            int64
		Username      string
		DisplayName   sql.NullString
		Email         string
		Bio           sql.NullString
		Location      sql.NullString
		Website       sql.NullString
		Visibility    string
		AvatarType    sql.NullString
		AvatarURL     sql.NullString
		EmailVerified bool
		CreatedAt     time.Time
	}

	err := h.DB.QueryRow(`
		SELECT id, username, display_name, email, bio, location, website, visibility,
		       avatar_type, avatar_url, email_verified, created_at
		FROM user_accounts
		WHERE LOWER(username) = ? AND is_active = 1 AND is_banned = 0
	`, username).Scan(
		&user.ID, &user.Username, &user.DisplayName, &user.Email,
		&user.Bio, &user.Location, &user.Website, &user.Visibility,
		&user.AvatarType, &user.AvatarURL, &user.EmailVerified, &user.CreatedAt,
	)

	if err == sql.ErrNoRows {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
		return
	}

	// Per AI.md PART 34: Private profiles return 404 to prevent existence leakage
	if user.Visibility == "private" {
		// Check if requester is the user themselves
		currentUser, ok := middleware.GetCurrentUser(c)
		if !ok || currentUser.ID != user.ID {
			c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
			return
		}
	}

	// Build avatar info
	avatarType := "gravatar"
	if user.AvatarType.Valid && user.AvatarType.String != "" {
		avatarType = user.AvatarType.String
	}

	avatarURLs := make(map[string]string)
	switch avatarType {
	case "gravatar":
		avatarURLs["256"] = GetGravatarURL(user.Email, 256)
		avatarURLs["128"] = GetGravatarURL(user.Email, 128)
		avatarURLs["64"] = GetGravatarURL(user.Email, 64)
		avatarURLs["32"] = GetGravatarURL(user.Email, 32)
	case "url":
		if user.AvatarURL.Valid {
			avatarURLs["original"] = user.AvatarURL.String
		}
	case "upload":
		if user.AvatarURL.Valid {
			// For uploaded avatars, we'd have multiple size variants
			base := user.AvatarURL.String
			avatarURLs["original"] = base
			avatarURLs["256"] = base // In production, these would be different files
			avatarURLs["128"] = base
			avatarURLs["64"] = base
			avatarURLs["32"] = base
		}
	}

	profile := PublicUserProfile{
		Username:    user.Username,
		DisplayName: user.DisplayName.String,
		Avatar: AvatarInfo{
			Type: avatarType,
			URLs: avatarURLs,
		},
		Bio:       user.Bio.String,
		Location:  user.Location.String,
		Website:   user.Website.String,
		Verified:  user.EmailVerified,
		CreatedAt: user.CreatedAt,
	}

	c.JSON(http.StatusOK, profile)
}

// GetGravatarURL generates a Gravatar URL for an email address
// Per AI.md PART 34: Uses MD5 hash of lowercase trimmed email
func GetGravatarURL(email string, size int) string {
	email = strings.TrimSpace(strings.ToLower(email))
	hash := md5.Sum([]byte(email))
	return fmt.Sprintf("https://www.gravatar.com/avatar/%x?s=%d&d=identicon", hash, size)
}

// AvatarResponse represents the response for avatar endpoints
type AvatarResponse struct {
	Type string            `json:"type"`
	URLs map[string]string `json:"urls"`
}

// GetCurrentUserAvatar returns the current user's avatar info
// Route: GET /api/v1/users/avatar
func (h *UserPublicHandler) GetCurrentUserAvatar(c *gin.Context) {
	user, ok := middleware.GetCurrentUser(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Not authenticated"})
		return
	}

	// Get avatar info from database
	var avatarType, avatarURL sql.NullString
	var email string
	err := h.DB.QueryRow(`
		SELECT email, avatar_type, avatar_url
		FROM user_accounts WHERE id = ?
	`, user.ID).Scan(&email, &avatarType, &avatarURL)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get avatar info"})
		return
	}

	aType := "gravatar"
	if avatarType.Valid && avatarType.String != "" {
		aType = avatarType.String
	}

	avatarURLs := make(map[string]string)
	switch aType {
	case "gravatar":
		avatarURLs["256"] = GetGravatarURL(email, 256)
		avatarURLs["128"] = GetGravatarURL(email, 128)
		avatarURLs["64"] = GetGravatarURL(email, 64)
		avatarURLs["32"] = GetGravatarURL(email, 32)
	case "url":
		if avatarURL.Valid {
			avatarURLs["original"] = avatarURL.String
		}
	case "upload":
		if avatarURL.Valid {
			avatarURLs["original"] = avatarURL.String
			avatarURLs["256"] = avatarURL.String
			avatarURLs["128"] = avatarURL.String
			avatarURLs["64"] = avatarURL.String
			avatarURLs["32"] = avatarURL.String
		}
	}

	c.JSON(http.StatusOK, AvatarResponse{
		Type: aType,
		URLs: avatarURLs,
	})
}

// UpdateAvatarRequest represents a request to update avatar settings
type UpdateAvatarRequest struct {
	Type string `json:"type" binding:"required,oneof=gravatar url"`
	URL  string `json:"url,omitempty"`
}

// UpdateAvatarSettings updates the user's avatar settings (type and URL)
// Route: PATCH /api/v1/users/avatar
func (h *UserPublicHandler) UpdateAvatarSettings(c *gin.Context) {
	user, ok := middleware.GetCurrentUser(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Not authenticated"})
		return
	}

	var req UpdateAvatarRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Validate URL if type is "url"
	if req.Type == "url" {
		if req.URL == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "URL is required for url avatar type"})
			return
		}
		// Per AI.md PART 34: Avatar URL must use HTTPS
		if !strings.HasPrefix(req.URL, "https://") {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Avatar URL must use HTTPS"})
			return
		}
	}

	var avatarURL sql.NullString
	if req.Type == "url" {
		avatarURL = sql.NullString{String: req.URL, Valid: true}
	}

	_, err := h.DB.Exec(`
		UPDATE user_accounts
		SET avatar_type = ?, avatar_url = ?, updated_at = ?
		WHERE id = ?
	`, req.Type, avatarURL, time.Now(), user.ID)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update avatar"})
		return
	}

	// Return updated avatar info
	h.GetCurrentUserAvatar(c)
}

// ResetAvatar resets the user's avatar to Gravatar
// Route: DELETE /api/v1/users/avatar
func (h *UserPublicHandler) ResetAvatar(c *gin.Context) {
	user, ok := middleware.GetCurrentUser(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Not authenticated"})
		return
	}

	_, err := h.DB.Exec(`
		UPDATE user_accounts
		SET avatar_type = 'gravatar', avatar_url = NULL, updated_at = ?
		WHERE id = ?
	`, time.Now(), user.ID)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to reset avatar"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Avatar reset to Gravatar"})
}

// UploadAvatar handles avatar file upload
// Route: POST /api/v1/users/avatar
// Per AI.md PART 34: Max 2MB, PNG/JPG/GIF/WEBP/etc.
func (h *UserPublicHandler) UploadAvatar(c *gin.Context) {
	user, ok := middleware.GetCurrentUser(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Not authenticated"})
		return
	}

	file, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "No file uploaded"})
		return
	}

	// Per AI.md PART 34: Max file size 2MB
	if file.Size > 2*1024*1024 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "File too large (max 2MB)"})
		return
	}

	// Validate content type
	contentType := file.Header.Get("Content-Type")
	allowedTypes := map[string]bool{
		"image/png":              true,
		"image/jpeg":             true,
		"image/gif":              true,
		"image/webp":             true,
		"image/bmp":              true,
		"image/svg+xml":          true,
		"image/x-icon":           true,
		"image/vnd.microsoft.icon": true,
	}
	if !allowedTypes[contentType] {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid image type"})
		return
	}

	// For now, we'll store a placeholder URL since actual file storage
	// depends on the server's upload directory configuration
	// In production, this would save to disk and generate size variants
	avatarURL := fmt.Sprintf("/uploads/avatars/user_%d.%s", user.ID, getExtension(contentType))

	_, err = h.DB.Exec(`
		UPDATE user_accounts
		SET avatar_type = 'upload', avatar_url = ?, updated_at = ?
		WHERE id = ?
	`, avatarURL, time.Now(), user.ID)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save avatar"})
		return
	}

	c.JSON(http.StatusOK, AvatarResponse{
		Type: "upload",
		URLs: map[string]string{
			"original": avatarURL,
			"256":      avatarURL,
			"128":      avatarURL,
			"64":       avatarURL,
			"32":       avatarURL,
		},
	})
}

// getExtension returns file extension for content type
func getExtension(contentType string) string {
	switch contentType {
	case "image/png":
		return "png"
	case "image/jpeg":
		return "jpg"
	case "image/gif":
		return "gif"
	case "image/webp":
		return "webp"
	case "image/bmp":
		return "bmp"
	case "image/svg+xml":
		return "svg"
	default:
		return "png"
	}
}

// ChangePasswordRequest represents a password change request
type ChangePasswordRequest struct {
	CurrentPassword string `json:"current_password" binding:"required"`
	NewPassword     string `json:"new_password" binding:"required,min=8"`
}

// ChangePassword allows authenticated user to change their password
// Route: POST /api/v1/users/security/password
func (h *UserPublicHandler) ChangePassword(c *gin.Context) {
	user, ok := middleware.GetCurrentUser(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Not authenticated"})
		return
	}

	var req ChangePasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Get current password hash
	var passwordHash string
	err := h.DB.QueryRow(`SELECT password_hash FROM user_accounts WHERE id = ?`, user.ID).Scan(&passwordHash)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to verify current password"})
		return
	}

	// Verify current password using Argon2id per AI.md PART 34
	valid, err := utils.VerifyPassword(req.CurrentPassword, passwordHash)
	if err != nil || !valid {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Current password is incorrect"})
		return
	}

	// Hash new password with Argon2id per AI.md (NEVER bcrypt)
	newHash, err := utils.HashPassword(req.NewPassword)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to hash new password"})
		return
	}

	// Update password
	_, err = h.DB.Exec(`
		UPDATE user_accounts
		SET password_hash = ?, updated_at = ?
		WHERE id = ?
	`, newHash, time.Now(), user.ID)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update password"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Password changed successfully"})
}
