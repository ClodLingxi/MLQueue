package routes

import (
	"MLQueue/internal/handlers"
	"MLQueue/internal/middleware"

	"github.com/gin-gonic/gin"
)

// SetupV2Routes 配置V2版本路由（Python客户端驱动架构）
func SetupV2Routes(router *gin.Engine) {
	v2 := router.Group("/v2")
	{
		// 需要认证
		v2.Use(middleware.AuthMiddleware())

		// ============ 组管理 ============
		groupHandler := handlers.NewGroupHandler()
		groups := v2.Group("/groups")
		{
			groups.POST("", middleware.RateLimitMiddleware(false), groupHandler.CreateGroup)
			groups.GET("", middleware.RateLimitMiddleware(false), groupHandler.ListGroups)
			groups.GET("/:group_id", middleware.RateLimitMiddleware(false), groupHandler.GetGroup)
			groups.PUT("/:group_id", middleware.RateLimitMiddleware(false), groupHandler.UpdateGroup)
			groups.DELETE("/:group_id", middleware.RateLimitMiddleware(false), groupHandler.DeleteGroup)
		}

		// ============ 训练单元管理 ============
		unitHandler := handlers.NewUnitHandler()

		// 在组下创建训练单元
		v2.POST("/groups/:group_id/units", middleware.RateLimitMiddleware(false), unitHandler.CreateTrainingUnit)
		v2.GET("/groups/:group_id/units", middleware.RateLimitMiddleware(false), unitHandler.ListTrainingUnits)

		// 训练单元操作
		units := v2.Group("/units")
		{
			units.GET("/:unit_id", middleware.RateLimitMiddleware(false), unitHandler.GetTrainingUnit)
			units.PUT("/:unit_id", middleware.RateLimitMiddleware(false), unitHandler.UpdateTrainingUnit)
			units.DELETE("/:unit_id", middleware.RateLimitMiddleware(false), unitHandler.DeleteTrainingUnit)

			// Python客户端同步端点
			units.POST("/:unit_id/sync", middleware.RateLimitMiddleware(false), unitHandler.SyncTrainingUnit)
			// Python客户端心跳端点
			units.POST("/:unit_id/heartbeat", middleware.RateLimitMiddleware(false), unitHandler.Heartbeat)
		}

		// ============ 训练队列管理 ============
		queueHandler := handlers.NewQueueHandlerV2()

		// 在训练单元下创建队列
		v2.POST("/units/:unit_id/queues", middleware.RateLimitMiddleware(false), queueHandler.CreateTrainingQueue)
		v2.POST("/units/:unit_id/queues/batch", middleware.RateLimitMiddleware(true), queueHandler.BatchCreateQueues)
		v2.GET("/units/:unit_id/queues", middleware.RateLimitMiddleware(false), queueHandler.ListTrainingQueues)

		// 重新排序队列
		v2.POST("/units/:unit_id/queues/reorder", middleware.RateLimitMiddleware(false), queueHandler.ReorderQueues)

		// 训练队列操作
		queues := v2.Group("/queues")
		{
			queues.GET("/:queue_id", middleware.RateLimitMiddleware(false), queueHandler.GetTrainingQueue)
			queues.PUT("/:queue_id", middleware.RateLimitMiddleware(false), queueHandler.UpdateTrainingQueue)
			queues.DELETE("/:queue_id", middleware.RateLimitMiddleware(false), queueHandler.DeleteTrainingQueue)

			// Python客户端专用端点（执行控制）
			queues.POST("/:queue_id/start", middleware.RateLimitMiddleware(false), queueHandler.StartQueue)
			queues.POST("/:queue_id/complete", middleware.RateLimitMiddleware(false), queueHandler.CompleteQueue)
			queues.POST("/:queue_id/fail", middleware.RateLimitMiddleware(false), queueHandler.FailQueue)
		}
	}
}
