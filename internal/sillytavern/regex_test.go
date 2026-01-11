package sillytavern

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/tbxark/ChatGPT-Telegram-Workers/go_version/internal/storage"
)

// MockStorage for testing RegexProcessor
type mockRegexStorage struct {
	patterns map[uint]*storage.RegexPattern
	nextID   uint
}

func newMockRegexStorage() *mockRegexStorage {
	return &mockRegexStorage{
		patterns: make(map[uint]*storage.RegexPattern),
		nextID:   1,
	}
}

func (m *mockRegexStorage) CreateRegexPattern(pattern *storage.RegexPattern) error {
	pattern.ID = m.nextID
	m.nextID++
	m.patterns[pattern.ID] = pattern
	return nil
}

func (m *mockRegexStorage) GetRegexPattern(id uint) (*storage.RegexPattern, error) {
	pattern, ok := m.patterns[id]
	if !ok {
		return nil, storage.ErrNotFound
	}
	return pattern, nil
}

func (m *mockRegexStorage) ListRegexPatterns(userID *int64, patternType string) ([]*storage.RegexPattern, error) {
	var result []*storage.RegexPattern
	for _, pattern := range m.patterns {
		// Include global patterns and user's own patterns
		if pattern.UserID == nil || (userID != nil && pattern.UserID != nil && *pattern.UserID == *userID) {
			if patternType == "" || pattern.Type == patternType {
				result = append(result, pattern)
			}
		}
	}
	return result, nil
}

func (m *mockRegexStorage) UpdateRegexPattern(pattern *storage.RegexPattern) error {
	if _, ok := m.patterns[pattern.ID]; !ok {
		return storage.ErrNotFound
	}
	m.patterns[pattern.ID] = pattern
	return nil
}

func (m *mockRegexStorage) DeleteRegexPattern(id uint) error {
	delete(m.patterns, id)
	return nil
}

func (m *mockRegexStorage) UpdateRegexPatternStatus(id uint, enabled bool) error {
	pattern, ok := m.patterns[id]
	if !ok {
		return storage.ErrNotFound
	}
	pattern.Enabled = enabled
	return nil
}

