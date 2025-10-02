package models

import (
	"database/sql"
	"fmt"
	"time"
)

// Notification represents a user notification
type Notification struct {
	ID        int       `json:"id"`
	UserID    int       `json:"user_id"`
	Type      string    `json:"type"`
	Title     string    `json:"title"`
	Message   string    `json:"message"`
	Link      string    `json:"link,omitempty"`
	Read      bool      `json:"read"`
	CreatedAt time.Time `json:"created_at"`
}

// NotificationModel handles notification database operations
type NotificationModel struct {
	DB *sql.DB
}

// Create creates a new notification
func (m *NotificationModel) Create(userID int, notifType, title, message, link string) (*Notification, error) {
	result, err := m.DB.Exec(`
		INSERT INTO notifications (user_id, type, title, message, link, read, created_at)
		VALUES (?, ?, ?, ?, ?, ?, ?)
	`, userID, notifType, title, message, link, false, time.Now())

	if err != nil {
		return nil, fmt.Errorf("failed to create notification: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return nil, err
	}

	return m.GetByID(int(id))
}

// GetByID retrieves a notification by ID
func (m *NotificationModel) GetByID(id int) (*Notification, error) {
	notif := &Notification{}
	var link sql.NullString

	err := m.DB.QueryRow(`
		SELECT id, user_id, type, title, message, link, read, created_at
		FROM notifications WHERE id = ?
	`, id).Scan(&notif.ID, &notif.UserID, &notif.Type, &notif.Title,
		&notif.Message, &link, &notif.Read, &notif.CreatedAt)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("notification not found")
	}
	if err != nil {
		return nil, err
	}

	if link.Valid {
		notif.Link = link.String
	}

	return notif, nil
}

// GetByUserID retrieves all notifications for a user
func (m *NotificationModel) GetByUserID(userID int, limit int) ([]*Notification, error) {
	query := `
		SELECT id, user_id, type, title, message, link, read, created_at
		FROM notifications WHERE user_id = ?
		ORDER BY created_at DESC
	`
	if limit > 0 {
		query += fmt.Sprintf(" LIMIT %d", limit)
	}

	rows, err := m.DB.Query(query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var notifications []*Notification
	for rows.Next() {
		notif := &Notification{}
		var link sql.NullString
		err := rows.Scan(&notif.ID, &notif.UserID, &notif.Type, &notif.Title,
			&notif.Message, &link, &notif.Read, &notif.CreatedAt)
		if err != nil {
			return nil, err
		}
		if link.Valid {
			notif.Link = link.String
		}
		notifications = append(notifications, notif)
	}

	return notifications, nil
}

// GetUnreadCount returns the count of unread notifications for a user
func (m *NotificationModel) GetUnreadCount(userID int) (int, error) {
	var count int
	err := m.DB.QueryRow(`
		SELECT COUNT(*) FROM notifications
		WHERE user_id = ? AND read = 0
	`, userID).Scan(&count)
	return count, err
}

// MarkAsRead marks a notification as read
func (m *NotificationModel) MarkAsRead(id int) error {
	_, err := m.DB.Exec("UPDATE notifications SET read = 1 WHERE id = ?", id)
	return err
}

// MarkAllAsRead marks all notifications as read for a user
func (m *NotificationModel) MarkAllAsRead(userID int) error {
	_, err := m.DB.Exec("UPDATE notifications SET read = 1 WHERE user_id = ?", userID)
	return err
}

// Delete deletes a notification
func (m *NotificationModel) Delete(id int) error {
	_, err := m.DB.Exec("DELETE FROM notifications WHERE id = ?", id)
	return err
}

// DeleteOld deletes notifications older than specified days
func (m *NotificationModel) DeleteOld(days int) error {
	cutoff := time.Now().AddDate(0, 0, -days)
	_, err := m.DB.Exec("DELETE FROM notifications WHERE created_at < ?", cutoff)
	return err
}
