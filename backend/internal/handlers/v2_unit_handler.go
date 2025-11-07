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

type UnitHandler struct{}

func NewUnitHandler() *UnitHandler {
	return &UnitHandler{}
}

// CreateTrainingUnit 创建训练单元（Python客户端调用）
func (h *UnitHandler) CreateTrainingUnit(c *gin.Context) {
	groupID := c.Param("group_id")
	userID := middleware.GetUserID(c)

	var req struct {
		Name        string                 `json:"name" binding:"required"`
		Description string                 `json:"description"`
		Config      map[string]interface{} `json:"config"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "无效的请求参数",
		})
		return
	}

	// 验证组存在
	var group models.Group
	if err := database.DB.Where("id = ? AND user_id = ?", groupID, userID).
		First(&group).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"success": false,
			"error":   "组不存在",
		})
		return
	}

	unit := models.TrainingUnit{
		ID:          "unit_" + uuid.New().String()[:8],
		GroupID:     groupID,
		Name:        req.Name,
		Description: req.Description,
		Config:      models.JSONB(req.Config),
		Version:     1,
		Status:      "idle",
		UserID:      userID,
	}

	if err := database.DB.Create(&unit).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "创建训练单元失败",
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"success": true,
		"unit_id": unit.ID,
		"version": unit.Version,
	})
}

// ListTrainingUnits 列出组内的训练单元
func (h *UnitHandler) ListTrainingUnits(c *gin.Context) {
	groupID := c.Param("group_id")
	userID := middleware.GetUserID(c)

	// 验证组存在
	var group models.Group
	if err := database.DB.Where("id = ? AND user_id = ?", groupID, userID).
		First(&group).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"success": false,
			"error":   "组不存在",
		})
		return
	}

	var units []models.TrainingUnit
	if err := database.DB.Where("group_id = ?", groupID).
		Order("created_at DESC").
		Find(&units).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "查询训练单元失败",
		})
		return
	}

	// 为每个单元统计队列数
	type UnitWithCount struct {
		models.TrainingUnit
		QueueCount int64 `json:"queue_count"`
	}

	unitsWithCount := make([]UnitWithCount, len(units))
	for i, unit := range units {
		// 检查并更新连接状态
		checkConnectionStatus(&unit)

		var count int64
		database.DB.Model(&models.TrainingQueue{}).
			Where("unit_id = ?", unit.ID).
			Count(&count)

		unitsWithCount[i] = UnitWithCount{
			TrainingUnit: unit,
			QueueCount:   count,
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"units":   unitsWithCount,
	})
}

// GetTrainingUnit 获取训练单元详情
func (h *UnitHandler) GetTrainingUnit(c *gin.Context) {
	unitID := c.Param("unit_id")
	userID := middleware.GetUserID(c)

	var unit models.TrainingUnit
	if err := database.DB.Where("id = ? AND user_id = ?", unitID, userID).
		First(&unit).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"success": false,
			"error":   "训练单元不存在",
		})
		return
	}

	// 检查并更新连接状态
	checkConnectionStatus(&unit)

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"unit":    unit,
	})
}

// SyncTrainingUnit Python客户端同步训练单元（拉取云端最新配置）
func (h *UnitHandler) SyncTrainingUnit(c *gin.Context) {
	unitID := c.Param("unit_id")
	userID := middleware.GetUserID(c)

	var req struct {
		ClientVersion int `json:"client_version"` // Python客户端当前版本
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "无效的请求参数",
		})
		return
	}

	var unit models.TrainingUnit
	if err := database.DB.Where("id = ? AND user_id = ?", unitID, userID).
		First(&unit).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"success": false,
			"error":   "训练单元不存在",
		})
		return
	}

	// 检查是否需要同步
	needSync := unit.Version > req.ClientVersion

	// 获取所有训练队列
	var queues []models.TrainingQueue
	database.DB.Where("unit_id = ?", unitID).
		Order("priority DESC, created_at ASC").
		Find(&queues)

	c.JSON(http.StatusOK, gin.H{
		"success":       true,
		"need_sync":     needSync,
		"cloud_version": unit.Version,
		"unit":          unit,
		"queues":        queues,
	})
}

// UpdateTrainingUnit 更新训练单元（前端或Python客户端）
func (h *UnitHandler) UpdateTrainingUnit(c *gin.Context) {
	unitID := c.Param("unit_id")
	userID := middleware.GetUserID(c)

	var req struct {
		Name        string                 `json:"name"`
		Description string                 `json:"description"`
		Config      map[string]interface{} `json:"config"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "无效的请求参数",
		})
		return
	}

	var unit models.TrainingUnit
	if err := database.DB.Where("id = ? AND user_id = ?", unitID, userID).
		First(&unit).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"success": false,
			"error":   "训练单元不存在",
		})
		return
	}

	// 更新字段
	if req.Name != "" {
		unit.Name = req.Name
	}
	unit.Description = req.Description
	if req.Config != nil {
		unit.Config = models.JSONB(req.Config)
	}

	// 版本号递增
	unit.Version++

	if err := database.DB.Save(&unit).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "更新训练单元失败",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"unit":    unit,
		"version": unit.Version,
	})
}

// DeleteTrainingUnit 删除训练单元
func (h *UnitHandler) DeleteTrainingUnit(c *gin.Context) {
	unitID := c.Param("unit_id")
	userID := middleware.GetUserID(c)

	// 检查是否有训练队列
	var count int64
	database.DB.Model(&models.TrainingQueue{}).
		Where("unit_id = ?", unitID).
		Count(&count)

	if count > 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "训练单元内还有训练队列，无法删除",
		})
		return
	}

	if err := database.DB.Where("id = ? AND user_id = ?", unitID, userID).
		Delete(&models.TrainingUnit{}).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "删除训练单元失败",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "训练单元已删除",
	})
}

// Heartbeat Python客户端心跳（保持连接状态）
func (h *UnitHandler) Heartbeat(c *gin.Context) {
	unitID := c.Param("unit_id")
	userID := middleware.GetUserID(c)

	var unit models.TrainingUnit
	if err := database.DB.Where("id = ? AND user_id = ?", unitID, userID).
		First(&unit).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"success": false,
			"error":   "训练单元不存在",
		})
		return
	}

	// 更新心跳时间和连接状态
	now := time.Now()
	unit.LastHeartbeat = &now
	unit.ConnectionStatus = "connected"

	if err := database.DB.Save(&unit).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "更新心跳失败",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success":           true,
		"connection_status": unit.ConnectionStatus,
		"last_heartbeat":    unit.LastHeartbeat,
	})
}

// checkConnectionStatus 检查并更新连接状态（10秒无心跳则标记为断开）
func checkConnectionStatus(unit *models.TrainingUnit) {
	if unit.LastHeartbeat == nil {
		unit.ConnectionStatus = "disconnected"
		return
	}

	// 如果超过10秒没有心跳，标记为断开
	if time.Since(*unit.LastHeartbeat) > 10*time.Second {
		if unit.ConnectionStatus != "disconnected" {
			unit.ConnectionStatus = "disconnected"
			database.DB.Model(unit).Update("connection_status", "disconnected")
		}
	}
}
