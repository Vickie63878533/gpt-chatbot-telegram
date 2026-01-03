package handler

import (
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/tbxark/ChatGPT-Telegram-Workers/go_version/internal/config"
)

// UpdateHandler processes Telegram updates
type UpdateHandler interface {
	Handle(update *tgbotapi.Update, ctx *config.WorkerContext) error
}

// MessageHandler processes Telegram messages
type MessageHandler interface {
	Handle(message *tgbotapi.Message, ctx *config.WorkerContext) error
}

// UpdateHandlerFunc is a function adapter for UpdateHandler
type UpdateHandlerFunc func(update *tgbotapi.Update, ctx *config.WorkerContext) error

// Handle implements UpdateHandler
func (f UpdateHandlerFunc) Handle(update *tgbotapi.Update, ctx *config.WorkerContext) error {
	return f(update, ctx)
}

// MessageHandlerFunc is a function adapter for MessageHandler
type MessageHandlerFunc func(message *tgbotapi.Message, ctx *config.WorkerContext) error

// Handle implements MessageHandler
func (f MessageHandlerFunc) Handle(message *tgbotapi.Message, ctx *config.WorkerContext) error {
	return f(message, ctx)
}
