package models

import "database/sql"

// NotificationModel handles notification-related database operations
type NotificationModel struct {
	DB *sql.DB
}

// GetUnreadCount returns the count of unread notifications for a user
func (m *NotificationModel) GetUnreadCount(userID int64) (int, error) {
	// Stub implementation - return 0 for now
	return 0, nil
}
