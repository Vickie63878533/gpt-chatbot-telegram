package storage

import (
	"os"
	"testing"
	"time"
)

// TestGORMStorage_CharacterCard tests character card CRUD operations
func TestGORMStorage_CharacterCard(t *testing.T) {
	// Create temporary database
	dbPath := "./test_character_card.db"
	defer os.Remove(dbPath)

	storage, err := NewStorage("", dbPath)
	if err != nil {
		t.Fatalf("Failed to create storage: %v", err)
	}
	defer storage.Close()

	userID := int64(123)

	// Test Create
	card := &CharacterCard{
		UserID: &userID,
		Name:   "Test Character",
		Avatar: "https://example.com/avatar.png",
		Data:   `{"name":"Test","description":"A test character"}`,
	}

	err = storage.CreateCharacterCard(card)
	if err != nil {
		t.Fatalf("Failed to create character card: %v", err)
	}

	if card.ID == 0 {
		t.Fatal("Character card ID should be set after creation")
	}

	// Test Get
	retrieved, err := storage.GetCharacterCard(card.ID)
	if err != nil {
		t.Fatalf("Failed to get character card: %v", err)
	}

	if retrieved.Name != card.Name {
		t.Errorf("Expected name %s, got %s", card.Name, retrieved.Name)
	}

	// Test List
	cards, err := storage.ListCharacterCards(&userID)
	if err != nil {
		t.Fatalf("Failed to list character cards: %v", err)
	}

	if len(cards) != 1 {
		t.Errorf("Expected 1 card, got %d", len(cards))
	}

	// Test Update
	card.Name = "Updated Character"
	err = storage.UpdateCharacterCard(card)
	if err != nil {
		t.Fatalf("Failed to update character card: %v", err)
	}

	updated, err := storage.GetCharacterCard(card.ID)
	if err != nil {
		t.Fatalf("Failed to get updated character card: %v", err)
	}

	if updated.Name != "Updated Character" {
		t.Errorf("Expected name 'Updated Character', got %s", updated.Name)
	}

	// Test Activate
	err = storage.ActivateCharacterCard(&userID, card.ID)
	if err != nil {
		t.Fatalf("Failed to activate character card: %v", err)
	}

	active, err := storage.GetActiveCharacterCard(&userID)
	if err != nil {
		t.Fatalf("Failed to get active character card: %v", err)
	}

	if active == nil || active.ID != card.ID {
		t.Error("Expected card to be active")
	}

	// Test Delete
	err = storage.DeleteCharacterCard(card.ID)
	if err != nil {
		t.Fatalf("Failed to delete character card: %v", err)
	}

	_, err = storage.GetCharacterCard(card.ID)
	if err == nil {
		t.Error("Expected error when getting deleted card")
	}
}

// TestGORMStorage_WorldBook tests world book CRUD operations
func TestGORMStorage_WorldBook(t *testing.T) {
	dbPath := "./test_world_book.db"
	defer os.Remove(dbPath)

	storage, err := NewStorage("", dbPath)
	if err != nil {
		t.Fatalf("Failed to create storage: %v", err)
	}
	defer storage.Close()

	userID := int64(456)

	// Test Create
	book := &WorldBook{
		UserID: &userID,
		Name:   "Test World",
		Data:   `{"name":"Test World","entries":{}}`,
	}

	err = storage.CreateWorldBook(book)
	if err != nil {
		t.Fatalf("Failed to create world book: %v", err)
	}

	// Test Get
	retrieved, err := storage.GetWorldBook(book.ID)
	if err != nil {
		t.Fatalf("Failed to get world book: %v", err)
	}

	if retrieved.Name != book.Name {
		t.Errorf("Expected name %s, got %s", book.Name, retrieved.Name)
	}

	// Test Activate
	err = storage.ActivateWorldBook(&userID, book.ID)
	if err != nil {
		t.Fatalf("Failed to activate world book: %v", err)
	}

	active, err := storage.GetActiveWorldBook(&userID)
	if err != nil {
		t.Fatalf("Failed to get active world book: %v", err)
	}

	if active == nil || active.ID != book.ID {
		t.Error("Expected book to be active")
	}
}

