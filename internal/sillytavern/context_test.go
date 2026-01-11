package sillytavern

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/tbxark/ChatGPT-Telegram-Workers/go_version/internal/agent"
	"github.com/tbxark/ChatGPT-Telegram-Workers/go_version/internal/config"
	"github.com/tbxark/ChatGPT-Telegram-Workers/go_version/internal/storage"
)

// MockContextStorage is a simple mock for context manager testing
type MockContextStorage struct {
	history map[string][]storage.HistoryItem
}

func NewMockContextStorage() *MockContextStorage {
	return &MockContextStorage{
		history: make(map[string][]storage.HistoryItem),
	}
}

func (m *MockContextStorage) sessionKey(ctx *storage.SessionContext) string {
	userID := int64(0)
	if ctx.UserID != nil {
		userID = *ctx.UserID
	}
	threadID := int64(0)
	if ctx.ThreadID != nil {
		threadID = *ctx.ThreadID
	}
	return string(rune(ctx.ChatID)) + string(rune(ctx.BotID)) + string(rune(userID)) + string(rune(threadID))
}

func (m *MockContextStorage) GetChatHistory(ctx *storage.SessionContext) ([]storage.HistoryItem, error) {
	key := m.sessionKey(ctx)
	if history, ok := m.history[key]; ok {
		return history, nil
	}
	return []storage.HistoryItem{}, nil
}

func (m *MockContextStorage) SaveChatHistory(ctx *storage.SessionContext, history []storage.HistoryItem) error {
	key := m.sessionKey(ctx)
	m.history[key] = history
	return nil
}

func (m *MockContextStorage) DeleteChatHistory(ctx *storage.SessionContext) error {
	key := m.sessionKey(ctx)
	delete(m.history, key)
	return nil
}

// Stub implementations for other methods
func (m *MockContextStorage) GetUserConfig(ctx *storage.SessionContext) (*storage.UserConfig, error) {
	return &storage.UserConfig{}, nil
}
func (m *MockContextStorage) SaveUserConfig(ctx *storage.SessionContext, config *storage.UserConfig) error {
	return nil
}
func (m *MockContextStorage) GetMessageIDs(ctx *storage.SessionContext) ([]int, error) {
	return []int{}, nil
}
func (m *MockContextStorage) SaveMessageIDs(ctx *storage.SessionContext, ids []int) error {
	return nil
}
func (m *MockContextStorage) GetGroupAdmins(chatID int64) ([]storage.ChatMember, error) {
	return []storage.ChatMember{}, nil
}
func (m *MockContextStorage) SaveGroupAdmins(chatID int64, admins []storage.ChatMember, ttl int) error {
	return nil
}
func (m *MockContextStorage) CreateCharacterCard(card *storage.CharacterCard) error {
	return nil
}
func (m *MockContextStorage) GetCharacterCard(id uint) (*storage.CharacterCard, error) {
	return nil, storage.ErrNotFound
}
func (m *MockContextStorage) ListCharacterCards(userID *int64) ([]*storage.CharacterCard, error) {
	return []*storage.CharacterCard{}, nil
}
func (m *MockContextStorage) UpdateCharacterCard(card *storage.CharacterCard) error {
	return nil
}
func (m *MockContextStorage) DeleteCharacterCard(id uint) error {
	return nil
}
func (m *MockContextStorage) ActivateCharacterCard(userID *int64, cardID uint) error {
	return nil
}
func (m *MockContextStorage) GetActiveCharacterCard(userID *int64) (*storage.CharacterCard, error) {
	return nil, storage.ErrNotFound
}
func (m *MockContextStorage) CreateWorldBook(book *storage.WorldBook) error {
	return nil
}
func (m *MockContextStorage) GetWorldBook(id uint) (*storage.WorldBook, error) {
	return nil, storage.ErrNotFound
}
func (m *MockContextStorage) ListWorldBooks(userID *int64) ([]*storage.WorldBook, error) {
	return []*storage.WorldBook{}, nil
}
func (m *MockContextStorage) UpdateWorldBook(book *storage.WorldBook) error {
	return nil
}
func (m *MockContextStorage) DeleteWorldBook(id uint) error {
	return nil
}
func (m *MockContextStorage) ActivateWorldBook(userID *int64, bookID uint) error {
	return nil
}
func (m *MockContextStorage) GetActiveWorldBook(userID *int64) (*storage.WorldBook, error) {
	return nil, storage.ErrNotFound
}
func (m *MockContextStorage) CreateWorldBookEntry(entry *storage.WorldBookEntry) error {
	return nil
}
func (m *MockContextStorage) GetWorldBookEntry(id uint) (*storage.WorldBookEntry, error) {
	return nil, storage.ErrNotFound
}
func (m *MockContextStorage) ListWorldBookEntries(bookID uint) ([]*storage.WorldBookEntry, error) {
	return []*storage.WorldBookEntry{}, nil
}
func (m *MockContextStorage) UpdateWorldBookEntry(entry *storage.WorldBookEntry) error {
	return nil
}
func (m *MockContextStorage) DeleteWorldBookEntry(id uint) error {
	return nil
}
func (m *MockContextStorage) UpdateWorldBookEntryStatus(id uint, enabled bool) error {
	return nil
}
func (m *MockContextStorage) CreatePreset(preset *storage.Preset) error {
	return nil
}
func (m *MockContextStorage) GetPreset(id uint) (*storage.Preset, error) {
	return nil, storage.ErrNotFound
}
func (m *MockContextStorage) ListPresets(userID *int64, apiType string) ([]*storage.Preset, error) {
	return []*storage.Preset{}, nil
}
func (m *MockContextStorage) UpdatePreset(preset *storage.Preset) error {
	return nil
}
func (m *MockContextStorage) DeletePreset(id uint) error {
	return nil
}
func (m *MockContextStorage) ActivatePreset(userID *int64, presetID uint) error {
	return nil
}
func (m *MockContextStorage) GetActivePreset(userID *int64, apiType string) (*storage.Preset, error) {
	return nil, storage.ErrNotFound
}
func (m *MockContextStorage) CreateRegexPattern(pattern *storage.RegexPattern) error {
	return nil
}
func (m *MockContextStorage) GetRegexPattern(id uint) (*storage.RegexPattern, error) {
	return nil, storage.ErrNotFound
}
func (m *MockContextStorage) ListRegexPatterns(userID *int64, patternType string) ([]*storage.RegexPattern, error) {
	return []*storage.RegexPattern{}, nil
}
func (m *MockContextStorage) UpdateRegexPattern(pattern *storage.RegexPattern) error {
	return nil
}
func (m *MockContextStorage) DeleteRegexPattern(id uint) error {
	return nil
}
func (m *MockContextStorage) UpdateRegexPatternStatus(id uint, enabled bool) error {
	return nil
}
func (m *MockContextStorage) CreateLoginToken(token *storage.LoginToken) error {
	return nil
}
func (m *MockContextStorage) GetLoginToken(userID int64) (*storage.LoginToken, error) {
	return nil, storage.ErrNotFound
}
func (m *MockContextStorage) ValidateLoginToken(userID int64, token string) (bool, error) {
	return false, nil
}
func (m *MockContextStorage) DeleteLoginToken(userID int64) error {
	return nil
}
func (m *MockContextStorage) CleanupExpiredTokens() error {
	return nil
}
func (m *MockContextStorage) DeleteAllChatHistory() error {
	return nil
}
func (m *MockContextStorage) CleanupExpired() error {
	return nil
}
func (m *MockContextStorage) Close() error {
	return nil
}

