package command

import (
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/tbxark/ChatGPT-Telegram-Workers/go_version/internal/config"
	"github.com/tbxark/ChatGPT-Telegram-Workers/go_version/internal/storage"
)

// NewSessionContext creates a SessionContext from a Telegram message
func NewSessionContext(message *tgbotapi.Message, botID int64, shareMode bool) *storage.SessionContext {
	chatID := message.Chat.ID
	isGroup := message.Chat.IsGroup() || message.Chat.IsSuperGroup()

	var userID *int64
	var threadID *int64

	// In group non-shared mode, include user ID
	if isGroup && !shareMode && message.From != nil {
		uid := message.From.ID
		userID = &uid
	}

	// Include thread ID for forum/topic messages
	// Note: Forum features may not be available in all versions of telegram-bot-api

	return config.NewSessionContextFromChat(chatID, botID, isGroup, shareMode, userID, threadID)
}
