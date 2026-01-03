package handler

import (
	"github.com/tbxark/ChatGPT-Telegram-Workers/go_version/internal/config"
	"github.com/tbxark/ChatGPT-Telegram-Workers/go_version/internal/i18n"
)

// BuildUpdateHandlerChain builds the complete update handler chain
// according to the requirements (12.1, 12.2)
func BuildUpdateHandlerChain(cfg *config.Config, i18n *i18n.I18n, commandRegistry CommandRegistry) *UpdateHandlerChain {
	// Build message handler chain
	messageHandlers := BuildMessageHandlerChain(cfg, commandRegistry)

	// Build update handler chain
	return NewUpdateHandlerChain(
		NewEnvChecker(),
		NewWhiteListFilter(cfg, i18n),
		NewUpdate2MessageHandler(messageHandlers),
		NewCallbackQueryHandler(cfg, i18n),
	)
}

// BuildMessageHandlerChain builds the complete message handler chain
// according to the requirements (2.12, 12.3, 12.4)
func BuildMessageHandlerChain(cfg *config.Config, commandRegistry CommandRegistry) []MessageHandler {
	cmdHandler := NewCommandHandler(cfg)
	cmdHandler.SetRegistry(commandRegistry)

	return []MessageHandler{
		NewSaveLastMessage(cfg),
		NewOldMessageFilter(cfg),
		NewMessageFilter(cfg),
		cmdHandler,
		NewChatHandler(cfg),
	}
}
