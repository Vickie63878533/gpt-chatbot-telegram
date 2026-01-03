package command

import (
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/tbxark/ChatGPT-Telegram-Workers/go_version/internal/config"
)

// Command represents a bot command handler
type Command interface {
	// Name returns the command name (without the leading /)
	Name() string

	// Description returns the command description for the given language
	Description(lang string) string

	// Scopes returns the command scopes (e.g., "all_private_chats", "all_group_chats", "all_chat_administrators")
	Scopes() []string

	// NeedAuth returns the auth checker for this command
	NeedAuth() AuthChecker

	// Handle processes the command
	Handle(message *tgbotapi.Message, args string, ctx *config.WorkerContext) error
}

// AuthChecker checks if a user is authorized to execute a command
type AuthChecker interface {
	Check(message *tgbotapi.Message, ctx *config.WorkerContext) (bool, error)
}

// AuthCheckerFunc is a function adapter for AuthChecker
type AuthCheckerFunc func(message *tgbotapi.Message, ctx *config.WorkerContext) (bool, error)

// Check implements AuthChecker
func (f AuthCheckerFunc) Check(message *tgbotapi.Message, ctx *config.WorkerContext) (bool, error) {
	return f(message, ctx)
}

// NoAuthRequired is an auth checker that always allows access
var NoAuthRequired = AuthCheckerFunc(func(message *tgbotapi.Message, ctx *config.WorkerContext) (bool, error) {
	return true, nil
})

// AdminOnly is an auth checker that requires admin permissions in groups
var AdminOnly = AuthCheckerFunc(func(message *tgbotapi.Message, ctx *config.WorkerContext) (bool, error) {
	// In private chats, always allow
	if message.Chat.IsPrivate() {
		return true, nil
	}

	// In groups, check if user is admin
	return IsGroupAdmin(message.Chat.ID, message.From.ID, ctx)
})

// ShareModeGroup is an auth checker for commands that work in share mode groups
var ShareModeGroup = AuthCheckerFunc(func(message *tgbotapi.Message, ctx *config.WorkerContext) (bool, error) {
	// In private chats, always allow
	if message.Chat.IsPrivate() {
		return true, nil
	}

	// In groups, check if user is admin
	return IsGroupAdmin(message.Chat.ID, message.From.ID, ctx)
})
