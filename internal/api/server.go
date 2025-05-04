package api

import (
	"MikoNews/internal/api/handler"
	"MikoNews/internal/api/router"
	"MikoNews/internal/config"
	"MikoNews/internal/database"
	"MikoNews/internal/repository/impl/mysql"
	"MikoNews/internal/service/impl"
	"fmt"
	"log"

	"github.com/gin-gonic/gin"
)

// Server 是API服务器结构体
type Server struct {
	config  *config.Config // 配置对象
	db      *database.DB   // 数据库连接
	engine  *gin.Engine    // Gin引擎
	started bool           // 是否已启动
}

// New 创建新的API服务器
func New(config *config.Config, db *database.DB) *Server {
	// 创建Gin引擎
	gin.SetMode(gin.ReleaseMode) // 生产模式
	engine := gin.New()

	s := &Server{
		config: config,
		db:     db,
		engine: engine,
	}

	// 初始化服务器
	s.init()

	return s
}

// init 初始化服务器
func (s *Server) init() {
	// 在这里创建 Repository 和 Service
	articleRepo := mysql.NewArticleRepository(s.db.DB)
	articleService := impl.NewArticleService(articleRepo)

	// 创建处理器
	articleHandler := handler.NewArticleHandler(articleService)

	// 配置路由
	router.Setup(s.engine, articleHandler, s.config)
}

// Start 启动HTTP服务器
func (s *Server) Start() error {
	if s.started {
		return fmt.Errorf("服务器已经启动")
	}

	s.started = true
	addr := fmt.Sprintf(":%d", s.config.Server.Port)
	log.Printf("API服务器开始监听 %s", addr)
	return s.engine.Run(addr)
}

// GetEngine 获取Gin引擎，方便测试
func (s *Server) GetEngine() *gin.Engine {
	return s.engine
}
