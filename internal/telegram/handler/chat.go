package handler

import (
	"context"
	"encoding/base64"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"strings"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/tbxark/ChatGPT-Telegram-Workers/go_version/internal/agent"
	"github.com/tbxark/ChatGPT-Telegram-Workers/go_version/internal/config"
	"github.com/tbxark/ChatGPT-Telegram-Workers/go_version/internal/storage"
	"github.com/tbxark/ChatGPT-Telegram-Workers/go_version/internal/telegram/api"
	"github.com/tbxark/ChatGPT-Telegram-Workers/go_version/internal/telegram/sender"
)

// extractUserMessageItem extracts a user message from a Telegram message
// Supports text messages, photo messages, and messages with captions
func extractUserMessageItem(message *tgbotapi.Message, cfg *config.Config) (storage.HistoryItem, error) {
	var contentParts []storage.ContentPart

	// Extract text content
	text := message.Text
	if text == "" && message.Caption != "" {
		text = message.Caption
	}

	// Add text part if present
	if text != "" {
		contentParts = append(contentParts, storage.ContentPart{
			Type: "text",
			Text: text,
		})
	}

	// Extract photo if present
	if message.Photo != nil && len(message.Photo) > 0 {
		photoURL, err := extractPhotoURL(message, cfg)
		if err != nil {
			return storage.HistoryItem{}, fmt.Errorf("failed to extract photo: %w", err)
		}

		// Convert to base64 if configured
		imageData := photoURL
		if cfg.TelegramImageTransferMode == "base64" {
			base64Data, err := convertImageToBase64(photoURL)
			if err != nil {
				slog.Warn("Failed to convert image to base64, using URL", "error", err)
			} else {
				imageData = base64Data
			}
		}

		contentParts = append(contentParts, storage.ContentPart{
			Type:  "image",
			Image: imageData,
		})
	}

	// If no content parts, return error
	if len(contentParts) == 0 {
		return storage.HistoryItem{}, fmt.Errorf("no content in message")
	}

	// If only text, return simple string content
	if len(contentParts) == 1 && contentParts[0].Type == "text" {
		return storage.HistoryItem{
			Role:    "user",
			Content: contentParts[0].Text,
		}, nil
	}

	// Return multi-part content
	return storage.HistoryItem{
		Role:    "user",
		Content: contentParts,
	}, nil
}

// extractPhotoURL extracts the photo URL from a Telegram message
func extractPhotoURL(message *tgbotapi.Message, cfg *config.Config) (string, error) {
	if message.Photo == nil || len(message.Photo) == 0 {
		return "", fmt.Errorf("no photo in message")
	}

	// Select photo size based on offset
	// Offset can be negative (from end) or positive (from start)
	offset := cfg.TelegramPhotoSizeOffset
	photoSizes := message.Photo

	var selectedPhoto *tgbotapi.PhotoSize
	if offset < 0 {
		// Negative offset: count from end
		index := len(photoSizes) + offset
		if index < 0 {
			index = 0
		}
		selectedPhoto = &photoSizes[index]
	} else {
		// Positive offset: count from start
		index := offset
		if index >= len(photoSizes) {
			index = len(photoSizes) - 1
		}
		selectedPhoto = &photoSizes[index]
	}

	// Get file path from Telegram
	fileID := selectedPhoto.FileID

	// Construct file URL using Telegram API domain
	// Note: This requires the bot token, which should be available in the context
	// For now, we'll return the file ID and let the caller handle URL construction
	return fileID, nil
}

// getPhotoURL gets the full URL for a photo file ID
func getPhotoURL(client *api.Client, fileID string) (string, error) {
	// Get the direct URL for the file
	url, err := client.GetFileDirectURL(fileID)
	if err != nil {
		return "", fmt.Errorf("failed to get file URL: %w", err)
	}
	return url, nil
}

// convertImageToBase64 downloads an image from URL and converts it to base64
func convertImageToBase64(imageURL string) (string, error) {
	// Download the image
	resp, err := http.Get(imageURL)
	if err != nil {
		return "", fmt.Errorf("failed to download image: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("failed to download image: status %d", resp.StatusCode)
	}

	// Read image data
	imageData, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read image data: %w", err)
	}

	// Convert to base64
	base64Data := base64.StdEncoding.EncodeToString(imageData)

	// Add data URI prefix based on content type
	contentType := resp.Header.Get("Content-Type")
	if contentType == "" {
		contentType = "image/jpeg" // Default
	}

	return fmt.Sprintf("data:%s;base64,%s", contentType, base64Data), nil
}

