package bot

import (
	"MikoNews/internal/config"
	"MikoNews/internal/pkg/logger"
	"MikoNews/internal/service"
	"context"

	larkevent "github.com/larksuite/oapi-sdk-go/v3/event"
	"github.com/larksuite/oapi-sdk-go/v3/event/dispatcher"
	larkapplication "github.com/larksuite/oapi-sdk-go/v3/service/application/v6"
	larkim "github.com/larksuite/oapi-sdk-go/v3/service/im/v1"
)

// FeishuEventDispatcher 负责分发和处理飞书事件
type FeishuEventDispatcher struct {
	conf                   *config.FeishuConfig
	bot                    *FeishuBot
	messageHandlingService service.MessageHandlingService
}

// NewFeishuEventDispatcher 创建一个新的事件分发器
func NewFeishuEventDispatcher(conf *config.FeishuConfig, bot *FeishuBot, msgHandler service.MessageHandlingService) *FeishuEventDispatcher {
	return &FeishuEventDispatcher{
		conf:                   conf,
		bot:                    bot,
		messageHandlingService: msgHandler,
	}
}

// GetEventDispatcher 返回事件处理函数
func (d *FeishuEventDispatcher) GetEventDispatcher() *dispatcher.EventDispatcher {
	return dispatcher.NewEventDispatcher(d.conf.VerificationToken, d.conf.EncryptKey).
		OnP2MessageReceiveV1(func(ctx context.Context, event *larkim.P2MessageReceiveV1) error {
			if err := d.messageHandlingService.ProcessReceivedMessage(ctx, event); err != nil {
				logger.Error("Error processing received message event", "error", err)
				return nil
			}
			return nil
		}).
		OnCustomizedEvent("create_post", func(ctx context.Context, event *larkevent.EventReq) error {
			logger.Infof("收到自定义事件: %v", event)
			return nil
		}).
		OnP2BotMenuV6(func(ctx context.Context, event *larkapplication.P2BotMenuV6) error {
			logger.Infof("收到机器人菜单事件: %v", event)
			return nil
		})
}
