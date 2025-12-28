package handler

import (
	"net/http"
	"strconv"
	"github.com/apimgr/weather/src/scheduler"

	"github.com/gin-gonic/gin"
)

// SchedulerHandler handles scheduler-related requests
type SchedulerHandler struct {
	Scheduler *scheduler.Scheduler
}

// NewSchedulerHandler creates a new scheduler handler
func NewSchedulerHandler(s *scheduler.Scheduler) *SchedulerHandler {
	return &SchedulerHandler{Scheduler: s}
}

// GetAllTasks returns all tasks with their status and history
func (h *SchedulerHandler) GetAllTasks(c *gin.Context) {
	tasks, err := h.Scheduler.GetAllTaskInfo()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to get task information",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"tasks": tasks,
		"count": len(tasks),
	})
}

// GetTaskHistory returns execution history for a specific task
func (h *SchedulerHandler) GetTaskHistory(c *gin.Context) {
	taskName := c.Param("name")
	limitStr := c.DefaultQuery("limit", "50")

	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit <= 0 {
		limit = 50
	}

	history, err := h.Scheduler.GetTaskHistory(taskName, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to get task history",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"task_name": taskName,
		"history":   history,
		"count":     len(history),
	})
}

// EnableTask enables a specific task
func (h *SchedulerHandler) EnableTask(c *gin.Context) {
	taskName := c.Param("name")

	err := h.Scheduler.EnableTask(taskName)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":   "Task enabled successfully",
		"task_name": taskName,
	})
}

// DisableTask disables a specific task
func (h *SchedulerHandler) DisableTask(c *gin.Context) {
	taskName := c.Param("name")

	err := h.Scheduler.DisableTask(taskName)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":   "Task disabled successfully",
		"task_name": taskName,
	})
}

// TriggerTask manually triggers a task to run immediately
func (h *SchedulerHandler) TriggerTask(c *gin.Context) {
	taskName := c.Param("name")

	err := h.Scheduler.TriggerTask(taskName)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":   "Task triggered successfully",
		"task_name": taskName,
	})
}
