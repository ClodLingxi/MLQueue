package handlers

import (
	"net/http"
	"time"

	"MLQueue/internal/database"
	"MLQueue/internal/middleware"
	"MLQueue/internal/models"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type QueueHandlerV2 struct{}

func NewQueueHandlerV2() *QueueHandlerV2 {
	return &QueueHandlerV2{}
}

// CreateTrainingQueue 创建训练队列（Python客户端或前端）
func (h *QueueHandlerV2) CreateTrainingQueue(c *gin.Context) {
	unitID := c.Param("unit_id")
	userID := middleware.GetUserID(c)

	var req struct {
		Name       string                 `json:"name" binding:"required"`
		Parameters map[string]interface{} `json:"parameters" binding:"required"`
		CreatedBy  string                 `json:"created_by"` // 'client' or 'web'
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "无效的请求参数",
		})
		return
	}

	// 验证训练单元存在
	var unit models.TrainingUnit
	if err := database.DB.Where("id = ? AND user_id = ?", unitID, userID).
		First(&unit).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"success": false,
			"error":   "训练单元不存在",
		})
		return
	}

	// 计算新队列的order值（追加到末尾）
	var maxOrder int
	database.DB.Model(&models.TrainingQueue{}).
		Where("unit_id = ?", unitID).
		Select("COALESCE(MAX(\"order\"), -1)").
		Scan(&maxOrder)

	newOrder := maxOrder + 1

	// 默认创建来源
	createdBy := req.CreatedBy
	if createdBy == "" {
		createdBy = "web"
	}

	queue := models.TrainingQueue{
		ID:         "queue_" + uuid.New().String()[:8],
		UnitID:     unitID,
		Name:       req.Name,
		Parameters: models.JSONB(req.Parameters),
		Order:      newOrder,
		Status:     "pending",
		CreatedBy:  createdBy,
		UserID:     userID,
	}

	if err := database.DB.Create(&queue).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "创建训练队列失败",
		})
		return
	}

	// 更新训练单元版本号（通知Python客户端有新队列）
	database.DB.Model(&unit).Update("version", unit.Version+1)

	c.JSON(http.StatusCreated, gin.H{
		"success":  true,
		"queue_id": queue.ID,
		"queue":    queue,
	})
}

