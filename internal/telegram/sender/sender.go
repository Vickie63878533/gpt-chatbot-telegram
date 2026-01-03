package sender

import (
	"fmt"
	"sync"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/tbxark/ChatGPT-Telegram-Workers/go_version/internal/telegram/api"
)

// MessageSender handles sending messages to Telegram with streaming support
type MessageSender struct {
	client    *api.Client
	chatID    int64
	messageID int // Used for streaming updates
	context   map[string]interface{}
	mu        sync.Mutex

	// Streaming configuration
	minStreamInterval time.Duration
	lastUpdateTime    time.Time
}

// NewMessageSender creates a new MessageSender
func NewMessageSender(client *api.Client, chatID int64) *MessageSender {
	return &MessageSender{
		client:            client,
		chatID:            chatID,
		messageID:         0,
		context:           make(map[string]interface{}),
		minStreamInterval: 0,
		lastUpdateTime:    time.Time{},
	}
}

// SetMinStreamInterval sets the minimum interval between stream updates
func (s *MessageSender) SetMinStreamInterval(interval time.Duration) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.minStreamInterval = interval
}

// Update sets the message ID for streaming updates
func (s *MessageSender) Update(messageID int) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.messageID = messageID
}

// GetMessageID returns the current message ID
func (s *MessageSender) GetMessageID() int {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.messageID
}

// SetContext sets a context value
func (s *MessageSender) SetContext(key string, value interface{}) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.context[key] = value
}

// GetContext gets a context value
func (s *MessageSender) GetContext(key string) (interface{}, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	val, ok := s.context[key]
	return val, ok
}

// SendPlainText sends a plain text message or updates an existing one
func (s *MessageSender) SendPlainText(text string) error {
	return s.SendRichText(text, "")
}

// SendRichText sends a formatted text message with optional parse mode
func (s *MessageSender) SendRichText(text string, parseMode string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Check if we need to respect the minimum stream interval
	if s.minStreamInterval > 0 && !s.lastUpdateTime.IsZero() {
		elapsed := time.Since(s.lastUpdateTime)
		if elapsed < s.minStreamInterval {
			// Skip this update if it's too soon
			return nil
		}
	}

	if s.messageID == 0 {
		// First message - send a new message
		msg := tgbotapi.NewMessage(s.chatID, text)
		if parseMode != "" {
			msg.ParseMode = parseMode
		}

		sent, err := s.client.Send(msg)
		if err != nil {
			return fmt.Errorf("failed to send message: %w", err)
		}

		s.messageID = sent.MessageID
		s.lastUpdateTime = time.Now()
		return nil
	}

	// Update existing message
	edit := tgbotapi.NewEditMessageText(s.chatID, s.messageID, text)
	if parseMode != "" {
		edit.ParseMode = parseMode
	}

	_, err := s.client.Send(edit)
	if err != nil {
		// If edit fails (e.g., message is too old or identical), ignore the error
		// This is common in streaming scenarios
		return nil
	}

	s.lastUpdateTime = time.Now()
	return nil
}

// SendPhoto sends a photo message
func (s *MessageSender) SendPhoto(photoURL string) error {
	return s.SendPhotoWithCaption(photoURL, "")
}

// SendPhotoWithCaption sends a photo message with a caption
func (s *MessageSender) SendPhotoWithCaption(photoURL string, caption string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	var msg tgbotapi.PhotoConfig

	// Check if it's a URL or file ID
	if len(photoURL) > 0 && (photoURL[0] == 'h' || photoURL[0] == 'H') {
		// Assume it's a URL
		msg = tgbotapi.NewPhoto(s.chatID, tgbotapi.FileURL(photoURL))
	} else {
		// Assume it's a file ID or local path
		msg = tgbotapi.NewPhoto(s.chatID, tgbotapi.FileID(photoURL))
	}

	if caption != "" {
		msg.Caption = caption
	}

	sent, err := s.client.Send(msg)
	if err != nil {
		return fmt.Errorf("failed to send photo: %w", err)
	}

	// Store the message ID in case we need it later
	s.messageID = sent.MessageID
	return nil
}