// extractExtraContext extracts extra context from replied-to messages
func extractExtraContext(message *tgbotapi.Message, cfg *config.Config, ctx *config.WorkerContext) []storage.HistoryItem {
	if message.ReplyToMessage == nil {
		return nil
	}

	if !cfg.ExtraMessageContext {
		return nil
	}

	// Get client from context
	client, ok := ctx.Bot.(*api.Client)
	if !ok {
		slog.Warn("Bot client not available in context")
		return nil
	}

	var extraItems []storage.HistoryItem

	// Extract text from replied message
	replyText := message.ReplyToMessage.Text
	if replyText == "" && message.ReplyToMessage.Caption != "" {
		replyText = message.ReplyToMessage.Caption
	}

	if replyText != "" {
		extraItems = append(extraItems, storage.HistoryItem{
			Role:    "user",
			Content: replyText,
		})
	}

	// Extract media if compatible
	if message.ReplyToMessage.Photo != nil && len(message.ReplyToMessage.Photo) > 0 {
		// Check if image is in compatible media types
		isCompatible := false
		for _, mediaType := range cfg.ExtraMessageMediaCompatible {
			if strings.ToLower(mediaType) == "image" {
				isCompatible = true
				break
			}
		}

		if isCompatible {
			photoURL, err := extractPhotoURL(message.ReplyToMessage, cfg)
			if err == nil {
				// Get full URL
				fullURL, err := getPhotoURL(client, photoURL)
				if err == nil {
					imageData := fullURL
					if cfg.TelegramImageTransferMode == "base64" {
						base64Data, err := convertImageToBase64(fullURL)
						if err == nil {
							imageData = base64Data
						}
					}

					var contentParts []storage.ContentPart
					if replyText != "" {
						contentParts = append(contentParts, storage.ContentPart{
							Type: "text",
							Text: replyText,
						})
					}
					contentParts = append(contentParts, storage.ContentPart{
						Type:  "image",
						Image: imageData,
					})

					// Replace the last item with multi-part content
					if len(extraItems) > 0 {
						extraItems[len(extraItems)-1] = storage.HistoryItem{
							Role:    "user",
							Content: contentParts,
						}
					} else {
						extraItems = append(extraItems, storage.HistoryItem{
							Role:    "user",
							Content: contentParts,
						})
					}
				}
			}
		}
	}

	return extraItems
}

// loadHistory loads conversation history from storage
func loadHistory(ctx *storage.SessionContext, db storage.Storage) ([]storage.HistoryItem, error) {
	history, err := db.GetChatHistory(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to load history: %w", err)
	}
	return history, nil
}

// saveHistory saves conversation history to storage
func saveHistory(ctx *storage.SessionContext, db storage.Storage, history []storage.HistoryItem) error {
	if err := db.SaveChatHistory(ctx, history); err != nil {
		return fmt.Errorf("failed to save history: %w", err)
	}
	return nil
}

// trimHistory trims the conversation history based on configuration
func trimHistory(history []storage.HistoryItem, cfg *config.Config) []storage.HistoryItem {
	if !cfg.AutoTrimHistory {
		return history
	}

	maxLength := cfg.MaxHistoryLength
	if maxLength <= 0 {
		return history
	}

	// Keep system messages and trim user/assistant messages
	var systemMessages []storage.HistoryItem
	var conversationMessages []storage.HistoryItem

	for _, item := range history {
		if item.Role == "system" {
			systemMessages = append(systemMessages, item)
		} else {
			conversationMessages = append(conversationMessages, item)
		}
	}

	// Trim conversation messages if needed
	if len(conversationMessages) > maxLength {
		conversationMessages = conversationMessages[len(conversationMessages)-maxLength:]
	}

	// Combine system messages and trimmed conversation
	result := make([]storage.HistoryItem, 0, len(systemMessages)+len(conversationMessages))
	result = append(result, systemMessages...)
	result = append(result, conversationMessages...)

	return result
}

