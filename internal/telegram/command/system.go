package command

import (
	"fmt"
	"strings"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/tbxark/ChatGPT-Telegram-Workers/go_version/internal/config"
	"github.com/tbxark/ChatGPT-Telegram-Workers/go_version/internal/i18n"
)

// StartCommand implements the /start command
type StartCommand struct {
	config *config.Config
	i18n   *i18n.I18n
}

// NewStartCommand creates a new /start command
func NewStartCommand(cfg *config.Config, i18n *i18n.I18n) *StartCommand {
	return &StartCommand{
		config: cfg,
		i18n:   i18n,
	}
}

func (c *StartCommand) Name() string {
	return "start"
}

func (c *StartCommand) Description(lang string) string {
	i18n := i18n.LoadI18n(lang)
	return i18n.Command.Help.Start
}

func (c *StartCommand) Scopes() []string {
	return []string{"all"}
}

func (c *StartCommand) NeedAuth() AuthChecker {
	return NoAuthRequired
}

func (c *StartCommand) Handle(message *tgbotapi.Message, args string, ctx *config.WorkerContext) error {
	// Show chat ID and start new conversation
	chatID := message.Chat.ID
	userID := message.From.ID

	// Clear history to start new conversation
	sessionCtx := NewSessionContext(message, ctx.ShareContext.BotID, c.config.GroupChatBotShareMode)
	if err := ctx.DB.DeleteChatHistory(sessionCtx); err != nil {
		return fmt.Errorf("failed to clear history: %w", err)
	}

	// Send welcome message with chat ID
	text := fmt.Sprintf("🤖 Welcome!\n\nYour Chat ID: `%d`\nYour User ID: `%d`\n\n%s",
		chatID, userID, c.i18n.Command.New.NewChatStart)

	msg := tgbotapi.NewMessage(chatID, text)
	msg.ParseMode = c.config.DefaultParseMode

	// Add reply keyboard if enabled and in private chat
	isPrivateChat := message.Chat.Type == "private"
	if c.config.ShowReplyButton && isPrivateChat {
		keyboard := tgbotapi.NewReplyKeyboard(
			tgbotapi.NewKeyboardButtonRow(
				tgbotapi.NewKeyboardButton("/new"),
				tgbotapi.NewKeyboardButton("/redo"),
			),
		)
		keyboard.Selective = true
		keyboard.ResizeKeyboard = true
		keyboard.OneTimeKeyboard = false
		msg.ReplyMarkup = keyboard
	} else {
		// Remove keyboard if not showing reply buttons
		msg.ReplyMarkup = tgbotapi.NewRemoveKeyboard(true)
	}

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

// HelpCommand implements the /help command
type HelpCommand struct {
	config   *config.Config
	i18n     *i18n.I18n
	registry *Registry
}

// NewHelpCommand creates a new /help command
func NewHelpCommand(cfg *config.Config, i18n *i18n.I18n, registry *Registry) *HelpCommand {
	return &HelpCommand{
		config:   cfg,
		i18n:     i18n,
		registry: registry,
	}
}

func (c *HelpCommand) Name() string {
	return "help"
}

func (c *HelpCommand) Description(lang string) string {
	i18n := i18n.LoadI18n(lang)
	return i18n.Command.Help.Help
}

func (c *HelpCommand) Scopes() []string {
	return []string{"all_private_chats", "all_chat_administrators"}
}

func (c *HelpCommand) NeedAuth() AuthChecker {
	return NoAuthRequired
}

func (c *HelpCommand) Handle(message *tgbotapi.Message, args string, ctx *config.WorkerContext) error {
	var sb strings.Builder

	// Add summary
	sb.WriteString(c.i18n.Command.Help.Summary)
	sb.WriteString("\n")

	// Add command list
	lang := c.config.Language
	for _, cmd := range c.registry.commands {
		sb.WriteString(fmt.Sprintf("/%s - %s\n", cmd.Name(), cmd.Description(lang)))
	}

	msg := tgbotapi.NewMessage(message.Chat.ID, sb.String())
	msg.ParseMode = c.config.DefaultParseMode

	// Get bot instance
	bot, ok := ctx.Bot.(*tgbotapi.BotAPI)
	if !ok || bot == nil {
		return fmt.Errorf("bot instance not available")
	}

	if _, err := bot.Send(msg); err != nil {
		return fmt.Errorf("failed to send help message: %w", err)
	}

	return nil
}

// NewCommand implements the /new command
type NewCommand struct {
	config *config.Config
	i18n   *i18n.I18n
}

// NewNewCommand creates a new /new command
func NewNewCommand(cfg *config.Config, i18n *i18n.I18n) *NewCommand {
	return &NewCommand{
		config: cfg,
		i18n:   i18n,
	}
}

func (c *NewCommand) Name() string {
	return "new"
}

func (c *NewCommand) Description(lang string) string {
	i18n := i18n.LoadI18n(lang)
	return i18n.Command.Help.New
}

func (c *NewCommand) Scopes() []string {
	return []string{"all_private_chats", "all_group_chats", "all_chat_administrators"}
}

func (c *NewCommand) NeedAuth() AuthChecker {
	return NoAuthRequired
}

func (c *NewCommand) Handle(message *tgbotapi.Message, args string, ctx *config.WorkerContext) error {
	// Clear conversation history
	sessionCtx := NewSessionContext(message, ctx.ShareContext.BotID, c.config.GroupChatBotShareMode)
	if err := ctx.DB.DeleteChatHistory(sessionCtx); err != nil {
		return fmt.Errorf("failed to clear history: %w", err)
	}

	// Send confirmation
	msg := tgbotapi.NewMessage(message.Chat.ID, c.i18n.Command.New.NewChatStart)
	msg.ParseMode = c.config.DefaultParseMode

	// Add reply keyboard if enabled and in private chat
	isPrivateChat := message.Chat.Type == "private"
	if c.config.ShowReplyButton && isPrivateChat {
		keyboard := tgbotapi.NewReplyKeyboard(
			tgbotapi.NewKeyboardButtonRow(
				tgbotapi.NewKeyboardButton("/new"),
				tgbotapi.NewKeyboardButton("/redo"),
			),
		)
		keyboard.Selective = true
		keyboard.ResizeKeyboard = true
		keyboard.OneTimeKeyboard = false
		msg.ReplyMarkup = keyboard
	} else {
		// Remove keyboard if not showing reply buttons
		msg.ReplyMarkup = tgbotapi.NewRemoveKeyboard(true)
	}

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

// formatTimestamp formats a Unix timestamp to a human-readable string
func formatTimestamp(timestamp int64) string {
	if timestamp == 0 {
		return "unknown"
	}
	t := time.Unix(timestamp, 0)
	return t.Format("2006-01-02 15:04:05 MST")
}

// RedoCommand implements the /redo command
// It regenerates the last response, optionally with modified user input
type RedoCommand struct {
	config *config.Config
	i18n   *i18n.I18n
}

// NewRedoCommand creates a new /redo command
func NewRedoCommand(cfg *config.Config, i18n *i18n.I18n) *RedoCommand {
	return &RedoCommand{
		config: cfg,
		i18n:   i18n,
	}
}

func (c *RedoCommand) Name() string {
	return "redo"
}

func (c *RedoCommand) Description(lang string) string {
	i18n := i18n.LoadI18n(lang)
	return i18n.Command.Help.Redo
}

func (c *RedoCommand) Scopes() []string {
	return []string{"all_private_chats", "all_group_chats", "all_chat_administrators"}
}

func (c *RedoCommand) NeedAuth() AuthChecker {
	return NoAuthRequired
}

func (c *RedoCommand) Handle(message *tgbotapi.Message, args string, ctx *config.WorkerContext) error {
	// Mark this as a redo operation in the context
	// The message handler will check for this flag and apply history modification
	if ctx.Context == nil {
		ctx.Context = make(map[string]interface{})
	}
	ctx.Context["redo_mode"] = true
	ctx.Context["redo_text"] = args

	// Return nil to allow the message handler chain to continue
	// The ChatHandler will detect the redo_mode flag and apply the history modifier
	return nil
}