// SendPhotoBytes sends a photo from byte data
func (s *MessageSender) SendPhotoBytes(photoData []byte, caption string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	msg := tgbotapi.NewPhoto(s.chatID, tgbotapi.FileBytes{
		Name:  "photo.jpg",
		Bytes: photoData,
	})

	if caption != "" {
		msg.Caption = caption
	}

	sent, err := s.client.Send(msg)
	if err != nil {
		return fmt.Errorf("failed to send photo bytes: %w", err)
	}

	s.messageID = sent.MessageID
	return nil
}

// SendRawMessage sends a raw Telegram message config
func (s *MessageSender) SendRawMessage(config tgbotapi.Chattable) (tgbotapi.Message, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	sent, err := s.client.Send(config)
	if err != nil {
		return tgbotapi.Message{}, fmt.Errorf("failed to send raw message: %w", err)
	}

	s.messageID = sent.MessageID
	return sent, nil
}

// SendMessageWithKeyboard sends a message with an inline keyboard
func (s *MessageSender) SendMessageWithKeyboard(text string, keyboard tgbotapi.InlineKeyboardMarkup, parseMode string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	msg := tgbotapi.NewMessage(s.chatID, text)
	msg.ReplyMarkup = keyboard
	if parseMode != "" {
		msg.ParseMode = parseMode
	}

	sent, err := s.client.Send(msg)
	if err != nil {
		return fmt.Errorf("failed to send message with keyboard: %w", err)
	}

	s.messageID = sent.MessageID
	return nil
}

// SendMessageWithReplyKeyboard sends a message with a reply keyboard
func (s *MessageSender) SendMessageWithReplyKeyboard(text string, keyboard tgbotapi.ReplyKeyboardMarkup, parseMode string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	msg := tgbotapi.NewMessage(s.chatID, text)
	msg.ReplyMarkup = keyboard
	if parseMode != "" {
		msg.ParseMode = parseMode
	}

	sent, err := s.client.Send(msg)
	if err != nil {
		return fmt.Errorf("failed to send message with reply keyboard: %w", err)
	}

	s.messageID = sent.MessageID
	return nil
}

// SendMessageWithReplyKeyboardRemove sends a message and removes the reply keyboard
func (s *MessageSender) SendMessageWithReplyKeyboardRemove(text string, parseMode string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	msg := tgbotapi.NewMessage(s.chatID, text)
	msg.ReplyMarkup = tgbotapi.NewRemoveKeyboard(true)
	if parseMode != "" {
		msg.ParseMode = parseMode
	}

	sent, err := s.client.Send(msg)
	if err != nil {
		return fmt.Errorf("failed to send message with keyboard removal: %w", err)
	}

	s.messageID = sent.MessageID
	return nil
}

// SendChatAction sends a chat action (typing, upload_photo, etc.)
func (s *MessageSender) SendChatAction(action string) error {
	return s.client.SendChatAction(s.chatID, action)
}

// EditMessageReplyMarkup edits the reply markup of the current message
func (s *MessageSender) EditMessageReplyMarkup(keyboard tgbotapi.InlineKeyboardMarkup) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.messageID == 0 {
		return fmt.Errorf("no message to edit")
	}

	edit := tgbotapi.NewEditMessageReplyMarkup(s.chatID, s.messageID, keyboard)
	_, err := s.client.Send(edit)
	if err != nil {
		return fmt.Errorf("failed to edit message reply markup: %w", err)
	}

	return nil
}

// DeleteMessage deletes the current message
func (s *MessageSender) DeleteMessage() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.messageID == 0 {
		return fmt.Errorf("no message to delete")
	}

	deleteConfig := tgbotapi.NewDeleteMessage(s.chatID, s.messageID)
	_, err := s.client.Request(deleteConfig)
	if err != nil {
		return fmt.Errorf("failed to delete message: %w", err)
	}

	s.messageID = 0
	return nil
}

// Reset resets the sender state (clears message ID and context)
func (s *MessageSender) Reset() {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.messageID = 0
	s.context = make(map[string]interface{})
	s.lastUpdateTime = time.Time{}
}