// BatchCreateQueues 批量创建训练队列（用于超参数搜索）
func (h *QueueHandlerV2) BatchCreateQueues(c *gin.Context) {
	unitID := c.Param("unit_id")
	userID := middleware.GetUserID(c)

	var req struct {
		Queues []struct {
			Name       string                 `json:"name" binding:"required"`
			Parameters map[string]interface{} `json:"parameters" binding:"required"`
		} `json:"queues" binding:"required"`
		CreatedBy string `json:"created_by"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "无效的请求参数",
		})
		return
	}

	// 验证训练单元存在
	var unit models.TrainingUnit
	if err := database.DB.Where("id = ? AND user_id = ?", unitID, userID).
		First(&unit).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"success": false,
			"error":   "训练单元不存在",
		})
		return
	}

	// 获取当前最大order值
	var maxOrder int
	database.DB.Model(&models.TrainingQueue{}).
		Where("unit_id = ?", unitID).
		Select("COALESCE(MAX(\"order\"), -1)").
		Scan(&maxOrder)

	createdBy := req.CreatedBy
	if createdBy == "" {
		createdBy = "web"
	}

	queueIDs := make([]string, 0, len(req.Queues))

	for i, queueReq := range req.Queues {
		queue := models.TrainingQueue{
			ID:         "queue_" + uuid.New().String()[:8],
			UnitID:     unitID,
			Name:       queueReq.Name,
			Parameters: models.JSONB(queueReq.Parameters),
			Order:      maxOrder + 1 + i,
			Status:     "pending",
			CreatedBy:  createdBy,
			UserID:     userID,
		}

		if err := database.DB.Create(&queue).Error; err != nil {
			continue
		}

		queueIDs = append(queueIDs, queue.ID)
	}

	// 更新训练单元版本号
	database.DB.Model(&unit).Update("version", unit.Version+1)

	c.JSON(http.StatusCreated, gin.H{
		"success":       true,
		"queue_ids":     queueIDs,
		"created_count": len(queueIDs),
	})
}

// ListTrainingQueues 列出训练单元的所有队列
func (h *QueueHandlerV2) ListTrainingQueues(c *gin.Context) {
	unitID := c.Param("unit_id")
	userID := middleware.GetUserID(c)

	// 验证权限
	var unit models.TrainingUnit
	if err := database.DB.Where("id = ? AND user_id = ?", unitID, userID).
		First(&unit).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"success": false,
			"error":   "训练单元不存在",
		})
		return
	}

	status := c.Query("status")

	query := database.DB.Where("unit_id = ?", unitID)
	if status != "" {
		query = query.Where("status = ?", status)
	}

	var queues []models.TrainingQueue
	if err := query.Order("\"order\" ASC").
		Find(&queues).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "查询训练队列失败",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"queues":  queues,
		"count":   len(queues),
	})
}

// GetTrainingQueue 获取队列详情
func (h *QueueHandlerV2) GetTrainingQueue(c *gin.Context) {
	queueID := c.Param("queue_id")
	userID := middleware.GetUserID(c)

	var queue models.TrainingQueue
	if err := database.DB.Where("id = ? AND user_id = ?", queueID, userID).
		First(&queue).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"success": false,
			"error":   "训练队列不存在",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"queue":   queue,
	})
}

// UpdateTrainingQueue 更新队列参数（仅前端，不能修改运行中的）
func (h *QueueHandlerV2) UpdateTrainingQueue(c *gin.Context) {
	queueID := c.Param("queue_id")
	userID := middleware.GetUserID(c)

	var req struct {
		Name       string                 `json:"name"`
		Parameters map[string]interface{} `json:"parameters"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "无效的请求参数",
		})
		return
	}

	var queue models.TrainingQueue
	if err := database.DB.Where("id = ? AND user_id = ?", queueID, userID).
		First(&queue).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"success": false,
			"error":   "训练队列不存在",
		})
		return
	}

	// 不允许修改运行中或已完成的队列
	if queue.Status == "running" {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "无法修改运行中的队列",
		})
		return
	}

	if queue.Status == "completed" {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "无法修改已完成的队列",
		})
		return
	}

	// 更新字段
	if req.Name != "" {
		queue.Name = req.Name
	}
	if req.Parameters != nil {
		queue.Parameters = models.JSONB(req.Parameters)
	}

	if err := database.DB.Save(&queue).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "更新队列失败",
		})
		return
	}

	// 更新训练单元版本号
	database.DB.Model(&models.TrainingUnit{}).
		Where("id = ?", queue.UnitID).
		Update("version", database.DB.Raw("version + 1"))

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"queue":   queue,
	})
}

// DeleteTrainingQueue 删除队列
func (h *QueueHandlerV2) DeleteTrainingQueue(c *gin.Context) {
	queueID := c.Param("queue_id")
	userID := middleware.GetUserID(c)

	var queue models.TrainingQueue
	if err := database.DB.Where("id = ? AND user_id = ?", queueID, userID).
		First(&queue).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"success": false,
			"error":   "训练队列不存在",
		})
		return
	}

	// 不允许删除运行中的队列
	if queue.Status == "running" {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "无法删除运行中的队列",
		})
		return
	}

	unitID := queue.UnitID

	if err := database.DB.Delete(&queue).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "删除队列失败",
		})
		return
	}

	// 更新训练单元版本号
	database.DB.Model(&models.TrainingUnit{}).
		Where("id = ?", unitID).
		Update("version", database.DB.Raw("version + 1"))

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "队列已删除",
	})
}

// ============ Python客户端专用API ============

// StartQueue Python客户端开始执行队列
func (h *QueueHandlerV2) StartQueue(c *gin.Context) {
	queueID := c.Param("queue_id")
	userID := middleware.GetUserID(c)

	var queue models.TrainingQueue
	if err := database.DB.Where("id = ? AND user_id = ?", queueID, userID).
		First(&queue).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"success": false,
			"error":   "训练队列不存在",
		})
		return
	}

	if queue.Status != "pending" {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "队列状态不是pending，无法开始",
		})
		return
	}

	now := time.Now()
	queue.Status = "running"
	queue.StartedAt = &now

	if err := database.DB.Save(&queue).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "更新队列状态失败",
		})
		return
	}

	// 更新训练单元状态为running
	database.DB.Model(&models.TrainingUnit{}).
		Where("id = ?", queue.UnitID).
		Update("status", "running")

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"queue":   queue,
	})
}

