package service

import (
	"MikoNews/internal/config"
	"context"

	lark "github.com/larksuite/oapi-sdk-go/v3"
	larkim "github.com/larksuite/oapi-sdk-go/v3/service/im/v1"
)

// MessageCardContent represents the structure for Lark interactive message cards.
// It uses interface{} for flexibility as the exact structure can vary greatly.
type MessageCardContent struct {
	Config   interface{} `json:"config,omitempty"`
	Header   interface{} `json:"header,omitempty"`
	Elements interface{} `json:"elements,omitempty"`
	// Add other top-level card fields if needed, e.g., CardLink
}

// FeishuMessageService defines the interface for sending messages via Feishu/Lark.
type FeishuMessageService interface {
	// SendTextMessage sends a plain text message to a specified chat ID.
	SendTextMessage(ctx context.Context, chatID string, text string) (*larkim.CreateMessageResp, error)

	// SendCardMessage sends an interactive card message to a specified chat ID.
	SendCardMessage(ctx context.Context, chatID string, card *MessageCardContent) (*larkim.CreateMessageResp, error)

	// ReplyTextMessage replies to a specific message with plain text.
	ReplyTextMessage(ctx context.Context, msgID string, text string) (*larkim.ReplyMessageResp, error)

	// ReplyCardMessage replies to a specific message with an interactive card.
	// Note: The implementation needs this method added if required.
	// ReplyCardMessage(ctx context.Context, msgID string, card *MessageCardContent) (*larkim.ReplyMessageResp, error)

	// TODO: Consider adding methods for updating cards, sending other message types etc. if needed.
}

// FeishuMessageServiceProvider defines the type for the constructor function
type FeishuMessageServiceProvider func(conf *config.FeishuConfig, client *lark.Client) FeishuMessageService
