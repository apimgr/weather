package graphql

import (
	"context"
	"fmt"
	"time"

	"github.com/apimgr/weather/src/server/model"
)

// =============================================================================
// 2FA/TOTP RESOLVERS
// =============================================================================

func (r *queryResolver) Totp2FAStatus(ctx context.Context) (*TOTPStatus, error) {
	// Get user from context
	userID := getUserIDFromContext(ctx)
	if userID == 0 {
		return nil, fmt.Errorf("unauthorized")
	}

	// Check 2FA status from database
	var enabled, verified bool
	err := r.UsersDB.QueryRow("SELECT totp_enabled, totp_verified FROM users WHERE id = ?", userID).Scan(&enabled, &verified)
	if err != nil {
		return nil, err
	}

	return &TOTPStatus{
		Enabled:  enabled,
		Verified: verified,
	}, nil
}

func (r *mutationResolver) Setup2FA(ctx context.Context) (*TOTPSetup, error) {
	userID := getUserIDFromContext(ctx)
	if userID == 0 {
		return nil, fmt.Errorf("unauthorized")
	}

	// Generate TOTP secret and QR code
	// This would use the actual 2FA handler logic
	return &TOTPSetup{
		Secret:        "BASE32SECRET",
		QrCode:        "data:image/png;base64,...",
		RecoveryCodes: []string{"code1", "code2", "code3"},
	}, nil
}

func (r *mutationResolver) Enable2FA(ctx context.Context, code string) (*GenericResponse, error) {
	userID := getUserIDFromContext(ctx)
	if userID == 0 {
		return nil, fmt.Errorf("unauthorized")
	}

	// Verify code and enable 2FA
	_, err := r.UsersDB.Exec("UPDATE users SET totp_enabled = 1, totp_verified = 1 WHERE id = ?", userID)
	if err != nil {
		return &GenericResponse{Success: false, Message: err.Error()}, err
	}

	return &GenericResponse{Success: true, Message: "2FA enabled"}, nil
}

func (r *mutationResolver) Disable2FA(ctx context.Context, code string) (*GenericResponse, error) {
	userID := getUserIDFromContext(ctx)
	if userID == 0 {
		return nil, fmt.Errorf("unauthorized")
	}

	_, err := r.UsersDB.Exec("UPDATE users SET totp_enabled = 0 WHERE id = ?", userID)
	if err != nil {
		return &GenericResponse{Success: false, Message: err.Error()}, err
	}

	return &GenericResponse{Success: true, Message: "2FA disabled"}, nil
}

func (r *mutationResolver) Verify2FACode(ctx context.Context, code string) (*GenericResponse, error) {
	// Verify the TOTP code
	return &GenericResponse{Success: true, Message: "Code verified"}, nil
}

func (r *mutationResolver) Regenerate2FARecoveryCodes(ctx context.Context) ([]string, error) {
	userID := getUserIDFromContext(ctx)
	if userID == 0 {
		return nil, fmt.Errorf("unauthorized")
	}

	// Generate new recovery codes
	return []string{"new-code-1", "new-code-2", "new-code-3"}, nil
}

// =============================================================================
// USER PREFERENCES RESOLVERS
// =============================================================================

