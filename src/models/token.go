package models

import (
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"fmt"
	"time"
)

// APIToken represents an API token
type APIToken struct {
	ID         int       `json:"id"`
	UserID     int       `json:"user_id"`
	Token      string    `json:"token,omitempty"` // Only shown once on creation
	Name       string    `json:"name"`
	CreatedAt  time.Time `json:"created_at"`
	LastUsedAt *time.Time `json:"last_used_at,omitempty"`
}

// TokenModel handles API token database operations
type TokenModel struct {
	DB *sql.DB
}

// GenerateToken creates a cryptographically secure random token
func GenerateToken() (string, error) {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}

// Create creates a new API token for a user
func (m *TokenModel) Create(userID int, name string) (*APIToken, error) {
	token, err := GenerateToken()
	if err != nil {
		return nil, fmt.Errorf("failed to generate token: %w", err)
	}

	result, err := m.DB.Exec(`
		INSERT INTO api_tokens (user_id, token, name, created_at)
		VALUES (?, ?, ?, ?)
	`, userID, token, name, time.Now())

	if err != nil {
		return nil, fmt.Errorf("failed to create token: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return nil, err
	}

	return &APIToken{
		ID:        int(id),
		UserID:    userID,
		Token:     token, // Only returned once
		Name:      name,
		CreatedAt: time.Now(),
	}, nil
}

// GetByToken retrieves a token by its value
func (m *TokenModel) GetByToken(token string) (*APIToken, error) {
	apiToken := &APIToken{}
	var lastUsed sql.NullTime

	err := m.DB.QueryRow(`
		SELECT id, user_id, token, name, created_at, last_used_at
		FROM api_tokens WHERE token = ?
	`, token).Scan(&apiToken.ID, &apiToken.UserID, &apiToken.Token, &apiToken.Name,
		&apiToken.CreatedAt, &lastUsed)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("token not found")
	}
	if err != nil {
		return nil, err
	}

	if lastUsed.Valid {
		apiToken.LastUsedAt = &lastUsed.Time
	}

	return apiToken, nil
}

// GetByUserID retrieves all tokens for a user
func (m *TokenModel) GetByUserID(userID int) ([]*APIToken, error) {
	rows, err := m.DB.Query(`
		SELECT id, user_id, name, created_at, last_used_at
		FROM api_tokens WHERE user_id = ?
		ORDER BY created_at DESC
	`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tokens []*APIToken
	for rows.Next() {
		token := &APIToken{}
		var lastUsed sql.NullTime
		err := rows.Scan(&token.ID, &token.UserID, &token.Name, &token.CreatedAt, &lastUsed)
		if err != nil {
			return nil, err
		}
		if lastUsed.Valid {
			token.LastUsedAt = &lastUsed.Time
		}
		// Don't return the actual token value for security
		tokens = append(tokens, token)
	}

	return tokens, nil
}

// UpdateLastUsed updates the last_used_at timestamp
func (m *TokenModel) UpdateLastUsed(tokenID int) error {
	_, err := m.DB.Exec(`
		UPDATE api_tokens SET last_used_at = ?
		WHERE id = ?
	`, time.Now(), tokenID)
	return err
}

// Delete deletes a token
func (m *TokenModel) Delete(id int) error {
	_, err := m.DB.Exec("DELETE FROM api_tokens WHERE id = ?", id)
	return err
}

// DeleteByUserID deletes all tokens for a user
func (m *TokenModel) DeleteByUserID(userID int) error {
	_, err := m.DB.Exec("DELETE FROM api_tokens WHERE user_id = ?", userID)
	return err
}
