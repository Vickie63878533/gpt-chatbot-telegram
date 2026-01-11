package sillytavern

import (
	"encoding/json"
	"testing"

	"github.com/tbxark/ChatGPT-Telegram-Workers/go_version/internal/storage"
)

// MockStorage is a simple mock for testing
type MockStorage struct {
	cards       map[uint]*storage.CharacterCard
	nextID      uint
	activeCards map[int64]uint // userID -> cardID
}

func NewMockStorage() *MockStorage {
	return &MockStorage{
		cards:       make(map[uint]*storage.CharacterCard),
		nextID:      1,
		activeCards: make(map[int64]uint),
	}
}

func (m *MockStorage) CreateCharacterCard(card *storage.CharacterCard) error {
	card.ID = m.nextID
	m.nextID++
	m.cards[card.ID] = card
	return nil
}

func (m *MockStorage) GetCharacterCard(id uint) (*storage.CharacterCard, error) {
	card, ok := m.cards[id]
	if !ok {
		return nil, storage.ErrNotFound
	}
	return card, nil
}

func (m *MockStorage) ListCharacterCards(userID *int64) ([]*storage.CharacterCard, error) {
	var result []*storage.CharacterCard
	for _, card := range m.cards {
		// Include global cards and user's own cards
		if card.UserID == nil || (userID != nil && card.UserID != nil && *card.UserID == *userID) {
			result = append(result, card)
		}
	}
	return result, nil
}

func (m *MockStorage) UpdateCharacterCard(card *storage.CharacterCard) error {
	if _, ok := m.cards[card.ID]; !ok {
		return storage.ErrNotFound
	}
	m.cards[card.ID] = card
	return nil
}

func (m *MockStorage) DeleteCharacterCard(id uint) error {
	delete(m.cards, id)
	return nil
}

func (m *MockStorage) ActivateCharacterCard(userID *int64, cardID uint) error {
	// Deactivate all cards for this user
	for _, card := range m.cards {
		if card.UserID == nil || (userID != nil && card.UserID != nil && *card.UserID == *userID) {
			card.IsActive = false
		}
	}
	
	// Activate the specified card
	card, ok := m.cards[cardID]
	if !ok {
		return storage.ErrNotFound
	}
	card.IsActive = true
	
	if userID != nil {
		m.activeCards[*userID] = cardID
	} else {
		m.activeCards[0] = cardID // Use 0 for global
	}
	
	return nil
}

func (m *MockStorage) GetActiveCharacterCard(userID *int64) (*storage.CharacterCard, error) {
	var activeID uint
	var ok bool
	
	if userID != nil {
		activeID, ok = m.activeCards[*userID]
	} else {
		activeID, ok = m.activeCards[0]
	}
	
	if !ok {
		return nil, storage.ErrNotFound
	}
	
	return m.GetCharacterCard(activeID)
}

