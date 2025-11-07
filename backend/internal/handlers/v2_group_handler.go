package handlers

import (
	"net/http"

	"MLQueue/internal/database"
	"MLQueue/internal/middleware"
	"MLQueue/internal/models"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type GroupHandler struct{}

func NewGroupHandler() *GroupHandler {
	return &GroupHandler{}
}

// CreateGroup 创建组（由Python客户端调用）
func (h *GroupHandler) CreateGroup(c *gin.Context) {
	userID := middleware.GetUserID(c)

	var req struct {
		Name        string `json:"name" binding:"required"`
		Description string `json:"description"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "无效的请求参数",
		})
		return
	}

	group := models.Group{
		ID:          "group_" + uuid.New().String()[:8],
		Name:        req.Name,
		Description: req.Description,
		UserID:      userID,
	}

	if err := database.DB.Create(&group).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "创建组失败",
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"success":  true,
		"group_id": group.ID,
		"name":     group.Name,
	})
}

// ListGroups 列出所有组
func (h *GroupHandler) ListGroups(c *gin.Context) {
	userID := middleware.GetUserID(c)

	var groups []models.Group
	if err := database.DB.Where("user_id = ?", userID).
		Order("created_at DESC").
		Find(&groups).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "查询组失败",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"groups":  groups,
	})
}

// GetGroup 获取组详情
func (h *GroupHandler) GetGroup(c *gin.Context) {
	groupID := c.Param("group_id")
	userID := middleware.GetUserID(c)

	var group models.Group
	if err := database.DB.Where("id = ? AND user_id = ?", groupID, userID).
		First(&group).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"success": false,
			"error":   "组不存在",
		})
		return
	}

	// 统计训练单元数量
	var unitCount int64
	database.DB.Model(&models.TrainingUnit{}).
		Where("group_id = ?", groupID).
		Count(&unitCount)

	c.JSON(http.StatusOK, gin.H{
		"success":    true,
		"group":      group,
		"unit_count": unitCount,
	})
}

// UpdateGroup 更新组信息
func (h *GroupHandler) UpdateGroup(c *gin.Context) {
	groupID := c.Param("group_id")
	userID := middleware.GetUserID(c)

	var req struct {
		Name        string `json:"name"`
		Description string `json:"description"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "无效的请求参数",
		})
		return
	}

	var group models.Group
	if err := database.DB.Where("id = ? AND user_id = ?", groupID, userID).
		First(&group).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"success": false,
			"error":   "组不存在",
		})
		return
	}

	if req.Name != "" {
		group.Name = req.Name
	}
	group.Description = req.Description

	if err := database.DB.Save(&group).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "更新组失败",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"group":   group,
	})
}

// DeleteGroup 删除组
func (h *GroupHandler) DeleteGroup(c *gin.Context) {
	groupID := c.Param("group_id")
	userID := middleware.GetUserID(c)

	// 检查是否有训练单元
	var count int64
	database.DB.Model(&models.TrainingUnit{}).
		Where("group_id = ?", groupID).
		Count(&count)

	//if count > 0 {
	//	c.JSON(http.StatusBadRequest, gin.H{
	//		"success": false,
	//		"error":   "组内还有训练单元，无法删除",
	//	})
	//	return
	//}

	if err := database.DB.Where("id = ? AND user_id = ?", groupID, userID).
		Delete(&models.Group{}).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "删除组失败",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "组已删除",
	})
}
