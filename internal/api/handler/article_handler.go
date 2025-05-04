package handler

import (
	"MikoNews/internal/pkg/errors"
	"MikoNews/internal/pkg/logger"
	"MikoNews/internal/pkg/response"
	"MikoNews/internal/service"
	"strconv"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// ArticleHandler 处理文章相关的HTTP请求
type ArticleHandler struct {
	articleService service.ArticleService
}

// NewArticleHandler 创建文章处理器
func NewArticleHandler(articleService service.ArticleService) *ArticleHandler {
	return &ArticleHandler{
		articleService: articleService,
	}
}

// GetArticle godoc
// @Summary      获取指定ID的文章
// @Description  根据提供的文章ID获取详细信息
// @Tags         Articles
// @Accept       json
// @Produce      json
// @Param        id   path      int  true  "文章ID"
// @Success      200  {object}  response.Response{data=model.Article} "成功响应"
// @Failure      400  {object}  response.Response "无效的文章ID"
// @Failure      404  {object}  response.Response "文章未找到"
// @Failure      500  {object}  response.Response "服务器内部错误"
// @Router       /articles/{id} [get]
func (h *ArticleHandler) GetArticle(c *gin.Context) {
	// 获取文章ID
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		response.BadRequest(c, "无效的文章ID")
		return
	}

	// 获取文章
	article, err := h.articleService.FindArticleByID(c.Request.Context(), id)
	if err != nil {
		logger.Error("获取文章失败", zap.Error(err), zap.Int64("id", id))
		handleError(c, err)
		return
	}

	response.Success(c, article)
}

// 处理错误
func handleError(c *gin.Context, err error) {
	// 尝试转换为应用错误
	if appErr, ok := err.(*errors.AppError); ok {
		response.Error(c, appErr.HTTPStatus, appErr.Message, appErr.Code)
		return
	}

	// 默认为内部服务器错误
	response.InternalServerError(c, "服务器内部错误")
}
