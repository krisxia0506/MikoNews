package messagehandler

import (
	"MikoNews/internal/pkg/logger"
	"MikoNews/internal/service"
	"context"
	"fmt"

	larkim "github.com/larksuite/oapi-sdk-go/v3/service/im/v1"
	"go.uber.org/zap"
)

type messageHandlingServiceImpl struct {
	strategies []service.MessageHandlerStrategy
}

// NewMessageHandlingService creates a new MessageHandlingService instance.
// It takes a list of strategies to be used.
func NewMessageHandlingService(strategies ...service.MessageHandlerStrategy) service.MessageHandlingService {
	return &messageHandlingServiceImpl{
		strategies: strategies,
	}
}

func (s *messageHandlingServiceImpl) ProcessReceivedMessage(ctx context.Context, event *larkim.P2MessageReceiveV1) error {
	messageID := "unknown"
	if event.Event != nil && event.Event.Message != nil && event.Event.Message.MessageId != nil {
		messageID = *event.Event.Message.MessageId
	}
	logger.Info("Processing received message", zap.String("messageID", messageID))

	for _, strategy := range s.strategies {
		if strategy.ShouldHandle(ctx, event) {
			logger.Info("Found matching strategy",
				zap.String("messageID", messageID),
				zap.String("strategy", fmt.Sprintf("%T", strategy)), // Log strategy type
			)
			err := strategy.Handle(ctx, event)
			if err != nil {
				logger.Error("Error handling message with strategy",
					zap.String("messageID", messageID),
					zap.String("strategy", fmt.Sprintf("%T", strategy)),
					zap.Error(err),
				)
				return fmt.Errorf("strategy %T failed: %w", strategy, err)
			}
			logger.Info("Message handled successfully by strategy",
				zap.String("messageID", messageID),
				zap.String("strategy", fmt.Sprintf("%T", strategy)),
			)
			return nil // Strategy handled the message, stop processing
		}
	}

	logger.Warn("No suitable strategy found for message", zap.String("messageID", messageID))
	// Optionally, implement a default action here if no strategy matches
	return nil // Or return an error if unhandled messages are considered an error
}
