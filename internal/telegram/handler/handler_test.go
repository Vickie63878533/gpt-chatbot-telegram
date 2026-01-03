package handler

import (
	"testing"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/tbxark/ChatGPT-Telegram-Workers/go_version/internal/config"
	"github.com/tbxark/ChatGPT-Telegram-Workers/go_version/internal/i18n"
	"github.com/tbxark/ChatGPT-Telegram-Workers/go_version/internal/storage"
)

// mockStorage is a simple mock implementation of storage.Storage for testing
type mockStorage struct{}

func (m *mockStorage) GetChatHistory(ctx *storage.SessionContext) ([]storage.HistoryItem, error) {
	return []storage.HistoryItem{}, nil
}

func (m *mockStorage) SaveChatHistory(ctx *storage.SessionContext, history []storage.HistoryItem) error {
	return nil
}

func (m *mockStorage) DeleteChatHistory(ctx *storage.SessionContext) error {
	return nil
}

func (m *mockStorage) GetUserConfig(ctx *storage.SessionContext) (*storage.UserConfig, error) {
	return &storage.UserConfig{
		DefineKeys: []string{},
		Values:     make(map[string]interface{}),
	}, nil
}

func (m *mockStorage) SaveUserConfig(ctx *storage.SessionContext, config *storage.UserConfig) error {
	return nil
}

func (m *mockStorage) GetMessageIDs(ctx *storage.SessionContext) ([]int, error) {
	return []int{}, nil
}

func (m *mockStorage) SaveMessageIDs(ctx *storage.SessionContext, ids []int) error {
	return nil
}

func (m *mockStorage) GetGroupAdmins(chatID int64) ([]storage.ChatMember, error) {
	return []storage.ChatMember{}, nil
}

func (m *mockStorage) SaveGroupAdmins(chatID int64, admins []storage.ChatMember, ttl int) error {
	return nil
}

func (m *mockStorage) CleanupExpired() error {
	return nil
}

func (m *mockStorage) Close() error {
	return nil
}

func TestEnvChecker(t *testing.T) {
	checker := NewEnvChecker()

	// Test with nil database
	ctx := &config.WorkerContext{
		DB: nil,
	}
	update := &tgbotapi.Update{}

	err := checker.Handle(update, ctx)
	if err == nil {
		t.Error("Expected error when database is nil")
	}

	// Test with valid database
	ctx.DB = &mockStorage{}
	err = checker.Handle(update, ctx)
	if err != nil {
		t.Errorf("Expected no error with valid database, got: %v", err)
	}
}

func TestWhiteListFilter_GenerousMode(t *testing.T) {
	cfg := &config.Config{
		IAmAGenerousPerson: true,
	}
	i18nInstance := i18n.LoadI18n("en")
	filter := NewWhiteListFilter(cfg, i18nInstance)

	ctx := &config.WorkerContext{
		DB: &mockStorage{},
	}

	// Create a message from a non-whitelisted user
	update := &tgbotapi.Update{
		Message: &tgbotapi.Message{
			Chat: &tgbotapi.Chat{
				ID:   12345,
				Type: "private",
			},
		},
	}

	err := filter.Handle(update, ctx)
	if err != nil {
		t.Errorf("Expected no error in generous mode, got: %v", err)
	}
}

func TestWhiteListFilter_PrivateChat(t *testing.T) {
	cfg := &config.Config{
		IAmAGenerousPerson: false,
		ChatWhiteList:      []string{"12345"},
	}
	i18nInstance := i18n.LoadI18n("en")
	filter := NewWhiteListFilter(cfg, i18nInstance)

	ctx := &config.WorkerContext{
		DB: &mockStorage{},
	}

	// Test whitelisted user
	update := &tgbotapi.Update{
		Message: &tgbotapi.Message{
			Chat: &tgbotapi.Chat{
				ID:   12345,
				Type: "private",
			},
		},
	}

	err := filter.Handle(update, ctx)
	if err != nil {
		t.Errorf("Expected no error for whitelisted user, got: %v", err)
	}

	// Test non-whitelisted user
	update.Message.Chat.ID = 99999
	err = filter.Handle(update, ctx)
	if err == nil {
		t.Error("Expected error for non-whitelisted user")
	}
}

