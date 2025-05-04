package bot

import (
	"MikoNews/internal/config"
	"MikoNews/internal/database"
	"MikoNews/internal/pkg/logger"
	"MikoNews/internal/repository"
	"MikoNews/internal/repository/impl/mysql"
	"MikoNews/internal/service"
	articleServiceImpl "MikoNews/internal/service/impl"
	mh "MikoNews/internal/service/impl/messagehandler"
	"context"

	lark "github.com/larksuite/oapi-sdk-go/v3"
	larkcore "github.com/larksuite/oapi-sdk-go/v3/core"
	larkws "github.com/larksuite/oapi-sdk-go/v3/ws"
)

// FeishuBot 结构体表示飞书机器人
type FeishuBot struct {
	client                 *larkws.Client
	conf                   *config.FeishuConfig
	db                     *database.DB
	apiClient              *lark.Client
	articleRepo            repository.ArticleRepository
	articleService         service.ArticleService
	messageHandlingService service.MessageHandlingService
	dispatcher             *FeishuEventDispatcher
	msgService             service.FeishuMessageService
}

// NewFeishuBot 创建一个新的 FeishuBot 实例
func NewFeishuBot(conf *config.FeishuConfig, db *database.DB) *FeishuBot {

	// Create API client
	apiClient := lark.NewClient(conf.AppID, conf.AppSecret,
		lark.WithLogLevel(larkcore.LogLevelDebug),
		lark.WithLogReqAtDebug(true),
	)

	// --- Create Dependencies ---
	// Repository
	articleRepo := mysql.NewArticleRepository(db.DB)

	// Services
	articleService := articleServiceImpl.NewArticleService(articleRepo)
	msgService := articleServiceImpl.NewFeishuMessageService(apiClient)
	feishuContactService := articleServiceImpl.NewFeishuContactService(apiClient)
	// Message Handling Strategies (Use alias 'mh')
	submissionStrategy := mh.NewSubmissionHandlerStrategy(articleService, msgService, feishuContactService)
	defaultStrategy := mh.NewDefaultMessageHandlerStrategy()

	// Message Handling Service (Use alias 'mh')
	messageHandlingService := mh.NewMessageHandlingService(submissionStrategy, defaultStrategy)

	// --- Create Bot and Dispatcher ---
	bot := &FeishuBot{
		conf:                   conf,
		db:                     db,
		apiClient:              apiClient,
		articleRepo:            articleRepo,
		articleService:         articleService,
		messageHandlingService: messageHandlingService,
		msgService:             msgService,
	}

	// Event Dispatcher (injects the handling service)
	bot.dispatcher = NewFeishuEventDispatcher(conf, bot, messageHandlingService)

	// WebSocket Client
	bot.client = larkws.NewClient(conf.AppID, conf.AppSecret,
		larkws.WithEventHandler(bot.dispatcher.GetEventDispatcher()),
		larkws.WithLogLevel(larkcore.LogLevelDebug),
	)

	logger.Info("FeishuBot initialized")
	return bot
}

// Start 启动飞书机器人，建立 WebSocket 长连接
func (b *FeishuBot) Start(ctx context.Context) error {
	return b.client.Start(ctx) // 启动客户端
}

// GetClient 获取API客户端
func (b *FeishuBot) GetClient() *lark.Client {
	return b.apiClient
}

// GetArticleRepo 获取文章仓库
func (b *FeishuBot) GetArticleRepo() repository.ArticleRepository {
	return b.articleRepo
}

// GetArticleService returns the article service
func (b *FeishuBot) GetArticleService() service.ArticleService {
	return b.articleService
}

// GetMessageHandlingService returns the message handling service
func (b *FeishuBot) GetMessageHandlingService() service.MessageHandlingService {
	return b.messageHandlingService
}

// GetMessageService 获取消息服务
func (b *FeishuBot) GetMessageService() service.FeishuMessageService {
	return b.msgService
}

// GetConfig 返回配置
func (b *FeishuBot) GetConfig() *config.FeishuConfig {
	return b.conf
}

// GetDB 返回数据库连接
func (b *FeishuBot) GetDB() *database.DB {
	return b.db
}

// GetEventDispatcher 获取事件分发器
func (b *FeishuBot) GetEventDispatcher() *FeishuEventDispatcher {
	return b.dispatcher
}
