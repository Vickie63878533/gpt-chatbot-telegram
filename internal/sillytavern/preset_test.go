package sillytavern

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/tbxark/ChatGPT-Telegram-Workers/go_version/internal/storage"
)

// MockStorage for testing PresetManager
type mockPresetStorage struct {
	presets       map[uint]*storage.Preset
	nextID        uint
	activePresets map[string]uint // key: userID_apiType
}

func newMockPresetStorage() *mockPresetStorage {
	return &mockPresetStorage{
		presets:       make(map[uint]*storage.Preset),
		nextID:        1,
		activePresets: make(map[string]uint),
	}
}

func (m *mockPresetStorage) CreatePreset(preset *storage.Preset) error {
	preset.ID = m.nextID
	m.nextID++
	m.presets[preset.ID] = preset
	return nil
}

func (m *mockPresetStorage) GetPreset(id uint) (*storage.Preset, error) {
	preset, ok := m.presets[id]
	if !ok {
		return nil, storage.ErrNotFound
	}
	return preset, nil
}

func (m *mockPresetStorage) ListPresets(userID *int64, apiType string) ([]*storage.Preset, error) {
	var result []*storage.Preset
	for _, preset := range m.presets {
		// Include global presets and user's own presets
		if preset.UserID == nil || (userID != nil && preset.UserID != nil && *preset.UserID == *userID) {
			if apiType == "" || preset.APIType == apiType {
				result = append(result, preset)
			}
		}
	}
	return result, nil
}

func (m *mockPresetStorage) UpdatePreset(preset *storage.Preset) error {
	if _, ok := m.presets[preset.ID]; !ok {
		return storage.ErrNotFound
	}
	m.presets[preset.ID] = preset
	return nil
}

func (m *mockPresetStorage) DeletePreset(id uint) error {
	delete(m.presets, id)
	return nil
}

func (m *mockPresetStorage) ActivatePreset(userID *int64, presetID uint) error {
	preset, ok := m.presets[presetID]
	if !ok {
		return storage.ErrNotFound
	}

	// Deactivate all presets for this user and API type
	for _, p := range m.presets {
		if p.APIType == preset.APIType {
			if (p.UserID == nil && userID == nil) || (p.UserID != nil && userID != nil && *p.UserID == *userID) {
				p.IsActive = false
			}
		}
	}

	// Activate the selected preset
	preset.IsActive = true

	// Store in active presets map
	key := ""
	if userID != nil {
		key = string(rune(*userID)) + "_" + preset.APIType
	} else {
		key = "global_" + preset.APIType
	}
	m.activePresets[key] = presetID

	return nil
}

func (m *mockPresetStorage) GetActivePreset(userID *int64, apiType string) (*storage.Preset, error) {
	for _, preset := range m.presets {
		if preset.IsActive && preset.APIType == apiType {
			if (preset.UserID == nil && userID == nil) || (preset.UserID != nil && userID != nil && *preset.UserID == *userID) {
				return preset, nil
			}
		}
	}
	return nil, storage.ErrNotFound
}

