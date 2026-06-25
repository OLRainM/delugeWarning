package api

import (
	"delugewarning/internal/middleware"
	"delugewarning/internal/model"

	"github.com/gin-gonic/gin"
)

// Register 挂载所有路由。
func (h *Handler) Register(r *gin.Engine) {
	r.GET("/healthz", h.health)

	v1 := r.Group("/api/v1")
	{
		// 公开
		v1.POST("/auth/wx-login", h.wxLogin)
		// 设备上报：走设备签名校验，不需要 JWT
		v1.POST("/devices/:id/readings", h.ingestReading)

		// 需登录
		authed := v1.Group("")
		authed.Use(middleware.Auth(h.jwt))
		{
			authed.GET("/profile", h.profile)

			// 设备/趋势（网格员查看）
			gw := authed.Group("")
			gw.Use(middleware.RequireRole(model.RoleGridworker))
			{
				gw.GET("/devices", h.listDevices)
				gw.GET("/devices/:id/readings/latest", h.latestReading)
				gw.GET("/devices/:id/trend", h.deviceTrend)

				// 预警复核
				gw.GET("/alerts", h.listAlerts)
				gw.GET("/alerts/:id", h.getAlert)
				gw.POST("/alerts/:id/review", h.reviewAlert)
				gw.POST("/alerts", h.createManualAlert)
				gw.POST("/alerts/:id/archive", h.archiveAlert)

				// 任务处置
				gw.GET("/tasks", h.listTasks)
				gw.POST("/tasks/:id/confirm", h.confirmTask)
				gw.POST("/tasks/:id/handle", h.handleTask)
				gw.POST("/uploads/presign", h.presignUpload)
			}

			// 村民端
			authed.GET("/village/alerts", h.villageAlerts)
			authed.GET("/alerts/:id/broadcast", h.broadcast)
			authed.POST("/reports", h.submitReport)
			authed.GET("/guides", h.guides)
		}
	}
}
