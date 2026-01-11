package manager

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/tbxark/ChatGPT-Telegram-Workers/go_version/internal/config"
	"github.com/tbxark/ChatGPT-Telegram-Workers/go_version/internal/storage"
)

// Extended MockStorage for character card testing
type MockStorageWithCharacters struct {
	*MockStorage
	cards       map[uint]*storage.CharacterCard
	nextID      uint
	activeCards map[int64]uint // userID -> cardID
}

func NewMockStorageWithCharacters() *MockStorageWithCharacters {
	return &MockStorageWithCharacters{
		MockStorage: NewMockStorage(),
		cards:       make(map[uint]*storage.CharacterCard),
		nextID:      1,
		activeCards: make(map[int64]uint),
	}
}

func (m *MockStorageWithCharacters) CreateCharacterCard(card *storage.CharacterCard) error {
	card.ID = m.nextID
	m.nextID++
	m.cards[card.ID] = card
	return nil
}

func (m *MockStorageWithCharacters) GetCharacterCard(id uint) (*storage.CharacterCard, error) {
	card, exists := m.cards[id]
	if !exists {
		return nil, storage.ErrNotFound
	}
	return card, nil
}

func (m *MockStorageWithCharacters) ListCharacterCards(userID *int64) ([]*storage.CharacterCard, error) {
	var result []*storage.CharacterCard
	for _, card := range m.cards {
		if userID == nil && card.UserID == nil {
			result = append(result, card)
		} else if userID != nil && card.UserID != nil && *card.UserID == *userID {
			result = append(result, card)
		}
	}
	return result, nil
}

func (m *MockStorageWithCharacters) UpdateCharacterCard(card *storage.CharacterCard) error {
	if _, exists := m.cards[card.ID]; !exists {
		return storage.ErrNotFound
	}
	m.cards[card.ID] = card
	return nil
}

func (m *MockStorageWithCharacters) DeleteCharacterCard(id uint) error {
	if _, exists := m.cards[id]; !exists {
		return storage.ErrNotFound
	}
	delete(m.cards, id)
	return nil
}

func (m *MockStorageWithCharacters) ActivateCharacterCard(userID *int64, cardID uint) error {
	card, exists := m.cards[cardID]
	if !exists {
		return storage.ErrNotFound
	}

	// Deactivate all other cards for this user
	for _, c := range m.cards {
		if (userID == nil && c.UserID == nil) || (userID != nil && c.UserID != nil && *c.UserID == *userID) {
			c.IsActive = false
		}
	}

	// Activate this card
	card.IsActive = true
	if userID != nil {
		m.activeCards[*userID] = cardID
	}
	return nil
}

func (m *MockStorageWithCharacters) GetActiveCharacterCard(userID *int64) (*storage.CharacterCard, error) {
	if userID == nil {
		return nil, storage.ErrNotFound
	}

	cardID, exists := m.activeCards[*userID]
	if !exists {
		return nil, storage.ErrNotFound
	}

	return m.GetCharacterCard(cardID)
}

// Helper to create a valid character card JSON
func createValidCharacterCardJSON() string {
	card := map[string]interface{}{
		"spec":         "chara_card_v2",
		"spec_version": "2.0",
		"data": map[string]interface{}{
			"name":        "Test Character",
			"description": "A test character",
			"personality": "Friendly and helpful",
			"scenario":    "A helpful assistant",
			"first_mes":   "Hello!",
			"mes_example": "How can I help?",
			"tags":        []string{"test"},
		},
	}
	data, _ := json.Marshal(card)
	return string(data)
}

