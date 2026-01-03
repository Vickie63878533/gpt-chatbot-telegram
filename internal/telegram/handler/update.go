package handler

import (
	"fmt"
	"log/slog"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/tbxark/ChatGPT-Telegram-Workers/go_version/internal/config"
	"github.com/tbxark/ChatGPT-Telegram-Workers/go_version/internal/i18n"
)

// EnvChecker verifies that required environment variables are set
type EnvChecker struct{}

// NewEnvChecker creates a new EnvChecker
func NewEnvChecker() *EnvChecker {
	return &EnvChecker{}
}

// Handle checks if DATABASE is configured
func (h *EnvChecker) Handle(update *tgbotapi.Update, ctx *config.WorkerContext) error {
	if ctx.DB == nil {
		slog.Error("DATABASE not configured")
		return fmt.Errorf("database not configured")
	}
	return nil
}

// WhiteListFilter filters updates based on whitelist configuration
type WhiteListFilter struct {
	config *config.Config
	i18n   *i18n.I18n
}

// NewWhiteListFilter creates a new WhiteListFilter
func NewWhiteListFilter(cfg *config.Config, i18n *i18n.I18n) *WhiteListFilter {
	return &WhiteListFilter{
		config: cfg,
		i18n:   i18n,
	}
}

// Handle filters updates based on whitelist
func (h *WhiteListFilter) Handle(update *tgbotapi.Update, ctx *config.WorkerContext) error {
	// If generous mode is enabled, allow all
	if h.config.IAmAGenerousPerson {
		return nil
	}

	var chatID int64
	var isGroup bool

	// Extract chat info from update
	if update.Message != nil {
		chatID = update.Message.Chat.ID
		isGroup = update.Message.Chat.IsGroup() || update.Message.Chat.IsSuperGroup()
	} else if update.CallbackQuery != nil && update.CallbackQuery.Message != nil {
		chatID = update.CallbackQuery.Message.Chat.ID
		isGroup = update.CallbackQuery.Message.Chat.IsGroup() || update.CallbackQuery.Message.Chat.IsSuperGroup()
	} else {
		// No chat info, allow by default
		return nil
	}

	// Check whitelist
	chatIDStr := fmt.Sprintf("%d", chatID)

	if isGroup {
		// Check group whitelist
		if len(h.config.ChatGroupWhiteList) > 0 {
			allowed := false
			for _, id := range h.config.ChatGroupWhiteList {
				if id == chatIDStr {
					allowed = true
					break
				}
			}
			if !allowed {
				slog.Warn("Group not in whitelist", "chat_id", chatID)
				return fmt.Errorf("unauthorized group: %d", chatID)
			}
		}
	} else {
		// Check private chat whitelist
		if len(h.config.ChatWhiteList) > 0 {
			allowed := false
			for _, id := range h.config.ChatWhiteList {
				if id == chatIDStr {
					allowed = true
					break
				}
			}
			if !allowed {
				slog.Warn("User not in whitelist", "chat_id", chatID)
				// Send unauthorized message with chat_id
				return fmt.Errorf("unauthorized user: %d", chatID)
			}
		}
	}

	return nil
}

// Update2MessageHandler converts Update to Message and delegates to message handlers
type Update2MessageHandler struct {
	messageHandlers []MessageHandler
}

// NewUpdate2MessageHandler creates a new Update2MessageHandler
func NewUpdate2MessageHandler(handlers []MessageHandler) *Update2MessageHandler {
	return &Update2MessageHandler{
		messageHandlers: handlers,
	}
}

// Handle processes the update and delegates to message handlers
func (h *Update2MessageHandler) Handle(update *tgbotapi.Update, ctx *config.WorkerContext) error {
	// Ignore edited messages (Requirement 2.11)
	if update.EditedMessage != nil {
		slog.Debug("Ignoring edited message")
		return nil
	}

	// Extract message from update
	var message *tgbotapi.Message
	if update.Message != nil {
		message = update.Message
	} else {
		// No message to process
		return nil
	}

	// Process through message handler chain
	for _, handler := range h.messageHandlers {
		if err := handler.Handle(message, ctx); err != nil {
			return err
		}
	}

	return nil
}

// UpdateHandlerChain chains multiple update handlers
type UpdateHandlerChain struct {
	handlers []UpdateHandler
}

// NewUpdateHandlerChain creates a new handler chain
func NewUpdateHandlerChain(handlers ...UpdateHandler) *UpdateHandlerChain {
	return &UpdateHandlerChain{
		handlers: handlers,
	}
}

// Handle processes the update through all handlers in sequence
func (c *UpdateHandlerChain) Handle(update *tgbotapi.Update, ctx *config.WorkerContext) error {
	for _, handler := range c.handlers {
		if err := handler.Handle(update, ctx); err != nil {
			return err
		}
	}
	return nil
}
