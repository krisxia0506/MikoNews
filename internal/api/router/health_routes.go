package router

import (
	"MikoNews/internal/pkg/response"
	"github.com/gin-gonic/gin"
	"time"
)

// setupHealthRoutes 配置健康检查路由
func setupHealthRoutes(router *gin.Engine) {
	// 健康检查路由
	router.GET("/health", func(c *gin.Context) {
		response.Success(c, map[string]string{
			"status": "ok",
			"time":   time.Now().Format(time.RFC3339),
		})
	})
	
	// 简单的ping-pong路由
	router.GET("/ping", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "pong",
		})
	})
} 