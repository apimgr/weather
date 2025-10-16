package database

import (
	"database/sql"
	"fmt"
	"strings"
	"time"

	_ "github.com/go-sql-driver/mysql"
	_ "github.com/lib/pq"
	_ "github.com/microsoft/go-mssqldb"
	_ "modernc.org/sqlite"
)

// DatabaseConfig holds database connection configuration
type DatabaseConfig struct {
	Type     string // sqlite, postgres, mysql, mssql
	Host     string
	Port     int
	Database string
	Username string
	Password string
	SSLMode  string // For PostgreSQL
	Options  map[string]string
}

// InitDBWithConfig initializes database connection with explicit configuration
func InitDBWithConfig(config *DatabaseConfig) (*DB, error) {
	var dsn string
	var driver string

	switch strings.ToLower(config.Type) {
	case "sqlite":
		driver = "sqlite"
		if config.Database == "" {
			return nil, fmt.Errorf("database path required for SQLite")
		}
		dsn = config.Database

	case "postgres", "postgresql":
		driver = "postgres"
		sslMode := config.SSLMode
		if sslMode == "" {
			sslMode = "disable"
		}
		dsn = fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
			config.Host, config.Port, config.Username, config.Password, config.Database, sslMode)

	case "mysql", "mariadb":
		driver = "mysql"
		dsn = fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=True&loc=Local",
			config.Username, config.Password, config.Host, config.Port, config.Database)

	case "mssql", "sqlserver":
		driver = "mssql"
		dsn = fmt.Sprintf("server=%s;port=%d;database=%s;user id=%s;password=%s;encrypt=disable",
			config.Host, config.Port, config.Database, config.Username, config.Password)

	default:
		return nil, fmt.Errorf("unsupported database type: %s", config.Type)
	}

	// Open database connection
	db, err := sql.Open(driver, dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// Test connection with timeout
	db.SetConnMaxLifetime(time.Minute * 3)
	db.SetMaxOpenConns(10)
	db.SetMaxIdleConns(5)

	// Test connection
	if err := db.Ping(); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	// Enable SQLite-specific optimizations
	if driver == "sqlite" {
		if _, err := db.Exec("PRAGMA foreign_keys = ON"); err != nil {
			return nil, fmt.Errorf("failed to enable foreign keys: %w", err)
		}
		if _, err := db.Exec("PRAGMA journal_mode = WAL"); err != nil {
			return nil, fmt.Errorf("failed to enable WAL mode: %w", err)
		}
	}

	// Create schema
	if _, err := db.Exec(Schema); err != nil {
		return nil, fmt.Errorf("failed to create schema: %w", err)
	}

	// Check and apply migrations
	var currentVersion int
	err = db.QueryRow("SELECT COALESCE(MAX(version), 0) FROM schema_version").Scan(&currentVersion)
	if err != nil {
		return nil, fmt.Errorf("failed to check schema version: %w", err)
	}

	if currentVersion == 0 {
		// New database - insert schema version and defaults
		if _, err := db.Exec("INSERT INTO schema_version (version) VALUES (?)", SchemaVersion); err != nil {
			return nil, fmt.Errorf("failed to insert schema version: %w", err)
		}

		// Insert default settings
		for key, value := range DefaultSettings {
			_, err := db.Exec(`
				INSERT INTO settings (key, value) VALUES (?, ?)
				ON CONFLICT(key) DO NOTHING
			`, key, value)
			if err != nil {
				return nil, fmt.Errorf("failed to insert default setting %s: %w", key, err)
			}
		}
	} else if currentVersion < SchemaVersion {
		// Run migrations
		if err := runMigrations(db, currentVersion, SchemaVersion); err != nil {
			return nil, fmt.Errorf("failed to run migrations: %w", err)
		}
	}

	return &DB{db}, nil
}

// ParseConnectionString parses a database connection string and returns config
func ParseConnectionString(connString string) (*DatabaseConfig, error) {
	config := &DatabaseConfig{
		Options: make(map[string]string),
	}

	if strings.HasPrefix(connString, "sqlite://") || strings.HasPrefix(connString, "sqlite:") {
		config.Type = "sqlite"
		config.Database = strings.TrimPrefix(connString, "sqlite://")
		config.Database = strings.TrimPrefix(config.Database, "sqlite:")
		config.Database = strings.TrimPrefix(config.Database, "//")
		return config, nil
	}

	if strings.HasPrefix(connString, "postgres://") || strings.HasPrefix(connString, "postgresql://") {
		config.Type = "postgres"
		// Parse postgres://user:pass@host:port/dbname?sslmode=disable
		// This is a simplified parser - for production use url.Parse
		return nil, fmt.Errorf("PostgreSQL connection string parsing not yet implemented - use DB_TYPE, DB_HOST, etc.")
	}

	if strings.HasPrefix(connString, "mysql://") {
		config.Type = "mysql"
		return nil, fmt.Errorf("MySQL connection string parsing not yet implemented - use DB_TYPE, DB_HOST, etc.")
	}

	if strings.HasPrefix(connString, "sqlserver://") || strings.HasPrefix(connString, "mssql://") {
		config.Type = "mssql"
		return nil, fmt.Errorf("MSSQL connection string parsing not yet implemented - use DB_TYPE, DB_HOST, etc.")
	}

	// Assume raw SQLite path
	config.Type = "sqlite"
	config.Database = connString
	return config, nil
}