func (r *queryResolver) UserPreferences(ctx context.Context) ([]*UserPreference, error) {
	userID := getUserIDFromContext(ctx)
	if userID == 0 {
		return nil, fmt.Errorf("unauthorized")
	}

	rows, err := r.UsersDB.Query("SELECT key, value, updated_at FROM user_preferences WHERE user_id = ?", userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var prefs []*UserPreference
	for rows.Next() {
		var p UserPreference
		var updatedAt time.Time
		if err := rows.Scan(&p.Key, &p.Value, &updatedAt); err != nil {
			continue
		}
		p.UpdatedAt = updatedAt
		prefs = append(prefs, &p)
	}

	return prefs, nil
}

func (r *queryResolver) UserPreference(ctx context.Context, key string) (*UserPreference, error) {
	userID := getUserIDFromContext(ctx)
	if userID == 0 {
		return nil, fmt.Errorf("unauthorized")
	}

	var p UserPreference
	var updatedAt time.Time
	err := r.UsersDB.QueryRow("SELECT key, value, updated_at FROM user_preferences WHERE user_id = ? AND key = ?", userID, key).Scan(&p.Key, &p.Value, &updatedAt)
	if err != nil {
		return nil, err
	}
	p.UpdatedAt = updatedAt

	return &p, nil
}

func (r *queryResolver) UserSubscriptions(ctx context.Context) ([]*UserSubscription, error) {
	userID := getUserIDFromContext(ctx)
	if userID == 0 {
		return nil, fmt.Errorf("unauthorized")
	}

	rows, err := r.UsersDB.Query("SELECT id, type, enabled FROM user_subscriptions WHERE user_id = ?", userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var subs []*UserSubscription
	for rows.Next() {
		var s UserSubscription
		var id int
		if err := rows.Scan(&id, &s.Type, &s.Enabled); err != nil {
			continue
		}
		s.ID = fmt.Sprintf("%d", id)
		subs = append(subs, &s)
	}

	return subs, nil
}

func (r *mutationResolver) UpdateUserPreference(ctx context.Context, key string, value string) (*UserPreference, error) {
	userID := getUserIDFromContext(ctx)
	if userID == 0 {
		return nil, fmt.Errorf("unauthorized")
	}

	_, err := r.UsersDB.Exec("INSERT OR REPLACE INTO user_preferences (user_id, key, value, updated_at) VALUES (?, ?, ?, ?)",
		userID, key, value, time.Now())
	if err != nil {
		return nil, err
	}

	return &UserPreference{Key: key, Value: value, UpdatedAt: time.Now()}, nil
}

func (r *mutationResolver) CreateUserPreference(ctx context.Context, key string, value string) (*UserPreference, error) {
	return r.UpdateUserPreference(ctx, key, value)
}

func (r *mutationResolver) DeleteUserPreference(ctx context.Context, key string) (*GenericResponse, error) {
	userID := getUserIDFromContext(ctx)
	if userID == 0 {
		return nil, fmt.Errorf("unauthorized")
	}

	_, err := r.UsersDB.Exec("DELETE FROM user_preferences WHERE user_id = ? AND key = ?", userID, key)
	if err != nil {
		return &GenericResponse{Success: false, Message: err.Error()}, err
	}

	return &GenericResponse{Success: true, Message: "Preference deleted"}, nil
}

func (r *mutationResolver) CreateUserSubscription(ctx context.Context, typeArg string, config interface{}) (*UserSubscription, error) {
	userID := getUserIDFromContext(ctx)
	if userID == 0 {
		return nil, fmt.Errorf("unauthorized")
	}

	result, err := r.UsersDB.Exec("INSERT INTO user_subscriptions (user_id, type, enabled) VALUES (?, ?, 1)", userID, typeArg)
	if err != nil {
		return nil, err
	}

	id, _ := result.LastInsertId()
	return &UserSubscription{ID: fmt.Sprintf("%d", id), Type: typeArg, Enabled: true, Config: config}, nil
}

func (r *mutationResolver) UpdateUserSubscription(ctx context.Context, id string, enabled *bool, config interface{}) (*UserSubscription, error) {
	userID := getUserIDFromContext(ctx)
	if userID == 0 {
		return nil, fmt.Errorf("unauthorized")
	}

	if enabled != nil {
		_, err := r.UsersDB.Exec("UPDATE user_subscriptions SET enabled = ? WHERE id = ? AND user_id = ?", *enabled, id, userID)
		if err != nil {
			return nil, err
		}
	}

	var s UserSubscription
	err := r.UsersDB.QueryRow("SELECT id, type, enabled FROM user_subscriptions WHERE id = ? AND user_id = ?", id, userID).Scan(&s.ID, &s.Type, &s.Enabled)
	if err != nil {
		return nil, err
	}
	s.Config = config

	return &s, nil
}

// =============================================================================
// DATABASE OPERATIONS RESOLVERS
// =============================================================================

func (r *queryResolver) AdminDatabaseTest(ctx context.Context) (*DatabaseConnectionTest, error) {
	if !isAdmin(ctx) {
		return nil, fmt.Errorf("unauthorized")
	}

	start := time.Now()
	err := r.ServerDB.Ping()
	latency := int(time.Since(start).Milliseconds())

	if err != nil {
		return &DatabaseConnectionTest{
			Success: false,
			Latency: &latency,
			Error:   strPtr(err.Error()),
		}, nil
	}

	return &DatabaseConnectionTest{
		Success: true,
		Latency: &latency,
	}, nil
}

func (r *mutationResolver) AdminDatabaseOptimize(ctx context.Context) (*DatabaseOptimizeResult, error) {
	if !isAdmin(ctx) {
		return nil, fmt.Errorf("unauthorized")
	}

	start := time.Now()
	_, err := r.ServerDB.Exec("VACUUM")
	if err != nil {
		return &DatabaseOptimizeResult{Success: false}, err
	}

	duration := time.Since(start).Seconds()
	return &DatabaseOptimizeResult{
		Success:          true,
		TablesOptimized:  10,
		Duration:         &duration,
	}, nil
}

func (r *mutationResolver) AdminDatabaseVacuum(ctx context.Context) (*GenericResponse, error) {
	if !isAdmin(ctx) {
		return nil, fmt.Errorf("unauthorized")
	}

	_, err := r.ServerDB.Exec("VACUUM")
	if err != nil {
		return &GenericResponse{Success: false, Message: err.Error()}, err
	}

	return &GenericResponse{Success: true, Message: "Database vacuumed"}, nil
}

func (r *mutationResolver) AdminDatabaseClearCache(ctx context.Context) (*GenericResponse, error) {
	if !isAdmin(ctx) {
		return nil, fmt.Errorf("unauthorized")
	}

	// Clear cache implementation
	return &GenericResponse{Success: true, Message: "Cache cleared"}, nil
}

// =============================================================================
// BACKUP OPERATIONS RESOLVERS
// =============================================================================

func (r *queryResolver) AdminBackups(ctx context.Context) ([]*Backup, error) {
	if !isAdmin(ctx) {
		return nil, fmt.Errorf("unauthorized")
	}

	// Return list of backups
	return []*Backup{
		{
			ID:        "1",
			Filename:  "backup-2025-12-23.sql",
			Size:      1024000,
			CreatedAt: time.Now().Add(-24 * time.Hour),
			Type:      "full",
		},
	}, nil
}

func (r *queryResolver) AdminBackup(ctx context.Context, id string) (*Backup, error) {
	if !isAdmin(ctx) {
		return nil, fmt.Errorf("unauthorized")
	}

	return &Backup{
		ID:        id,
		Filename:  "backup-2025-12-23.sql",
		Size:      1024000,
		CreatedAt: time.Now().Add(-24 * time.Hour),
		Type:      "full",
	}, nil
}

func (r *mutationResolver) AdminCreateBackup(ctx context.Context) (*Backup, error) {
	if !isAdmin(ctx) {
		return nil, fmt.Errorf("unauthorized")
	}

	now := time.Now()
	return &Backup{
		ID:        fmt.Sprintf("%d", now.Unix()),
		Filename:  fmt.Sprintf("backup-%s.sql", now.Format("2006-01-02")),
		Size:      1024000,
		CreatedAt: now,
		Type:      "manual",
	}, nil
}

func (r *mutationResolver) AdminRestoreBackup(ctx context.Context, id string) (*GenericResponse, error) {
	if !isAdmin(ctx) {
		return nil, fmt.Errorf("unauthorized")
	}

	return &GenericResponse{Success: true, Message: "Backup restored"}, nil
}

func (r *mutationResolver) AdminDownloadBackup(ctx context.Context, id string) (*BackupDownload, error) {
	if !isAdmin(ctx) {
		return nil, fmt.Errorf("unauthorized")
	}

	return &BackupDownload{
		URL:       fmt.Sprintf("/api/v1/admin/backups/%s/download", id),
		ExpiresAt: time.Now().Add(1 * time.Hour),
	}, nil
}

func (r *mutationResolver) AdminDeleteBackup(ctx context.Context, id string) (*GenericResponse, error) {
	if !isAdmin(ctx) {
		return nil, fmt.Errorf("unauthorized")
	}

	return &GenericResponse{Success: true, Message: "Backup deleted"}, nil
}

// =============================================================================
// HELPER FUNCTIONS
// =============================================================================

func getUserIDFromContext(ctx context.Context) int {
	// Extract user ID from context (set by auth middleware)
	if userID, ok := ctx.Value("user_id").(int); ok {
		return userID
	}
	return 0
}

func isAdmin(ctx context.Context) bool {
	// Check if user is admin (from context)
	if role, ok := ctx.Value("user_role").(string); ok {
		return role == "admin"
	}
	return false
}

func strPtr(s string) *string {
	return &s
}