// replaceImagePlaceholder replaces images in history with placeholder text
func replaceImagePlaceholder(history []storage.HistoryItem, placeholder string) []storage.HistoryItem {
	if placeholder == "" {
		return history
	}

	result := make([]storage.HistoryItem, len(history))
	for i, item := range history {
		result[i] = item

		// Check if content is multi-part
		if parts, ok := item.Content.([]storage.ContentPart); ok {
			newParts := make([]storage.ContentPart, 0, len(parts))
			for _, part := range parts {
				if part.Type == "image" {
					// Replace image with placeholder text
					newParts = append(newParts, storage.ContentPart{
						Type: "text",
						Text: placeholder,
					})
				} else {
					newParts = append(newParts, part)
				}
			}

			// If only one text part remains, simplify to string
			if len(newParts) == 1 && newParts[0].Type == "text" {
				result[i].Content = newParts[0].Text
			} else {
				result[i].Content = newParts
			}
		}
	}

	return result
}

// convertStorageToAgentHistory converts storage.HistoryItem to agent.HistoryItem
func convertStorageToAgentHistory(items []storage.HistoryItem) []agent.HistoryItem {
	result := make([]agent.HistoryItem, len(items))
	for i, item := range items {
		result[i] = agent.HistoryItem{
			Role:    item.Role,
			Content: item.Content,
		}
	}
	return result
}

// convertAgentToStorageHistory converts agent.HistoryItem to storage.HistoryItem
func convertAgentToStorageHistory(items []agent.HistoryItem) []storage.HistoryItem {
	result := make([]storage.HistoryItem, len(items))
	for i, item := range items {
		result[i] = storage.HistoryItem{
			Role:    item.Role,
			Content: item.Content,
		}
	}
	return result
}

// chatWithMessage processes a chat message and generates a response
func chatWithMessage(message *tgbotapi.Message, ctx *config.WorkerContext) error {
	cfg := ctx.Config

	// Get client from context
	client, ok := ctx.Bot.(*api.Client)
	if !ok {
		return fmt.Errorf("bot client not available in context")
	}

	// Create session context
	sessionCtx := NewSessionContext(message, ctx.ShareContext.BotID, cfg.GroupChatBotShareMode)

	// Load user config
	if err := ctx.LoadUserConfig(sessionCtx); err != nil {
		slog.Error("Failed to load user config", "error", err)
		// Continue with default config
	}

	// Load conversation history
	history, err := loadHistory(sessionCtx, ctx.DB)
	if err != nil {
		slog.Error("Failed to load history", "error", err)
		history = []storage.HistoryItem{}
	}

	// Check if this is a redo operation
	var userMessage storage.HistoryItem
	isRedoMode := false
	if ctx.Context != nil {
		if redoMode, ok := ctx.Context["redo_mode"].(bool); ok && redoMode {
			isRedoMode = true

			// Apply history modifier for redo
			modifiedHistory, lastUserMsg, err := applyRedoModifier(history, ctx.Context["redo_text"])
			if err != nil {
				return fmt.Errorf("failed to apply redo modifier: %w", err)
			}
			history = modifiedHistory
			userMessage = lastUserMsg

			// Clear the redo flags from context
			delete(ctx.Context, "redo_mode")
			delete(ctx.Context, "redo_text")
		}
	}

	// If not redo mode, extract user message normally
	if !isRedoMode {
		var err error
		userMessage, err = extractUserMessageItem(message, cfg)
		if err != nil {
			return fmt.Errorf("failed to extract user message: %w", err)
		}

		// Extract extra context from replied message
		extraContext := extractExtraContext(message, cfg, ctx)

		// Add extra context if present
		if len(extraContext) > 0 {
			history = append(history, extraContext...)
		}

		// Add user message to history
		history = append(history, userMessage)
	}

	// Trim history if needed
	history = trimHistory(history, cfg)

	// Replace image placeholders if configured
	if cfg.HistoryImagePlaceholder != "" {
		history = replaceImagePlaceholder(history, cfg.HistoryImagePlaceholder)
	}

	// Create message sender
	msgSender := sender.NewMessageSender(client, message.Chat.ID)
	if cfg.TelegramMinStreamInterval > 0 {
		msgSender.SetMinStreamInterval(time.Duration(cfg.TelegramMinStreamInterval) * time.Millisecond)
	}

	// Send typing action
	if err := msgSender.SendChatAction("typing"); err != nil {
		slog.Warn("Failed to send typing action", "error", err)
	}

	// Request completion from LLM
	response, err := requestCompletionsFromLLM(context.Background(), history, cfg, ctx.UserConfig, msgSender)
	if err != nil {
		return fmt.Errorf("failed to get LLM response: %w", err)
	}

	// Add assistant response to history (convert from agent to storage type)
	history = append(history, convertAgentToStorageHistory(response.Messages)...)

	// Trim history again after adding response
	history = trimHistory(history, cfg)

	// Save updated history
	if err := saveHistory(sessionCtx, ctx.DB, history); err != nil {
		slog.Error("Failed to save history", "error", err)
		// Don't fail the request if save fails
	}

	return nil
}

