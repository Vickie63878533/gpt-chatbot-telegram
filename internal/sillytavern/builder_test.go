package sillytavern

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tbxark/ChatGPT-Telegram-Workers/go_version/internal/storage"
)

// mockStorage is a minimal mock for testing
type mockStorage struct {
	storage.Storage
	activeCard   *storage.CharacterCard
	activeBook   *storage.WorldBook
	activePreset *storage.Preset
	bookEntries  []*storage.WorldBookEntry
}

func (m *mockStorage) GetActiveCharacterCard(userID *int64) (*storage.CharacterCard, error) {
	if m.activeCard == nil {
		return nil, storage.ErrNotFound
	}
	return m.activeCard, nil
}

func (m *mockStorage) GetActiveWorldBook(userID *int64) (*storage.WorldBook, error) {
	if m.activeBook == nil {
		return nil, storage.ErrNotFound
	}
	return m.activeBook, nil
}

func (m *mockStorage) GetActivePreset(userID *int64, apiType string) (*storage.Preset, error) {
	if m.activePreset == nil {
		return nil, storage.ErrNotFound
	}
	return m.activePreset, nil
}

func (m *mockStorage) ListRegexPatterns(userID *int64, patternType string) ([]*storage.RegexPattern, error) {
	return []*storage.RegexPattern{}, nil
}

func (m *mockStorage) ListWorldBookEntries(bookID uint) ([]*storage.WorldBookEntry, error) {
	return m.bookEntries, nil
}

func TestRequestBuilder_BuildRequest_Basic(t *testing.T) {
	// Create mock storage
	mock := &mockStorage{}

	// Create managers
	charManager := NewCharacterCardManager(mock)
	worldManager := NewWorldBookManager(mock)
	presetManager := NewPresetManager(mock)
	regexProcessor := NewRegexProcessor(mock)

	// Create builder
	builder := NewRequestBuilder(charManager, worldManager, presetManager, regexProcessor)

	// Create build context
	userID := int64(123)
	ctx := &BuildContext{
		UserID: &userID,
		History: []storage.HistoryItem{
			{Role: "user", Content: "Hello"},
			{Role: "assistant", Content: "Hi there!"},
		},
		CurrentInput: "How are you?",
		APIType:      "openai",
	}

	// Build request
	request, err := builder.BuildRequest(ctx)
	require.NoError(t, err)
	require.NotNil(t, request)

	// Verify messages
	assert.GreaterOrEqual(t, len(request.Messages), 3) // At least the history + current input
}

func TestRequestBuilder_BuildSystemPrompt(t *testing.T) {
	builder := &RequestBuilder{}

	t.Run("with system_prompt", func(t *testing.T) {
		cardData := &CharacterCardV2{
			Data: CharacterCardV2Data{
				SystemPrompt: "You are a helpful assistant.",
			},
		}

		prompt := builder.buildSystemPrompt(cardData)
		assert.Equal(t, "You are a helpful assistant.", prompt)
	})

	t.Run("without system_prompt", func(t *testing.T) {
		cardData := &CharacterCardV2{
			Data: CharacterCardV2Data{
				Description: "A friendly bot",
				Personality: "Helpful and kind",
				Scenario:    "Chatting with users",
			},
		}

		prompt := builder.buildSystemPrompt(cardData)
		assert.Contains(t, prompt, "A friendly bot")
		assert.Contains(t, prompt, "Helpful and kind")
		assert.Contains(t, prompt, "Chatting with users")
	})

	t.Run("nil character data", func(t *testing.T) {
		prompt := builder.buildSystemPrompt(nil)
		assert.Empty(t, prompt)
	})
}

func TestRequestBuilder_EnforceRoleAlternation(t *testing.T) {
	builder := &RequestBuilder{}

	t.Run("basic alternation", func(t *testing.T) {
		history := []storage.HistoryItem{
			{Role: "user", Content: "Hello"},
			{Role: "assistant", Content: "Hi"},
			{Role: "user", Content: "How are you?"},
			{Role: "assistant", Content: "I'm good"},
		}

		messages := builder.enforceRoleAlternation(history, "Great!")
		
		// Should have alternating roles
		for i := 1; i < len(messages); i++ {
			assert.NotEqual(t, messages[i-1].Role, messages[i].Role, "Messages should alternate roles")
		}

		// First should be user
		assert.Equal(t, "user", messages[0].Role)

		// Last should be user (current input)
		assert.Equal(t, "user", messages[len(messages)-1].Role)
	})

	t.Run("consecutive same roles", func(t *testing.T) {
		history := []storage.HistoryItem{
			{Role: "user", Content: "Hello"},
			{Role: "user", Content: "Are you there?"},
			{Role: "assistant", Content: "Yes"},
		}

		messages := builder.enforceRoleAlternation(history, "Good")

		// Should merge consecutive user messages
		assert.Equal(t, "user", messages[0].Role)
		assert.Contains(t, messages[0].Content, "Hello")
		assert.Contains(t, messages[0].Content, "Are you there?")
	})

	t.Run("starts with assistant", func(t *testing.T) {
		history := []storage.HistoryItem{
			{Role: "assistant", Content: "Hello there!"},
		}

		messages := builder.enforceRoleAlternation(history, "Hi")

		// Should add empty user message at start
		assert.Equal(t, "user", messages[0].Role)
	})

	t.Run("skips system messages", func(t *testing.T) {
		history := []storage.HistoryItem{
			{Role: "system", Content: "System message"},
			{Role: "user", Content: "Hello"},
			{Role: "assistant", Content: "Hi"},
		}

		messages := builder.enforceRoleAlternation(history, "")

		// System messages should be skipped
		for _, msg := range messages {
			assert.NotEqual(t, "system", msg.Role)
		}
	})
}