// MockContextChatAgent is a simple mock for chat agent testing
type MockContextChatAgent struct {
	response *agent.ChatAgentResponse
	err      error
}

func (m *MockContextChatAgent) Name() string {
	return "mock"
}

func (m *MockContextChatAgent) ModelKey() string {
	return "MOCK_MODEL"
}

func (m *MockContextChatAgent) Enable(config *config.Config) bool {
	return true
}

func (m *MockContextChatAgent) Model(config *config.Config) string {
	return "mock-model"
}

func (m *MockContextChatAgent) ModelList(config *config.Config) ([]string, error) {
	return []string{"mock-model"}, nil
}

func (m *MockContextChatAgent) Request(ctx context.Context, params *agent.LLMChatParams, config *config.Config, onStream agent.ChatStreamTextHandler) (*agent.ChatAgentResponse, error) {
	if m.err != nil {
		return nil, m.err
	}
	if m.response != nil {
		return m.response, nil
	}
	return &agent.ChatAgentResponse{
		Messages: []agent.HistoryItem{
			{Role: "assistant", Content: "This is a summary of the conversation."},
		},
	}, nil
}

// Test EstimateTokens
func TestEstimateTokens(t *testing.T) {
	mockStorage := NewMockContextStorage()
	manager := NewContextManager(mockStorage, nil, nil, nil)

	messages := []storage.HistoryItem{
		{Role: "user", Content: "Hello, how are you?"},
		{Role: "assistant", Content: "I'm doing well, thank you!"},
	}

	tokens := manager.EstimateTokens(messages)
	assert.Greater(t, tokens, 0, "Should estimate some tokens")
}

// Test AddMessage
func TestAddMessage(t *testing.T) {
	mockStorage := NewMockContextStorage()
	mockAgent := &MockContextChatAgent{}
	cfg := &config.Config{}
	
	manager := NewContextManager(mockStorage, nil, mockAgent, cfg)

	ctx := &storage.SessionContext{
		ChatID: 123,
		BotID:  456,
	}

	// Add initial message
	err := mockStorage.SaveChatHistory(ctx, []storage.HistoryItem{
		{Role: "user", Content: "Hello"},
	})
	assert.NoError(t, err)

	// Add new message
	err = manager.AddMessage(ctx, "assistant", "Hi there!")
	assert.NoError(t, err)

	// Verify history was updated
	history, err := mockStorage.GetChatHistory(ctx)
	assert.NoError(t, err)
	assert.Equal(t, 2, len(history))
	assert.Equal(t, "assistant", history[1].Role)
}