// TestHandleListCharacters tests listing character cards
func TestHandleListCharacters(t *testing.T) {
	mockStorage := NewMockStorageWithCharacters()
	cfg := &config.Config{
		Port:              8080,

		EnableUserSetting: true,
	}

	server := New(cfg, mockStorage)
	userID := int64(12345)

	// Create a test character card
	card := &storage.CharacterCard{
		Name:   "Test Character",
		Avatar: "avatar.png",
		Data:   createValidCharacterCardJSON(),
	}
	mockStorage.CreateCharacterCard(card)

	// Create request with auth context
	req := httptest.NewRequest("GET", "/api/manager/characters", nil)
	w := httptest.NewRecorder()

	server.handleListCharacters(w, req, userID)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	var resp CharacterCardsListResponse
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Errorf("Failed to decode response: %v", err)
	}

	if len(resp.Characters) == 0 {
		t.Error("Expected at least one character card")
	}
}

// TestHandleGetCharacter tests getting a specific character card
func TestHandleGetCharacter(t *testing.T) {
	mockStorage := NewMockStorageWithCharacters()
	cfg := &config.Config{
		Port:              8080,

		EnableUserSetting: true,
	}

	server := New(cfg, mockStorage)
	userID := int64(12345)

	// Create a test character card
	card := &storage.CharacterCard{
		Name:   "Test Character",
		Avatar: "avatar.png",
		Data:   createValidCharacterCardJSON(),
		UserID: &userID,
	}
	mockStorage.CreateCharacterCard(card)

	// Create request
	req := httptest.NewRequest("GET", fmt.Sprintf("/api/manager/characters/%d", card.ID), nil)
	w := httptest.NewRecorder()

	server.handleGetCharacter(w, req, userID)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	var resp CharacterCardDetailResponse
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Errorf("Failed to decode response: %v", err)
	}

	if resp.ID != card.ID {
		t.Errorf("Expected card ID %d, got %d", card.ID, resp.ID)
	}

	if resp.Name != "Test Character" {
		t.Errorf("Expected name 'Test Character', got '%s'", resp.Name)
	}
}

// TestHandleGetCharacter_NotFound tests getting a non-existent character card
func TestHandleGetCharacter_NotFound(t *testing.T) {
	mockStorage := NewMockStorageWithCharacters()
	cfg := &config.Config{
		Port:              8080,

		EnableUserSetting: true,
	}

	server := New(cfg, mockStorage)
	userID := int64(12345)

	req := httptest.NewRequest("GET", "/api/manager/characters/999", nil)
	w := httptest.NewRecorder()

	server.handleGetCharacter(w, req, userID)

	if w.Code != http.StatusNotFound {
		t.Errorf("Expected status 404, got %d", w.Code)
	}
}

// TestHandleUploadCharacter tests uploading a character card
func TestHandleUploadCharacter(t *testing.T) {
	mockStorage := NewMockStorageWithCharacters()
	cfg := &config.Config{
		Port:              8080,

		EnableUserSetting: true,
	}

	server := New(cfg, mockStorage)
	userID := int64(12345)

	// Create multipart form with character card file
	body := new(bytes.Buffer)
	writer := multipart.NewWriter(body)

	// Add file
	part, err := writer.CreateFormFile("file", "character.json")
	if err != nil {
		t.Fatalf("Failed to create form file: %v", err)
	}

	cardJSON := createValidCharacterCardJSON()
	if _, err := io.WriteString(part, cardJSON); err != nil {
		t.Fatalf("Failed to write file: %v", err)
	}

	writer.Close()

	// Create request
	req := httptest.NewRequest("POST", "/api/manager/characters", body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	w := httptest.NewRecorder()

	server.handleUploadCharacter(w, req, userID)

	if w.Code != http.StatusCreated {
		t.Errorf("Expected status 201, got %d. Body: %s", w.Code, w.Body.String())
	}

	var resp CharacterCardResponse
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Errorf("Failed to decode response: %v", err)
	}

	if resp.Name != "Test Character" {
		t.Errorf("Expected name 'Test Character', got '%s'", resp.Name)
	}

	if resp.ID == 0 {
		t.Error("Expected non-zero card ID")
	}
}

