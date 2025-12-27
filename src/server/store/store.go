package store

import (
	"context"
	"database/sql"
)

// Store defines the interface for data access operations
// Per AI.md specification: separate data access layer from business logic
type Store interface {
	// User operations
	GetUserByID(ctx context.Context, id int64) (interface{}, error)
	GetUserByEmail(ctx context.Context, email string) (interface{}, error)
	CreateUser(ctx context.Context, user interface{}) error
	UpdateUser(ctx context.Context, user interface{}) error
	DeleteUser(ctx context.Context, id int64) error
	ListUsers(ctx context.Context, limit, offset int) ([]interface{}, error)

	// Admin operations
	GetAdminByID(ctx context.Context, id int64) (interface{}, error)
	GetAdminByEmail(ctx context.Context, email string) (interface{}, error)
	CreateAdmin(ctx context.Context, admin interface{}) error
	UpdateAdmin(ctx context.Context, admin interface{}) error
	DeleteAdmin(ctx context.Context, id int64) error
	ListAdmins(ctx context.Context) ([]interface{}, error)

	// Session operations
	GetSession(ctx context.Context, token string) (interface{}, error)
	CreateSession(ctx context.Context, session interface{}) error
	UpdateSession(ctx context.Context, session interface{}) error
	DeleteSession(ctx context.Context, token string) error
	DeleteExpiredSessions(ctx context.Context) error

	// Token operations
	GetToken(ctx context.Context, token string) (interface{}, error)
	CreateToken(ctx context.Context, token interface{}) error
	DeleteToken(ctx context.Context, token string) error
	DeleteExpiredTokens(ctx context.Context) error

	// Location operations
	GetLocation(ctx context.Context, id int64) (interface{}, error)
	ListUserLocations(ctx context.Context, userID int64) ([]interface{}, error)
	CreateLocation(ctx context.Context, location interface{}) error
	UpdateLocation(ctx context.Context, location interface{}) error
	DeleteLocation(ctx context.Context, id int64) error

	// Notification operations
	GetNotification(ctx context.Context, id int64) (interface{}, error)
	ListUserNotifications(ctx context.Context, userID int64, limit int) ([]interface{}, error)
	CreateNotification(ctx context.Context, notification interface{}) error
	DeleteNotification(ctx context.Context, id int64) error
	DeleteOldNotifications(ctx context.Context, maxAge int) error

	// Settings operations
	GetSettings(ctx context.Context) (interface{}, error)
	UpdateSettings(ctx context.Context, settings interface{}) error

	// Transaction support
	Begin(ctx context.Context) (*sql.Tx, error)
	Commit(tx *sql.Tx) error
	Rollback(tx *sql.Tx) error

	// Health check
	Ping(ctx context.Context) error
}