// Implement other required Storage interface methods as no-ops
func (m *mockRegexStorage) GetChatHistory(ctx *storage.SessionContext) ([]storage.HistoryItem, error) {
	return nil, nil
}
func (m *mockRegexStorage) SaveChatHistory(ctx *storage.SessionContext, history []storage.HistoryItem) error {
	return nil
}
func (m *mockRegexStorage) DeleteChatHistory(ctx *storage.SessionContext) error { return nil }
func (m *mockRegexStorage) GetUserConfig(ctx *storage.SessionContext) (*storage.UserConfig, error) {
	return nil, nil
}
func (m *mockRegexStorage) SaveUserConfig(ctx *storage.SessionContext, config *storage.UserConfig) error {
	return nil
}
func (m *mockRegexStorage) GetMessageIDs(ctx *storage.SessionContext) ([]int, error) { return nil, nil }
func (m *mockRegexStorage) SaveMessageIDs(ctx *storage.SessionContext, ids []int) error {
	return nil
}
func (m *mockRegexStorage) GetGroupAdmins(chatID int64) ([]storage.ChatMember, error) {
	return nil, nil
}
func (m *mockRegexStorage) SaveGroupAdmins(chatID int64, admins []storage.ChatMember, ttl int) error {
	return nil
}
func (m *mockRegexStorage) CreateCharacterCard(card *storage.CharacterCard) error { return nil }
func (m *mockRegexStorage) GetCharacterCard(id uint) (*storage.CharacterCard, error) {
	return nil, storage.ErrNotFound
}
func (m *mockRegexStorage) ListCharacterCards(userID *int64) ([]*storage.CharacterCard, error) {
	return nil, nil
}
func (m *mockRegexStorage) UpdateCharacterCard(card *storage.CharacterCard) error { return nil }
func (m *mockRegexStorage) DeleteCharacterCard(id uint) error                     { return nil }
func (m *mockRegexStorage) ActivateCharacterCard(userID *int64, cardID uint) error {
	return nil
}
func (m *mockRegexStorage) GetActiveCharacterCard(userID *int64) (*storage.CharacterCard, error) {
	return nil, storage.ErrNotFound
}
func (m *mockRegexStorage) CreateWorldBook(book *storage.WorldBook) error { return nil }
func (m *mockRegexStorage) GetWorldBook(id uint) (*storage.WorldBook, error) {
	return nil, storage.ErrNotFound
}
func (m *mockRegexStorage) ListWorldBooks(userID *int64) ([]*storage.WorldBook, error) {
	return nil, nil
}
func (m *mockRegexStorage) UpdateWorldBook(book *storage.WorldBook) error { return nil }
func (m *mockRegexStorage) DeleteWorldBook(id uint) error                 { return nil }
func (m *mockRegexStorage) ActivateWorldBook(userID *int64, bookID uint) error {
	return nil
}
func (m *mockRegexStorage) GetActiveWorldBook(userID *int64) (*storage.WorldBook, error) {
	return nil, storage.ErrNotFound
}
func (m *mockRegexStorage) CreateWorldBookEntry(entry *storage.WorldBookEntry) error { return nil }
func (m *mockRegexStorage) GetWorldBookEntry(id uint) (*storage.WorldBookEntry, error) {
	return nil, storage.ErrNotFound
}
func (m *mockRegexStorage) ListWorldBookEntries(bookID uint) ([]*storage.WorldBookEntry, error) {
	return nil, nil
}
func (m *mockRegexStorage) UpdateWorldBookEntry(entry *storage.WorldBookEntry) error { return nil }
func (m *mockRegexStorage) DeleteWorldBookEntry(id uint) error                       { return nil }
func (m *mockRegexStorage) UpdateWorldBookEntryStatus(id uint, enabled bool) error   { return nil }
func (m *mockRegexStorage) CreatePreset(preset *storage.Preset) error                { return nil }
func (m *mockRegexStorage) GetPreset(id uint) (*storage.Preset, error) {
	return nil, storage.ErrNotFound
}
func (m *mockRegexStorage) ListPresets(userID *int64, apiType string) ([]*storage.Preset, error) {
	return nil, nil
}
func (m *mockRegexStorage) UpdatePreset(preset *storage.Preset) error { return nil }
func (m *mockRegexStorage) DeletePreset(id uint) error                { return nil }
func (m *mockRegexStorage) ActivatePreset(userID *int64, presetID uint) error {
	return nil
}
func (m *mockRegexStorage) GetActivePreset(userID *int64, apiType string) (*storage.Preset, error) {
	return nil, storage.ErrNotFound
}
func (m *mockRegexStorage) CreateLoginToken(token *storage.LoginToken) error { return nil }
func (m *mockRegexStorage) GetLoginToken(userID int64) (*storage.LoginToken, error) {
	return nil, storage.ErrNotFound
}
func (m *mockRegexStorage) ValidateLoginToken(userID int64, token string) (bool, error) {
	return false, nil
}
func (m *mockRegexStorage) DeleteLoginToken(userID int64) error { return nil }
func (m *mockRegexStorage) CleanupExpiredTokens() error         { return nil }
func (m *mockRegexStorage) DeleteAllChatHistory() error         { return nil }
func (m *mockRegexStorage) CleanupExpired() error               { return nil }
func (m *mockRegexStorage) Close() error                        { return nil }

func TestRegexProcessor_ProcessInput(t *testing.T) {
	mockStorage := newMockRegexStorage()
	processor := NewRegexProcessor(mockStorage)

	userID := int64(123)

	// Create a pattern to replace "hello" with "hi"
	pattern := &storage.RegexPattern{
		UserID:  &userID,
		Name:    "Replace Hello",
		Pattern: "hello",
		Replace: "hi",
		Type:    "input",
		Order:   1,
		Enabled: true,
	}
	mockStorage.CreateRegexPattern(pattern)

	// Process input
	result, err := processor.ProcessInput(&userID, "hello world")
	assert.NoError(t, err)
	assert.Equal(t, "hi world", result)
}

func TestRegexProcessor_ProcessOutput(t *testing.T) {
	mockStorage := newMockRegexStorage()
	processor := NewRegexProcessor(mockStorage)

	userID := int64(123)

	// Create a pattern to replace "goodbye" with "bye"
	pattern := &storage.RegexPattern{
		UserID:  &userID,
		Name:    "Replace Goodbye",
		Pattern: "goodbye",
		Replace: "bye",
		Type:    "output",
		Order:   1,
		Enabled: true,
	}
	mockStorage.CreateRegexPattern(pattern)

	// Process output
	result, err := processor.ProcessOutput(&userID, "goodbye world")
	assert.NoError(t, err)
	assert.Equal(t, "bye world", result)
}

