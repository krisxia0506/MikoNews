package response

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// Response 是API响应的标准格式
type Response struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data,omitempty"`
	Error   string      `json:"error,omitempty"`
	Code    int         `json:"code,omitempty"`
}

// Success 响应成功
func Success(c *gin.Context, data interface{}) {
	c.JSON(http.StatusOK, Response{
		Success: true,
		Data:    data,
	})
}

// Error 响应错误
func Error(c *gin.Context, statusCode int, errMsg string, errCode ...int) {
	resp := Response{
		Success: false,
		Error:   errMsg,
	}

	// 如果提供了错误码，则设置
	if len(errCode) > 0 {
		resp.Code = errCode[0]
	}

	c.JSON(statusCode, resp)
}

// Created 创建成功响应（HTTP 201）
func Created(c *gin.Context, data interface{}) {
	c.JSON(http.StatusCreated, Response{
		Success: true,
		Data:    data,
	})
}

// NoContent 无内容响应（HTTP 204）
func NoContent(c *gin.Context) {
	c.Status(http.StatusNoContent)
}

// BadRequest 无效请求错误响应（HTTP 400）
func BadRequest(c *gin.Context, errMsg string) {
	Error(c, http.StatusBadRequest, errMsg)
}

// Unauthorized 未授权错误响应（HTTP 401）
func Unauthorized(c *gin.Context, errMsg string) {
	Error(c, http.StatusUnauthorized, errMsg)
}

// Forbidden 禁止访问错误响应（HTTP 403）
func Forbidden(c *gin.Context, errMsg string) {
	Error(c, http.StatusForbidden, errMsg)
}

// NotFound 资源不存在错误响应（HTTP 404）
func NotFound(c *gin.Context, errMsg string) {
	Error(c, http.StatusNotFound, errMsg)
}

// InternalServerError 服务器内部错误响应（HTTP 500）
func InternalServerError(c *gin.Context, errMsg string) {
	Error(c, http.StatusInternalServerError, errMsg)
}
 