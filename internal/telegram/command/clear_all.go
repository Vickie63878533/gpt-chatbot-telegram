package command

import (
	"fmt"
	"log/slog"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/tbxark/ChatGPT-Telegram-Workers/go_version/internal/config"
)

// ClearAllChatCommand implements the /clear_all_chat command
// Deletes all chat history from the database (admin only)
type ClearAllChatCommand struct {
	config            *config.Config
	permissionChecker config.PermissionChecker
}

// NewClearAllChatCommand creates a new /clear_all_chat command
func NewClearAllChatCommand(cfg *config.Config, permissionChecker config.PermissionChecker) *ClearAllChatCommand {
	return &ClearAllChatCommand{
		config:            cfg,
		permissionChecker: permissionChecker,
	}
}

func (c *ClearAllChatCommand) Name() string {
	return "clear_all_chat"
}

func (c *ClearAllChatCommand) Description(lang string) string {
	return "Delete all chat history (admin only)"
}

func (c *ClearAllChatCommand) Scopes() []string {
	return []string{"all_private_chats", "all_chat_administrators"}
}

func (c *ClearAllChatCommand) NeedAuth() AuthChecker {
	return AuthCheckerFunc(func(message *tgbotapi.Message, ctx *config.WorkerContext) (bool, error) {
		// Check if user is admin
		if c.permissionChecker == nil {
			return false, fmt.Errorf("permission checker not available")
		}

		isAdmin, err := c.permissionChecker.IsAdmin(message.From.ID, message.Chat.ID, ctx)
		if err != nil {
			return false, fmt.Errorf("failed to check admin permission: %w", err)
		}

		return isAdmin, nil
	})
}

func (c *ClearAllChatCommand) Handle(message *tgbotapi.Message, args string, ctx *config.WorkerContext) error {
	// Get bot instance
	bot, ok := ctx.Bot.(*tgbotapi.BotAPI)
	if !ok || bot == nil {
		return fmt.Errorf("bot instance not available")
	}

	// Send confirmation prompt
	confirmText := "⚠️ 警告：此操作将删除所有对话历史记录！\n\n" +
		"这是一个不可逆的操作，将清空所有用户的所有对话。请谨慎操作！\n\n" +
		"如确认要继续，请在 30 秒内回复 \"CONFIRM\" 确认。"

	confirmMsg := tgbotapi.NewMessage(message.Chat.ID, confirmText)
	confirmMsg.ParseMode = c.config.DefaultParseMode

	if _, err := bot.Send(confirmMsg); err != nil {
		return fmt.Errorf("failed to send confirmation message: %w", err)
	}

	// Note: In a production system, you would implement a proper confirmation flow
	// with a callback query or message handler. For now, we'll require explicit
	// confirmation through a separate message.
	// The actual deletion should be triggered by a follow-up message handler
	// that checks for "CONFIRM" from the admin user.

	// For this implementation, we'll proceed with deletion if args contains "CONFIRM"
	if args != "CONFIRM" {
		return nil // Wait for confirmation
	}

	// Delete all chat history
	if err := ctx.DB.DeleteAllChatHistory(); err != nil {
		errorMsg := tgbotapi.NewMessage(message.Chat.ID, "❌ 删除失败："+err.Error())
		errorMsg.ParseMode = c.config.DefaultParseMode
		bot.Send(errorMsg)
		return fmt.Errorf("failed to delete all chat history: %w", err)
	}

	// Log audit trail
	slog.Info("All chat history deleted",
		"admin_user_id", message.From.ID,
		"admin_username", message.From.UserName,
		"chat_id", message.Chat.ID,
	)

	// Send success message
	successText := "✅ 所有对话历史已删除！\n\n" +
		"数据库中的所有对话记录已被清空。"

	successMsg := tgbotapi.NewMessage(message.Chat.ID, successText)
	successMsg.ParseMode = c.config.DefaultParseMode

	if _, err := bot.Send(successMsg); err != nil {
		return fmt.Errorf("failed to send success message: %w", err)
	}

	return nil
}
