package command

import (
	"fmt"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/tbxark/ChatGPT-Telegram-Workers/go_version/internal/config"
	"github.com/tbxark/ChatGPT-Telegram-Workers/go_version/internal/plugin"
	"github.com/tbxark/ChatGPT-Telegram-Workers/go_version/internal/telegram/api"
	"github.com/tbxark/ChatGPT-Telegram-Workers/go_version/internal/telegram/sender"
)

// PluginCommand represents a plugin-based command
type PluginCommand struct {
	command  string
	registry *plugin.PluginRegistry
}

// NewPluginCommand creates a new plugin command
func NewPluginCommand(command string, registry *plugin.PluginRegistry) *PluginCommand {
	return &PluginCommand{
		command:  command,
		registry: registry,
	}
}

// Name returns the command name
func (c *PluginCommand) Name() string {
	return strings.TrimPrefix(c.command, "/")
}

// Description returns the command description
func (c *PluginCommand) Description(lang string) string {
	config := c.registry.GetPluginConfig(c.command)
	if config != nil && config.Description != "" {
		return config.Description
	}
	return "Plugin command"
}

// Scopes returns the command scopes
func (c *PluginCommand) Scopes() []string {
	config := c.registry.GetPluginConfig(c.command)
	if config != nil && len(config.Scope) > 0 {
		return config.Scope
	}
	return []string{"all_private_chats", "all_chat_administrators"}
}

// NeedAuth returns the auth checker
func (c *PluginCommand) NeedAuth() AuthChecker {
	return NoAuthRequired
}

// Handle processes the plugin command
func (c *PluginCommand) Handle(message *tgbotapi.Message, args string, ctx *config.WorkerContext) error {
	// Get client from context
	client, ok := ctx.Bot.(*sender.MessageSender)
	if !ok {
		// Try to get as api.Client
		apiClient, ok := ctx.Bot.(*api.Client)
		if !ok || apiClient == nil {
			return fmt.Errorf("bot client not available")
		}
		client = sender.NewMessageSender(apiClient, message.Chat.ID)
	}

	s := client

	// Get the template
	template, err := c.registry.GetTemplate(c.command)
	if err != nil {
		help := c.Description(ctx.Config.Language)
		errorMsg := fmt.Sprintf("ERROR: %v", err)
		if help != "" {
			errorMsg += "\n" + help
		}
		return s.SendPlainText(errorMsg)
	}

	// Check if input is required
	if template.Input.Required && args == "" {
		help := c.Description(ctx.Config.Language)
		errorMsg := "ERROR: Input is required for this command"
		if help != "" {
			errorMsg += "\n" + help
		}
		return s.SendPlainText(errorMsg)
	}

	// Format input
	inputData := plugin.FormatInput(args, template.Input.Type)

	// Prepare data for template
	data := map[string]interface{}{
		"DATA": inputData,
		"ENV":  c.registry.Env,
	}

	// Execute the request
	result, err := plugin.ExecuteRequest(template, data)
	if err != nil {
		help := c.Description(ctx.Config.Language)
		errorMsg := fmt.Sprintf("ERROR: %v", err)
		if help != "" {
			errorMsg += "\n" + help
		}
		return s.SendPlainText(errorMsg)
	}

	// Send the result based on output type
	switch result.Type {
	case plugin.OutputTypeImage:
		// Send as photo
		photo := tgbotapi.NewPhoto(message.Chat.ID, tgbotapi.FileURL(result.Content))
		photo.ReplyToMessageID = message.MessageID
		_, err = s.SendRawMessage(photo)
		return err

	case plugin.OutputTypeHTML:
		return s.SendRichText(result.Content, "HTML")

	case plugin.OutputTypeMarkdown:
		return s.SendRichText(result.Content, "Markdown")

	case plugin.OutputTypeText:
		fallthrough
	default:
		return s.SendPlainText(result.Content)
	}
}
