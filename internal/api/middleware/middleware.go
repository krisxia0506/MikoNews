package middleware

import (
	"MikoNews/internal/pkg/logger"
	"MikoNews/internal/pkg/response"
	"bytes"
	"io"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// RequestID 为每个请求添加唯一的请求ID
func RequestID() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 从请求头获取请求ID，如果没有则生成一个
		requestID := c.GetHeader("X-Request-ID")
		if requestID == "" {
			requestID = generateRequestID()
		}

		// 设置请求ID
		c.Set("RequestID", requestID)
		c.Header("X-Request-ID", requestID)

		c.Next()
	}
}

// Logger 日志中间件，记录请求信息
func Logger() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 开始时间
		start := time.Now()

		// 获取请求ID
		requestID, exists := c.Get("RequestID")
		if !exists {
			requestID = "unknown"
		}

		// 读取请求体
		var requestBody []byte
		if c.Request.Body != nil {
			requestBody, _ = io.ReadAll(c.Request.Body)
			// 重置请求体，以便后续处理
			c.Request.Body = io.NopCloser(bytes.NewBuffer(requestBody))
		}

		// 处理请求
		c.Next()

		// 计算耗时
		latency := time.Since(start)

		// 使用 zap 记录请求信息
		logger.Info("Request processed",
			zap.String("requestID", requestID.(string)),
			zap.Int("status", c.Writer.Status()),
			zap.Duration("latency", latency),
			zap.String("clientIP", c.ClientIP()),
			zap.String("method", c.Request.Method),
			zap.String("path", c.Request.URL.Path),
		)

		// 如果发生错误，记录详细信息
		if len(c.Errors) > 0 {
			for _, e := range c.Errors {
				logger.Error("Gin error", zap.String("requestID", requestID.(string)), zap.Error(e.Err))
			}
		}
	}
}

// Recovery 错误恢复中间件，处理请求过程中的panic
func Recovery() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				// 记录panic错误
				logger.Error("Panic recovered", zap.Any("error", err), zap.Stack("stack"))

				// 返回500错误
				response.Error(c, http.StatusInternalServerError, "服务器内部错误")
				c.Abort()
			}
		}()

		c.Next()
	}
}

// CORS 跨域资源共享中间件
func CORS() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Origin, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization")
		c.Writer.Header().Set("Access-Control-Expose-Headers", "Content-Length")
		c.Writer.Header().Set("Access-Control-Max-Age", "86400")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		c.Next()
	}
}

// generateRequestID 生成请求ID
func generateRequestID() string {
	// 简单实现，实际项目中可以使用UUID或其他更复杂的方法
	return time.Now().Format("20060102150405.000") + "-" + randomString(8)
}

// randomString 生成随机字符串
func randomString(n int) string {
	// 简单实现，实际项目中可以使用更安全的方法
	var letterRunes = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789")
	b := make([]rune, n)
	for i := range b {
		b[i] = letterRunes[time.Now().UnixNano()%int64(len(letterRunes))]
		time.Sleep(time.Nanosecond)
	}
	return string(b)
}
