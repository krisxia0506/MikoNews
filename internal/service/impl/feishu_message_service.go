package impl

import (
	"MikoNews/internal/service"
	"context"
	"encoding/json"
	"fmt"

	"MikoNews/internal/pkg/logger"

	"go.uber.org/zap"

	lark "github.com/larksuite/oapi-sdk-go/v3"
	larkim "github.com/larksuite/oapi-sdk-go/v3/service/im/v1"
)

// MessageContent 消息内容结构
type MessageContent struct {
	Text string `json:"text"`
}

// feishuMessageServiceImpl implements service.FeishuMessageService
type feishuMessageServiceImpl struct {
	client *lark.Client
}

// NewFeishuMessageService creates a new Feishu message service implementation.
func NewFeishuMessageService(client *lark.Client) service.FeishuMessageService {
	return &feishuMessageServiceImpl{
		client: client,
	}
}

// SendTextMessage 发送文本消息
func (s *feishuMessageServiceImpl) SendTextMessage(ctx context.Context, chatID string, text string) (*larkim.CreateMessageResp, error) {
	content := &MessageContent{
		Text: text,
	}

	contentStr, err := json.Marshal(content)
	if err != nil {
		logger.Error("Failed to marshal text message content", zap.Error(err))
		return nil, fmt.Errorf("序列化文本消息失败: %w", err)
	}

	return s.createMessage(ctx, chatID, larkim.MsgTypeText, string(contentStr))
}

// SendCardMessage 发送卡片消息
// Note: The card parameter type is now service.MessageCardContent from the interface package
func (s *feishuMessageServiceImpl) SendCardMessage(ctx context.Context, chatID string, card *service.MessageCardContent) (*larkim.CreateMessageResp, error) {
	contentStr, err := json.Marshal(card)
	if err != nil {
		logger.Error("Failed to marshal card message content", zap.Error(err))
		return nil, fmt.Errorf("序列化卡片消息失败: %w", err)
	}

	return s.createMessage(ctx, chatID, larkim.MsgTypeInteractive, string(contentStr))
}

// ReplyMessage 回复消息 (internal helper, not part of the interface directly shown here)
func (s *feishuMessageServiceImpl) replyMessage(ctx context.Context, msgID, msgType, content string) (*larkim.ReplyMessageResp, error) {
	req := larkim.NewReplyMessageReqBuilder().
		MessageId(msgID).
		Body(larkim.NewReplyMessageReqBodyBuilder().
			MsgType(msgType).
			Content(content).
			Build()).
		Build()

	resp, err := s.client.Im.V1.Message.Reply(ctx, req)
	if err != nil {
		logger.Error("Failed to call Feishu reply API", zap.String("messageID", msgID), zap.Error(err))
		return nil, fmt.Errorf("飞书 API 调用失败: %w", err)
	}

	if !resp.Success() {
		logger.Error("Feishu reply API call unsuccessful",
			zap.String("messageID", msgID),
			zap.Int("code", resp.Code),
			zap.String("msg", resp.Msg),
		)
		return nil, fmt.Errorf("回复消息失败: %s (code: %d)", resp.Msg, resp.Code)
	}
	logger.Debug("Successfully replied to message", zap.String("messageID", msgID))
	return resp, nil
}

// ReplyTextMessage 回复文本消息
func (s *feishuMessageServiceImpl) ReplyTextMessage(ctx context.Context, msgID string, text string) (*larkim.ReplyMessageResp, error) {
	content := &MessageContent{
		Text: text,
	}

	contentStr, err := json.Marshal(content)
	if err != nil {
		logger.Error("Failed to marshal reply text message content", zap.String("messageID", msgID), zap.Error(err))
		return nil, fmt.Errorf("序列化回复文本消息失败: %w", err)
	}
	return s.replyMessage(ctx, msgID, larkim.MsgTypeText, string(contentStr))
}

// createMessage 创建并发送消息 (internal helper)
func (s *feishuMessageServiceImpl) createMessage(ctx context.Context, chatID, msgType, content string) (*larkim.CreateMessageResp, error) {
	req := larkim.NewCreateMessageReqBuilder().
		ReceiveIdType(larkim.ReceiveIdTypeChatId).
		Body(larkim.NewCreateMessageReqBodyBuilder().
			ReceiveId(chatID).
			MsgType(msgType).
			Content(content).
			Build()).
		Build()

	resp, err := s.client.Im.V1.Message.Create(ctx, req)
	if err != nil {
		logger.Error("Failed to call Feishu create message API", zap.String("chatID", chatID), zap.String("msgType", msgType), zap.Error(err))
		return nil, fmt.Errorf("飞书 API 调用失败: %w", err)
	}

	if !resp.Success() {
		logger.Error("Feishu create message API call unsuccessful",
			zap.String("chatID", chatID),
			zap.String("msgType", msgType),
			zap.Int("code", resp.Code),
			zap.String("msg", resp.Msg),
		)
		return nil, fmt.Errorf("发送消息失败: %s (code: %d)", resp.Msg, resp.Code)
	}
	logger.Debug("Successfully created message", zap.String("chatID", chatID), zap.String("msgType", msgType))
	return resp, nil
}

// Ensure feishuMessageServiceImpl implements FeishuMessageService
var _ service.FeishuMessageService = (*feishuMessageServiceImpl)(nil)
