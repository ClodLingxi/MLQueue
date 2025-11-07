package handlers

import (
	"net/http"

	"MLQueue/internal/database"
	"MLQueue/internal/middleware"
	"MLQueue/internal/models"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type ConfigHandler struct{}

func NewConfigHandler() *ConfigHandler {
	return &ConfigHandler{}
}

// GetTemplates retrieves configuration templates
func (h *ConfigHandler) GetTemplates(c *gin.Context) {
	userID := middleware.GetUserID(c)

	var templates []models.ConfigTemplate
	database.DB.Where("user_id = ?", userID).Find(&templates)

	templateList := make([]map[string]interface{}, len(templates))
	for i, t := range templates {
		templateList[i] = map[string]interface{}{
			"name":   t.Name,
			"config": t.Config,
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"success":   true,
		"templates": templateList,
	})
}

// CreateTemplate creates a configuration template
func (h *ConfigHandler) CreateTemplate(c *gin.Context) {
	userID := middleware.GetUserID(c)

	var req struct {
		Name        string                 `json:"name" binding:"required"`
		Config      map[string]interface{} `json:"config" binding:"required"`
		Description string                 `json:"description"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "无效的请求参数",
			"code":    "INVALID_CONFIG",
		})
		return
	}

	template := models.ConfigTemplate{
		ID:          "template_" + uuid.New().String()[:6],
		Name:        req.Name,
		Config:      models.JSONB(req.Config),
		Description: req.Description,
		UserID:      userID,
	}

	if err := database.DB.Create(&template).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "创建模板失败",
			"code":    "INTERNAL_ERROR",
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"success":     true,
		"template_id": template.ID,
		"name":        template.Name,
	})
}
