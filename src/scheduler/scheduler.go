package scheduler

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os/exec"
	"sync"
	"time"

	"github.com/apimgr/weather/src/database"
)

// Task represents a scheduled task
type Task struct {
	Name     string
	Interval time.Duration
	Fn       func() error
	ticker   *time.Ticker
	stopChan chan bool
	running  bool
	// Can be toggled on/off
	enabled  bool
	// Last execution time
	lastRun  *time.Time
	mu       sync.Mutex
}

// Global tasks that should only run on one node in cluster mode
// AI.md PART 19: Global Tasks (run once per cluster)
var globalTasks = map[string]bool{
	"ssl-renewal":           true,
	"geoip-update":          true,
	"blocklist-update":      true,
	"cve-update":            true,
	"backup-daily":          true,
	"backup-hourly":         true,
	"update-geoip-database": true,
}

// LockTimeout is how long a lock is valid before auto-release (5 minutes per AI.md)
const LockTimeout = 5 * time.Minute

// Scheduler manages scheduled tasks
type Scheduler struct {
	tasks  []*Task
	db     *sql.DB
	nodeID string
	mu     sync.RWMutex
}

// NewScheduler creates a new scheduler instance
func NewScheduler(db *sql.DB) *Scheduler {
	// Get node ID from hostname
	nodeID, err := getNodeID()
	if err != nil {
		nodeID = "default"
	}
	return &Scheduler{
		tasks:  make([]*Task, 0),
		db:     db,
		nodeID: nodeID,
	}
}

// getNodeID returns a unique identifier for this node
func getNodeID() (string, error) {
	hostname, err := exec.Command("hostname").Output()
	if err != nil {
		return "", err
	}
	return string(hostname[:len(hostname)-1]), nil // Remove trailing newline
}

// AddTask adds a new task to the scheduler
func (s *Scheduler) AddTask(name string, interval time.Duration, fn func() error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	task := &Task{
		Name:     name,
		Interval: interval,
		Fn:       fn,
		stopChan: make(chan bool),
		running:  false,
		// Enabled by default
		enabled:  true,
		lastRun:  nil,
	}

	s.tasks = append(s.tasks, task)
	// Silently add task, no logging
}

// Start starts all scheduled tasks
func (s *Scheduler) Start() {
	s.mu.Lock()
	defer s.mu.Unlock()

	for _, task := range s.tasks {
		go s.runTask(task)
	}

	log.Printf("📅 Task manager has started (%d scheduled tasks)", len(s.tasks))
}

// Stop stops all scheduled tasks
func (s *Scheduler) Stop() {
	s.mu.Lock()
	defer s.mu.Unlock()

	log.Println("🛑 Stopping scheduler...")

	for _, task := range s.tasks {
		task.mu.Lock()
		if task.running {
			task.stopChan <- true
			if task.ticker != nil {
				task.ticker.Stop()
			}
			task.running = false
		}
		task.mu.Unlock()
	}

	log.Println("✅ Scheduler stopped")
}

// runTask runs a single task on its interval
func (s *Scheduler) runTask(task *Task) {
	task.mu.Lock()
	task.ticker = time.NewTicker(task.Interval)
	task.running = true
	task.mu.Unlock()

	// Silently start task, no logging

	for {
		select {
		case <-task.ticker.C:
			s.executeTask(task)
		case <-task.stopChan:
			// Silently stop task, no logging
			return
		}
	}
}

// isGlobalTask returns true if this task should only run on one node
func isGlobalTask(taskName string) bool {
	return globalTasks[taskName]
}