// applyRedoModifier modifies the history for the /redo command
// It removes messages from the end until it finds the last user message,
// optionally replacing it with new text
func applyRedoModifier(history []storage.HistoryItem, redoTextRaw interface{}) ([]storage.HistoryItem, storage.HistoryItem, error) {
	if len(history) == 0 {
		return nil, storage.HistoryItem{}, fmt.Errorf("history not found")
	}

	// Make a copy of history
	historyCopy := make([]storage.HistoryItem, len(history))
	copy(historyCopy, history)

	// Find the last user message by popping from the end
	var lastUserMessage *storage.HistoryItem
	for len(historyCopy) > 0 {
		// Pop the last item
		lastItem := historyCopy[len(historyCopy)-1]
		historyCopy = historyCopy[:len(historyCopy)-1]

		if lastItem.Role == "user" {
			lastUserMessage = &lastItem
			break
		}
	}

	if lastUserMessage == nil {
		return nil, storage.HistoryItem{}, fmt.Errorf("redo message not found")
	}

	// If redo text is provided, replace the user message content
	if redoTextRaw != nil {
		if redoText, ok := redoTextRaw.(string); ok && redoText != "" {
			lastUserMessage.Content = redoText
		}
	}

	return historyCopy, *lastUserMessage, nil
}

// requestCompletionsFromLLM requests a completion from the configured LLM
func requestCompletionsFromLLM(
	ctx context.Context,
	history []storage.HistoryItem,
	cfg *config.Config,
	userConfig *storage.UserConfig,
	msgSender *sender.MessageSender,
) (*agent.ChatAgentResponse, error) {
	// Load chat agent
	chatAgent, err := agent.LoadChatLLM(cfg, userConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to load chat agent: %w", err)
	}

	// Prepare chat parameters (convert storage history to agent history)
	params := &agent.LLMChatParams{
		Prompt:   cfg.SystemInitMessage,
		Messages: convertStorageToAgentHistory(history),
	}

	// Create stream handler if stream mode is enabled
	var streamHandler agent.ChatStreamTextHandler
	if cfg.StreamMode {
		streamHandler = func(text string) error {
			return msgSender.SendRichText(text, cfg.DefaultParseMode)
		}
	}

	// Request completion
	response, err := chatAgent.Request(ctx, params, cfg, streamHandler)
	if err != nil {
		return nil, fmt.Errorf("chat agent request failed: %w", err)
	}

	// If not streaming, send the final response
	if !cfg.StreamMode && len(response.Messages) > 0 {
		// Get the last assistant message
		for i := len(response.Messages) - 1; i >= 0; i-- {
			if response.Messages[i].Role == "assistant" {
				content := ""
				switch v := response.Messages[i].Content.(type) {
				case string:
					content = v
				case []agent.ContentPart:
					// Concatenate text parts
					var parts []string
					for _, part := range v {
						if part.Type == "text" {
							parts = append(parts, part.Text)
						}
					}
					content = strings.Join(parts, "\n")
				}

				if content != "" {
					if err := msgSender.SendRichText(content, cfg.DefaultParseMode); err != nil {
						return nil, fmt.Errorf("failed to send response: %w", err)
					}
				}
				break
			}
		}
	}

	return response, nil
}
