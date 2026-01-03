package command

import (
	"context"
	"encoding/base64"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/tbxark/ChatGPT-Telegram-Workers/go_version/internal/agent"
	"github.com/tbxark/ChatGPT-Telegram-Workers/go_version/internal/config"
	"github.com/tbxark/ChatGPT-Telegram-Workers/go_version/internal/i18n"
	"github.com/tbxark/ChatGPT-Telegram-Workers/go_version/internal/telegram/api"
	"github.com/tbxark/ChatGPT-Telegram-Workers/go_version/internal/telegram/sender"
)

// ImgCommand implements the /img command for image generation
type ImgCommand struct {
	config *config.Config
	i18n   *i18n.I18n
}

// NewImgCommand creates a new /img command
func NewImgCommand(cfg *config.Config, i18n *i18n.I18n) *ImgCommand {
	return &ImgCommand{
		config: cfg,
		i18n:   i18n,
	}
}

func (c *ImgCommand) Name() string {
	return "img"
}

func (c *ImgCommand) Description(lang string) string {
	i18n := i18n.LoadI18n(lang)
	return i18n.Command.Help.Img
}

func (c *ImgCommand) Scopes() []string {
	return []string{"all_private_chats", "all_chat_administrators"}
}

func (c *ImgCommand) NeedAuth() AuthChecker {
	return NoAuthRequired
}

func (c *ImgCommand) Handle(message *tgbotapi.Message, args string, ctx *config.WorkerContext) error {
	// Check if prompt is provided
	prompt := strings.TrimSpace(args)
	if prompt == "" {
		return fmt.Errorf("please provide an image description, e.g., /img beach at moonlight")
	}

	// Get client from context
	client, ok := ctx.Bot.(*api.Client)
	if !ok || client == nil {
		return fmt.Errorf("bot client not available")
	}

	// Create message sender
	msgSender := sender.NewMessageSender(client, message.Chat.ID)

	// Send "upload_photo" action
	if err := msgSender.SendChatAction("upload_photo"); err != nil {
		// Non-fatal, just log
		fmt.Printf("Failed to send chat action: %v\n", err)
	}

	// Load image generation agent
	sessionCtx := NewSessionContext(message, ctx.ShareContext.BotID, c.config.GroupChatBotShareMode)
	userConfig, err := ctx.DB.GetUserConfig(sessionCtx)
	if err != nil {
		return fmt.Errorf("failed to load user config: %w", err)
	}

	imageAgent, err := agent.LoadImageGen(c.config, userConfig)
	if err != nil {
		return fmt.Errorf("no image generation provider available: %w", err)
	}

	// Generate image with timeout
	ctxWithTimeout, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	imageResult, err := imageAgent.Request(ctxWithTimeout, prompt, c.config)
	if err != nil {
		return fmt.Errorf("failed to generate image: %w", err)
	}

	// Send the image
	if err := c.sendImage(msgSender, imageResult); err != nil {
		return fmt.Errorf("failed to send image: %w", err)
	}

	return nil
}

// sendImage sends an image (URL or base64) to the user
func (c *ImgCommand) sendImage(sender *sender.MessageSender, imageData string) error {
	// Check if it's a URL or base64
	if strings.HasPrefix(imageData, "http://") || strings.HasPrefix(imageData, "https://") {
		// It's a URL - send directly
		return sender.SendPhoto(imageData)
	}

	// It's base64 - decode and send as bytes
	// Remove data URL prefix if present
	if strings.HasPrefix(imageData, "data:image/") {
		parts := strings.SplitN(imageData, ",", 2)
		if len(parts) == 2 {
			imageData = parts[1]
		}
	}

	// Decode base64
	photoBytes, err := base64.StdEncoding.DecodeString(imageData)
	if err != nil {
		// If decoding fails, try to download as URL
		return c.downloadAndSendImage(sender, imageData)
	}

	// Send as bytes
	return sender.SendPhotoBytes(photoBytes, "")
}