func TestRegexProcessor_MultiplePatterns(t *testing.T) {
	mockStorage := newMockRegexStorage()
	processor := NewRegexProcessor(mockStorage)

	userID := int64(123)

	// Create multiple patterns
	pattern1 := &storage.RegexPattern{
		UserID:  &userID,
		Name:    "Pattern 1",
		Pattern: "hello",
		Replace: "hi",
		Type:    "input",
		Order:   1,
		Enabled: true,
	}
	pattern2 := &storage.RegexPattern{
		UserID:  &userID,
		Name:    "Pattern 2",
		Pattern: "world",
		Replace: "universe",
		Type:    "input",
		Order:   2,
		Enabled: true,
	}
	mockStorage.CreateRegexPattern(pattern1)
	mockStorage.CreateRegexPattern(pattern2)

	// Process input
	result, err := processor.ProcessInput(&userID, "hello world")
	assert.NoError(t, err)
	assert.Equal(t, "hi universe", result)
}

func TestRegexProcessor_DisabledPattern(t *testing.T) {
	mockStorage := newMockRegexStorage()
	processor := NewRegexProcessor(mockStorage)

	userID := int64(123)

	// Create a disabled pattern
	pattern := &storage.RegexPattern{
		UserID:  &userID,
		Name:    "Disabled Pattern",
		Pattern: "hello",
		Replace: "hi",
		Type:    "input",
		Order:   1,
		Enabled: false,
	}
	mockStorage.CreateRegexPattern(pattern)

	// Process input - pattern should not be applied
	result, err := processor.ProcessInput(&userID, "hello world")
	assert.NoError(t, err)
	assert.Equal(t, "hello world", result)
}

func TestRegexProcessor_UpdatePatternStatus(t *testing.T) {
	mockStorage := newMockRegexStorage()
	processor := NewRegexProcessor(mockStorage)

	userID := int64(123)

	pattern := &storage.RegexPattern{
		UserID:  &userID,
		Name:    "Test Pattern",
		Pattern: "test",
		Replace: "TEST",
		Type:    "input",
		Order:   1,
		Enabled: true,
	}
	mockStorage.CreateRegexPattern(pattern)

	// Disable the pattern
	err := processor.UpdatePatternStatus(pattern.ID, false)
	assert.NoError(t, err)

	// Verify it's disabled
	loaded, _ := mockStorage.GetRegexPattern(pattern.ID)
	assert.False(t, loaded.Enabled)
}

func TestValidateRegexPattern_Valid(t *testing.T) {
	err := validateRegexPattern("hello")
	assert.NoError(t, err)

	err = validateRegexPattern("[a-z]+")
	assert.NoError(t, err)

	err = validateRegexPattern("\\d{3}-\\d{4}")
	assert.NoError(t, err)
}

func TestValidateRegexPattern_Invalid(t *testing.T) {
	// Empty pattern
	err := validateRegexPattern("")
	assert.Error(t, err)

	// Invalid regex syntax
	err = validateRegexPattern("[")
	assert.Error(t, err)

	// Too long pattern
	longPattern := ""
	for i := 0; i < 1001; i++ {
		longPattern += "a"
	}
	err = validateRegexPattern(longPattern)
	assert.Error(t, err)
}

func TestRegexProcessor_ListPatterns(t *testing.T) {
	mockStorage := newMockRegexStorage()
	processor := NewRegexProcessor(mockStorage)

	userID := int64(123)

	// Create multiple patterns
	for i := 0; i < 3; i++ {
		pattern := &storage.RegexPattern{
			UserID:  &userID,
			Name:    "Pattern",
			Pattern: "test",
			Replace: "TEST",
			Type:    "input",
			Order:   i,
			Enabled: true,
		}
		mockStorage.CreateRegexPattern(pattern)
	}

	// List patterns
	patterns, err := processor.ListPatterns(&userID, "input")
	assert.NoError(t, err)
	assert.Len(t, patterns, 3)
}
