package handlers

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

	"MLQueue/internal/database"
	"MLQueue/internal/middleware"
	"MLQueue/internal/models"
	"MLQueue/internal/queue"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type TaskHandler struct {
	queueManager *queue.Manager
}

func NewTaskHandler(qm *queue.Manager) *TaskHandler {
	return &TaskHandler{queueManager: qm}
}

// CreateTask creates a new training task
func (h *TaskHandler) CreateTask(c *gin.Context) {
	userID := middleware.GetUserID(c)

	var req struct {
		Name     string                 `json:"name" binding:"required"`
		Config   map[string]interface{} `json:"config" binding:"required"`
		Priority int                    `json:"priority"`
		Metadata map[string]interface{} `json:"metadata"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "无效的请求参数",
			"code":    "INVALID_CONFIG",
		})
		return
	}

	// Create task
	task := models.Task{
		ID:       "task_" + uuid.New().String()[:8],
		Name:     req.Name,
		Config:   models.JSONB(req.Config),
		Priority: req.Priority,
		Status:   models.TaskStatusQueued,
		Metadata: models.JSONB(req.Metadata),
		UserID:   userID,
	}

	if err := database.DB.Create(&task).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "创建任务失败",
			"code":    "INTERNAL_ERROR",
		})
		return
	}

	// Enqueue task
	if err := h.queueManager.EnqueueTask(task.ID, float64(req.Priority)); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "任务入队失败",
			"code":    "INTERNAL_ERROR",
		})
		return
	}

	position, _ := h.queueManager.GetQueuePosition(task.ID)

	c.JSON(http.StatusCreated, gin.H{
		"success":        true,
		"task_id":        task.ID,
		"status":         task.Status,
		"queue_position": position,
	})
}

// BatchCreateTasks creates multiple tasks
func (h *TaskHandler) BatchCreateTasks(c *gin.Context) {
	userID := middleware.GetUserID(c)

	var req struct {
		Tasks []struct {
			Name     string                 `json:"name" binding:"required"`
			Config   map[string]interface{} `json:"config" binding:"required"`
			Priority int                    `json:"priority"`
		} `json:"tasks" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "无效的请求参数",
			"code":    "INVALID_CONFIG",
		})
		return
	}

	taskIDs := make([]string, 0, len(req.Tasks))

	for _, taskReq := range req.Tasks {
		task := models.Task{
			ID:       "task_" + uuid.New().String()[:8],
			Name:     taskReq.Name,
			Config:   models.JSONB(taskReq.Config),
			Priority: taskReq.Priority,
			Status:   models.TaskStatusQueued,
			UserID:   userID,
		}

		if err := database.DB.Create(&task).Error; err != nil {
			continue
		}

		if err := h.queueManager.EnqueueTask(task.ID, float64(taskReq.Priority)); err != nil {
			continue
		}

		taskIDs = append(taskIDs, task.ID)
	}

	c.JSON(http.StatusCreated, gin.H{
		"success":       true,
		"task_ids":      taskIDs,
		"created_count": len(taskIDs),
	})
}

// GetTask retrieves task details
func (h *TaskHandler) GetTask(c *gin.Context) {
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

	c.JSON(http.StatusOK, gin.H{
		"success":       true,
		"task_id":       task.ID,
		"name":          task.Name,
		"config":        task.Config,
		"priority":      task.Priority,
		"status":        task.Status,
		"created_at":    task.CreatedAt,
		"started_at":    task.StartedAt,
		"completed_at":  task.CompletedAt,
		"result":        task.Result,
		"error_message": task.ErrorMessage,
	})
}