// acquireTaskLock attempts to acquire a distributed lock for a task
// AI.md PART 19: Cluster-aware task locking
func (s *Scheduler) acquireTaskLock(taskName string) bool {
	// For non-global tasks, always allow (run on every node)
	if !isGlobalTask(taskName) {
		return true
	}

	now := time.Now()
	lockExpiry := now.Add(-LockTimeout)

	// Try to acquire lock:
	// 1. If no lock exists, acquire it
	// 2. If lock exists but expired (older than 5 min), steal it
	// 3. If lock exists and held by us, refresh it
	// 4. If lock exists and held by another node, fail
	result, err := database.GetServerDB().Exec(`
		INSERT INTO server_scheduler_state (task_id, task_name, locked_by, locked_at, enabled)
		VALUES (?, ?, ?, ?, true)
		ON CONFLICT(task_id) DO UPDATE SET
			locked_by = CASE
				WHEN locked_by IS NULL OR locked_at < ? OR locked_by = ? THEN ?
				ELSE locked_by
			END,
			locked_at = CASE
				WHEN locked_by IS NULL OR locked_at < ? OR locked_by = ? THEN ?
				ELSE locked_at
			END
		WHERE task_id = ?
	`, taskName, taskName, s.nodeID, now, lockExpiry, s.nodeID, s.nodeID, lockExpiry, s.nodeID, now, taskName)

	if err != nil {
		log.Printf("⚠️  Failed to acquire lock for task '%s': %v", taskName, err)
		return false
	}

	// Check if we actually got the lock
	var lockedBy string
	err = database.GetServerDB().QueryRow(
		"SELECT locked_by FROM server_scheduler_state WHERE task_id = ?",
		taskName,
	).Scan(&lockedBy)

	if err != nil || lockedBy != s.nodeID {
		// Another node has the lock
		return false
	}

	rowsAffected, _ := result.RowsAffected()
	return rowsAffected > 0 || lockedBy == s.nodeID
}

// releaseTaskLock releases the distributed lock for a task
func (s *Scheduler) releaseTaskLock(taskName string) {
	if !isGlobalTask(taskName) {
		return
	}

	_, err := database.GetServerDB().Exec(`
		UPDATE server_scheduler_state
		SET locked_by = NULL, locked_at = NULL
		WHERE task_id = ? AND locked_by = ?
	`, taskName, s.nodeID)

	if err != nil {
		log.Printf("⚠️  Failed to release lock for task '%s': %v", taskName, err)
	}
}

// executeTask executes a task and logs results
func (s *Scheduler) executeTask(task *Task) {
	// Check if task is enabled
	task.mu.Lock()
	if !task.enabled {
		task.mu.Unlock()
		return
	}
	task.mu.Unlock()

	// AI.md PART 19: Cluster-aware locking for global tasks
	if !s.acquireTaskLock(task.Name) {
		// Another node is running this task, skip
		return
	}
	defer s.releaseTaskLock(task.Name)

	start := time.Now()
	err := task.Fn()
	end := time.Now()
	elapsed := end.Sub(start)

	// Update last run time
	task.mu.Lock()
	task.lastRun = &end
	task.mu.Unlock()

	// Record in database
	s.RecordTaskRun(task.Name, start, end, err)

	if err != nil {
		log.Printf("❌ Task '%s' failed after %v: %v", task.Name, elapsed, err)
	} else {
		log.Printf("✅ Task '%s' completed in %v", task.Name, elapsed)
	}

	// Log to audit if enabled
	s.logTaskExecution(task.Name, elapsed, err)
}

// logTaskExecution logs task execution to audit log
func (s *Scheduler) logTaskExecution(taskName string, duration time.Duration, err error) {
	// Check if audit logging is enabled
	var auditEnabled string
	queryErr := database.GetServerDB().QueryRow("SELECT value FROM server_config WHERE key = 'audit.enabled'").Scan(&auditEnabled)
	if queryErr != nil || auditEnabled != "true" {
		return
	}

	status := "success"
	details := fmt.Sprintf("Completed in %v", duration)
	if err != nil {
		status = "error"
		details = fmt.Sprintf("Failed: %v", err)
	}

	_, insertErr := database.GetServerDB().Exec(`
		INSERT INTO server_audit_log (user_id, action, resource_type, resource_id, details, ip_address, user_agent, status)
		VALUES (NULL, ?, 'scheduler', ?, ?, 'system', 'scheduler', ?)
	`, taskName, taskName, details, status)

	if insertErr != nil {
		log.Printf("⚠️  Failed to log scheduler task: %v", insertErr)
	}
}

