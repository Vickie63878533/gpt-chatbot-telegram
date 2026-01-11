package command

import (
	"encoding/json"
	"fmt"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/tbxark/ChatGPT-Telegram-Workers/go_version/internal/config"
	"github.com/tbxark/ChatGPT-Telegram-Workers/go_version/internal/i18n"
)

// SetenvCommand implements the /setenv command
type SetenvCommand struct {
	config *config.Config
	i18n   *i18n.I18n
}

// NewSetenvCommand creates a new /setenv command
func NewSetenvCommand(cfg *config.Config, i18n *i18n.I18n) *SetenvCommand {
	return &SetenvCommand{
		config: cfg,
		i18n:   i18n,
	}
}

func (c *SetenvCommand) Name() string {
	return "setenv"
}

func (c *SetenvCommand) Description(lang string) string {
	i18n := i18n.LoadI18n(lang)
	return i18n.Command.Help.Setenv
}

func (c *SetenvCommand) Scopes() []string {
	return []string{"shareModeGroup"}
}

func (c *SetenvCommand) NeedAuth() AuthChecker {
	return ShareModeGroup
}

func (c *SetenvCommand) Handle(message *tgbotapi.Message, args string, ctx *config.WorkerContext) error {
	// Parse KEY=VALUE format
	if args == "" {
		return fmt.Errorf("usage: /setenv KEY=VALUE")
	}

	parts := strings.SplitN(args, "=", 2)
	if len(parts) != 2 {
		return fmt.Errorf("invalid format, use: /setenv KEY=VALUE")
	}

	key := strings.TrimSpace(parts[0])
	value := strings.TrimSpace(parts[1])

	if key == "" {
		return fmt.Errorf("key cannot be empty")
	}

	// Load user config
	sessionCtx := NewSessionContext(message, ctx.ShareContext.BotID, c.config.GroupChatBotShareMode)
	if err := ctx.LoadUserConfig(sessionCtx); err != nil {
		return fmt.Errorf("failed to load user config: %w", err)
	}

	// Set the value with permission check
	if err := ctx.SetUserConfigValueWithPermission(key, value, c.config.LockUserConfigKeys, int64(message.From.ID), message.Chat.ID); err != nil {
		return fmt.Errorf("failed to set config: %w", err)
	}

	// Save user config
	if err := ctx.SaveUserConfig(sessionCtx); err != nil {
		return fmt.Errorf("failed to save user config: %w", err)
	}

	// Send confirmation
	text := fmt.Sprintf("‚ú?Configuration updated:\n`%s` = `%s`", key, value)
	msg := tgbotapi.NewMessage(message.Chat.ID, text)
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

// SetenvsCommand implements the /setenvs command
type SetenvsCommand struct {
	config *config.Config
	i18n   *i18n.I18n
}

// NewSetenvsCommand creates a new /setenvs command
func NewSetenvsCommand(cfg *config.Config, i18n *i18n.I18n) *SetenvsCommand {
	return &SetenvsCommand{
		config: cfg,
		i18n:   i18n,
	}
}

func (c *SetenvsCommand) Name() string {
	return "setenvs"
}

func (c *SetenvsCommand) Description(lang string) string {
	i18n := i18n.LoadI18n(lang)
	return i18n.Command.Help.Setenvs
}

func (c *SetenvsCommand) Scopes() []string {
	return []string{"shareModeGroup"}
}

func (c *SetenvsCommand) NeedAuth() AuthChecker {
	return ShareModeGroup
}

func (c *SetenvsCommand) Handle(message *tgbotapi.Message, args string, ctx *config.WorkerContext) error {
	// Parse JSON format: {"KEY1": "VALUE1", "KEY2": "VALUE2"}
	if args == "" {
		return fmt.Errorf("usage: /setenvs {\"KEY1\": \"VALUE1\", \"KEY2\": \"VALUE2\"}")
	}

	// Parse JSON
	var values map[string]interface{}
	if err := json.Unmarshal([]byte(args), &values); err != nil {
		return fmt.Errorf("invalid JSON format: %w", err)
	}

	if len(values) == 0 {
		return fmt.Errorf("no values provided")
	}

	// Load user config
	sessionCtx := NewSessionContext(message, ctx.ShareContext.BotID, c.config.GroupChatBotShareMode)
	if err := ctx.LoadUserConfig(sessionCtx); err != nil {
		return fmt.Errorf("failed to load user config: %w", err)
	}

	// Set all values with permission check
	var errors []string
	var updated []string
	for key, value := range values {
		if err := ctx.SetUserConfigValueWithPermission(key, value, c.config.LockUserConfigKeys, int64(message.From.ID), message.Chat.ID); err != nil {
			errors = append(errors, fmt.Sprintf("%s: %v", key, err))
		} else {
			updated = append(updated, key)
		}
	}

	// Save user config
	if err := ctx.SaveUserConfig(sessionCtx); err != nil {
		return fmt.Errorf("failed to save user config: %w", err)
	}

	// Build response
	var sb strings.Builder
	if len(updated) > 0 {
		sb.WriteString("‚ú?Configuration updated:\n")
		for _, key := range updated {
			sb.WriteString(fmt.Sprintf("- `%s`\n", key))
		}
	}
	if len(errors) > 0 {
		sb.WriteString("\n‚ù?Errors:\n")
		for _, err := range errors {
			sb.WriteString(fmt.Sprintf("- %s\n", err))
		}
	}

	msg := tgbotapi.NewMessage(message.Chat.ID, sb.String())
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

// DelenvCommand implements the /delenv command
type DelenvCommand struct {
	config *config.Config
	i18n   *i18n.I18n
}

// NewDelenvCommand creates a new /delenv command
func NewDelenvCommand(cfg *config.Config, i18n *i18n.I18n) *DelenvCommand {
	return &DelenvCommand{
		config: cfg,
		i18n:   i18n,
	}
}

func (c *DelenvCommand) Name() string {
	return "delenv"
}

func (c *DelenvCommand) Description(lang string) string {
	i18n := i18n.LoadI18n(lang)
	return i18n.Command.Help.Delenv
}

func (c *DelenvCommand) Scopes() []string {
	return []string{"shareModeGroup"}
}

func (c *DelenvCommand) NeedAuth() AuthChecker {
	return ShareModeGroup
}

func (c *DelenvCommand) Handle(message *tgbotapi.Message, args string, ctx *config.WorkerContext) error {
	// Parse KEY format
	if args == "" {
		return fmt.Errorf("usage: /delenv KEY")
	}

	key := strings.TrimSpace(args)
	if key == "" {
		return fmt.Errorf("key cannot be empty")
	}

	// Load user config
	sessionCtx := NewSessionContext(message, ctx.ShareContext.BotID, c.config.GroupChatBotShareMode)
	if err := ctx.LoadUserConfig(sessionCtx); err != nil {
		return fmt.Errorf("failed to load user config: %w", err)
	}

	// Delete the value with permission check
	if err := ctx.DeleteUserConfigValueWithPermission(key, c.config.LockUserConfigKeys, int64(message.From.ID), message.Chat.ID); err != nil {
		return fmt.Errorf("failed to delete config: %w", err)
	}

	// Save user config
	if err := ctx.SaveUserConfig(sessionCtx); err != nil {
		return fmt.Errorf("failed to save user config: %w", err)
	}

	// Send confirmation
	text := fmt.Sprintf("‚ú?Configuration deleted:\n`%s`", key)
	msg := tgbotapi.NewMessage(message.Chat.ID, text)
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

// ClearenvCommand implements the /clearenv command
type ClearenvCommand struct {
	config *config.Config
	i18n   *i18n.I18n
}

// NewClearenvCommand creates a new /clearenv command
func NewClearenvCommand(cfg *config.Config, i18n *i18n.I18n) *ClearenvCommand {
	return &ClearenvCommand{
		config: cfg,
		i18n:   i18n,
	}
}

func (c *ClearenvCommand) Name() string {
	return "clearenv"
}

func (c *ClearenvCommand) Description(lang string) string {
	i18n := i18n.LoadI18n(lang)
	return i18n.Command.Help.Clearenv
}

func (c *ClearenvCommand) Scopes() []string {
	return []string{"shareModeGroup"}
}

func (c *ClearenvCommand) NeedAuth() AuthChecker {
	return ShareModeGroup
}

func (c *ClearenvCommand) Handle(message *tgbotapi.Message, args string, ctx *config.WorkerContext) error {
	// Load user config
	sessionCtx := NewSessionContext(message, ctx.ShareContext.BotID, c.config.GroupChatBotShareMode)
	if err := ctx.LoadUserConfig(sessionCtx); err != nil {
		return fmt.Errorf("failed to load user config: %w", err)
	}

	// Clear all values with permission check
	if err := ctx.ClearUserConfigWithPermission(c.config.LockUserConfigKeys, int64(message.From.ID), message.Chat.ID); err != nil {
		return fmt.Errorf("failed to clear config: %w", err)
	}

	// Save user config
	if err := ctx.SaveUserConfig(sessionCtx); err != nil {
		return fmt.Errorf("failed to save user config: %w", err)
	}

	// Send confirmation
	text := "‚ú?All user configuration cleared (locked keys preserved)"
	msg := tgbotapi.NewMessage(message.Chat.ID, text)
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
