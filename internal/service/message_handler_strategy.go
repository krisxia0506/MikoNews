package service

import (
	"context"

	larkim "github.com/larksuite/oapi-sdk-go/v3/service/im/v1"
)

// MessageHandlerStrategy defines the interface for handling different P2P message types based on content.
type MessageHandlerStrategy interface {
	// ShouldHandle determines if this strategy is applicable to the given message event.
	// It should check message type, content patterns, etc.
	ShouldHandle(ctx context.Context, event *larkim.P2MessageReceiveV1) bool

	// Handle processes the message according to the specific strategy logic.
	Handle(ctx context.Context, event *larkim.P2MessageReceiveV1) error
}