// TestHandleUploadCharacter_InvalidFormat tests uploading invalid character card
func TestHandleUploadCharacter_InvalidFormat(t *testing.T) {
	mockStorage := NewMockStorageWithCharacters()
	cfg := &config.Config{
		Port:              8080,

		EnableUserSetting: true,
	}

	server := New(cfg, mockStorage)
	userID := int64(12345)

	// Create multipart form with invalid JSON
	body := new(bytes.Buffer)
	writer := multipart.NewWriter(body)

	part, err := writer.CreateFormFile("file", "character.json")
	if err != nil {
		t.Fatalf("Failed to create form file: %v", err)
	}

	if _, err := io.WriteString(part, "invalid json"); err != nil {
		t.Fatalf("Failed to write file: %v", err)
	}

	writer.Close()

	req := httptest.NewRequest("POST", "/api/manager/characters", body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	w := httptest.NewRecorder()

	server.handleUploadCharacter(w, req, userID)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d", w.Code)
	}
}

// TestHandleUpdateCharacter tests updating a character card
func TestHandleUpdateCharacter(t *testing.T) {
	mockStorage := NewMockStorageWithCharacters()
	cfg := &config.Config{
		Port:              8080,

		EnableUserSetting: true,
	}

	server := New(cfg, mockStorage)
	userID := int64(12345)

	// Create a test character card
	card := &storage.CharacterCard{
		Name:   "Original Name",
		Avatar: "avatar.png",
		Data:   createValidCharacterCardJSON(),
		UserID: &userID,
	}
	mockStorage.CreateCharacterCard(card)

	// Create update request
	updateReq := struct {
		Name string `json:"name"`
	}{
		Name: "Updated Name",
	}

	body, _ := json.Marshal(updateReq)
	req := httptest.NewRequest("PUT", fmt.Sprintf("/api/manager/characters/%d", card.ID), bytes.NewReader(body))
	w := httptest.NewRecorder()

	server.handleUpdateCharacter(w, req, userID)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	var resp CharacterCardResponse
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Errorf("Failed to decode response: %v", err)
	}

	if resp.Name != "Updated Name" {
		t.Errorf("Expected name 'Updated Name', got '%s'", resp.Name)
	}
}

// TestHandleDeleteCharacter tests deleting a character card
func TestHandleDeleteCharacter(t *testing.T) {
	mockStorage := NewMockStorageWithCharacters()
	cfg := &config.Config{
		Port:              8080,

		EnableUserSetting: true,
	}

	server := New(cfg, mockStorage)
	userID := int64(12345)

	// Create a test character card
	card := &storage.CharacterCard{
		Name:   "Test Character",
		Avatar: "avatar.png",
		Data:   createValidCharacterCardJSON(),
		UserID: &userID,
	}
	mockStorage.CreateCharacterCard(card)

	req := httptest.NewRequest("DELETE", fmt.Sprintf("/api/manager/characters/%d", card.ID), nil)
	w := httptest.NewRecorder()

	server.handleDeleteCharacter(w, req, userID)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	// Verify card is deleted
	_, err := mockStorage.GetCharacterCard(card.ID)
	if err != storage.ErrNotFound {
		t.Error("Expected card to be deleted")
	}
}

// TestHandleActivateCharacter tests activating a character card
func TestHandleActivateCharacter(t *testing.T) {
	mockStorage := NewMockStorageWithCharacters()
	cfg := &config.Config{
		Port:              8080,

		EnableUserSetting: true,
	}

	server := New(cfg, mockStorage)
	userID := int64(12345)

	// Create a test character card
	card := &storage.CharacterCard{
		Name:   "Test Character",
		Avatar: "avatar.png",
		Data:   createValidCharacterCardJSON(),
		UserID: &userID,
	}
	mockStorage.CreateCharacterCard(card)

	req := httptest.NewRequest("POST", fmt.Sprintf("/api/manager/characters/%d/activate", card.ID), nil)
	w := httptest.NewRecorder()

	server.handleActivateCharacter(w, req, userID)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	// Verify card is activated
	activeCard, err := mockStorage.GetActiveCharacterCard(&userID)
	if err != nil {
		t.Errorf("Failed to get active card: %v", err)
	}

	if activeCard.ID != card.ID {
		t.Errorf("Expected active card ID %d, got %d", card.ID, activeCard.ID)
	}
}

