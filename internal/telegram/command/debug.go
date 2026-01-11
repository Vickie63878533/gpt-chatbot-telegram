package command

import (
	"encoding/json"
	"fmt"
	"runtime"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/tbxark/ChatGPT-Telegram-Workers/go_version/internal/config"
	"github.com/tbxark/ChatGPT-Telegram-Workers/go_version/internal/i18n"
)

// SystemCommand implements the /system command
type SystemCommand struct {
	config *config.Config
	i18n   *i18n.I18n
}

// NewSystemCommand creates a new /system command
func NewSystemCommand(cfg *config.Config, i18n *i18n.I18n) *SystemCommand {
	return &SystemCommand{
		config: cfg,
		i18n:   i18n,
	}
}

func (c *SystemCommand) Name() string {
	return "system"
}

func (c *SystemCommand) Description(lang string) string {
	i18n := i18n.LoadI18n(lang)
	return i18n.Command.Help.System
}

func (c *SystemCommand) Scopes() []string {
	return []string{"all_private_chats", "all_chat_administrators"}
}

func (c *SystemCommand) NeedAuth() AuthChecker {
	return NoAuthRequired
}

func (c *SystemCommand) Handle(message *tgbotapi.Message, args string, ctx *config.WorkerContext) error {
	var sb strings.Builder

	sb.WriteString("ðŸ–¥ï¸?System Information\n\n")

	// Basic info
	sb.WriteString("**Runtime:**\n")
	sb.WriteString(fmt.Sprintf("- Go Version: `%s`\n", runtime.Version()))
	sb.WriteString(fmt.Sprintf("- OS/Arch: `%s/%s`\n", runtime.GOOS, runtime.GOARCH))
	sb.WriteString(fmt.Sprintf("- Goroutines: `%d`\n", runtime.NumGoroutine()))
	sb.WriteString("\n")

	// Bot info
	sb.WriteString("**Bot Configuration:**\n")
	sb.WriteString(fmt.Sprintf("- Bot ID: `%d`\n", ctx.ShareContext.BotID))
	sb.WriteString(fmt.Sprintf("- Language: `%s`\n", c.config.Language))
	sb.WriteString(fmt.Sprintf("- AI Provider: `%s`\n", c.config.AIProvider))
	sb.WriteString(fmt.Sprintf("- Stream Mode: `%v`\n", c.config.StreamMode))
	sb.WriteString(fmt.Sprintf("- Safe Mode: `%v`\n", c.config.SafeMode))
	sb.WriteString(fmt.Sprintf("- Debug Mode: `%v`\n", c.config.DebugMode))
	sb.WriteString("\n")

	// Additional info in DEV_MODE
	if c.config.DevMode {
		sb.WriteString("**Development Mode Info:**\n")
		sb.WriteString(fmt.Sprintf("- DB Path: `%s`\n", c.config.DBPath))
		sb.WriteString(fmt.Sprintf("- Port: `%d`\n", c.config.Port))
		sb.WriteString(fmt.Sprintf("- Max History Length: `%d`\n", c.config.MaxHistoryLength))
		sb.WriteString(fmt.Sprintf("- Group Bot Enable: `%v`\n", c.config.GroupChatBotEnable))
		sb.WriteString(fmt.Sprintf("- Group Share Mode: `%v`\n", c.config.GroupChatBotShareMode))
		sb.WriteString("\n")

		// User config info
		sb.WriteString("**User Configuration:**\n")
		if len(ctx.UserConfig.DefineKeys) > 0 {
			sb.WriteString(fmt.Sprintf("- Defined Keys: `%v`\n", ctx.UserConfig.DefineKeys))
		} else {
			sb.WriteString("- No user configuration set\n")
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
		return fmt.Errorf("failed to send system info: %w", err)
	}

	return nil
}

// EchoCommand implements the /echo command (for debugging)
type EchoCommand struct {
	config *config.Config
	i18n   *i18n.I18n
}

// NewEchoCommand creates a new /echo command
func NewEchoCommand(cfg *config.Config, i18n *i18n.I18n) *EchoCommand {
	return &EchoCommand{
		config: cfg,
		i18n:   i18n,
	}
}

func (c *EchoCommand) Name() string {
	return "echo"
}

func (c *EchoCommand) Description(lang string) string {
	i18n := i18n.LoadI18n(lang)
	return i18n.Command.Help.Echo
}

func (c *EchoCommand) Scopes() []string {
	return []string{"all"}
}

func (c *EchoCommand) NeedAuth() AuthChecker {
	return NoAuthRequired
}

func (c *EchoCommand) Handle(message *tgbotapi.Message, args string, ctx *config.WorkerContext) error {
	// Convert message to JSON for debugging
	messageJSON, err := json.MarshalIndent(message, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal message: %w", err)
	}

	// Truncate if too long (Telegram message limit is 4096 characters)
	text := string(messageJSON)
	if len(text) > 4000 {
		text = text[:4000] + "\n...(truncated)"
	}

	// Wrap in code block
	text = "```json\n" + text + "\n```"

	msg := tgbotapi.NewMessage(message.Chat.ID, text)
	msg.ParseMode = "Markdown"

	// Get bot instance
	bot, ok := ctx.Bot.(*tgbotapi.BotAPI)
	if !ok || bot == nil {
		return fmt.Errorf("bot instance not available")
	}

	if _, err := bot.Send(msg); err != nil {
		return fmt.Errorf("failed to send echo: %w", err)
	}

	return nil
}
