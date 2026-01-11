package command

import (
	"fmt"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/tbxark/ChatGPT-Telegram-Workers/go_version/internal/config"
	"github.com/tbxark/ChatGPT-Telegram-Workers/go_version/internal/sillytavern"
)

// ClearCommand implements the /clear command
// Creates a truncation marker without deleting history
type ClearCommand struct {
	config         *config.Config
	contextManager *sillytavern.ContextManager
}

// NewClearCommand creates a new /clear command
func NewClearCommand(cfg *config.Config, contextManager *sillytavern.ContextManager) *ClearCommand {
	return &ClearCommand{
		config:         cfg,
		contextManager: contextManager,
	}
}

func (c *ClearCommand) Name() string {
	return "clear"
}

func (c *ClearCommand) Description(lang string) string {
	return "Clear conversation context (history preserved)"
}

func (c *ClearCommand) Scopes() []string {
	return []string{"all_private_chats", "all_group_chats"}
}

func (c *ClearCommand) NeedAuth() AuthChecker {
	return NoAuthRequired
}

func (c *ClearCommand) Handle(message *tgbotapi.Message, args string, ctx *config.WorkerContext) error {
	// Get session context
	sessionCtx := NewSessionContext(message, ctx.ShareContext.BotID, c.config.GroupChatBotShareMode)

	// Call ContextManager.ClearHistory to create truncation marker
	if err := c.contextManager.ClearHistory(sessionCtx); err != nil {
		return fmt.Errorf("failed to clear history: %w", err)
	}

	// Send confirmation message
	responseText := "✅ 对话已清除！\n\n" +
		"历史记录已保留，但新对话将从空白状态开始。\n" +
		"你仍可使用 /share 命令查看完整历史记录。"

	msg := tgbotapi.NewMessage(message.Chat.ID, responseText)
	msg.ParseMode = c.config.DefaultParseMode

	// Get bot instance
	bot, ok := ctx.Bot.(*tgbotapi.BotAPI)
	if !ok || bot == nil {
		return fmt.Errorf("bot instance not available")
	}

	if _, err := bot.Send(msg); err != nil {
		return fmt.Errorf("failed to send message: %w", err)
	}

	return nil
}
