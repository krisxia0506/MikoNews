package messagehandler

import (
	"MikoNews/internal/pkg/logger"
	"MikoNews/internal/service"
	"context"
	"encoding/json"
	"fmt"
	"strings"

	larkim "github.com/larksuite/oapi-sdk-go/v3/service/im/v1"
	"go.uber.org/zap"
)

// Define a temporary struct to unmarshal the Post content for ShouldHandle check
type postContentTitleCheck struct {
	Title string `json:"title"`
}

// SubmissionHandlerStrategy handles messages starting with /投稿
type SubmissionHandlerStrategy struct {
	articleService       service.ArticleService
	feishuService        service.FeishuMessageService
	feishuContactService service.FeishuContactService
}

// NewSubmissionHandlerStrategy creates a new submission handler strategy.
func NewSubmissionHandlerStrategy(
	articleService service.ArticleService,
	feishuService service.FeishuMessageService,
	feishuContactService service.FeishuContactService,
) service.MessageHandlerStrategy {
	return &SubmissionHandlerStrategy{
		articleService:       articleService,
		feishuService:        feishuService,
		feishuContactService: feishuContactService,
	}
}

// ShouldHandle checks if the message is a P2P Post message with title "投稿"
func (s *SubmissionHandlerStrategy) ShouldHandle(ctx context.Context, event *larkim.P2MessageReceiveV1) bool {
	// Basic event and message structure checks
	if event.Event == nil || event.Event.Message == nil || event.Event.Sender == nil || event.Event.Sender.SenderId == nil ||
		*event.Event.Message.ChatType != "p2p" || *event.Event.Message.MessageType != larkim.MsgTypePost ||
		event.Event.Message.Content == nil {
		logger.Debug("SubmissionHandler: Event structure/type mismatch")
		return false
	}

	// Attempt to unmarshal the Content string to check the title field
	var contentCheck postContentTitleCheck
	rawContent := *event.Event.Message.Content
	if err := json.Unmarshal([]byte(rawContent), &contentCheck); err != nil {
		// Log the error if needed, but don't handle if parsing fails
		logger.Warn("SubmissionHandler: Failed to unmarshal post content for title check", zap.Error(err), zap.String("rawContent", rawContent))
		return false
	}

	// Check if the title field is exactly "投稿"
	if contentCheck.Title == "投稿" {
		logger.Debug("SubmissionHandler: Matched title '投稿'")
		return true
	}

	logger.Debug("SubmissionHandler: Title did not match '投稿'", zap.String("foundTitle", contentCheck.Title))
	return false
}

// Handle processes the submission.
func (s *SubmissionHandlerStrategy) Handle(ctx context.Context, event *larkim.P2MessageReceiveV1) error {
	msgID := *event.Event.Message.MessageId
	senderID := *event.Event.Sender.SenderId.OpenId
	rawContent := *event.Event.Message.Content

	logger.Info("Handling submission", zap.String("messageID", msgID), zap.String("senderOpenID", senderID))

	// 1. Parse Post content
	title, textContent, err := parsePostContentForSubmission(rawContent)
	if err != nil {
		logger.Error("Failed to parse post content", zap.String("messageID", msgID), zap.Error(err))
		// Reply to user about parsing error
		replyText := fmt.Sprintf("解析投稿内容失败：%s", err)
		if _, replyErr := s.feishuService.ReplyTextMessage(ctx, msgID, replyText); replyErr != nil {
			logger.Error("Failed to send error reply to user", zap.String("messageID", msgID), zap.Error(replyErr))
		}
		return fmt.Errorf("parsing post content failed: %w", err)
	}

	// 2. Get Author Name using FeishuContactService
	authorName := senderID // Default to senderID if fetching fails
	userInfo, err := s.feishuContactService.GetUserInfoByOpenID(ctx, senderID)
	if err != nil {
		// Log the error but continue with senderID as author name
		logger.Warn("Failed to get user info from Feishu API, using OpenID as author name",
			zap.String("senderOpenID", senderID),
			zap.Error(err),
		)
	} else if userInfo != nil && userInfo.Name != nil && *userInfo.Name != "" {
		authorName = *userInfo.Name
		logger.Info("Successfully fetched author name", zap.String("senderOpenID", senderID), zap.String("authorName", authorName))
	} else {
		logger.Warn("Fetched user info from Feishu API, but Name field is missing or empty, using OpenID as author name", zap.String("senderOpenID", senderID))
	}

	// 3. Call ArticleService to save the submission
	createdArticle, err := s.articleService.SaveSubmission(ctx, senderID, authorName, title, textContent)
	if err != nil {
		logger.Error("Failed to save submission", zap.String("messageID", msgID), zap.Error(err))
		// Reply to user about saving error
		replyText := fmt.Sprintf("保存投稿失败：%s", err)
		if _, replyErr := s.feishuService.ReplyTextMessage(ctx, msgID, replyText); replyErr != nil {
			logger.Error("Failed to send error reply to user", zap.String("messageID", msgID), zap.Error(replyErr))
		}
		return fmt.Errorf("failed to save article: %w", err)
	}

	// 4. Send confirmation reply to the user
	replyText := fmt.Sprintf("投稿 '%s' 已收到！感谢您的分享！(ID: %d)", createdArticle.Title, createdArticle.ID)
	if _, replyErr := s.feishuService.ReplyTextMessage(ctx, msgID, replyText); replyErr != nil {
		logger.Error("Failed to send confirmation reply to user", zap.String("messageID", msgID), zap.Error(replyErr))
	}

	// 5. TODO: Forward card to group chat (using s.feishuService.SendCardMessage)

	logger.Info("Submission handled successfully", zap.String("messageID", msgID), zap.String("title", title))
	return nil
}

// parsePostContentForSubmission extracts title and content based on bold style in the first line.
func parsePostContentForSubmission(rawContent string) (title string, textContent string, err error) {
	// Define structs matching the expected nested structure
	type PostElement struct {
		Tag   string   `json:"tag"`
		Text  string   `json:"text"`
		Style []string `json:"style"` // Add Style field
	}
	type PostBody struct {
		Title   string          `json:"title"`   // Keep the top-level title as potential fallback
		Content [][]PostElement `json:"content"` // Nested slice for content
	}

	var post PostBody
	if err = json.Unmarshal([]byte(rawContent), &post); err != nil {
		return "", "", fmt.Errorf("failed to unmarshal post content: %w", err)
	}

	extractedTitle := ""
	firstLineText := ""
	var textContentBuilder strings.Builder

	// Iterate through lines and elements
	for lineIdx, line := range post.Content {
		currentLineText := ""
		for _, element := range line {
			if element.Tag == "text" {
				// Append text to current line content
				currentLineText += element.Text

				// Check for bold style in the first line for the title
				if lineIdx == 0 && extractedTitle == "" {
					for _, style := range element.Style {
						if style == "bold" {
							extractedTitle = element.Text
							break // Found bold title, stop checking styles for this element
						}
					}
				}
			}
			// TODO: Potentially handle other tags like 'a' for links if needed
		}

		// Store the text of the first line for fallback title
		if lineIdx == 0 {
			firstLineText = currentLineText
		}

		// Append the full line text to the overall content
		if textContentBuilder.Len() > 0 {
			textContentBuilder.WriteString("\n")
		}
		textContentBuilder.WriteString(currentLineText)
	}

	// Determine the final title
	if extractedTitle != "" {
		title = extractedTitle
	} else if firstLineText != "" {
		title = firstLineText // Use first line text if no bold title found
	} else {
		title = "Untitled Submission" // Default title
	}

	textContent = textContentBuilder.String()

	return title, textContent, nil
}
