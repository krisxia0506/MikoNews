package bot

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
    "github.com/larksuite/oapi-sdk-go/v3"
	larkim "github.com/larksuite/oapi-sdk-go/v3/service/im/v1"
)

// MessageContent 消息内容结构
type MessageContent struct {
	Text string `json:"text"`
}

// MessageCardContent 卡片消息内容结构
type MessageCardContent struct {
	Config   interface{} `json:"config"`
	Header   interface{} `json:"header"`
	Elements interface{} `json:"elements"`
}

// FeishuMessageService 飞书消息服务
type FeishuMessageService struct {
	client *lark.Client
}

// NewFeishuMessageService 创建一个新的消息服务
func NewFeishuMessageService(client *lark.Client) *FeishuMessageService {
	return &FeishuMessageService{
		client: client,
	}
}

// SendTextMessage 发送文本消息
func (s *FeishuMessageService) SendTextMessage(ctx context.Context, chatID string, text string) (*larkim.CreateMessageResp, error) {
	content := &MessageContent{
		Text: text,
	}

	contentStr, err := json.Marshal(content)
	if err != nil {
		log.Printf("消息内容序列化失败: %v", err)
		return nil, err
	}

	return s.createMessage(ctx, chatID, "text", string(contentStr))
}

// SendCardMessage 发送卡片消息
func (s *FeishuMessageService) SendCardMessage(ctx context.Context, chatID string, card *MessageCardContent) (*larkim.CreateMessageResp, error) {
	contentStr, err := json.Marshal(card)
	if err != nil {
		log.Printf("卡片消息内容序列化失败: %v", err)
		return nil, err
	}

	return s.createMessage(ctx, chatID, "interactive", string(contentStr))
}


// ReplyMessage 回复消息
func (s *FeishuMessageService) ReplyMessage(ctx context.Context, msgID, chatID, msgType, content string) (*larkim.ReplyMessageResp, error) {
	req := larkim.NewReplyMessageReqBuilder().
		MessageId(msgID).
		Body(larkim.NewReplyMessageReqBodyBuilder().
			MsgType(msgType).
			Content(content).
			Build()).
		Build()

	resp, err := s.client.Im.V1.Message.Reply(ctx, req)
	if err != nil {
		log.Printf("回复消息失败: %v", err)
		return nil, err
	}

	if !resp.Success() {
		log.Printf("回复消息API调用失败: %s", resp.Msg)
		return nil, fmt.Errorf("回复消息失败: %s", resp.Msg)
	}

	return resp, nil
}

// ReplyTextMessage 回复文本消息
func (s *FeishuMessageService) ReplyTextMessage(ctx context.Context, msgID string, text string) (*larkim.ReplyMessageResp, error) {
	content := &MessageContent{
		Text: text,
	}

	contentStr, err := json.Marshal(content)
	if err != nil {
		log.Printf("消息内容序列化失败: %v", err)
		return nil, err
	}

	return s.ReplyMessage(ctx, msgID, "", "text", string(contentStr))
}

// createMessage 创建并发送消息
func (s *FeishuMessageService) createMessage(ctx context.Context, chatID, msgType, content string) (*larkim.CreateMessageResp, error) {
	req := larkim.NewCreateMessageReqBuilder().
		ReceiveIdType("chat_id").
		Body(larkim.NewCreateMessageReqBodyBuilder().
			ReceiveId(chatID).
			MsgType(msgType).
			Content(content).
			Build()).
		Build()

	resp, err := s.client.Im.V1.Message.Create(ctx, req)
	if err != nil {
		log.Printf("发送消息失败: %v", err)
		return nil, err
	}

	if !resp.Success() {
		log.Printf("发送消息API调用失败: %s", resp.Msg)
		return nil, fmt.Errorf("发送消息失败: %s", resp.Msg)
	}

	return resp, nil
}
