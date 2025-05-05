package router

import (
	"MikoNews/internal/api/handler"
	"MikoNews/internal/api/middleware"
	"MikoNews/internal/config"

	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"     // swagger embed files
	ginSwagger "github.com/swaggo/gin-swagger" // gin-swagger middleware
)

// Setup 配置路由
func Setup(
	engine *gin.Engine,
	articleHandler *handler.ArticleHandler,
	config *config.Config,
) {
	// 使用中间件
	engine.Use(middleware.RequestID())
	engine.Use(middleware.Logger())
	engine.Use(middleware.Recovery())
	engine.Use(middleware.CORS())

	// 健康检查路由
	setupHealthRoutes(engine)

	// API v1 路由组
	v1 := engine.Group("/api/v1")
	{
		// 文章相关路由
		setupArticleRoutes(v1, articleHandler, config)

		// 其他路由...
		// setupUserRoutes(v1, userHandler)
		// setupCommentRoutes(v1, commentHandler)
	}

	// 添加 Swagger UI 路由
	engine.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
}

// setupArticleRoutes 配置文章相关路由
func setupArticleRoutes(
	router *gin.RouterGroup,
	handler *handler.ArticleHandler,
	config *config.Config,
) {
	// 文章路由组
	articles := router.Group("/articles")
	{
		// 获取特定文章
		articles.GET("/:id", handler.GetArticle)
	}
}
