package models

import (
	"database/sql"
	"fmt"
	"time"

	"golang.org/x/crypto/bcrypt"
)

// User represents a user in the system
type User struct {
	ID           int       `json:"id"`
	Email        string    `json:"email"`
	PasswordHash string    `json:"-"` // Never expose password hash
	Role         string    `json:"role"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

// UserModel handles user database operations
type UserModel struct {
	DB *sql.DB
}

// Create creates a new user with hashed password
func (m *UserModel) Create(email, password, role string) (*User, error) {
	// Validate role
	if role != "admin" && role != "user" {
		return nil, fmt.Errorf("invalid role: must be 'admin' or 'user'")
	}

	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}

	// Insert user
	result, err := m.DB.Exec(`
		INSERT INTO users (email, password_hash, role, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?)
	`, email, string(hashedPassword), role, time.Now(), time.Now())

	if err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return nil, err
	}

	return m.GetByID(int(id))
}

// GetByID retrieves a user by ID
func (m *UserModel) GetByID(id int) (*User, error) {
	user := &User{}
	err := m.DB.QueryRow(`
		SELECT id, email, password_hash, role, created_at, updated_at
		FROM users WHERE id = ?
	`, id).Scan(&user.ID, &user.Email, &user.PasswordHash, &user.Role, &user.CreatedAt, &user.UpdatedAt)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("user not found")
	}
	if err != nil {
		return nil, err
	}

	return user, nil
}

// GetByEmail retrieves a user by email
func (m *UserModel) GetByEmail(email string) (*User, error) {
	user := &User{}
	err := m.DB.QueryRow(`
		SELECT id, email, password_hash, role, created_at, updated_at
		FROM users WHERE email = ?
	`, email).Scan(&user.ID, &user.Email, &user.PasswordHash, &user.Role, &user.CreatedAt, &user.UpdatedAt)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("user not found")
	}
	if err != nil {
		return nil, err
	}

	return user, nil
}

// GetAll retrieves all users
func (m *UserModel) GetAll() ([]*User, error) {
	rows, err := m.DB.Query(`
		SELECT id, email, password_hash, role, created_at, updated_at
		FROM users ORDER BY created_at DESC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []*User
	for rows.Next() {
		user := &User{}
		err := rows.Scan(&user.ID, &user.Email, &user.PasswordHash, &user.Role, &user.CreatedAt, &user.UpdatedAt)
		if err != nil {
			return nil, err
		}
		users = append(users, user)
	}

	return users, nil
}

// Update updates a user's information
func (m *UserModel) Update(id int, email, role string) error {
	_, err := m.DB.Exec(`
		UPDATE users SET email = ?, role = ?, updated_at = ?
		WHERE id = ?
	`, email, role, time.Now(), id)
	return err
}

// UpdatePassword updates a user's password
func (m *UserModel) UpdatePassword(id int, newPassword string) error {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("failed to hash password: %w", err)
	}

	_, err = m.DB.Exec(`
		UPDATE users SET password_hash = ?, updated_at = ?
		WHERE id = ?
	`, string(hashedPassword), time.Now(), id)
	return err
}

// Delete deletes a user
func (m *UserModel) Delete(id int) error {
	_, err := m.DB.Exec("DELETE FROM users WHERE id = ?", id)
	return err
}

// CheckPassword verifies a password against the stored hash
func (m *UserModel) CheckPassword(user *User, password string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password))
	return err == nil
}

// Count returns the total number of users
func (m *UserModel) Count() (int, error) {
	var count int
	err := m.DB.QueryRow("SELECT COUNT(*) FROM users").Scan(&count)
	return count, err
}

// CountByRole returns the number of users with a specific role
func (m *UserModel) CountByRole(role string) (int, error) {
	var count int
	err := m.DB.QueryRow("SELECT COUNT(*) FROM users WHERE role = ?", role).Scan(&count)
	return count, err
}
