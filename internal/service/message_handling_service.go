package service

import (
	"context"

	larkim "github.com/larksuite/oapi-sdk-go/v3/service/im/v1"
)

// MessageHandlingService defines the interface for processing received P2P messages using strategies.
type MessageHandlingService interface {
	// ProcessReceivedMessage selects and executes the appropriate strategy for a given message event.
	ProcessReceivedMessage(ctx context.Context, event *larkim.P2MessageReceiveV1) error
}