// TestGORMStorage_Preset tests preset CRUD operations
func TestGORMStorage_Preset(t *testing.T) {
	dbPath := "./test_preset.db"
	defer os.Remove(dbPath)

	storage, err := NewStorage("", dbPath)
	if err != nil {
		t.Fatalf("Failed to create storage: %v", err)
	}
	defer storage.Close()

	userID := int64(789)

	// Test Create
	preset := &Preset{
		UserID:  &userID,
		Name:    "Test Preset",
		APIType: "openai",
		Data:    `{"temperature":0.7,"max_tokens":2048}`,
	}

	err = storage.CreatePreset(preset)
	if err != nil {
		t.Fatalf("Failed to create preset: %v", err)
	}

	// Test List with API type filter
	presets, err := storage.ListPresets(&userID, "openai")
	if err != nil {
		t.Fatalf("Failed to list presets: %v", err)
	}

	if len(presets) != 1 {
		t.Errorf("Expected 1 preset, got %d", len(presets))
	}

	// Test Activate
	err = storage.ActivatePreset(&userID, preset.ID)
	if err != nil {
		t.Fatalf("Failed to activate preset: %v", err)
	}

	active, err := storage.GetActivePreset(&userID, "openai")
	if err != nil {
		t.Fatalf("Failed to get active preset: %v", err)
	}

	if active == nil || active.ID != preset.ID {
		t.Error("Expected preset to be active")
	}
}

// TestGORMStorage_RegexPattern tests regex pattern CRUD operations
func TestGORMStorage_RegexPattern(t *testing.T) {
	dbPath := "./test_regex.db"
	defer os.Remove(dbPath)

	storage, err := NewStorage("", dbPath)
	if err != nil {
		t.Fatalf("Failed to create storage: %v", err)
	}
	defer storage.Close()

	userID := int64(111)

	// Test Create
	pattern := &RegexPattern{
		UserID:  &userID,
		Name:    "Test Pattern",
		Pattern: `\btest\b`,
		Replace: "TEST",
		Type:    "input",
		Order:   100,
		Enabled: true,
	}

	err = storage.CreateRegexPattern(pattern)
	if err != nil {
		t.Fatalf("Failed to create regex pattern: %v", err)
	}

	// Test List with type filter
	patterns, err := storage.ListRegexPatterns(&userID, "input")
	if err != nil {
		t.Fatalf("Failed to list regex patterns: %v", err)
	}

	if len(patterns) != 1 {
		t.Errorf("Expected 1 pattern, got %d", len(patterns))
	}

	// Test Update
	pattern.Enabled = false
	err = storage.UpdateRegexPattern(pattern)
	if err != nil {
		t.Fatalf("Failed to update regex pattern: %v", err)
	}

	updated, err := storage.GetRegexPattern(pattern.ID)
	if err != nil {
		t.Fatalf("Failed to get updated regex pattern: %v", err)
	}

	if updated.Enabled {
		t.Error("Expected pattern to be disabled")
	}
}

// TestGORMStorage_LoginToken tests login token operations
func TestGORMStorage_LoginToken(t *testing.T) {
	dbPath := "./test_login_token.db"
	defer os.Remove(dbPath)

	storage, err := NewStorage("", dbPath)
	if err != nil {
		t.Fatalf("Failed to create storage: %v", err)
	}
	defer storage.Close()

	userID := int64(999)

	// Test Create
	token := &LoginToken{
		UserID:    userID,
		Token:     "test-token-123",
		ExpiresAt: time.Now().Add(24 * time.Hour),
	}

	err = storage.CreateLoginToken(token)
	if err != nil {
		t.Fatalf("Failed to create login token: %v", err)
	}

	// Test Validate - valid token
	valid, err := storage.ValidateLoginToken(userID, "test-token-123")
	if err != nil {
		t.Fatalf("Failed to validate login token: %v", err)
	}

	if !valid {
		t.Error("Expected token to be valid")
	}

	// Test Validate - invalid token
	valid, err = storage.ValidateLoginToken(userID, "wrong-token")
	if err != nil {
		t.Fatalf("Failed to validate login token: %v", err)
	}

	if valid {
		t.Error("Expected token to be invalid")
	}

	// Test Delete
	err = storage.DeleteLoginToken(userID)
	if err != nil {
		t.Fatalf("Failed to delete login token: %v", err)
	}

	valid, err = storage.ValidateLoginToken(userID, "test-token-123")
	if err != nil {
		t.Fatalf("Failed to validate login token: %v", err)
	}

	if valid {
		t.Error("Expected token to be invalid after deletion")
	}
}