// Implement other required Storage interface methods as no-ops
func (m *mockPresetStorage) GetChatHistory(ctx *storage.SessionContext) ([]storage.HistoryItem, error) {
	return nil, nil
}
func (m *mockPresetStorage) SaveChatHistory(ctx *storage.SessionContext, history []storage.HistoryItem) error {
	return nil
}
func (m *mockPresetStorage) DeleteChatHistory(ctx *storage.SessionContext) error { return nil }
func (m *mockPresetStorage) GetUserConfig(ctx *storage.SessionContext) (*storage.UserConfig, error) {
	return nil, nil
}
func (m *mockPresetStorage) SaveUserConfig(ctx *storage.SessionContext, config *storage.UserConfig) error {
	return nil
}
func (m *mockPresetStorage) GetMessageIDs(ctx *storage.SessionContext) ([]int, error) { return nil, nil }
func (m *mockPresetStorage) SaveMessageIDs(ctx *storage.SessionContext, ids []int) error {
	return nil
}
func (m *mockPresetStorage) GetGroupAdmins(chatID int64) ([]storage.ChatMember, error) {
	return nil, nil
}
func (m *mockPresetStorage) SaveGroupAdmins(chatID int64, admins []storage.ChatMember, ttl int) error {
	return nil
}
func (m *mockPresetStorage) CreateCharacterCard(card *storage.CharacterCard) error { return nil }
func (m *mockPresetStorage) GetCharacterCard(id uint) (*storage.CharacterCard, error) {
	return nil, storage.ErrNotFound
}
func (m *mockPresetStorage) ListCharacterCards(userID *int64) ([]*storage.CharacterCard, error) {
	return nil, nil
}
func (m *mockPresetStorage) UpdateCharacterCard(card *storage.CharacterCard) error { return nil }
func (m *mockPresetStorage) DeleteCharacterCard(id uint) error                     { return nil }
func (m *mockPresetStorage) ActivateCharacterCard(userID *int64, cardID uint) error {
	return nil
}
func (m *mockPresetStorage) GetActiveCharacterCard(userID *int64) (*storage.CharacterCard, error) {
	return nil, storage.ErrNotFound
}
func (m *mockPresetStorage) CreateWorldBook(book *storage.WorldBook) error { return nil }
func (m *mockPresetStorage) GetWorldBook(id uint) (*storage.WorldBook, error) {
	return nil, storage.ErrNotFound
}
func (m *mockPresetStorage) ListWorldBooks(userID *int64) ([]*storage.WorldBook, error) {
	return nil, nil
}
func (m *mockPresetStorage) UpdateWorldBook(book *storage.WorldBook) error { return nil }
func (m *mockPresetStorage) DeleteWorldBook(id uint) error                 { return nil }
func (m *mockPresetStorage) ActivateWorldBook(userID *int64, bookID uint) error {
	return nil
}
func (m *mockPresetStorage) GetActiveWorldBook(userID *int64) (*storage.WorldBook, error) {
	return nil, storage.ErrNotFound
}
func (m *mockPresetStorage) CreateWorldBookEntry(entry *storage.WorldBookEntry) error { return nil }
func (m *mockPresetStorage) GetWorldBookEntry(id uint) (*storage.WorldBookEntry, error) {
	return nil, storage.ErrNotFound
}
func (m *mockPresetStorage) ListWorldBookEntries(bookID uint) ([]*storage.WorldBookEntry, error) {
	return nil, nil
}
func (m *mockPresetStorage) UpdateWorldBookEntry(entry *storage.WorldBookEntry) error { return nil }
func (m *mockPresetStorage) DeleteWorldBookEntry(id uint) error                       { return nil }
func (m *mockPresetStorage) UpdateWorldBookEntryStatus(id uint, enabled bool) error   { return nil }
func (m *mockPresetStorage) CreateRegexPattern(pattern *storage.RegexPattern) error   { return nil }
func (m *mockPresetStorage) GetRegexPattern(id uint) (*storage.RegexPattern, error) {
	return nil, storage.ErrNotFound
}
func (m *mockPresetStorage) ListRegexPatterns(userID *int64, patternType string) ([]*storage.RegexPattern, error) {
	return nil, nil
}
func (m *mockPresetStorage) UpdateRegexPattern(pattern *storage.RegexPattern) error { return nil }
func (m *mockPresetStorage) DeleteRegexPattern(id uint) error                       { return nil }
func (m *mockPresetStorage) UpdateRegexPatternStatus(id uint, enabled bool) error   { return nil }
func (m *mockPresetStorage) CreateLoginToken(token *storage.LoginToken) error       { return nil }
func (m *mockPresetStorage) GetLoginToken(userID int64) (*storage.LoginToken, error) {
	return nil, storage.ErrNotFound
}
func (m *mockPresetStorage) ValidateLoginToken(userID int64, token string) (bool, error) {
	return false, nil
}
func (m *mockPresetStorage) DeleteLoginToken(userID int64) error { return nil }
func (m *mockPresetStorage) CleanupExpiredTokens() error         { return nil }
func (m *mockPresetStorage) DeleteAllChatHistory() error         { return nil }
func (m *mockPresetStorage) CleanupExpired() error               { return nil }
func (m *mockPresetStorage) Close() error                        { return nil }

func TestPresetManager_SaveAndLoad(t *testing.T) {
	mockStorage := newMockPresetStorage()
	manager := NewPresetManager(mockStorage)

	userID := int64(123)
	presetData := PresetData{
		Name:        "Test Preset",
		Temperature: 0.7,
		TopP:        0.9,
		MaxTokens:   2048,
	}

	dataJSON, err := json.Marshal(presetData)
	assert.NoError(t, err)

	preset := &storage.Preset{
		UserID:  &userID,
		Name:    "Test Preset",
		APIType: "openai",
		Data:    string(dataJSON),
	}

	// Save preset
	err = manager.SavePreset(preset)
	assert.NoError(t, err)
	assert.NotZero(t, preset.ID)

	// Load preset
	loaded, err := manager.LoadPreset(&userID, preset.ID)
	assert.NoError(t, err)
	assert.Equal(t, preset.Name, loaded.Name)
	assert.Equal(t, preset.APIType, loaded.APIType)
}

func TestPresetManager_ActivatePreset(t *testing.T) {
	mockStorage := newMockPresetStorage()
	manager := NewPresetManager(mockStorage)

	userID := int64(123)
	presetData := PresetData{
		Name:        "Test Preset",
		Temperature: 0.7,
	}

	dataJSON, _ := json.Marshal(presetData)

	preset := &storage.Preset{
		UserID:  &userID,
		Name:    "Test Preset",
		APIType: "openai",
		Data:    string(dataJSON),
	}

	manager.SavePreset(preset)

	// Activate preset
	err := manager.ActivatePreset(&userID, preset.ID)
	assert.NoError(t, err)

	// Get active preset
	active, err := manager.GetActivePreset(&userID, "openai")
	assert.NoError(t, err)
	assert.Equal(t, preset.ID, active.ID)
}

func TestPresetManager_ListPresets(t *testing.T) {
	mockStorage := newMockPresetStorage()
	manager := NewPresetManager(mockStorage)

	userID := int64(123)

	// Create multiple presets
	for i := 0; i < 3; i++ {
		presetData := PresetData{
			Name:        "Preset " + string(rune(i)),
			Temperature: 0.7,
		}
		dataJSON, _ := json.Marshal(presetData)

		preset := &storage.Preset{
			UserID:  &userID,
			Name:    presetData.Name,
			APIType: "openai",
			Data:    string(dataJSON),
		}
		manager.SavePreset(preset)
	}

	// List presets
	presets, err := manager.ListPresets(&userID, "openai")
	assert.NoError(t, err)
	assert.Len(t, presets, 3)
}
