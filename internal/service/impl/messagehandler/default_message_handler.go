package messagehandler

import (
	"MikoNews/internal/pkg/logger"
	"MikoNews/internal/service"
	"context"

	larkim "github.com/larksuite/oapi-sdk-go/v3/service/im/v1"
	"go.uber.org/zap"
)

// DefaultMessageHandlerStrategy handles any message not handled by other strategies.
type DefaultMessageHandlerStrategy struct {
}

// NewDefaultMessageHandlerStrategy creates a new default handler strategy.
func NewDefaultMessageHandlerStrategy() service.MessageHandlerStrategy {
	return &DefaultMessageHandlerStrategy{
	}
}

// ShouldHandle always returns true for P2P messages, acting as a fallback.
func (s *DefaultMessageHandlerStrategy) ShouldHandle(ctx context.Context, event *larkim.P2MessageReceiveV1) bool {
	// Only handle P2P messages that weren't handled by more specific strategies.
	return event.Event != nil && event.Event.Message != nil &&
		event.Event.Message.ChatType != nil && *event.Event.Message.ChatType == "p2p"
}

// Handle logs the unhandled message.
func (s *DefaultMessageHandlerStrategy) Handle(ctx context.Context, event *larkim.P2MessageReceiveV1) error {
	msgID := "unknown"
	if event.Event.Message.MessageId != nil {
		msgID = *event.Event.Message.MessageId
	}
	senderID := "unknown"
	if event.Event.Sender != nil && event.Event.Sender.SenderId != nil && event.Event.Sender.SenderId.OpenId != nil {
		senderID = *event.Event.Sender.SenderId.OpenId
	}

	logger.Warn("Unhandled P2P message",
		zap.String("messageID", msgID),
		zap.String("senderID", senderID),
		zap.Stringp("messageType", event.Event.Message.MessageType),
		zap.Stringp("contentPreview", event.Event.Message.Content), // Log raw content for debugging
	)

	// TODO: Optionally send a default reply like "Sorry, I didn't understand that command."

	return nil
}
