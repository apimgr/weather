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
	Username     string    `json:"username"`
	Email        string    `json:"email"`
	Phone        string    `json:"phone,omitempty"`         // Optional, omit if empty
	DisplayName  string    `json:"display_name,omitempty"`  // Optional display name
	PasswordHash string    `json:"-"`                       // Never expose password hash
	Role         string    `json:"role"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

// UserModel handles user database operations
type UserModel struct {
	DB *sql.DB
}

// Create creates a new user with hashed password
func (m *UserModel) Create(username, email, password, role string) (*User, error) {
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
		INSERT INTO users (username, email, password_hash, role, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?)
	`, username, email, string(hashedPassword), role, time.Now(), time.Now())

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
	var phone, displayName sql.NullString
	err := m.DB.QueryRow(`
		SELECT id, username, email, phone, display_name, password_hash, role, created_at, updated_at
		FROM users WHERE id = ?
	`, id).Scan(&user.ID, &user.Username, &user.Email, &phone, &displayName, &user.PasswordHash, &user.Role, &user.CreatedAt, &user.UpdatedAt)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("user not found")
	}
	if err != nil {
		return nil, err
	}

	if phone.Valid {
		user.Phone = phone.String
	}
	if displayName.Valid {
		user.DisplayName = displayName.String
	}

	return user, nil
}

// GetByEmail retrieves a user by email
func (m *UserModel) GetByEmail(email string) (*User, error) {
	user := &User{}
	var phone, displayName sql.NullString
	err := m.DB.QueryRow(`
		SELECT id, username, email, phone, display_name, password_hash, role, created_at, updated_at
		FROM users WHERE email = ?
	`, email).Scan(&user.ID, &user.Username, &user.Email, &phone, &displayName, &user.PasswordHash, &user.Role, &user.CreatedAt, &user.UpdatedAt)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("user not found")
	}
	if err != nil {
		return nil, err
	}

	if phone.Valid {
		user.Phone = phone.String
	}
	if displayName.Valid {
		user.DisplayName = displayName.String
	}

	return user, nil
}

// GetByUsername retrieves a user by username
func (m *UserModel) GetByUsername(username string) (*User, error) {
	user := &User{}
	var phone, displayName sql.NullString
	err := m.DB.QueryRow(`
		SELECT id, username, email, phone, display_name, password_hash, role, created_at, updated_at
		FROM users WHERE username = ?
	`, username).Scan(&user.ID, &user.Username, &user.Email, &phone, &displayName, &user.PasswordHash, &user.Role, &user.CreatedAt, &user.UpdatedAt)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("user not found")
	}
	if err != nil {
		return nil, err
	}

	if phone.Valid {
		user.Phone = phone.String
	}
	if displayName.Valid {
		user.DisplayName = displayName.String
	}

	return user, nil
}

// GetByPhone retrieves a user by phone number
func (m *UserModel) GetByPhone(phone string) (*User, error) {
	user := &User{}
	var phoneNull, displayName sql.NullString
	err := m.DB.QueryRow(`
		SELECT id, username, email, phone, display_name, password_hash, role, created_at, updated_at
		FROM users WHERE phone = ?
	`, phone).Scan(&user.ID, &user.Username, &user.Email, &phoneNull, &displayName, &user.PasswordHash, &user.Role, &user.CreatedAt, &user.UpdatedAt)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("user not found")
	}
	if err != nil {
		return nil, err
	}

	if phoneNull.Valid {
		user.Phone = phoneNull.String
	}
	if displayName.Valid {
		user.DisplayName = displayName.String
	}

	return user, nil
}

// GetByIdentifier retrieves a user by username, email, or phone
func (m *UserModel) GetByIdentifier(identifier string) (*User, error) {
	user := &User{}
	var phone, displayName sql.NullString
	err := m.DB.QueryRow(`
		SELECT id, username, email, phone, display_name, password_hash, role, created_at, updated_at
		FROM users WHERE username = ? OR email = ? OR phone = ?
	`, identifier, identifier, identifier).Scan(&user.ID, &user.Username, &user.Email, &phone, &displayName, &user.PasswordHash, &user.Role, &user.CreatedAt, &user.UpdatedAt)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("user not found")
	}
	if err != nil {
		return nil, err
	}

	if phone.Valid {
		user.Phone = phone.String
	}
	if displayName.Valid {
		user.DisplayName = displayName.String
	}

	return user, nil
}

// GetAll retrieves all users
func (m *UserModel) GetAll() ([]*User, error) {
	rows, err := m.DB.Query(`
		SELECT id, username, email, phone, display_name, password_hash, role, created_at, updated_at
		FROM users ORDER BY created_at DESC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []*User
	for rows.Next() {
		user := &User{}
		var phone, displayName sql.NullString
		err := rows.Scan(&user.ID, &user.Username, &user.Email, &phone, &displayName, &user.PasswordHash, &user.Role, &user.CreatedAt, &user.UpdatedAt)
		if err != nil {
			return nil, err
		}
		if phone.Valid {
			user.Phone = phone.String
		}
		if displayName.Valid {
			user.DisplayName = displayName.String
		}
		users = append(users, user)
	}

	return users, nil
}

// Update updates a user's information
func (m *UserModel) Update(id int, username, email, role string) error {
	_, err := m.DB.Exec(`
		UPDATE users SET username = ?, email = ?, role = ?, updated_at = ?
		WHERE id = ?
	`, username, email, role, time.Now(), id)
	return err
}

// UpdatePhone updates a user's phone number
func (m *UserModel) UpdatePhone(id int, phone string) error {
	_, err := m.DB.Exec(`
		UPDATE users SET phone = ?, updated_at = ?
		WHERE id = ?
	`, phone, time.Now(), id)
	return err
}

// UpdateProfile updates a user's display name and phone (safe fields)
func (m *UserModel) UpdateProfile(id int, displayName, phone string) error {
	_, err := m.DB.Exec(`
		UPDATE users SET display_name = ?, phone = ?, updated_at = ?
		WHERE id = ?
	`, displayName, phone, time.Now(), id)
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
