package command

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/tbxark/ChatGPT-Telegram-Workers/go_version/internal/config"
	"github.com/tbxark/ChatGPT-Telegram-Workers/go_version/internal/storage"
)

// LoginCommand implements the /login command
// Generates a secure token for web manager authentication
type LoginCommand struct {
	config *config.Config
}

// NewLoginCommand creates a new /login command
func NewLoginCommand(cfg *config.Config) *LoginCommand {
	return &LoginCommand{
		config: cfg,
	}
}

func (c *LoginCommand) Name() string {
	return "login"
}

func (c *LoginCommand) Description(lang string) string {
	return "Generate login token for web manager"
}

func (c *LoginCommand) Scopes() []string {
	return []string{"all_private_chats"}
}

func (c *LoginCommand) NeedAuth() AuthChecker {
	return NoAuthRequired
}

func (c *LoginCommand) Handle(message *tgbotapi.Message, args string, ctx *config.WorkerContext) error {
	// 1. Check if this is a private chat
	if !message.Chat.IsPrivate() {
		msg := tgbotapi.NewMessage(message.Chat.ID, "�?此命令仅支持私聊使用")
		msg.ParseMode = c.config.DefaultParseMode

		bot, ok := ctx.Bot.(*tgbotapi.BotAPI)
		if !ok || bot == nil {
			return fmt.Errorf("bot instance not available")
		}

		if _, err := bot.Send(msg); err != nil {
			return fmt.Errorf("failed to send message: %w", err)
		}
		return nil
	}

	// 2. Generate secure token
	token, err := generateSecureToken()
	if err != nil {
		return fmt.Errorf("failed to generate token: %w", err)
	}

	// 3. Save to database with 24 hour expiry
	userID := message.From.ID
	loginToken := &storage.LoginToken{
		UserID:    userID,
		Token:     token,
		ExpiresAt: time.Now().Add(24 * time.Hour),
	}

	// Delete any existing token for this user first
	_ = ctx.DB.DeleteLoginToken(userID)

	if err := ctx.DB.CreateLoginToken(loginToken); err != nil {
		return fmt.Errorf("failed to save login token: %w", err)
	}

	// 4. Return formatted message
	responseText := fmt.Sprintf(
		"🔐 登录凭证已生成\n\n"+
			"用户名：`%d`\n"+
			"密码：`%s`\n"+
			"有效期：24小时\n\n"+
			"请使用这些凭证登�?Web 管理器�?,
		userID,
		token,
	)

	msg := tgbotapi.NewMessage(message.Chat.ID, responseText)
	msg.ParseMode = "Markdown"

	bot, ok := ctx.Bot.(*tgbotapi.BotAPI)
	if !ok || bot == nil {
		return fmt.Errorf("bot instance not available")
	}

	if _, err := bot.Send(msg); err != nil {
		return fmt.Errorf("failed to send message: %w", err)
	}

	return nil
}

// generateSecureToken generates a cryptographically secure random token
func generateSecureToken() (string, error) {
	// Generate 32 bytes of random data
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}

	// Convert to hex string (64 characters)
	return hex.EncodeToString(bytes), nil
}