// ListTasks lists tasks with filtering
func (h *TaskHandler) ListTasks(c *gin.Context) {
	userID := middleware.GetUserID(c)

	status := c.Query("status")
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "100"))
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))
	sortBy := c.DefaultQuery("sort", "created_at")

	query := database.DB.Where("user_id = ?", userID)

	if status != "" {
		query = query.Where("status = ?", status)
	}

	var total int64
	query.Model(&models.Task{}).Count(&total)

	var tasks []models.Task
	query = query.Order(sortBy + " DESC").Limit(limit).Offset(offset)

	if err := query.Find(&tasks).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "查询任务失败",
			"code":    "INTERNAL_ERROR",
		})
		return
	}

	taskList := make([]map[string]interface{}, len(tasks))
	for i, task := range tasks {
		taskList[i] = map[string]interface{}{
			"task_id":    task.ID,
			"name":       task.Name,
			"status":     task.Status,
			"priority":   task.Priority,
			"created_at": task.CreatedAt,
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"tasks":   taskList,
		"total":   total,
		"limit":   limit,
		"offset":  offset,
	})
}

// UpdateTaskPriority updates task priority
func (h *TaskHandler) UpdateTaskPriority(c *gin.Context) {
	taskID := c.Param("task_id")
	userID := middleware.GetUserID(c)

	var req struct {
		Priority int `json:"priority" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "无效的优先级",
			"code":    "INVALID_PRIORITY",
		})
		return
	}

	var task models.Task
	if err := database.DB.Where("id = ? AND user_id = ?", taskID, userID).First(&task).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"success": false,
			"error":   "任务不存在",
			"code":    "TASK_NOT_FOUND",
		})
		return
	}

	if task.Status != models.TaskStatusQueued && task.Status != models.TaskStatusPending {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "只能修改队列中任务的优先级",
			"code":    "TASK_ALREADY_RUNNING",
		})
		return
	}

	task.Priority = req.Priority
	database.DB.Save(&task)

	if err := h.queueManager.UpdatePriority(taskID, float64(req.Priority)); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}
	position, _ := h.queueManager.GetQueuePosition(taskID)

	c.JSON(http.StatusOK, gin.H{
		"success":            true,
		"task_id":            taskID,
		"new_priority":       req.Priority,
		"new_queue_position": position,
	})
}

// CancelTask cancels a task
func (h *TaskHandler) CancelTask(c *gin.Context) {
	taskID := c.Param("task_id")
	userID := middleware.GetUserID(c)

	var req struct {
		Reason string `json:"reason"`
	}
	err := c.ShouldBindJSON(&req)
	if err != nil {
		return
	}

	var task models.Task
	if err := database.DB.Where("id = ? AND user_id = ?", taskID, userID).First(&task).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"success": false,
			"error":   "任务不存在",
			"code":    "TASK_NOT_FOUND",
		})
		return
	}

	if task.Status == models.TaskStatusCompleted || task.Status == models.TaskStatusCancelled {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "任务已完成或已取消",
			"code":    "TASK_ALREADY_COMPLETED",
		})
		return
	}

	task.Status = models.TaskStatusCancelled
	task.ErrorMessage = fmt.Sprintf("用户取消: %s", req.Reason)
	database.DB.Save(&task)

	if err := h.queueManager.RemoveTask(taskID); err != nil {
		//c.JSON(http.StatusOK, gin.H{
		//	"success": false,
		//	"error":   "任务移除失败，或已被移除",
		//})
		//return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"task_id": taskID,
		"status":  task.Status,
	})
}

// UploadResult uploads task result
func (h *TaskHandler) UploadResult(c *gin.Context) {
	taskID := c.Param("task_id")
	userID := middleware.GetUserID(c)

	var req struct {
		Result    map[string]interface{} `json:"result" binding:"required"`
		Artifacts map[string]interface{} `json:"artifacts"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "无效的结果数据",
			"code":    "INVALID_CONFIG",
		})
		return
	}

	var task models.Task
	if err := database.DB.Where("id = ? AND user_id = ?", taskID, userID).First(&task).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"success": false,
			"error":   "任务不存在",
			"code":    "TASK_NOT_FOUND",
		})
		return
	}

	// Merge result with artifacts
	result := req.Result
	if req.Artifacts != nil {
		result["artifacts"] = req.Artifacts
	}

	task.Result = models.JSONB(result)
	task.Status = models.TaskStatusCompleted
	now := time.Now()
	task.CompletedAt = &now

	database.DB.Save(&task)

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"task_id": taskID,
		"status":  task.Status,
	})
}