func TestRequestBuilder_ApplyPresetParameters(t *testing.T) {
	builder := &RequestBuilder{}

	request := &AIRequest{}
	presetData := &PresetData{
		Temperature:       0.7,
		TopP:              0.9,
		TopK:              40,
		MaxTokens:         2048,
		PresencePenalty:   0.1,
		FrequencyPenalty:  0.2,
		StopSequences:     []string{"\n\nHuman:", "\n\nAssistant:"},
	}

	builder.applyPresetParameters(request, presetData)

	assert.Equal(t, 0.7, request.Temperature)
	assert.Equal(t, 0.9, request.TopP)
	assert.Equal(t, 40, request.TopK)
	assert.Equal(t, 2048, request.MaxTokens)
	assert.Equal(t, 0.1, request.PresencePenalty)
	assert.Equal(t, 0.2, request.FrequencyPenalty)
	assert.Equal(t, []string{"\n\nHuman:", "\n\nAssistant:"}, request.StopSequences)
}

func TestRequestBuilder_WithCharacterCard(t *testing.T) {
	// Create character card data
	cardData := CharacterCardV2{
		Spec:        "chara_card_v2",
		SpecVersion: "2.0",
		Data: CharacterCardV2Data{
			Name:         "TestBot",
			SystemPrompt: "You are a test bot.",
		},
	}
	cardJSON, _ := json.Marshal(cardData)

	// Create mock storage with active card
	mock := &mockStorage{
		activeCard: &storage.CharacterCard{
			ID:   1,
			Name: "TestBot",
			Data: string(cardJSON),
		},
	}

	// Create managers
	charManager := NewCharacterCardManager(mock)
	worldManager := NewWorldBookManager(mock)
	presetManager := NewPresetManager(mock)
	regexProcessor := NewRegexProcessor(mock)

	// Create builder
	builder := NewRequestBuilder(charManager, worldManager, presetManager, regexProcessor)

	// Create build context
	userID := int64(123)
	ctx := &BuildContext{
		UserID:       &userID,
		History:      []storage.HistoryItem{},
		CurrentInput: "Hello",
		APIType:      "openai",
	}

	// Build request
	request, err := builder.BuildRequest(ctx)
	require.NoError(t, err)
	require.NotNil(t, request)

	// Should have system message from character card
	assert.GreaterOrEqual(t, len(request.Messages), 1)
	assert.Equal(t, "system", request.Messages[0].Role)
	assert.Equal(t, "You are a test bot.", request.Messages[0].Content)
}

func TestRequestBuilder_WithPreset(t *testing.T) {
	// Create preset data
	presetData := PresetData{
		Name:        "TestPreset",
		Temperature: 0.8,
		MaxTokens:   1024,
	}
	presetJSON, _ := json.Marshal(presetData)

	// Create mock storage with active preset
	mock := &mockStorage{
		activePreset: &storage.Preset{
			ID:      1,
			Name:    "TestPreset",
			APIType: "openai",
			Data:    string(presetJSON),
		},
	}

	// Create managers
	charManager := NewCharacterCardManager(mock)
	worldManager := NewWorldBookManager(mock)
	presetManager := NewPresetManager(mock)
	regexProcessor := NewRegexProcessor(mock)

	// Create builder
	builder := NewRequestBuilder(charManager, worldManager, presetManager, regexProcessor)

	// Create build context
	userID := int64(123)
	ctx := &BuildContext{
		UserID:       &userID,
		History:      []storage.HistoryItem{},
		CurrentInput: "Hello",
		APIType:      "openai",
	}

	// Build request
	request, err := builder.BuildRequest(ctx)
	require.NoError(t, err)
	require.NotNil(t, request)

	// Should have preset parameters applied
	assert.Equal(t, 0.8, request.Temperature)
	assert.Equal(t, 1024, request.MaxTokens)
}
