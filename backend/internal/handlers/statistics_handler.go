package handlers

import (
	"net/http"
	"time"

	"MLQueue/internal/database"
	"MLQueue/internal/middleware"
	"MLQueue/internal/models"

	"github.com/gin-gonic/gin"
)

type StatisticsHandler struct{}

func NewStatisticsHandler() *StatisticsHandler {
	return &StatisticsHandler{}
}

// GetTaskStatistics returns task statistics
func (h *StatisticsHandler) GetTaskStatistics(c *gin.Context) {
	userID := middleware.GetUserID(c)

	startDateStr := c.Query("start_date")
	endDateStr := c.Query("end_date")

	startDate := time.Now().AddDate(0, 0, -30) // Default 30 days ago
	endDate := time.Now()

	if startDateStr != "" {
		if parsed, err := time.Parse("2006-01-02", startDateStr); err == nil {
			startDate = parsed
		}
	}
	if endDateStr != "" {
		if parsed, err := time.Parse("2006-01-02", endDateStr); err == nil {
			endDate = parsed
		}
	}

	query := database.DB.Model(&models.Task{}).
		Where("user_id = ? AND created_at >= ? AND created_at <= ?", userID, startDate, endDate)

	var totalTasks int64
	var completedTasks int64
	var failedTasks int64

	query.Count(&totalTasks)
	database.DB.Model(&models.Task{}).
		Where("user_id = ? AND status = ? AND created_at >= ? AND created_at <= ?",
			userID, models.TaskStatusCompleted, startDate, endDate).
		Count(&completedTasks)
	database.DB.Model(&models.Task{}).
		Where("user_id = ? AND status = ? AND created_at >= ? AND created_at <= ?",
			userID, models.TaskStatusFailed, startDate, endDate).
		Count(&failedTasks)

	// Calculate average duration
	var tasks []models.Task
	database.DB.Where("user_id = ? AND status = ? AND completed_at IS NOT NULL AND started_at IS NOT NULL",
		userID, models.TaskStatusCompleted).
		Limit(100).
		Find(&tasks)

	var totalDuration time.Duration
	for _, task := range tasks {
		if task.CompletedAt != nil && task.StartedAt != nil {
			totalDuration += task.CompletedAt.Sub(*task.StartedAt)
		}
	}

	avgDuration := "0s"
	if len(tasks) > 0 {
		avgDuration = (totalDuration / time.Duration(len(tasks))).String()
	}

	successRate := 0.0
	if totalTasks > 0 {
		successRate = float64(completedTasks) / float64(totalTasks)
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"period": gin.H{
			"start": startDate.Format("2006-01-02"),
			"end":   endDate.Format("2006-01-02"),
		},
		"statistics": gin.H{
			"total_tasks":      totalTasks,
			"completed_tasks":  completedTasks,
			"failed_tasks":     failedTasks,
			"average_duration": avgDuration,
			"success_rate":     successRate,
		},
	})
}

// GetTaskLogs returns task execution logs
func (h *StatisticsHandler) GetTaskLogs(c *gin.Context) {
	taskID := c.Param("task_id")
	userID := middleware.GetUserID(c)
	var task models.Task
	if err := database.DB.Where("id = ? AND user_id = ?", taskID, userID).First(&task).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"success": false,
			"error":   "任务不存在",
			"code":    "TASK_NOT_FOUND",
		})
		return
	}

	// Simulated logs (in real scenario, fetch from logging system)
	logs := []map[string]interface{}{
		{
			"timestamp": task.CreatedAt.Format(time.RFC3339),
			"level":     "INFO",
			"message":   "任务已创建",
		},
	}

	if task.StartedAt != nil {
		logs = append(logs, map[string]interface{}{
			"timestamp": task.StartedAt.Format(time.RFC3339),
			"level":     "INFO",
			"message":   "开始训练...",
		})
	}

	if task.CompletedAt != nil {
		logs = append(logs, map[string]interface{}{
			"timestamp": task.CompletedAt.Format(time.RFC3339),
			"level":     "INFO",
			"message":   "训练完成",
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"task_id": taskID,
		"logs":    logs,
	})
}
