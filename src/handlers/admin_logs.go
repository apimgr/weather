package handlers

import (
	"bufio"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

type LogsHandler struct {
	logsDir string
	logFile string
}

func NewLogsHandler(logsDir string) *LogsHandler {
	return &LogsHandler{
		logsDir: logsDir,
		logFile: filepath.Join(logsDir, "weather.log"),
	}
}

type LogEntry struct {
	Timestamp string `json:"timestamp"`
	Level     string `json:"level"`
	Source    string `json:"source"`
	Message   string `json:"message"`
}

// GetLogs retrieves application logs with filtering
func (h *LogsHandler) GetLogs(c *gin.Context) {
	// Get tail parameter (number of lines to return)
	tailParam := c.DefaultQuery("tail", "250")
	var tailLines int
	if tailParam == "all" {
		tailLines = -1 // All lines
	} else {
		var err error
		tailLines, err = strconv.Atoi(tailParam)
		if err != nil {
			tailLines = 250
		}
	}

	// Read log file
	logs, err := h.readLogs(tailLines)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to read logs"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"logs":  logs,
		"count": len(logs),
	})
}

// DownloadLogs allows downloading the complete log file
func (h *LogsHandler) DownloadLogs(c *gin.Context) {
	// Check if log file exists
	if _, err := os.Stat(h.logFile); os.IsNotExist(err) {
		c.JSON(http.StatusNotFound, gin.H{"error": "Log file not found"})
		return
	}

	// Set headers for download
	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=weather_%s.log", time.Now().Format("2006-01-02")))
	c.Header("Content-Type", "text/plain")
	c.File(h.logFile)
}

// ClearLogs clears all log entries
func (h *LogsHandler) ClearLogs(c *gin.Context) {
	// Truncate log file
	if err := os.Truncate(h.logFile, 0); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to clear logs"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Logs cleared successfully"})
}

// GetLogStats returns statistics about the logs
func (h *LogsHandler) GetLogStats(c *gin.Context) {
	logs, err := h.readLogs(-1) // Read all logs
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to read logs"})
		return
	}

	stats := map[string]interface{}{
		"total":  len(logs),
		"debug":  0,
		"info":   0,
		"warn":   0,
		"error":  0,
		"fatal":  0,
		"sources": make(map[string]int),
	}

	for _, log := range logs {
		switch log.Level {
		case "DEBUG":
			stats["debug"] = stats["debug"].(int) + 1
		case "INFO":
			stats["info"] = stats["info"].(int) + 1
		case "WARN":
			stats["warn"] = stats["warn"].(int) + 1
		case "ERROR":
			stats["error"] = stats["error"].(int) + 1
		case "FATAL":
			stats["fatal"] = stats["fatal"].(int) + 1
		}

		sources := stats["sources"].(map[string]int)
		sources[log.Source]++
	}

	c.JSON(http.StatusOK, stats)
}

// StreamLogs streams logs in real-time (Server-Sent Events)
func (h *LogsHandler) StreamLogs(c *gin.Context) {
	c.Header("Content-Type", "text/event-stream")
	c.Header("Cache-Control", "no-cache")
	c.Header("Connection", "keep-alive")
	c.Header("Access-Control-Allow-Origin", "*")

	// Get current file size
	fileInfo, err := os.Stat(h.logFile)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to stat log file"})
		return
	}

	lastSize := fileInfo.Size()
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-c.Request.Context().Done():
			return
		case <-ticker.C:
			// Check for new content
			fileInfo, err := os.Stat(h.logFile)
			if err != nil {
				continue
			}

			if fileInfo.Size() > lastSize {
				// Read new content
				file, err := os.Open(h.logFile)
				if err != nil {
					continue
				}

				file.Seek(lastSize, 0)
				scanner := bufio.NewScanner(file)

				for scanner.Scan() {
					line := scanner.Text()
					if line != "" {
						entry := h.parseLine(line)
						c.SSEvent("log", entry)
					}
				}

				file.Close()
				lastSize = fileInfo.Size()
				c.Writer.Flush()
			}
		}
	}
}

// Helper: Read logs from file
func (h *LogsHandler) readLogs(tailLines int) ([]LogEntry, error) {
	// Check if log file exists
	if _, err := os.Stat(h.logFile); os.IsNotExist(err) {
		// Return empty logs if file doesn't exist yet
		return []LogEntry{}, nil
	}

	file, err := os.Open(h.logFile)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var lines []string
	scanner := bufio.NewScanner(file)

	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	// Apply tail if requested
	if tailLines > 0 && len(lines) > tailLines {
		lines = lines[len(lines)-tailLines:]
	}

	// Parse lines into log entries
	logs := make([]LogEntry, 0, len(lines))
	for _, line := range lines {
		if line != "" {
			logs = append(logs, h.parseLine(line))
		}
	}

	return logs, nil
}

// Helper: Parse a log line into a LogEntry
func (h *LogsHandler) parseLine(line string) LogEntry {
	// Expected format: [2025-12-13 10:30:45] [INFO] [server] Message here
	entry := LogEntry{
		Timestamp: time.Now().Format("2006-01-02 15:04:05"),
		Level:     "INFO",
		Source:    "unknown",
		Message:   line,
	}

	// Try to parse structured log
	if strings.HasPrefix(line, "[") {
		parts := strings.SplitN(line, "]", 4)
		if len(parts) >= 4 {
			entry.Timestamp = strings.TrimPrefix(parts[0], "[")
			entry.Level = strings.TrimSpace(strings.TrimPrefix(parts[1], "["))
			entry.Source = strings.TrimSpace(strings.TrimPrefix(parts[2], "["))
			entry.Message = strings.TrimSpace(parts[3])
		}
	}

	return entry
}

// RotateLogs rotates the current log file
func (h *LogsHandler) RotateLogs(c *gin.Context) {
	// Create archive filename
	timestamp := time.Now().Format("2006-01-02_15-04-05")
	archivePath := filepath.Join(h.logsDir, fmt.Sprintf("weather_%s.log", timestamp))

	// Copy current log to archive
	if err := copyFile(h.logFile, archivePath); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to archive logs"})
		return
	}

	// Truncate current log file
	if err := os.Truncate(h.logFile, 0); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to truncate log file"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Logs rotated successfully",
		"archive": archivePath,
	})
}

// ListArchivedLogs lists all archived log files
func (h *LogsHandler) ListArchivedLogs(c *gin.Context) {
	files, err := os.ReadDir(h.logsDir)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to read logs directory"})
		return
	}

	archives := make([]map[string]interface{}, 0)
	for _, file := range files {
		if file.IsDir() || !strings.HasPrefix(file.Name(), "weather_") {
			continue
		}

		info, err := file.Info()
		if err != nil {
			continue
		}

		archives = append(archives, map[string]interface{}{
			"name": file.Name(),
			"size": info.Size(),
			"date": info.ModTime().Format("2006-01-02 15:04:05"),
		})
	}

	c.JSON(http.StatusOK, gin.H{"archives": archives})
}

// Helper: Copy file
func copyFile(src, dst string) error {
	sourceFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer sourceFile.Close()

	destFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer destFile.Close()

	_, err = destFile.ReadFrom(sourceFile)
	return err
}