// Implement other required Storage interface methods as no-ops
func (m *MockStorage) GetChatHistory(ctx *storage.SessionContext) ([]storage.HistoryItem, error) {
	return nil, nil
}
func (m *MockStorage) SaveChatHistory(ctx *storage.SessionContext, history []storage.HistoryItem) error {
	return nil
}
func (m *MockStorage) DeleteChatHistory(ctx *storage.SessionContext) error { return nil }
func (m *MockStorage) GetUserConfig(ctx *storage.SessionContext) (*storage.UserConfig, error) {
	return nil, nil
}
func (m *MockStorage) SaveUserConfig(ctx *storage.SessionContext, config *storage.UserConfig) error {
	return nil
}
func (m *MockStorage) GetMessageIDs(ctx *storage.SessionContext) ([]int, error) { return nil, nil }
func (m *MockStorage) SaveMessageIDs(ctx *storage.SessionContext, ids []int) error {
	return nil
}
func (m *MockStorage) GetGroupAdmins(chatID int64) ([]storage.ChatMember, error) { return nil, nil }
func (m *MockStorage) SaveGroupAdmins(chatID int64, admins []storage.ChatMember, ttl int) error {
	return nil
}
func (m *MockStorage) CreateWorldBook(book *storage.WorldBook) error                  { return nil }
func (m *MockStorage) GetWorldBook(id uint) (*storage.WorldBook, error)               { return nil, nil }
func (m *MockStorage) ListWorldBooks(userID *int64) ([]*storage.WorldBook, error)     { return nil, nil }
func (m *MockStorage) UpdateWorldBook(book *storage.WorldBook) error                  { return nil }
func (m *MockStorage) DeleteWorldBook(id uint) error                                  { return nil }
func (m *MockStorage) ActivateWorldBook(userID *int64, bookID uint) error             { return nil }
func (m *MockStorage) GetActiveWorldBook(userID *int64) (*storage.WorldBook, error)   { return nil, nil }
func (m *MockStorage) CreateWorldBookEntry(entry *storage.WorldBookEntry) error       { return nil }
func (m *MockStorage) GetWorldBookEntry(id uint) (*storage.WorldBookEntry, error)     { return nil, nil }
func (m *MockStorage) ListWorldBookEntries(bookID uint) ([]*storage.WorldBookEntry, error) {
	return nil, nil
}
func (m *MockStorage) UpdateWorldBookEntry(entry *storage.WorldBookEntry) error { return nil }
func (m *MockStorage) DeleteWorldBookEntry(id uint) error                       { return nil }
func (m *MockStorage) UpdateWorldBookEntryStatus(id uint, enabled bool) error   { return nil }
func (m *MockStorage) CreatePreset(preset *storage.Preset) error                { return nil }
func (m *MockStorage) GetPreset(id uint) (*storage.Preset, error)               { return nil, nil }
func (m *MockStorage) ListPresets(userID *int64, apiType string) ([]*storage.Preset, error) {
	return nil, nil
}
func (m *MockStorage) UpdatePreset(preset *storage.Preset) error                      { return nil }
func (m *MockStorage) DeletePreset(id uint) error                                     { return nil }
func (m *MockStorage) ActivatePreset(userID *int64, presetID uint) error              { return nil }
func (m *MockStorage) GetActivePreset(userID *int64, apiType string) (*storage.Preset, error) {
	return nil, nil
}
func (m *MockStorage) CreateRegexPattern(pattern *storage.RegexPattern) error { return nil }
func (m *MockStorage) GetRegexPattern(id uint) (*storage.RegexPattern, error) { return nil, nil }
func (m *MockStorage) ListRegexPatterns(userID *int64, patternType string) ([]*storage.RegexPattern, error) {
	return nil, nil
}
func (m *MockStorage) UpdateRegexPattern(pattern *storage.RegexPattern) error     { return nil }
func (m *MockStorage) DeleteRegexPattern(id uint) error                           { return nil }
func (m *MockStorage) UpdateRegexPatternStatus(id uint, enabled bool) error       { return nil }
func (m *MockStorage) CreateLoginToken(token *storage.LoginToken) error           { return nil }
func (m *MockStorage) GetLoginToken(userID int64) (*storage.LoginToken, error)    { return nil, nil }
func (m *MockStorage) ValidateLoginToken(userID int64, token string) (bool, error) { return false, nil }
func (m *MockStorage) DeleteLoginToken(userID int64) error                        { return nil }
func (m *MockStorage) CleanupExpiredTokens() error                                { return nil }
func (m *MockStorage) DeleteAllChatHistory() error                                { return nil }
func (m *MockStorage) CleanupExpired() error                                      { return nil }
func (m *MockStorage) Close() error                                               { return nil }