// TestGORMStorage_CleanupExpiredTokens tests expired token cleanup
func TestGORMStorage_CleanupExpiredTokens(t *testing.T) {
	dbPath := "./test_cleanup_tokens.db"
	defer os.Remove(dbPath)

	storage, err := NewStorage("", dbPath)
	if err != nil {
		t.Fatalf("Failed to create storage: %v", err)
	}
	defer storage.Close()

	// Create expired token
	expiredToken := &LoginToken{
		UserID:    int64(111),
		Token:     "expired-token",
		ExpiresAt: time.Now().Add(-1 * time.Hour),
	}

	err = storage.CreateLoginToken(expiredToken)
	if err != nil {
		t.Fatalf("Failed to create expired token: %v", err)
	}

	// Create valid token
	validToken := &LoginToken{
		UserID:    int64(222),
		Token:     "valid-token",
		ExpiresAt: time.Now().Add(24 * time.Hour),
	}

	err = storage.CreateLoginToken(validToken)
	if err != nil {
		t.Fatalf("Failed to create valid token: %v", err)
	}

	// Cleanup expired tokens
	err = storage.CleanupExpiredTokens()
	if err != nil {
		t.Fatalf("Failed to cleanup expired tokens: %v", err)
	}

	// Verify expired token is gone
	valid, err := storage.ValidateLoginToken(111, "expired-token")
	if err != nil {
		t.Fatalf("Failed to validate expired token: %v", err)
	}

	if valid {
		t.Error("Expected expired token to be invalid")
	}

	// Verify valid token still exists
	valid, err = storage.ValidateLoginToken(222, "valid-token")
	if err != nil {
		t.Fatalf("Failed to validate valid token: %v", err)
	}

	if !valid {
		t.Error("Expected valid token to still be valid")
	}
}

// TestGORMStorage_WorldBookEntry tests world book entry operations
func TestGORMStorage_WorldBookEntry(t *testing.T) {
	dbPath := "./test_wb_entry.db"
	defer os.Remove(dbPath)

	storage, err := NewStorage("", dbPath)
	if err != nil {
		t.Fatalf("Failed to create storage: %v", err)
	}
	defer storage.Close()

	userID := int64(333)

	// Create a world book first
	book := &WorldBook{
		UserID: &userID,
		Name:   "Test World",
		Data:   `{"name":"Test World"}`,
	}

	err = storage.CreateWorldBook(book)
	if err != nil {
		t.Fatalf("Failed to create world book: %v", err)
	}

	// Test Create Entry
	entry := &WorldBookEntry{
		WorldBookID:   book.ID,
		UID:           "entry-001",
		Keys:          `["keyword1","keyword2"]`,
		SecondaryKeys: `["secondary1"]`,
		Content:       "This is test content",
		Comment:       "Test comment",
		Constant:      false,
		Selective:     true,
		Order:         100,
		Position:      "after_char",
		Enabled:       true,
		Extensions:    `{}`,
	}

	err = storage.CreateWorldBookEntry(entry)
	if err != nil {
		t.Fatalf("Failed to create world book entry: %v", err)
	}

	// Test Get Entry
	retrieved, err := storage.GetWorldBookEntry(entry.ID)
	if err != nil {
		t.Fatalf("Failed to get world book entry: %v", err)
	}

	if retrieved.UID != entry.UID {
		t.Errorf("Expected UID %s, got %s", entry.UID, retrieved.UID)
	}

	// Test List Entries
	entries, err := storage.ListWorldBookEntries(book.ID)
	if err != nil {
		t.Fatalf("Failed to list world book entries: %v", err)
	}

	if len(entries) != 1 {
		t.Errorf("Expected 1 entry, got %d", len(entries))
	}

	// Test Update Entry
	entry.Content = "Updated content"
	err = storage.UpdateWorldBookEntry(entry)
	if err != nil {
		t.Fatalf("Failed to update world book entry: %v", err)
	}

	updated, err := storage.GetWorldBookEntry(entry.ID)
	if err != nil {
		t.Fatalf("Failed to get updated entry: %v", err)
	}

	if updated.Content != "Updated content" {
		t.Errorf("Expected content 'Updated content', got %s", updated.Content)
	}

	// Test Delete Entry
	err = storage.DeleteWorldBookEntry(entry.ID)
	if err != nil {
		t.Fatalf("Failed to delete world book entry: %v", err)
	}

	_, err = storage.GetWorldBookEntry(entry.ID)
	if err == nil {
		t.Error("Expected error when getting deleted entry")
	}
}