// Test GetBuildHistory with truncation
func TestGetBuildHistory_WithTruncation(t *testing.T) {
	mockStorage := NewMockContextStorage()
	manager := NewContextManager(mockStorage, nil, nil, nil)

	ctx := &storage.SessionContext{
		ChatID: 123,
		BotID:  456,
	}

	fullHistory := []storage.HistoryItem{
		{Role: "user", Content: "Old message 1"},
		{Role: "assistant", Content: "Old response 1"},
		{Role: "system", Content: "[Conversation cleared by user]", Truncated: true},
		{Role: "user", Content: "New message 1"},
		{Role: "assistant", Content: "New response 1"},
	}

	mockStorage.SaveChatHistory(ctx, fullHistory)

	buildHistory, err := manager.GetBuildHistory(ctx)
	assert.NoError(t, err)
	assert.Equal(t, 2, len(buildHistory), "Should only include messages after truncation")
	assert.Equal(t, "New message 1", buildHistory[0].Content)
}

// Test GetFullHistory
func TestGetFullHistory(t *testing.T) {
	mockStorage := NewMockContextStorage()
	manager := NewContextManager(mockStorage, nil, nil, nil)

	ctx := &storage.SessionContext{
		ChatID: 123,
		BotID:  456,
	}

	fullHistory := []storage.HistoryItem{
		{Role: "system", Content: "System message"},
		{Role: "user", Content: "User message 1"},
		{Role: "assistant", Content: "Assistant response 1"},
		{Role: "summary", Content: "Summary"},
		{Role: "user", Content: "User message 2"},
		{Role: "assistant", Content: "Assistant response 2"},
	}

	mockStorage.SaveChatHistory(ctx, fullHistory)

	result, err := manager.GetFullHistory(ctx)
	assert.NoError(t, err)
	assert.Equal(t, 4, len(result), "Should only include user and assistant messages")
	
	for _, msg := range result {
		assert.True(t, msg.Role == "user" || msg.Role == "assistant")
	}
}

// Test ClearHistory
func TestClearHistory(t *testing.T) {
	mockStorage := NewMockContextStorage()
	manager := NewContextManager(mockStorage, nil, nil, nil)

	ctx := &storage.SessionContext{
		ChatID: 123,
		BotID:  456,
	}

	existingHistory := []storage.HistoryItem{
		{Role: "user", Content: "Message 1"},
		{Role: "assistant", Content: "Response 1"},
	}

	mockStorage.SaveChatHistory(ctx, existingHistory)

	err := manager.ClearHistory(ctx)
	assert.NoError(t, err)

	// Verify truncation marker was added
	history, err := mockStorage.GetChatHistory(ctx)
	assert.NoError(t, err)
	assert.Equal(t, 3, len(history))
	assert.True(t, history[2].Truncated)
}

// Test TriggerSummary
func TestTriggerSummary(t *testing.T) {
	mockStorage := NewMockContextStorage()
	mockAgent := &MockContextChatAgent{}
	cfg := &config.Config{}
	
	contextConfig := &ContextConfig{
		MaxContextLength: 8000,
		SummaryThreshold: 0.8,
		MinRecentPairs:   2,
		TokensPerMessage: 10,
		TokensPerChar:    0.25,
	}
	
	manager := NewContextManager(mockStorage, contextConfig, mockAgent, cfg)

	ctx := &storage.SessionContext{
		ChatID: 123,
		BotID:  456,
	}

	// Create history with enough messages to trigger summary
	existingHistory := []storage.HistoryItem{
		{Role: "system", Content: "System prompt"},
		{Role: "user", Content: "Message 1"},
		{Role: "assistant", Content: "Response 1"},
		{Role: "user", Content: "Message 2"},
		{Role: "assistant", Content: "Response 2"},
		{Role: "user", Content: "Message 3"},
		{Role: "assistant", Content: "Response 3"},
		{Role: "user", Content: "Message 4"},
		{Role: "assistant", Content: "Response 4"},
	}

	mockStorage.SaveChatHistory(ctx, existingHistory)

	err := manager.TriggerSummary(ctx)
	assert.NoError(t, err)

	// Verify summary was added
	history, err := mockStorage.GetChatHistory(ctx)
	assert.NoError(t, err)
	
	hasSummary := false
	for _, msg := range history {
		if msg.Role == "summary" {
			hasSummary = true
			break
		}
	}
	assert.True(t, hasSummary, "Should have a summary message")
}
