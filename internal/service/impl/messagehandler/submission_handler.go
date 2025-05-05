package messagehandler

import (
	"MikoNews/internal/config"
	"MikoNews/internal/pkg/logger"
	"MikoNews/internal/service"
	"context"
	"encoding/json"
	"fmt"
	"math/rand"
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
	cfg                  *config.FeishuConfig
}

// NewSubmissionHandlerStrategy creates a new submission handler strategy.
func NewSubmissionHandlerStrategy(
	articleService service.ArticleService,
	feishuService service.FeishuMessageService,
	feishuContactService service.FeishuContactService,
	cfg *config.FeishuConfig,
) service.MessageHandlerStrategy {
	return &SubmissionHandlerStrategy{
		articleService:       articleService,
		feishuService:        feishuService,
		feishuContactService: feishuContactService,
		cfg:                  cfg,
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
	replyText := fmt.Sprintf("投稿 '%s' 已收到！感谢您的分享！(ID: %d) 正在转发到群聊...", createdArticle.Title, createdArticle.ID)
	if _, replyErr := s.feishuService.ReplyTextMessage(ctx, msgID, replyText); replyErr != nil {
		logger.Error("Failed to send confirmation reply to user", zap.String("messageID", msgID), zap.Error(replyErr))
	}

	// 5. Build and Forward card to group chat(s)
	cardContent, err := s.buildForwardingCard(rawContent)
	if err != nil {
		logger.Error("Failed to build forwarding card content", zap.String("messageID", msgID), zap.Error(err))
		// Don't return error here, submission is saved, just forwarding failed.
	} else {
		// Send to all configured group chats
		if len(s.cfg.GroupChats) == 0 {
			logger.Warn("No group chats configured for forwarding", zap.String("messageID", msgID))
		} else {
			for _, groupID := range s.cfg.GroupChats {
				if _, sendErr := s.feishuService.SendCardMessage(ctx, groupID, cardContent); sendErr != nil {
					logger.Error("Failed to forward card to group chat",
						zap.String("messageID", msgID),
						zap.String("groupID", groupID),
						zap.Error(sendErr),
					)
				} else {
					logger.Info("Successfully forwarded card to group chat",
						zap.String("messageID", msgID),
						zap.String("groupID", groupID),
					)
				}
			}
		}
	}

	logger.Info("Submission handled successfully", zap.String("messageID", msgID), zap.String("title", title))
	return nil
}

// buildForwardingCard constructs the interactive card content for forwarding.
func (s *SubmissionHandlerStrategy) buildForwardingCard(rawContent string) (*service.MessageCardContent, error) {
	// Define input structure (can reuse/adapt from parsePostContentForSubmission)
	type PostElement struct {
		Tag      string   `json:"tag"`
		Text     string   `json:"text"`
		Style    []string `json:"style"`
		ImageKey string   `json:"image_key"` // For img tags
		Href     string   `json:"href"`      // For a tags
	}
	type PostBody struct {
		Title   string          `json:"title"`
		Content [][]PostElement `json:"content"`
	}

	// Define output card structure elements (using map for flexibility in elements)
	type CardConfig struct {
		WideScreenMode bool `json:"wide_screen_mode"`
	}
	type CardHeaderTitle struct {
		Content string `json:"content"`
		Tag     string `json:"tag"`
	}
	type CardHeader struct {
		Template string          `json:"template"`
		Title    CardHeaderTitle `json:"title"`
	}

	var post PostBody
	if err := json.Unmarshal([]byte(rawContent), &post); err != nil {
		return nil, fmt.Errorf("failed to unmarshal input post content: %w", err)
	}

	// --- Build Card Header ---
	cardTitle := ""
	// Extract title (first bold text in first line)
	if len(post.Content) > 0 {
		for _, element := range post.Content[0] {
			if element.Tag == "text" {
				for _, style := range element.Style {
					if style == "bold" {
						cardTitle = element.Text
						break
					}
				}
			}
			if cardTitle != "" {
				break
			}
		}
	}
	if cardTitle == "" { // Fallback if no bold title found
		if len(post.Content) > 0 {
			for _, element := range post.Content[0] {
				if element.Tag == "text" {
					cardTitle += element.Text
				}
			}
		}
		if cardTitle == "" {
			cardTitle = "分享内容" // Ultimate fallback
		}
	}

	// Choose random header color
	colors := []string{"blue", "wathet", "turquoise", "green", "yellow", "orange", "red", "carmine", "violet", "purple", "indigo", "grey"}
	randomColor := colors[rand.Intn(len(colors))]

	cardHeader := CardHeader{
		Template: randomColor,
		Title: CardHeaderTitle{
			Content: cardTitle,
			Tag:     "plain_text",
		},
	}

	// --- Build Card Elements ---
	cardElements := make([]interface{}, 0)
	var mdContentBuilder strings.Builder

	for lineIdx, line := range post.Content {
		lineHasContent := false // Track if line contributes to markdown
		for _, element := range line {
			switch element.Tag {
			case "img":
				// If there's pending markdown text, add it as a div first
				if mdContentBuilder.Len() > 0 {
					cardElements = append(cardElements, map[string]interface{}{
						"tag":  "div",
						"text": map[string]string{"tag": "lark_md", "content": mdContentBuilder.String()},
					})
					mdContentBuilder.Reset()
				}
				// Add the image element
				cardElements = append(cardElements, map[string]interface{}{
					"tag":     "img",
					"img_key": element.ImageKey,
					"alt":     map[string]string{"tag": "plain_text", "content": "图片"}, // Add alt text
				})
			case "text":
				isBold := false
				for _, style := range element.Style {
					if style == "bold" {
						isBold = true
						break
					}
				}
				if isBold {
					mdContentBuilder.WriteString(fmt.Sprintf("**%s**", element.Text))
				} else {
					mdContentBuilder.WriteString(element.Text)
				}
				lineHasContent = true
			case "a":
				mdContentBuilder.WriteString(fmt.Sprintf("[%s](%s)", element.Text, element.Href))
				lineHasContent = true
			}
		}
		// Add newline after processing each line that had markdown content
		if lineHasContent && lineIdx < len(post.Content)-1 {
			mdContentBuilder.WriteString("\n")
		}
	}

	// Add any remaining markdown content as a final div
	if mdContentBuilder.Len() > 0 {
		cardElements = append(cardElements, map[string]interface{}{
			"tag":  "div",
			"text": map[string]string{"tag": "lark_md", "content": mdContentBuilder.String()},
		})
	}

	// Handle case where there are no elements (e.g., empty post)
	if len(cardElements) == 0 {
		cardElements = append(cardElements, map[string]interface{}{
			"tag":  "div",
			"text": map[string]string{"tag": "lark_md", "content": "(无内容)"},
		})
	}

	// --- Assemble Final Card ---
	finalCard := &service.MessageCardContent{
		Config:   CardConfig{WideScreenMode: true},
		Header:   cardHeader,
		Elements: cardElements,
	}

	return finalCard, nil
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
