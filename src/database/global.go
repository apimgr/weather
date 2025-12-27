package database

import (
	"database/sql"
	"sync"
)

// Global database instances for handler access
var (
	globalDualDB *DualDB
	globalMutex  sync.RWMutex
)

// SetGlobalDualDB sets the global dual database instance
func SetGlobalDualDB(ddb *DualDB) {
	globalMutex.Lock()
	defer globalMutex.Unlock()
	globalDualDB = ddb
}

// GetGlobalDualDB returns the global dual database instance
func GetGlobalDualDB() *DualDB {
	globalMutex.RLock()
	defer globalMutex.RUnlock()
	return globalDualDB
}

// GetServerDB returns the server database connection (server.db in SQLite mode)
// Returns nil if not using dual database mode
func GetServerDB() *sql.DB {
	globalMutex.RLock()
	defer globalMutex.RUnlock()
	if globalDualDB != nil {
		return globalDualDB.Server
	}
	return nil
}

// GetUsersDB returns the users database connection (users.db in SQLite mode)
// Returns nil if not using dual database mode
func GetUsersDB() *sql.DB {
	globalMutex.RLock()
	defer globalMutex.RUnlock()
	if globalDualDB != nil {
		return globalDualDB.Users
	}
	return nil
}

// IsDualMode returns true if running in dual database mode (SQLite standalone)
func IsDualMode() bool {
	globalMutex.RLock()
	defer globalMutex.RUnlock()
	return globalDualDB != nil
}