func TestUpdate2MessageHandler_IgnoreEditedMessages(t *testing.T) {
	handler := NewUpdate2MessageHandler([]MessageHandler{})

	ctx := &config.WorkerContext{
		DB: &mockStorage{},
	}

	// Test with edited message
	update := &tgbotapi.Update{
		EditedMessage: &tgbotapi.Message{
			Text: "edited",
		},
	}

	err := handler.Handle(update, ctx)
	if err != nil {
		t.Errorf("Expected no error for edited message (should be ignored), got: %v", err)
	}
}

func TestMessageFilter_SupportedTypes(t *testing.T) {
	cfg := &config.Config{}
	filter := NewMessageFilter(cfg)

	ctx := &config.WorkerContext{
		DB: &mockStorage{},
	}

	// Test text message
	message := &tgbotapi.Message{
		Text: "Hello",
		Chat: &tgbotapi.Chat{ID: 12345},
	}

	err := filter.Handle(message, ctx)
	if err != nil {
		t.Errorf("Expected no error for text message, got: %v", err)
	}

	// Test photo message
	message = &tgbotapi.Message{
		Photo: []tgbotapi.PhotoSize{{FileID: "photo1"}},
		Chat:  &tgbotapi.Chat{ID: 12345},
	}

	err = filter.Handle(message, ctx)
	if err != nil {
		t.Errorf("Expected no error for photo message, got: %v", err)
	}

	// Test command message
	message = &tgbotapi.Message{
		Text: "/start",
		Chat: &tgbotapi.Chat{ID: 12345},
		Entities: []tgbotapi.MessageEntity{
			{Type: "bot_command", Offset: 0, Length: 6},
		},
	}

	err = filter.Handle(message, ctx)
	if err != nil {
		t.Errorf("Expected no error for command message, got: %v", err)
	}
}

func TestMessageFilter_UnsupportedTypes(t *testing.T) {
	cfg := &config.Config{}
	filter := NewMessageFilter(cfg)

	ctx := &config.WorkerContext{
		DB: &mockStorage{},
	}

	// Test empty message (no text, no photo)
	message := &tgbotapi.Message{
		Chat: &tgbotapi.Chat{ID: 12345},
	}

	err := filter.Handle(message, ctx)
	if err == nil {
		t.Error("Expected error for unsupported message type")
	}
}

func TestOldMessageFilter_SafeMode(t *testing.T) {
	cfg := &config.Config{
		SafeMode:              true,
		GroupChatBotShareMode: true,
	}
	filter := NewOldMessageFilter(cfg)

	shareCtx := config.ShareContext{
		BotID: 123456,
	}

	ctx := &config.WorkerContext{
		ShareContext: shareCtx,
		DB:           &mockStorage{},
	}

	// First message should pass
	message := &tgbotapi.Message{
		MessageID: 1,
		Chat:      &tgbotapi.Chat{ID: 12345},
	}

	err := filter.Handle(message, ctx)
	if err != nil {
		t.Errorf("Expected no error for first message, got: %v", err)
	}
}

// mockCommandRegistry is a mock implementation for testing
type mockCommandRegistry struct{}

func (m *mockCommandRegistry) Handle(message *tgbotapi.Message, ctx *config.WorkerContext) error {
	return nil
}

func TestBuildHandlerChains(t *testing.T) {
	cfg := &config.Config{
		IAmAGenerousPerson: true,
		SafeMode:           false,
		DebugMode:          false,
	}
	i18nInstance := i18n.LoadI18n("en")

	// Create a mock command registry
	mockReg := &mockCommandRegistry{}

	// Test building update handler chain
	updateChain := BuildUpdateHandlerChain(cfg, i18nInstance, mockReg)
	if updateChain == nil {
		t.Error("Expected non-nil update handler chain")
	}

	// Test building message handler chain
	messageHandlers := BuildMessageHandlerChain(cfg, mockReg)
	if len(messageHandlers) == 0 {
		t.Error("Expected non-empty message handler chain")
	}

	// Verify we have the expected number of handlers
	expectedHandlers := 5 // SaveLastMessage, OldMessageFilter, MessageFilter, CommandHandler, ChatHandler
	if len(messageHandlers) != expectedHandlers {
		t.Errorf("Expected %d message handlers, got %d", expectedHandlers, len(messageHandlers))
	}
}
