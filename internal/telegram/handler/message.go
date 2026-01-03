package handler

import (
	"fmt"
	"log/slog"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/tbxark/ChatGPT-Telegram-Workers/go_version/internal/config"
)

// SaveLastMessage saves the last message for debugging
type SaveLastMessage struct {
	config *config.Config
}

// NewSaveLastMessage creates a new SaveLastMessage handler
func NewSaveLastMessage(cfg *config.Config) *SaveLastMessage {
	return &SaveLastMessage{
		config: cfg,
	}
}

// Handle saves the last message if DEBUG_MODE is enabled
func (h *SaveLastMessage) Handle(message *tgbotapi.Message, ctx *config.WorkerContext) error {
	if !h.config.DebugMode {
		return nil
	}

	// TODO: Implement debug message storage
	// For now, just log the message
	slog.Debug("Debug message",
		"chat_id", message.Chat.ID,
		"message_id", message.MessageID,
		"text", message.Text)

	return nil
}

// OldMessageFilter filters old messages in SAFE_MODE
type OldMessageFilter struct {
	config *config.Config
}

// NewOldMessageFilter creates a new OldMessageFilter
func NewOldMessageFilter(cfg *config.Config) *OldMessageFilter {
	return &OldMessageFilter{
		config: cfg,
	}
}

// Handle filters old messages based on SAFE_MODE
func (h *OldMessageFilter) Handle(message *tgbotapi.Message, ctx *config.WorkerContext) error {
	if !h.config.SafeMode {
		return nil
	}

	sessionCtx := NewSessionContext(message, ctx.ShareContext.BotID, h.config.GroupChatBotShareMode)

	// Get stored message IDs
	messageIDs, err := ctx.DB.GetMessageIDs(sessionCtx)
	if err != nil {
		slog.Error("Failed to get message IDs", "error", err)
		return nil // Don't fail on storage error
	}

	// Check if this message was already processed
	currentMessageID := message.MessageID
	for _, id := range messageIDs {
		if id == currentMessageID {
			slog.Debug("Duplicate message detected, ignoring", "message_id", currentMessageID)
			return fmt.Errorf("duplicate message")
		}
	}

	// Add current message ID to the list (keep last 100)
	messageIDs = append(messageIDs, currentMessageID)
	if len(messageIDs) > 100 {
		messageIDs = messageIDs[len(messageIDs)-100:]
	}

	// Save updated list
	if err := ctx.DB.SaveMessageIDs(sessionCtx, messageIDs); err != nil {
		slog.Error("Failed to save message IDs", "error", err)
	}

	return nil
}

// MessageFilter filters unsupported message types
type MessageFilter struct {
	config *config.Config
}

// NewMessageFilter creates a new MessageFilter
func NewMessageFilter(cfg *config.Config) *MessageFilter {
	return &MessageFilter{
		config: cfg,
	}
}

// Handle filters messages based on supported types
func (h *MessageFilter) Handle(message *tgbotapi.Message, ctx *config.WorkerContext) error {
	// Check if message is a command (starts with /)
	if message.IsCommand() {
		return nil // Commands are handled by CommandHandler
	}

	// Supported types: text, photo (with optional caption)
	hasText := message.Text != ""
	hasPhoto := message.Photo != nil && len(message.Photo) > 0
	hasCaption := message.Caption != ""

	// Accept if it has text, or photo with/without caption
	if hasText || hasPhoto || hasCaption {
		return nil
	}

	// Unsupported message type
	slog.Warn("Unsupported message type",
		"chat_id", message.Chat.ID,
		"has_text", hasText,
		"has_photo", hasPhoto,
		"has_caption", hasCaption)

	return fmt.Errorf("unsupported message type")
}

// CommandHandler processes commands
type CommandHandler struct {
	config   *config.Config
	registry CommandRegistry
}

// CommandRegistry is an interface for command registration and dispatching
type CommandRegistry interface {
	Handle(message *tgbotapi.Message, ctx *config.WorkerContext) error
}

// NewCommandHandler creates a new CommandHandler
func NewCommandHandler(cfg *config.Config) *CommandHandler {
	return &CommandHandler{
		config: cfg,
	}
}

// SetRegistry sets the command registry
func (h *CommandHandler) SetRegistry(registry CommandRegistry) {
	h.registry = registry
}

// Handle processes command messages
func (h *CommandHandler) Handle(message *tgbotapi.Message, ctx *config.WorkerContext) error {
	if !message.IsCommand() {
		return nil // Not a command, skip
	}

	// If no registry is set, log and skip
	if h.registry == nil {
		command := message.Command()
		slog.Debug("Command received but no registry set", "command", command, "chat_id", message.Chat.ID)
		return nil
	}

	// Delegate to registry
	if err := h.registry.Handle(message, ctx); err != nil {
		return err
	}

	// Check if this is a redo command - if so, don't stop the handler chain
	// The redo command sets a flag in the context and needs the ChatHandler to process it
	if ctx.Context != nil {
		if redoMode, ok := ctx.Context["redo_mode"].(bool); ok && redoMode {
			return nil // Continue to ChatHandler
		}
	}

	return nil
}

// ChatHandler processes chat messages
type ChatHandler struct {
	config *config.Config
}

// NewChatHandler creates a new ChatHandler
func NewChatHandler(cfg *config.Config) *ChatHandler {
	return &ChatHandler{
		config: cfg,
	}
}

// Handle processes chat messages
func (h *ChatHandler) Handle(message *tgbotapi.Message, ctx *config.WorkerContext) error {
	// Check if this is a redo command
	isRedoMode := false
	if ctx.Context != nil {
		if redoMode, ok := ctx.Context["redo_mode"].(bool); ok && redoMode {
			isRedoMode = true
		}
	}

	// Skip if it's a command (already handled), unless it's redo mode
	if message.IsCommand() && !isRedoMode {
		return nil
	}

	// Process chat message
	if err := chatWithMessage(message, ctx); err != nil {
		slog.Error("Failed to process chat message", "error", err, "chat_id", message.Chat.ID)
		return err
	}

	return nil
}

// MessageHandlerChain chains multiple message handlers
type MessageHandlerChain struct {
	handlers []MessageHandler
}

// NewMessageHandlerChain creates a new message handler chain
func NewMessageHandlerChain(handlers ...MessageHandler) *MessageHandlerChain {
	return &MessageHandlerChain{
		handlers: handlers,
	}
}

// Handle processes the message through all handlers in sequence
func (c *MessageHandlerChain) Handle(message *tgbotapi.Message, ctx *config.WorkerContext) error {
	for _, handler := range c.handlers {
		if err := handler.Handle(message, ctx); err != nil {
			// If a handler returns an error, stop the chain
			return err
		}
	}
	return nil
}

// Helper function to check if message is old (for SAFE_MODE)
func isOldMessage(message *tgbotapi.Message, maxAge time.Duration) bool {
	messageTime := time.Unix(int64(message.Date), 0)
	return time.Since(messageTime) > maxAge
}
