package command

import (
	"fmt"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/tbxark/ChatGPT-Telegram-Workers/go_version/internal/config"
	"github.com/tbxark/ChatGPT-Telegram-Workers/go_version/internal/sillytavern"
	"github.com/tbxark/ChatGPT-Telegram-Workers/go_version/internal/telegraph"
)

// ShareCommand implements the /share command
// Shares the conversation history via Telegraph
type ShareCommand struct {
	config         *config.Config
	contextManager *sillytavern.ContextManager
}

// NewShareCommand creates a new /share command
func NewShareCommand(cfg *config.Config, contextManager *sillytavern.ContextManager) *ShareCommand {
	return &ShareCommand{
		config:         cfg,
		contextManager: contextManager,
	}
}

func (c *ShareCommand) Name() string {
	return "share"
}

func (c *ShareCommand) Description(lang string) string {
	return "Share conversation via Telegraph"
}

func (c *ShareCommand) Scopes() []string {
	return []string{"all_private_chats", "all_group_chats"}
}

func (c *ShareCommand) NeedAuth() AuthChecker {
	return NoAuthRequired
}

func (c *ShareCommand) Handle(message *tgbotapi.Message, args string, ctx *config.WorkerContext) error {
	// Get bot instance
	bot, ok := ctx.Bot.(*tgbotapi.BotAPI)
	if !ok || bot == nil {
		return fmt.Errorf("bot instance not available")
	}

	// Send "processing" message
	processingMsg := tgbotapi.NewMessage(message.Chat.ID, "�?正在生成分享链接...")
	processingMsg.ParseMode = c.config.DefaultParseMode
	sentMsg, err := bot.Send(processingMsg)
	if err != nil {
		return fmt.Errorf("failed to send processing message: %w", err)
	}

	// 1. Get session context
	sessionCtx := NewSessionContext(message, ctx.ShareContext.BotID, c.config.GroupChatBotShareMode)

	// 2. Get full history (including summarized messages)
	history, err := c.contextManager.GetFullHistory(sessionCtx)
	if err != nil {
		// Delete processing message
		deleteMsg := tgbotapi.NewDeleteMessage(message.Chat.ID, sentMsg.MessageID)
		bot.Send(deleteMsg)

		errorMsg := tgbotapi.NewMessage(message.Chat.ID, "�?获取对话历史失败")
		errorMsg.ParseMode = c.config.DefaultParseMode
		bot.Send(errorMsg)
		return fmt.Errorf("failed to get full history: %w", err)
	}

	// 3. Filter user and assistant messages only
	var conversationMessages []telegraph.ConversationMessage
	for _, item := range history {
		if item.Role == "user" || item.Role == "assistant" {
			// Extract content as string
			content := ""
			switch v := item.Content.(type) {
			case string:
				content = v
			case []interface{}:
				// Handle multi-part content (text + images)
				for _, part := range v {
					if partMap, ok := part.(map[string]interface{}); ok {
						if partMap["type"] == "text" {
							if text, ok := partMap["text"].(string); ok {
								content += text
							}
						}
					}
				}
			}

			if content != "" {
				conversationMessages = append(conversationMessages, telegraph.ConversationMessage{
					Role:    item.Role,
					Content: content,
				})
			}
		}
	}

	// Check if there are any messages to share
	if len(conversationMessages) == 0 {
		// Delete processing message
		deleteMsg := tgbotapi.NewDeleteMessage(message.Chat.ID, sentMsg.MessageID)
		bot.Send(deleteMsg)

		errorMsg := tgbotapi.NewMessage(message.Chat.ID, "�?没有可分享的对话内容")
		errorMsg.ParseMode = c.config.DefaultParseMode
		bot.Send(errorMsg)
		return nil
	}

	// 4. Format as HTML
	htmlContent := telegraph.FormatConversation(conversationMessages)

	// 5. Create Telegraph page
	telegraphClient, err := telegraph.NewClient()
	if err != nil {
		// Delete processing message
		deleteMsg := tgbotapi.NewDeleteMessage(message.Chat.ID, sentMsg.MessageID)
		bot.Send(deleteMsg)

		errorMsg := tgbotapi.NewMessage(message.Chat.ID, "�?创建 Telegraph 客户端失�?)
		errorMsg.ParseMode = c.config.DefaultParseMode
		bot.Send(errorMsg)
		return fmt.Errorf("failed to create Telegraph client: %w", err)
	}

	// Generate title with timestamp
	title := fmt.Sprintf("Conversation - %s", time.Now().Format("2006-01-02 15:04"))

	url, err := telegraphClient.CreatePage(title, htmlContent)
	if err != nil {
		// Delete processing message
		deleteMsg := tgbotapi.NewDeleteMessage(message.Chat.ID, sentMsg.MessageID)
		bot.Send(deleteMsg)

		errorMsg := tgbotapi.NewMessage(message.Chat.ID, "�?创建 Telegraph 页面失败")
		errorMsg.ParseMode = c.config.DefaultParseMode
		bot.Send(errorMsg)
		return fmt.Errorf("failed to create Telegraph page: %w", err)
	}

	// 6. Delete processing message and send URL
	deleteMsg := tgbotapi.NewDeleteMessage(message.Chat.ID, sentMsg.MessageID)
	bot.Send(deleteMsg)

	responseText := fmt.Sprintf("�?对话已分享\n\n🔗 %s", url)
	msg := tgbotapi.NewMessage(message.Chat.ID, responseText)
	msg.ParseMode = c.config.DefaultParseMode

	if _, err := bot.Send(msg); err != nil {
		return fmt.Errorf("failed to send message: %w", err)
	}

	return nil
}