// TestHandleCharactersRoute_PermissionDenied tests permission checks
func TestHandleCharactersRoute_PermissionDenied(t *testing.T) {
	mockStorage := NewMockStorageWithCharacters()
	cfg := &config.Config{
		Port:              8080,

		EnableUserSetting: false, // Disable user settings
	}

	server := New(cfg, mockStorage)
	userID := int64(12345)
	otherUserID := int64(67890)

	// Create a character card for another user
	card := &storage.CharacterCard{
		Name:   "Other User's Character",
		Avatar: "avatar.png",
		Data:   createValidCharacterCardJSON(),
		UserID: &otherUserID,
	}
	mockStorage.CreateCharacterCard(card)

	// Try to delete as different user
	req := httptest.NewRequest("DELETE", fmt.Sprintf("/api/manager/characters/%d", card.ID), nil)
	w := httptest.NewRecorder()

	server.handleDeleteCharacter(w, req, userID)

	if w.Code != http.StatusForbidden {
		t.Errorf("Expected status 403, got %d", w.Code)
	}
}

// TestHandleCharactersRoute_GlobalCard tests accessing global cards
func TestHandleCharactersRoute_GlobalCard(t *testing.T) {
	mockStorage := NewMockStorageWithCharacters()
	cfg := &config.Config{
		Port:              8080,

		EnableUserSetting: true,
	}

	server := New(cfg, mockStorage)
	userID := int64(12345)

	// Create a global character card (UserID = nil)
	card := &storage.CharacterCard{
		Name:   "Global Character",
		Avatar: "avatar.png",
		Data:   createValidCharacterCardJSON(),
		UserID: nil,
	}
	mockStorage.CreateCharacterCard(card)

	// Any user should be able to access global cards
	req := httptest.NewRequest("GET", fmt.Sprintf("/api/manager/characters/%d", card.ID), nil)
	w := httptest.NewRecorder()

	server.handleGetCharacter(w, req, userID)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}
}

// TestHandleCharactersRoute_InvalidID tests invalid character ID
func TestHandleCharactersRoute_InvalidID(t *testing.T) {
	mockStorage := NewMockStorageWithCharacters()
	cfg := &config.Config{
		Port:              8080,

		EnableUserSetting: true,
	}

	server := New(cfg, mockStorage)
	userID := int64(12345)

	req := httptest.NewRequest("GET", "/api/manager/characters/invalid", nil)
	w := httptest.NewRecorder()

	server.handleGetCharacter(w, req, userID)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d", w.Code)
	}
}

// TestHandleCharactersRoute_MethodNotAllowed tests unsupported HTTP methods
func TestHandleCharactersRoute_MethodNotAllowed(t *testing.T) {
	mockStorage := NewMockStorageWithCharacters()
	cfg := &config.Config{
		Port:              8080,

		EnableUserSetting: true,
	}

	server := New(cfg, mockStorage)
	userID := int64(12345)

	// Create a valid token for auth
	mockStorage.CreateLoginToken(&storage.LoginToken{
		UserID:    userID,
		Token:     "test-token",
		ExpiresAt: time.Now().Add(24 * time.Hour),
	})

	req := httptest.NewRequest("PATCH", "/api/manager/characters", nil)
	req.Header.Set("X-User-ID", "12345")
	req.Header.Set("X-Auth-Token", "test-token")
	w := httptest.NewRecorder()

	// Call through the auth middleware
	server.withAuth(server.handleCharactersRoute).ServeHTTP(w, req)

	if w.Code != http.StatusMethodNotAllowed {
		t.Errorf("Expected status 405, got %d", w.Code)
	}
}