// downloadAndSendImage downloads an image from URL and sends it
func (c *ImgCommand) downloadAndSendImage(sender *sender.MessageSender, imageURL string) error {
	// Download the image
	resp, err := http.Get(imageURL)
	if err != nil {
		return fmt.Errorf("failed to download image: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to download image: status %d", resp.StatusCode)
	}

	// Read image data
	photoBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read image data: %w", err)
	}

	// Send as bytes
	return sender.SendPhotoBytes(photoBytes, "")
}

// ModelsCommand implements the /models command for model switching
type ModelsCommand struct {
	config *config.Config
	i18n   *i18n.I18n
}

// NewModelsCommand creates a new /models command
func NewModelsCommand(cfg *config.Config, i18n *i18n.I18n) *ModelsCommand {
	return &ModelsCommand{
		config: cfg,
		i18n:   i18n,
	}
}

func (c *ModelsCommand) Name() string {
	return "models"
}

func (c *ModelsCommand) Description(lang string) string {
	i18n := i18n.LoadI18n(lang)
	return i18n.Command.Help.Models
}

func (c *ModelsCommand) Scopes() []string {
	return []string{"all_private_chats", "all_group_chats", "all_chat_administrators"}
}

func (c *ModelsCommand) NeedAuth() AuthChecker {
	return NoAuthRequired
}

func (c *ModelsCommand) Handle(message *tgbotapi.Message, args string, ctx *config.WorkerContext) error {
	// Get client from context
	client, ok := ctx.Bot.(*api.Client)
	if !ok || client == nil {
		return fmt.Errorf("bot client not available")
	}

	// Create message sender
	msgSender := sender.NewMessageSender(client, message.Chat.ID)

	// Load user config to get current model
	sessionCtx := NewSessionContext(message, ctx.ShareContext.BotID, c.config.GroupChatBotShareMode)
	userConfig, err := ctx.DB.GetUserConfig(sessionCtx)
	if err != nil {
		return fmt.Errorf("failed to load user config: %w", err)
	}

	// Get all available chat agents
	chatAgents := agent.GetChatAgents()
	if len(chatAgents) == 0 {
		return fmt.Errorf("no chat providers available")
	}

	// Build inline keyboard with provider selection
	keyboard := c.buildProviderKeyboard(chatAgents)

	// Get current agent and model
	currentAgent, _ := agent.LoadChatLLM(c.config, userConfig)
	currentModel := ""
	if currentAgent != nil {
		currentModel = currentAgent.Model(c.config)
	}

	// Send message with keyboard
	text := fmt.Sprintf("%s\n\nCurrent Model: `%s`", c.i18n.CallbackQuery.SelectProvider, currentModel)
	if err := msgSender.SendMessageWithKeyboard(text, keyboard, c.config.DefaultParseMode); err != nil {
		return fmt.Errorf("failed to send models list: %w", err)
	}

	return nil
}

// buildProviderKeyboard builds an inline keyboard for provider selection
func (c *ModelsCommand) buildProviderKeyboard(agents []agent.ChatAgent) tgbotapi.InlineKeyboardMarkup {
	var rows [][]tgbotapi.InlineKeyboardButton

	// Get number of columns from config
	columns := c.config.ModelListColumns
	if columns < 1 {
		columns = 1
	}

	var currentRow []tgbotapi.InlineKeyboardButton

	for _, ag := range agents {
		if !ag.Enable(c.config) {
			continue
		}

		// Create button for this provider
		button := tgbotapi.NewInlineKeyboardButtonData(
			ag.Name(),
			fmt.Sprintf("al:%s", ag.Name()), // "al" = agent list
		)

		currentRow = append(currentRow, button)

		// If we've filled a row, add it and start a new one
		if len(currentRow) >= columns {
			rows = append(rows, currentRow)
			currentRow = []tgbotapi.InlineKeyboardButton{}
		}
	}

	// Add any remaining buttons
	if len(currentRow) > 0 {
		rows = append(rows, currentRow)
	}

	return tgbotapi.NewInlineKeyboardMarkup(rows...)
}
