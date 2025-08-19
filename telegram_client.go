package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/gotd/td/telegram"
	"github.com/gotd/td/telegram/auth"
	"github.com/gotd/td/telegram/updates"
	"github.com/gotd/td/tg"
	"go.uber.org/zap"
)

const (
	// Telegram API credentials - 需要从 https://my.telegram.org 获取
	AppID   = 12345678        // 替换为你的 API ID
	AppHash = "your_api_hash" // 替换为你的 API Hash
	
	// 用户登录信息
	PhoneNumber = "+1234567890" // 替换为你的手机号
	
	// 目标群组信息
	GroupChatID = -1001234567890 // 替换为目标群组的 Chat ID (负数)
	GroupTitle  = "Test Group"   // 群组名称，用于日志显示
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// 创建日志
	logger, _ := zap.NewDevelopment()
	defer logger.Sync()

	// 创建 Telegram 客户端
	client := telegram.NewClient(AppID, AppHash, telegram.Options{
		Logger: logger,
	})

	// 处理系统信号，优雅退出
	go func() {
		c := make(chan os.Signal, 1)
		signal.Notify(c, syscall.SIGINT, syscall.SIGTERM)
		<-c
		fmt.Println("\n收到退出信号，正在关闭...")
		cancel()
	}()

	// 运行客户端
	err := client.Run(ctx, func(ctx context.Context) error {
		// 设置认证流程
		flow := auth.NewFlow(
			&Terminal{}, // 用于处理验证码输入
			auth.SendCodeOptions{},
		)

		// 执行认证
		if err := client.Auth().IfNecessary(ctx, flow); err != nil {
			return fmt.Errorf("认证失败: %w", err)
		}

		fmt.Println("登录成功！")

		// 获取当前用户信息
		user, err := client.Self(ctx)
		if err != nil {
			return fmt.Errorf("获取用户信息失败: %w", err)
		}
		firstName, _ := user.GetFirstName()
		lastName, _ := user.GetLastName()
		username, _ := user.GetUsername()
		fmt.Printf("当前用户: %s %s (@%s)\n", firstName, lastName, username)

		// 设置消息更新处理器
		updatesHandler := &UpdatesHandler{
			client:      client.API(),
			targetChatID: GroupChatID,
		}

		// 创建更新管理器
		updateManager := updates.New(updates.Config{
			Handler: updatesHandler,
			Logger:  logger,
		})

		// 启动更新监听
		fmt.Printf("开始监听群组 '%s' (ID: %d) 的消息...\n", GroupTitle, GroupChatID)
		return updateManager.Run(ctx, client.API(), user.GetID(), updates.AuthOptions{
			IsBot: false, // 标识为用户账号，非机器人
		})
	})

	if err != nil {
		logger.Fatal("客户端运行失败", zap.Error(err))
	}
}

// Terminal 实现用户输入接口
type Terminal struct{}

func (Terminal) Phone(_ context.Context) (string, error) {
	return PhoneNumber, nil
}

func (Terminal) Code(_ context.Context, sentTo *tg.AuthSentCode) (string, error) {
	fmt.Printf("验证码已发送到: %s\n", sentTo.GetPhoneCodeHash())
	fmt.Print("请输入验证码: ")
	
	var code string
	if _, err := fmt.Scanln(&code); err != nil {
		return "", err
	}
	return code, nil
}

func (Terminal) Password(_ context.Context) (string, error) {
	fmt.Print("请输入两步验证密码: ")
	
	var password string
	if _, err := fmt.Scanln(&password); err != nil {
		return "", err
	}
	return password, nil
}

func (Terminal) AcceptTermsOfService(_ context.Context, _ tg.HelpTermsOfService) error {
	// 自动接受服务条款
	return nil
}

func (Terminal) SignUp(_ context.Context) (auth.UserInfo, error) {
	// 如果需要注册新用户，返回用户信息
	return auth.UserInfo{
		FirstName: "Test",
		LastName:  "User",
	}, nil
}

// UpdatesHandler 处理消息更新
type UpdatesHandler struct {
	client       *tg.Client
	targetChatID int64
}

func (h *UpdatesHandler) Handle(ctx context.Context, u tg.UpdatesClass) error {
	switch update := u.(type) {
	case *tg.Updates:
		// 处理常规更新
		for _, upd := range update.Updates {
			if err := h.handleUpdate(ctx, upd); err != nil {
				fmt.Printf("处理更新失败: %v\n", err)
			}
		}
	case *tg.UpdateShort:
		// 处理短更新
		if err := h.handleUpdate(ctx, update.Update); err != nil {
			fmt.Printf("处理短更新失败: %v\n", err)
		}
	}
	return nil
}

func (h *UpdatesHandler) handleUpdate(ctx context.Context, update tg.UpdateClass) error {
	switch u := update.(type) {
	case *tg.UpdateNewMessage:
		return h.handleNewMessage(ctx, u.Message)
	case *tg.UpdateNewChannelMessage:
		return h.handleNewMessage(ctx, u.Message)
	}
	return nil
}

func (h *UpdatesHandler) handleNewMessage(ctx context.Context, message tg.MessageClass) error {
	msg, ok := message.(*tg.Message)
	if !ok {
		return nil
	}

	// 检查是否来自目标群组
	var chatID int64
	switch peer := msg.GetPeerID().(type) {
	case *tg.PeerChannel:
		chatID = -1000000000000 - peer.ChannelID // 转换为标准 Chat ID 格式
	case *tg.PeerChat:
		chatID = -peer.ChatID
	case *tg.PeerUser:
		chatID = peer.UserID
	default:
		return nil
	}

	// 只处理目标群组的消息
	if chatID != h.targetChatID {
		return nil
	}

	// 获取消息文本
	messageText := msg.GetMessage()
	if messageText == "" {
		return nil // 跳过空消息或媒体消息
	}

	// 获取发送者信息
	var senderName string
	if msg.FromID != nil {
		switch fromID := msg.FromID.(type) {
		case *tg.PeerUser:
			// 这里可以调用 API 获取用户详细信息
			senderName = fmt.Sprintf("User_%d", fromID.UserID)
		case *tg.PeerChannel:
			senderName = fmt.Sprintf("Channel_%d", fromID.ChannelID)
		}
	}

	// 打印消息信息
	fmt.Printf("\n=== 新消息 ===\n")
	fmt.Printf("群组: %s (ID: %d)\n", GroupTitle, h.targetChatID)
	fmt.Printf("发送者: %s\n", senderName)
	fmt.Printf("消息ID: %d\n", msg.GetID())
	fmt.Printf("时间: %d\n", msg.GetDate())
	fmt.Printf("内容: %s\n", messageText)
	fmt.Printf("===============\n\n")

	return nil
}