package errors

import (
	"fmt"
	"net/http"
)

// AppError 应用程序错误
type AppError struct {
	Code       int    // 错误码
	Message    string // 错误消息
	HTTPStatus int    // HTTP状态码
	Err        error  // 原始错误
}

// Error 实现error接口
func (e *AppError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("%s: %v", e.Message, e.Err)
	}
	return e.Message
}

// Unwrap 返回原始错误
func (e *AppError) Unwrap() error {
	return e.Err
}

// 错误码定义
const (
	// 通用错误（1000-1999）
	ErrCodeInternal         = 1000
	ErrCodeInvalidRequest   = 1001
	ErrCodeUnauthorized     = 1002
	ErrCodeForbidden        = 1003
	ErrCodeNotFound         = 1004
	ErrCodeConflict         = 1005
	ErrCodeValidationFailed = 1006

	// 数据库错误（2000-2999）
	ErrCodeDBConnection = 2000
	ErrCodeDBQuery      = 2001
	ErrCodeDBInsert     = 2002
	ErrCodeDBUpdate     = 2003
	ErrCodeDBDelete     = 2004

	// 飞书API错误（3000-3999）
	ErrCodeFeishuAPI     = 3000
	ErrCodeFeishuAuth    = 3001
	ErrCodeFeishuMessage = 3002

	// 文章错误（4000-4999）
	ErrCodeArticleNotFound      = 4000
	ErrCodeArticleInvalidStatus = 4001
	ErrCodeArticlePublishFailed = 4002
)

// 创建各类错误的便捷函数

// NewInternalError 创建内部服务器错误
func NewInternalError(message string, err error) *AppError {
	return &AppError{
		Code:       ErrCodeInternal,
		Message:    message,
		HTTPStatus: http.StatusInternalServerError,
		Err:        err,
	}
}

// NewInvalidRequestError 创建无效请求错误
func NewInvalidRequestError(message string, err error) *AppError {
	return &AppError{
		Code:       ErrCodeInvalidRequest,
		Message:    message,
		HTTPStatus: http.StatusBadRequest,
		Err:        err,
	}
}

// NewUnauthorizedError 创建未授权错误
func NewUnauthorizedError(message string, err error) *AppError {
	return &AppError{
		Code:       ErrCodeUnauthorized,
		Message:    message,
		HTTPStatus: http.StatusUnauthorized,
		Err:        err,
	}
}

// NewForbiddenError 创建禁止访问错误
func NewForbiddenError(message string, err error) *AppError {
	return &AppError{
		Code:       ErrCodeForbidden,
		Message:    message,
		HTTPStatus: http.StatusForbidden,
		Err:        err,
	}
}

// NewNotFoundError 创建资源不存在错误
func NewNotFoundError(message string, err error) *AppError {
	return &AppError{
		Code:       ErrCodeNotFound,
		Message:    message,
		HTTPStatus: http.StatusNotFound,
		Err:        err,
	}
}

// NewDBError 创建数据库错误
func NewDBError(message string, err error, code int) *AppError {
	return &AppError{
		Code:       code,
		Message:    message,
		HTTPStatus: http.StatusInternalServerError,
		Err:        err,
	}
}

// NewArticleError 创建文章相关错误
func NewArticleError(message string, err error, code int, httpStatus int) *AppError {
	return &AppError{
		Code:       code,
		Message:    message,
		HTTPStatus: httpStatus,
		Err:        err,
	}
}
 