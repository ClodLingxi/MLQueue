package routes

import (
	"MLQueue/internal/handlers"
	"MLQueue/internal/middleware"
	"MLQueue/internal/queue"

	"github.com/gin-gonic/gin"
)

func SetupRouter(qm *queue.Manager) *gin.Engine {
	router := gin.Default()

	// Global middleware
	router.Use(middleware.CORSMiddleware())

	// Health check
	router.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	// API v1 routes
	v1 := router.Group("/v1")
	{
		// Authentication required for all routes
		v1.Use(middleware.AuthMiddleware())

		// Task routes
		taskHandler := handlers.NewTaskHandler(qm)
		tasks := v1.Group("/tasks")
		{
			tasks.POST("", middleware.RateLimitMiddleware(false), taskHandler.CreateTask)
			tasks.POST("/batch", middleware.RateLimitMiddleware(true), taskHandler.BatchCreateTasks)
			tasks.GET("", middleware.RateLimitMiddleware(false), taskHandler.ListTasks)
			tasks.GET("/:task_id", middleware.RateLimitMiddleware(false), taskHandler.GetTask)
			tasks.PATCH("/:task_id/priority", middleware.RateLimitMiddleware(false), taskHandler.UpdateTaskPriority)
			tasks.POST("/:task_id/cancel", middleware.RateLimitMiddleware(false), taskHandler.CancelTask)
			tasks.POST("/:task_id/result", middleware.RateLimitMiddleware(false), taskHandler.UploadResult)
		}

		// Queue routes
		queueHandler := handlers.NewQueueHandler(qm)
		queueGroup := v1.Group("/queue")
		{
			queueGroup.GET("/status", middleware.RateLimitMiddleware(false), queueHandler.GetQueueStatus)
			queueGroup.POST("/reorder", middleware.RateLimitMiddleware(false), queueHandler.ReorderQueue)
			queueGroup.POST("/pause", middleware.RateLimitMiddleware(false), queueHandler.PauseQueue)
			queueGroup.POST("/resume", middleware.RateLimitMiddleware(false), queueHandler.ResumeQueue)
		}

		// Config routes
		configHandler := handlers.NewConfigHandler()
		configs := v1.Group("/configs")
		{
			configs.GET("/templates", middleware.RateLimitMiddleware(false), configHandler.GetTemplates)
			configs.POST("/templates", middleware.RateLimitMiddleware(false), configHandler.CreateTemplate)
		}

		// Statistics routes
		statsHandler := handlers.NewStatisticsHandler()
		statistics := v1.Group("/statistics")
		{
			statistics.GET("/tasks", middleware.RateLimitMiddleware(false), statsHandler.GetTaskStatistics)
		}

		// Task logs
		v1.GET("/tasks/:task_id/logs", middleware.RateLimitMiddleware(false), statsHandler.GetTaskLogs)
	}

	return router
}