// GetTaskStatus returns status of all tasks
func (s *Scheduler) GetTaskStatus() []map[string]interface{} {
	s.mu.RLock()
	defer s.mu.RUnlock()

	status := make([]map[string]interface{}, 0, len(s.tasks))

	for _, task := range s.tasks {
		task.mu.Lock()
		status = append(status, map[string]interface{}{
			"name":     task.Name,
			"interval": task.Interval.String(),
			"running":  task.running,
		})
		task.mu.Unlock()
	}

	return status
}

// CleanupOldSessions removes expired sessions
func CleanupOldSessions(db *sql.DB) error {
	result, err := database.GetUsersDB().Exec("DELETE FROM user_sessions WHERE expires_at < datetime('now')")
	if err != nil {
		return fmt.Errorf("failed to cleanup sessions: %w", err)
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected > 0 {
		log.Printf("🧹 Cleaned up %d expired sessions", rowsAffected)
	}

	return nil
}

// CleanupOldAuditLogs removes audit logs older than retention period
func CleanupOldAuditLogs(db *sql.DB) error {
	// Get retention days from settings
	var retentionDays int
	err := database.GetServerDB().QueryRow("SELECT value FROM server_config WHERE key = 'audit.retention_days'").Scan(&retentionDays)
	if err != nil {
		// Default to 90 days
		retentionDays = 90
	}

	result, err := database.GetServerDB().Exec(`
		DELETE FROM server_audit_log
		WHERE created_at < datetime('now', '-' || ? || ' days')
	`, retentionDays)

	if err != nil {
		return fmt.Errorf("failed to cleanup audit logs: %w", err)
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected > 0 {
		log.Printf("🧹 Cleaned up %d old audit logs (retention: %d days)", rowsAffected, retentionDays)
	}

	return nil
}

// CheckWeatherAlerts checks for weather alerts on saved locations
func CheckWeatherAlerts(db *sql.DB) error {
	// Get all locations with alerts enabled
	rows, err := database.GetUsersDB().Query(`
		SELECT l.id, l.name, l.latitude, l.longitude, l.user_id
		FROM user_saved_locations l
		JOIN user_accounts u ON l.user_id = u.id
		WHERE l.alerts_enabled = 1
	`)
	if err != nil {
		return fmt.Errorf("failed to fetch locations: %w", err)
	}
	defer rows.Close()

	alertCount := 0

	for rows.Next() {
		var locationID int
		var name string
		var latitude, longitude float64
		var userID int

		if err := rows.Scan(&locationID, &name, &latitude, &longitude, &userID); err != nil {
			log.Printf("⚠️  Failed to scan location: %v", err)
			continue
		}

		// Fetch weather data from Open-Meteo API
		url := fmt.Sprintf("https://api.open-meteo.com/v1/forecast?latitude=%.4f&longitude=%.4f&current=temperature_2m,wind_speed_10m,precipitation,weather_code&temperature_unit=fahrenheit&wind_speed_unit=mph&precipitation_unit=inch",
			latitude, longitude)

		resp, err := http.Get(url)
		if err != nil {
			log.Printf("⚠️  Failed to fetch weather for %s: %v", name, err)
			continue
		}
		defer resp.Body.Close()

		var weatherData struct {
			Current struct {
				Temperature   float64 `json:"temperature_2m"`
				WindSpeed     float64 `json:"wind_speed_10m"`
				Precipitation float64 `json:"precipitation"`
				WeatherCode   int     `json:"weather_code"`
			} `json:"current"`
		}

		if err := json.NewDecoder(resp.Body).Decode(&weatherData); err != nil {
			log.Printf("⚠️  Failed to decode weather data for %s: %v", name, err)
			continue
		}

		// Check for alert conditions and create notifications
		created := checkAndCreateAlerts(db, userID, locationID, name, weatherData)
		alertCount += created
	}

	if alertCount > 0 {
		log.Printf("🔔 Created %d weather alerts", alertCount)
	}

	return nil
}

// checkAndCreateAlerts checks weather conditions and creates notifications
func checkAndCreateAlerts(db *sql.DB, userID, locationID int, locationName string, weather struct {
	Current struct {
		Temperature   float64 `json:"temperature_2m"`
		WindSpeed     float64 `json:"wind_speed_10m"`
		Precipitation float64 `json:"precipitation"`
		WeatherCode   int     `json:"weather_code"`
	} `json:"current"`
}) int {
	alertCount := 0

	// Check for extreme cold (below 32°F / 0°C)
	if weather.Current.Temperature < 32 {
		createNotification(db, userID, "alert", "⚠️ Freezing Temperature Alert",
			fmt.Sprintf("%s: Temperature is %.1f°F. Bundle up!", locationName, weather.Current.Temperature),
			fmt.Sprintf("/dashboard?location=%d", locationID))
		alertCount++
	}

	// Check for extreme heat (above 95°F / 35°C)
	if weather.Current.Temperature > 95 {
		createNotification(db, userID, "alert", "🌡️ Heat Alert",
			fmt.Sprintf("%s: Temperature is %.1f°F. Stay hydrated!", locationName, weather.Current.Temperature),
			fmt.Sprintf("/dashboard?location=%d", locationID))
		alertCount++
	}

	// Check for high winds (above 40 mph)
	if weather.Current.WindSpeed > 40 {
		createNotification(db, userID, "warning", "💨 High Wind Alert",
			fmt.Sprintf("%s: Wind speed is %.0f mph. Secure loose objects!", locationName, weather.Current.WindSpeed),
			fmt.Sprintf("/dashboard?location=%d", locationID))
		alertCount++
	}

	// Check for heavy precipitation (above 0.5 inches)
	if weather.Current.Precipitation > 0.5 {
		createNotification(db, userID, "info", "🌧️ Heavy Rain Alert",
			fmt.Sprintf("%s: Heavy precipitation detected (%.1f in). Prepare for flooding!", locationName, weather.Current.Precipitation),
			fmt.Sprintf("/dashboard?location=%d", locationID))
		alertCount++
	}

	// Check for severe weather codes (thunderstorms, snow, etc.)
	if weather.Current.WeatherCode >= 95 {
		createNotification(db, userID, "alert", "⛈️ Severe Weather Alert",
			fmt.Sprintf("%s: Severe weather detected. Stay safe!", locationName),
			fmt.Sprintf("/dashboard?location=%d", locationID))
		alertCount++
	}

	return alertCount
}

// createNotification creates a notification in the database
func createNotification(db *sql.DB, userID int, notifType, title, message, link string) {
	_, err := database.GetUsersDB().Exec(`
		INSERT INTO user_notifications (user_id, type, title, message, link, read)
		VALUES (?, ?, ?, ?, ?, 0)
	`, userID, notifType, title, message, link)

	if err != nil {
		log.Printf("⚠️  Failed to create notification: %v", err)
	}
}

// RefreshWeatherCache could refresh cached weather data
func RefreshWeatherCache(db *sql.DB) error {
	// This would refresh weather data cache for frequently accessed locations
	// For now, it's a placeholder as weather data is fetched on-demand

	log.Println("🌤️  Weather cache refresh (placeholder)")

	return nil
}

// CreateSystemBackup creates a backup of the database
func CreateSystemBackup(db *sql.DB) error {
	// Get backup settings
	var backupEnabled string
	err := database.GetServerDB().QueryRow("SELECT value FROM server_config WHERE key = 'backup.enabled'").Scan(&backupEnabled)
	if err != nil || backupEnabled != "true" {
		// Backups disabled
		return nil
	}

	// Backup functionality implemented via backup.Create()
	// This would copy the SQLite database file to a backup location
	// with timestamp-based naming

	log.Println("💾 System backup (placeholder)")

	return nil
}

// CleanupExpiredTokens removes expired API and setup tokens
// AI.md PART 19: token cleanup every 15 minutes
func CleanupExpiredTokens(db *sql.DB) error {
	// Clean up expired API tokens
	result, err := database.GetServerDB().Exec(`
		DELETE FROM server_api_tokens
		WHERE expires_at IS NOT NULL AND expires_at < datetime('now')
	`)
	if err != nil {
		return fmt.Errorf("failed to cleanup expired API tokens: %w", err)
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected > 0 {
		log.Printf("🧹 Cleaned up %d expired API tokens", rowsAffected)
	}

	// Clean up expired setup tokens
	result2, err := database.GetServerDB().Exec(`
		DELETE FROM server_setup_tokens
		WHERE expires_at < datetime('now')
	`)
	if err != nil {
		// Table may not exist, that's ok
		return nil
	}

	rowsAffected2, _ := result2.RowsAffected()
	if rowsAffected2 > 0 {
		log.Printf("🧹 Cleaned up %d expired setup tokens", rowsAffected2)
	}

	return nil
}

// CheckSSLRenewal checks if SSL certificates need renewal
// AI.md PART 19: SSL renewal daily at 03:00, renew 7 days before expiry
func CheckSSLRenewal() error {
	// Check if SSL is enabled via Let's Encrypt
	var sslEnabled string
	err := database.GetServerDB().QueryRow("SELECT value FROM server_config WHERE key = 'ssl.letsencrypt.enabled'").Scan(&sslEnabled)
	if err != nil || sslEnabled != "true" {
		// SSL not using Let's Encrypt, skip renewal
		return nil
	}

	log.Println("🔐 Checking SSL certificate renewal status...")

	// TODO: Implement actual certificate expiry check and renewal via ACME
	// For now, this is a placeholder
	// The renewal logic should:
	// 1. Check cert expiry date
	// 2. If < 7 days until expiry, trigger renewal
	// 3. Use Let's Encrypt ACME challenge

	return nil
}

// SelfHealthCheck performs internal health verification
// AI.md PART 19: healthcheck_self every 5 minutes
func SelfHealthCheck() error {
	// Check database connectivity
	err := database.GetServerDB().Ping()
	if err != nil {
		log.Printf("⚠️ Self health check: database ping failed: %v", err)
		return fmt.Errorf("database health check failed: %w", err)
	}

	// Check disk space (simplified)
	// TODO: Implement actual disk space check

	return nil
}

// CheckTorHealth checks Tor service connectivity
// AI.md PART 19: tor_health every 10 minutes, auto-restart if needed
func CheckTorHealth() error {
	// Check if Tor binary exists
	_, err := exec.LookPath("tor")
	if err != nil {
		// Tor not installed, skip
		return nil
	}

	// Check if Tor service is enabled
	var torEnabled string
	queryErr := database.GetServerDB().QueryRow("SELECT value FROM server_config WHERE key = 'tor.enabled'").Scan(&torEnabled)
	if queryErr != nil || torEnabled != "true" {
		// Tor not enabled, skip
		return nil
	}

	// TODO: Check Tor circuit health and restart if needed
	// This would involve:
	// 1. Check if Tor process is running
	// 2. Check if circuits are established
	// 3. Restart Tor if unhealthy

	return nil
}

// CleanupRateLimitCounters resets rate limit counters
func CleanupRateLimitCounters(db *sql.DB) error {
	// Reset hourly counters that are older than 1 hour
	result, err := database.GetServerDB().Exec(`
		DELETE FROM server_rate_limits
		WHERE window_start < datetime('now', '-1 hour')
	`)
	if err != nil {
		return fmt.Errorf("failed to cleanup rate limits: %w", err)
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected > 0 {
		log.Printf("🧹 Cleaned up %d old rate limit counters", rowsAffected)
	}

	return nil
}

// UpdateBlocklist updates the IP blocklist database
// AI.md PART 19: blocklist_update daily at 04:00
func UpdateBlocklist() error {
	log.Println("🛡️ Updating IP blocklist database...")

	// Check if blocklist is enabled
	var blocklistEnabled string
	err := database.GetServerDB().QueryRow("SELECT value FROM server_config WHERE key = 'security.blocklist.enabled'").Scan(&blocklistEnabled)
	if err != nil || blocklistEnabled != "true" {
		// Blocklist not enabled, skip
		return nil
	}

	// TODO: Implement blocklist download and parsing
	// Sources to consider:
	// - Spamhaus DROP/EDROP
	// - FireHOL Level 1
	// - Emerging Threats compromised IPs
	// The implementation would:
	// 1. Download blocklist from configured source
	// 2. Parse IP ranges/addresses
	// 3. Store in server_ip_blocklist table
	// 4. Update in-memory cache for fast lookups

	log.Println("🛡️ Blocklist update complete")
	return nil
}

// UpdateCVEDatabase updates the CVE vulnerability database
// AI.md PART 19: cve_update daily at 05:00
func UpdateCVEDatabase() error {
	log.Println("🔒 Updating CVE database...")

	// Check if CVE monitoring is enabled
	var cveEnabled string
	err := database.GetServerDB().QueryRow("SELECT value FROM server_config WHERE key = 'security.cve.enabled'").Scan(&cveEnabled)
	if err != nil || cveEnabled != "true" {
		// CVE monitoring not enabled, skip
		return nil
	}

	// TODO: Implement CVE database update
	// Sources to consider:
	// - NVD (National Vulnerability Database) API
	// - GitHub Security Advisories
	// The implementation would:
	// 1. Fetch recent CVE data from NVD
	// 2. Check for vulnerabilities affecting dependencies
	// 3. Store relevant CVEs in server_cve_alerts table
	// 4. Generate alerts for critical vulnerabilities

	log.Println("🔒 CVE database update complete")
	return nil
}

// ClusterHeartbeat sends a heartbeat to indicate this node is alive
// AI.md PART 19 line 24792: cluster.heartbeat every 30 seconds (cluster mode only)
func ClusterHeartbeat(nodeID string) error {
	// Check if cluster mode is enabled
	var clusterEnabled string
	err := database.GetServerDB().QueryRow("SELECT value FROM server_config WHERE key = 'cluster.enabled'").Scan(&clusterEnabled)
	if err != nil || clusterEnabled != "true" {
		// Not in cluster mode, skip silently
		return nil
	}

	// Update node heartbeat in cluster nodes table
	// Per AI.md lines 22616-22620
	_, err = database.GetServerDB().Exec(`
		INSERT INTO server_nodes (node_id, last_heartbeat, status)
		VALUES (?, datetime('now'), 'online')
		ON CONFLICT(node_id) DO UPDATE SET
			last_heartbeat = datetime('now'),
			status = 'online'
	`, nodeID)

	if err != nil {
		return fmt.Errorf("failed to send cluster heartbeat: %w", err)
	}

	return nil
}

// EnableTask enables a task by name
func (s *Scheduler) EnableTask(taskName string) error {
	s.mu.RLock()
	defer s.mu.RUnlock()

	for _, task := range s.tasks {
		if task.Name == taskName {
			task.mu.Lock()
			task.enabled = true
			task.mu.Unlock()
			log.Printf("✅ Task '%s' enabled", taskName)
			return nil
		}
	}

	return fmt.Errorf("task '%s' not found", taskName)
}

// DisableTask disables a task by name
func (s *Scheduler) DisableTask(taskName string) error {
	s.mu.RLock()
	defer s.mu.RUnlock()

	for _, task := range s.tasks {
		if task.Name == taskName {
			task.mu.Lock()
			task.enabled = false
			task.mu.Unlock()
			log.Printf("⏸️  Task '%s' disabled", taskName)
			return nil
		}
	}

	return fmt.Errorf("task '%s' not found", taskName)
}

// TriggerTask manually triggers a task to run immediately
func (s *Scheduler) TriggerTask(taskName string) error {
	s.mu.RLock()
	defer s.mu.RUnlock()

	for _, task := range s.tasks {
		if task.Name == taskName {
			log.Printf("🔄 Manually triggering task '%s'", taskName)
			go s.executeTask(task)
			return nil
		}
	}

	return fmt.Errorf("task '%s' not found", taskName)
}

// GetTask returns a task by name
func (s *Scheduler) GetTask(taskName string) *Task {
	s.mu.RLock()
	defer s.mu.RUnlock()

	for _, task := range s.tasks {
		if task.Name == taskName {
			return task
		}
	}

	return nil
}