// CompleteQueue Python客户端标记队列完成
func (h *QueueHandlerV2) CompleteQueue(c *gin.Context) {
	queueID := c.Param("queue_id")
	userID := middleware.GetUserID(c)

	var req struct {
		Result  map[string]interface{} `json:"result"`
		Metrics map[string]interface{} `json:"metrics"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "无效的请求参数",
		})
		return
	}

	var queue models.TrainingQueue
	if err := database.DB.Where("id = ? AND user_id = ?", queueID, userID).
		First(&queue).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"success": false,
			"error":   "训练队列不存在",
		})
		return
	}

	now := time.Now()
	queue.Status = "completed"
	queue.CompletedAt = &now
	queue.Result = models.JSONB(req.Result)
	queue.Metrics = models.JSONB(req.Metrics)

	if err := database.DB.Save(&queue).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "更新队列状态失败",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"queue":   queue,
	})
}

// FailQueue Python客户端标记队列失败
func (h *QueueHandlerV2) FailQueue(c *gin.Context) {
	queueID := c.Param("queue_id")
	userID := middleware.GetUserID(c)

	var req struct {
		ErrorMsg string `json:"error_msg"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "无效的请求参数",
		})
		return
	}

	var queue models.TrainingQueue
	if err := database.DB.Where("id = ? AND user_id = ?", queueID, userID).
		First(&queue).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"success": false,
			"error":   "训练队列不存在",
		})
		return
	}

	now := time.Now()
	queue.Status = "failed"
	queue.CompletedAt = &now
	queue.ErrorMsg = req.ErrorMsg

	if err := database.DB.Save(&queue).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "更新队列状态失败",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"queue":   queue,
	})
}

// ReorderQueues 重新排序队列
// 只能调整pending队列，不能调整到running/completed之前
func (h *QueueHandlerV2) ReorderQueues(c *gin.Context) {
	unitID := c.Param("unit_id")
	userID := middleware.GetUserID(c)

	var req struct {
		QueueIDs []string `json:"queue_ids" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "无效的请求参数",
		})
		return
	}

	// 验证训练单元存在
	var unit models.TrainingUnit
	if err := database.DB.Where("id = ? AND user_id = ?", unitID, userID).
		First(&unit).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"success": false,
			"error":   "训练单元不存在",
		})
		return
	}

	// 获取所有待调整的队列
	var queuesToReorder []models.TrainingQueue
	if err := database.DB.Where("id IN ? AND user_id = ?", req.QueueIDs, userID).
		Find(&queuesToReorder).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "查询队列失败",
		})
		return
	}

	// 验证所有队列都属于该训练单元
	for _, queue := range queuesToReorder {
		if queue.UnitID != unitID {
			c.JSON(http.StatusBadRequest, gin.H{
				"success": false,
				"error":   "部分队列不属于该训练单元",
			})
			return
		}
	}

	// 验证所有队列都是pending状态
	for _, queue := range queuesToReorder {
		if queue.Status != "pending" {
			c.JSON(http.StatusBadRequest, gin.H{
				"success": false,
				"error":   "只能调整pending状态的队列",
				"code":    "INVALID_QUEUE_STATUS",
			})
			return
		}
	}

	// 获取所有非pending的队列数量（这些队列的order不能被占用）
	var nonPendingCount int64
	database.DB.Model(&models.TrainingQueue{}).
		Where("unit_id = ? AND status IN ?", unitID, []string{"running", "completed", "failed"}).
		Count(&nonPendingCount)

	// 重新分配order值
	// pending队列必须从nonPendingCount开始
	startOrder := int(nonPendingCount)

	// 创建ID到队列的映射，保持请求的顺序
	queueMap := make(map[string]*models.TrainingQueue)
	for i := range queuesToReorder {
		queueMap[queuesToReorder[i].ID] = &queuesToReorder[i]
	}

	// 按照请求的顺序更新order
	for i, queueID := range req.QueueIDs {
		if queue, ok := queueMap[queueID]; ok {
			queue.Order = startOrder + i
			if err := database.DB.Save(queue).Error; err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{
					"success": false,
					"error":   "更新队列顺序失败",
				})
				return
			}
		}
	}

	// 更新训练单元版本号
	database.DB.Model(&unit).Update("version", unit.Version+1)

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "队列顺序已更新",
		"count":   len(queuesToReorder),
	})
}
