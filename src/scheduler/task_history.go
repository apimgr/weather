package scheduler

import (
	"database/sql"
	"fmt"
	"log"
	"time"
)

// TaskRun represents a single execution of a task
type TaskRun struct {
	ID        int64     `json:"id"`
	TaskName  string    `json:"task_name"`
	StartTime time.Time `json:"start_time"`
	EndTime   time.Time `json:"end_time"`
	Duration  int64     `json:"duration_ms"` // milliseconds
	Status    string    `json:"status"`      // "success", "error"
	Error     string    `json:"error,omitempty"`
}

// TaskInfo provides detailed information about a task
type TaskInfo struct {
	Name         string     `json:"name"`
	Interval     string     `json:"interval"`
	Enabled      bool       `json:"enabled"`
	Running      bool       `json:"running"`
	NextRun      *time.Time `json:"next_run,omitempty"`
	LastRun      *TaskRun   `json:"last_run,omitempty"`
	RunCount     int        `json:"run_count"`
	SuccessCount int        `json:"success_count"`
	ErrorCount   int        `json:"error_count"`
}

// InitTaskHistoryTable creates the task_history table if it doesn't exist
func (s *Scheduler) InitTaskHistoryTable() error {
	query := `
	CREATE TABLE IF NOT EXISTS task_history (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		task_name TEXT NOT NULL,
		start_time DATETIME NOT NULL,
		end_time DATETIME NOT NULL,
		duration_ms INTEGER NOT NULL,
		status TEXT NOT NULL,
		error TEXT,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);
	CREATE INDEX IF NOT EXISTS idx_task_history_name ON task_history(task_name);
	CREATE INDEX IF NOT EXISTS idx_task_history_start ON task_history(start_time DESC);
	`

	_, err := s.db.Exec(query)
	return err
}

// RecordTaskRun records a task execution in the database
func (s *Scheduler) RecordTaskRun(taskName string, startTime, endTime time.Time, err error) error {
	duration := endTime.Sub(startTime).Milliseconds()
	status := "success"
	errorMsg := ""

	if err != nil {
		status = "error"
		errorMsg = err.Error()
	}

	_, dbErr := s.db.Exec(`
		INSERT INTO task_history (task_name, start_time, end_time, duration_ms, status, error)
		VALUES (?, ?, ?, ?, ?, ?)
	`, taskName, startTime, endTime, duration, status, errorMsg)

	return dbErr
}

// GetTaskHistory returns execution history for a specific task
func (s *Scheduler) GetTaskHistory(taskName string, limit int) ([]TaskRun, error) {
	if limit <= 0 {
		limit = 50
	}

	rows, err := s.db.Query(`
		SELECT id, task_name, start_time, end_time, duration_ms, status, COALESCE(error, '')
		FROM task_history
		WHERE task_name = ?
		ORDER BY start_time DESC
		LIMIT ?
	`, taskName, limit)

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var history []TaskRun
	for rows.Next() {
		var run TaskRun
		err := rows.Scan(&run.ID, &run.TaskName, &run.StartTime, &run.EndTime,
			&run.Duration, &run.Status, &run.Error)
		if err != nil {
			continue
		}
		history = append(history, run)
	}

	return history, nil
}

// GetLastTaskRun returns the most recent run for a task
func (s *Scheduler) GetLastTaskRun(taskName string) (*TaskRun, error) {
	var run TaskRun
	err := s.db.QueryRow(`
		SELECT id, task_name, start_time, end_time, duration_ms, status, COALESCE(error, '')
		FROM task_history
		WHERE task_name = ?
		ORDER BY start_time DESC
		LIMIT 1
	`, taskName).Scan(&run.ID, &run.TaskName, &run.StartTime, &run.EndTime,
		&run.Duration, &run.Status, &run.Error)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	return &run, nil
}

// GetTaskStats returns statistics for a task
func (s *Scheduler) GetTaskStats(taskName string) (total, success, errors int, err error) {
	err = s.db.QueryRow(`
		SELECT
			COUNT(*) as total,
			SUM(CASE WHEN status = 'success' THEN 1 ELSE 0 END) as success,
			SUM(CASE WHEN status = 'error' THEN 1 ELSE 0 END) as errors
		FROM task_history
		WHERE task_name = ?
	`, taskName).Scan(&total, &success, &errors)

	return
}

// GetAllTaskInfo returns detailed information about all tasks
func (s *Scheduler) GetAllTaskInfo() ([]TaskInfo, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	info := make([]TaskInfo, 0, len(s.tasks))

	for _, task := range s.tasks {
		task.mu.Lock()

		taskInfo := TaskInfo{
			Name:     task.Name,
			Interval: task.Interval.String(),
			Enabled:  task.enabled,
			Running:  task.running,
		}

		// Calculate next run time
		if task.enabled && task.lastRun != nil {
			nextRun := task.lastRun.Add(task.Interval)
			taskInfo.NextRun = &nextRun
		} else if task.enabled {
			// First run
			nextRun := time.Now().Add(task.Interval)
			taskInfo.NextRun = &nextRun
		}

		task.mu.Unlock()

		// Get last run from database
		lastRun, _ := s.GetLastTaskRun(task.Name)
		taskInfo.LastRun = lastRun

		// Get stats
		total, success, errors, _ := s.GetTaskStats(task.Name)
		taskInfo.RunCount = total
		taskInfo.SuccessCount = success
		taskInfo.ErrorCount = errors

		info = append(info, taskInfo)
	}

	return info, nil
}

// CleanupOldTaskHistory removes task history older than the specified days
func (s *Scheduler) CleanupOldTaskHistory(days int) error {
	if days <= 0 {
		days = 90 // Default: keep 90 days
	}

	result, err := s.db.Exec(`
		DELETE FROM task_history
		WHERE start_time < datetime('now', ?)
	`, fmt.Sprintf("-%d days", days))

	if err != nil {
		return err
	}

	rows, _ := result.RowsAffected()
	if rows > 0 {
		log.Printf("🗑️  Cleaned up %d old task history records (older than %d days)", rows, days)
	}

	return nil
}