func TestCharacterCardManager_SaveAndLoad(t *testing.T) {
	mockStorage := NewMockStorage()
	manager := NewCharacterCardManager(mockStorage)

	// Create a valid V2 character card
	cardData := CharacterCardV2{
		Spec:        "chara_card_v2",
		SpecVersion: "2.0",
		Data: CharacterCardV2Data{
			Name:        "Test Character",
			Description: "A test character",
			Personality: "Friendly",
			Scenario:    "Test scenario",
			FirstMes:    "Hello!",
		},
	}

	cardJSON, err := json.Marshal(cardData)
	if err != nil {
		t.Fatalf("Failed to marshal card data: %v", err)
	}

	userID := int64(123)
	card := &storage.CharacterCard{
		UserID: &userID,
		Data:   string(cardJSON),
	}

	// Save the card
	err = manager.SaveCard(card)
	if err != nil {
		t.Fatalf("Failed to save card: %v", err)
	}

	// Verify the card was saved with correct name
	if card.Name != "Test Character" {
		t.Errorf("Expected name 'Test Character', got '%s'", card.Name)
	}

	// Load the card
	loadedCard, err := manager.LoadCard(&userID, card.ID)
	if err != nil {
		t.Fatalf("Failed to load card: %v", err)
	}

	if loadedCard.Name != card.Name {
		t.Errorf("Expected name '%s', got '%s'", card.Name, loadedCard.Name)
	}
}

func TestCharacterCardManager_ActivateCard(t *testing.T) {
	mockStorage := NewMockStorage()
	manager := NewCharacterCardManager(mockStorage)

	// Create and save a card
	cardData := CharacterCardV2{
		Spec:        "chara_card_v2",
		SpecVersion: "2.0",
		Data: CharacterCardV2Data{
			Name: "Test Character",
		},
	}

	cardJSON, _ := json.Marshal(cardData)
	userID := int64(123)
	card := &storage.CharacterCard{
		UserID: &userID,
		Data:   string(cardJSON),
	}

	manager.SaveCard(card)

	// Activate the card
	err := manager.ActivateCard(&userID, card.ID)
	if err != nil {
		t.Fatalf("Failed to activate card: %v", err)
	}

	// Get active card
	activeCard, err := manager.GetActiveCard(&userID)
	if err != nil {
		t.Fatalf("Failed to get active card: %v", err)
	}

	if activeCard.ID != card.ID {
		t.Errorf("Expected active card ID %d, got %d", card.ID, activeCard.ID)
	}
}

func TestCharacterCardManager_ListCards(t *testing.T) {
	mockStorage := NewMockStorage()
	manager := NewCharacterCardManager(mockStorage)

	userID := int64(123)

	// Create multiple cards
	for i := 0; i < 3; i++ {
		cardData := CharacterCardV2{
			Spec:        "chara_card_v2",
			SpecVersion: "2.0",
			Data: CharacterCardV2Data{
				Name: "Test Character",
			},
		}

		cardJSON, _ := json.Marshal(cardData)
		card := &storage.CharacterCard{
			UserID: &userID,
			Data:   string(cardJSON),
		}

		manager.SaveCard(card)
	}

	// List cards
	cards, err := manager.ListCards(&userID)
	if err != nil {
		t.Fatalf("Failed to list cards: %v", err)
	}

	if len(cards) != 3 {
		t.Errorf("Expected 3 cards, got %d", len(cards))
	}
}

func TestPNGParser_ValidateCharacterCardV2(t *testing.T) {
	parser := NewPNGParser()

	// Valid card
	validCard := CharacterCardV2{
		Spec:        "chara_card_v2",
		SpecVersion: "2.0",
		Data: CharacterCardV2Data{
			Name: "Test Character",
		},
	}

	validJSON, _ := json.Marshal(validCard)
	err := parser.ValidateCharacterCardV2(string(validJSON))
	if err != nil {
		t.Errorf("Valid card should not produce error: %v", err)
	}

	// Invalid spec
	invalidCard := CharacterCardV2{
		Spec:        "invalid_spec",
		SpecVersion: "2.0",
		Data: CharacterCardV2Data{
			Name: "Test Character",
		},
	}

	invalidJSON, _ := json.Marshal(invalidCard)
	err = parser.ValidateCharacterCardV2(string(invalidJSON))
	if err == nil {
		t.Error("Invalid spec should produce error")
	}

	// Missing name
	noNameCard := CharacterCardV2{
		Spec:        "chara_card_v2",
		SpecVersion: "2.0",
		Data:        CharacterCardV2Data{},
	}

	noNameJSON, _ := json.Marshal(noNameCard)
	err = parser.ValidateCharacterCardV2(string(noNameJSON))
	if err == nil {
		t.Error("Missing name should produce error")
	}
}
