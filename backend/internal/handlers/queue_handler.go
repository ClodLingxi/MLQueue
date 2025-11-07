package handlers

import (
	"net/http"
	"time"

	"MLQueue/internal/database"
	"MLQueue/internal/middleware"
	"MLQueue/internal/models"
	"MLQueue/internal/queue"

	"github.com/gin-gonic/gin"
)

type QueueHandler struct {
	queueManager *queue.Manager
}

func NewQueueHandler(qm *queue.Manager) *QueueHandler {
	return &QueueHandler{queueManager: qm}
}

// GetQueueStatus returns queue statistics
func (h *QueueHandler) GetQueueStatus(c *gin.Context) {
	userID := middleware.GetUserID(c)

	// Get statistics
	var stats struct {
		Pending   int64 `json:"pending"`
		Queued    int64 `json:"queued"`
		Running   int64 `json:"running"`
		Completed int64 `json:"completed"`
		Failed    int64 `json:"failed"`
		Cancelled int64 `json:"cancelled"`
	}

	database.DB.Model(&models.Task{}).Where("user_id = ? AND status = ?", userID, models.TaskStatusPending).Count(&stats.Pending)
	database.DB.Model(&models.Task{}).Where("user_id = ? AND status = ?", userID, models.TaskStatusQueued).Count(&stats.Queued)
	database.DB.Model(&models.Task{}).Where("user_id = ? AND status = ?", userID, models.TaskStatusRunning).Count(&stats.Running)
	database.DB.Model(&models.Task{}).Where("user_id = ? AND status = ?", userID, models.TaskStatusCompleted).Count(&stats.Completed)
	database.DB.Model(&models.Task{}).Where("user_id = ? AND status = ?", userID, models.TaskStatusFailed).Count(&stats.Failed)
	database.DB.Model(&models.Task{}).Where("user_id = ? AND status = ?", userID, models.TaskStatusCancelled).Count(&stats.Cancelled)

	// Get current running tasks
	var currentTasks []models.Task
	database.DB.Where("user_id = ? AND status = ?", userID, models.TaskStatusRunning).
		Order("started_at DESC").
		Limit(10).
		Find(&currentTasks)

	currentTasksList := make([]map[string]interface{}, len(currentTasks))
	for i, task := range currentTasks {
		currentTasksList[i] = map[string]interface{}{
			"task_id":    task.ID,
			"name":       task.Name,
			"status":     task.Status,
			"started_at": task.StartedAt,
		}
	}

	queueLength, _ := h.queueManager.GetQueueLength()

	// Calculate estimated wait time (simplified)
	avgTaskTime := 5 * time.Minute // Example average
	estimatedWait := time.Duration(queueLength) * avgTaskTime

	c.JSON(http.StatusOK, gin.H{
		"success":             true,
		"queue_name":          "default",
		"statistics":          stats,
		"current_tasks":       currentTasksList,
		"queue_length":        queueLength,
		"estimated_wait_time": estimatedWait.String(),
	})
}

// ReorderQueue manually reorders queue
func (h *QueueHandler) ReorderQueue(c *gin.Context) {
	userID := middleware.GetUserID(c)

	var req struct {
		TaskIDs []string `json:"task_ids" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "无效的请求参数",
			"code":    "INVALID_CONFIG",
		})
		return
	}

	// Verify all tasks belong to user
	var count int64
	database.DB.Model(&models.Task{}).
		Where("id IN ? AND user_id = ?", req.TaskIDs, userID).
		Count(&count)

	if int(count) != len(req.TaskIDs) {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "部分任务不存在或无权限",
			"code":    "TASK_NOT_FOUND",
		})
		return
	}

	// Reorder by updating priorities
	newOrder := make([]map[string]interface{}, len(req.TaskIDs))
	for i, taskID := range req.TaskIDs {
		priority := len(req.TaskIDs) - i
		h.queueManager.UpdatePriority(taskID, float64(priority))

		var task models.Task
		database.DB.First(&task, "id = ?", taskID)
		task.Priority = priority
		database.DB.Save(&task)

		newOrder[i] = map[string]interface{}{
			"task_id":  taskID,
			"position": i + 1,
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"success":   true,
		"message":   "队列已重新排序",
		"new_order": newOrder,
	})
}

// PauseQueue pauses queue processing
func (h *QueueHandler) PauseQueue(c *gin.Context) {
	h.queueManager.Pause()

	c.JSON(http.StatusOK, gin.H{
		"success":      true,
		"queue_status": "paused",
		"message":      "队列已暂停",
	})
}

// ResumeQueue resumes queue processing
func (h *QueueHandler) ResumeQueue(c *gin.Context) {
	h.queueManager.Resume()

	c.JSON(http.StatusOK, gin.H{
		"success":      true,
		"queue_status": "active",
		"message":      "队列已恢复",
	})
}
