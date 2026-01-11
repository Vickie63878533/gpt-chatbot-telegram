package manager

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/tbxark/ChatGPT-Telegram-Workers/go_version/internal/storage"
)

// MockStorage implements a mock storage for testing
type MockStorage struct {
	tokens map[int64]*storage.LoginToken
}

func NewMockStorage() *MockStorage {
	return &MockStorage{
		tokens: make(map[int64]*storage.LoginToken),
	}
}

func (m *MockStorage) ValidateLoginToken(userID int64, token string) (bool, error) {
	t, exists := m.tokens[userID]
	if !exists {
		return false, nil
	}

	if t.Token != token {
		return false, nil
	}

	if time.Now().After(t.ExpiresAt) {
		return false, nil
	}

	return true, nil
}

func (m *MockStorage) CreateLoginToken(token *storage.LoginToken) error {
	m.tokens[token.UserID] = token
	return nil
}

func (m *MockStorage) GetLoginToken(userID int64) (*storage.LoginToken, error) {
	t, exists := m.tokens[userID]
	if !exists {
		return nil, storage.ErrNotFound
	}
	return t, nil
}

func (m *MockStorage) DeleteLoginToken(userID int64) error {
	delete(m.tokens, userID)
	return nil
}

func (m *MockStorage) CleanupExpiredTokens() error {
	return nil
}

// Implement other required methods as no-ops
func (m *MockStorage) GetChatHistory(ctx *storage.SessionContext) ([]storage.HistoryItem, error) {
	return nil, nil
}
func (m *MockStorage) SaveChatHistory(ctx *storage.SessionContext, history []storage.HistoryItem) error {
	return nil
}
func (m *MockStorage) DeleteChatHistory(ctx *storage.SessionContext) error {
	return nil
}
func (m *MockStorage) DeleteAllChatHistory() error {
	return nil
}
func (m *MockStorage) GetUserConfig(ctx *storage.SessionContext) (*storage.UserConfig, error) {
	return nil, nil
}
func (m *MockStorage) SaveUserConfig(ctx *storage.SessionContext, config *storage.UserConfig) error {
	return nil
}
func (m *MockStorage) GetMessageIDs(ctx *storage.SessionContext) ([]int, error) {
	return nil, nil
}
func (m *MockStorage) SaveMessageIDs(ctx *storage.SessionContext, ids []int) error {
	return nil
}
func (m *MockStorage) GetGroupAdmins(chatID int64) ([]storage.ChatMember, error) {
	return nil, nil
}
func (m *MockStorage) SaveGroupAdmins(chatID int64, admins []storage.ChatMember, ttl int) error {
	return nil
}
func (m *MockStorage) CreateCharacterCard(card *storage.CharacterCard) error {
	return nil
}
func (m *MockStorage) GetCharacterCard(id uint) (*storage.CharacterCard, error) {
	return nil, nil
}
func (m *MockStorage) ListCharacterCards(userID *int64) ([]*storage.CharacterCard, error) {
	return nil, nil
}
func (m *MockStorage) UpdateCharacterCard(card *storage.CharacterCard) error {
	return nil
}
func (m *MockStorage) DeleteCharacterCard(id uint) error {
	return nil
}
func (m *MockStorage) ActivateCharacterCard(userID *int64, cardID uint) error {
	return nil
}
func (m *MockStorage) GetActiveCharacterCard(userID *int64) (*storage.CharacterCard, error) {
	return nil, nil
}
func (m *MockStorage) CreateWorldBook(book *storage.WorldBook) error {
	return nil
}
func (m *MockStorage) GetWorldBook(id uint) (*storage.WorldBook, error) {
	return nil, nil
}
func (m *MockStorage) ListWorldBooks(userID *int64) ([]*storage.WorldBook, error) {
	return nil, nil
}
func (m *MockStorage) UpdateWorldBook(book *storage.WorldBook) error {
	return nil
}
func (m *MockStorage) DeleteWorldBook(id uint) error {
	return nil
}
func (m *MockStorage) ActivateWorldBook(userID *int64, bookID uint) error {
	return nil
}
func (m *MockStorage) GetActiveWorldBook(userID *int64) (*storage.WorldBook, error) {
	return nil, nil
}
func (m *MockStorage) CreateWorldBookEntry(entry *storage.WorldBookEntry) error {
	return nil
}
func (m *MockStorage) GetWorldBookEntry(id uint) (*storage.WorldBookEntry, error) {
	return nil, nil
}
func (m *MockStorage) ListWorldBookEntries(bookID uint) ([]*storage.WorldBookEntry, error) {
	return nil, nil
}
func (m *MockStorage) UpdateWorldBookEntry(entry *storage.WorldBookEntry) error {
	return nil
}
func (m *MockStorage) DeleteWorldBookEntry(id uint) error {
	return nil
}
func (m *MockStorage) UpdateWorldBookEntryStatus(id uint, enabled bool) error {
	return nil
}
func (m *MockStorage) CreatePreset(preset *storage.Preset) error {
	return nil
}
func (m *MockStorage) GetPreset(id uint) (*storage.Preset, error) {
	return nil, nil
}
func (m *MockStorage) ListPresets(userID *int64, apiType string) ([]*storage.Preset, error) {
	return nil, nil
}
func (m *MockStorage) UpdatePreset(preset *storage.Preset) error {
	return nil
}
func (m *MockStorage) DeletePreset(id uint) error {
	return nil
}
func (m *MockStorage) ActivatePreset(userID *int64, presetID uint) error {
	return nil
}
func (m *MockStorage) GetActivePreset(userID *int64, apiType string) (*storage.Preset, error) {
	return nil, nil
}
func (m *MockStorage) CreateRegexPattern(pattern *storage.RegexPattern) error {
	return nil
}
func (m *MockStorage) GetRegexPattern(id uint) (*storage.RegexPattern, error) {
	return nil, nil
}
func (m *MockStorage) ListRegexPatterns(userID *int64, patternType string) ([]*storage.RegexPattern, error) {
	return nil, nil
}
func (m *MockStorage) UpdateRegexPattern(pattern *storage.RegexPattern) error {
	return nil
}
func (m *MockStorage) DeleteRegexPattern(id uint) error {
	return nil
}
func (m *MockStorage) UpdateRegexPatternStatus(id uint, enabled bool) error {
	return nil
}
func (m *MockStorage) CleanupExpired() error {
	return nil
}
func (m *MockStorage) Close() error {
	return nil
}

// TestAuthMiddleware_ValidToken tests authentication with a valid token
func TestAuthMiddleware_ValidToken(t *testing.T) {
	mockStorage := NewMockStorage()
	userID := int64(12345)
	token := "valid-token-123"

	// Create a valid token
	mockStorage.tokens[userID] = &storage.LoginToken{
		UserID:    userID,
		Token:     token,
		ExpiresAt: time.Now().Add(24 * time.Hour),
	}

	auth := NewAuthMiddleware(mockStorage)

	// Create a test handler
	handlerCalled := false
	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		handlerCalled = true
		// Verify user ID is in context
		if userID, ok := GetUserIDFromContext(r); !ok || userID != 12345 {
			t.Errorf("User ID not found in context or incorrect value")
		}
		w.WriteHeader(http.StatusOK)
	})

	// Create request with valid headers
	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("X-User-ID", "12345")
	req.Header.Set("X-Auth-Token", token)

	// Create response recorder
	w := httptest.NewRecorder()

	// Call middleware
	auth.Authenticate(testHandler).ServeHTTP(w, req)

	if !handlerCalled {
		t.Error("Handler was not called")
	}

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}
}

// TestAuthMiddleware_MissingHeaders tests authentication with missing headers
func TestAuthMiddleware_MissingHeaders(t *testing.T) {
	mockStorage := NewMockStorage()
	auth := NewAuthMiddleware(mockStorage)

	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()

	auth.Authenticate(testHandler).ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("Expected status 401, got %d", w.Code)
	}
}

// TestAuthMiddleware_InvalidToken tests authentication with invalid token
func TestAuthMiddleware_InvalidToken(t *testing.T) {
	mockStorage := NewMockStorage()
	auth := NewAuthMiddleware(mockStorage)

	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("X-User-ID", "12345")
	req.Header.Set("X-Auth-Token", "invalid-token")

	w := httptest.NewRecorder()

	auth.Authenticate(testHandler).ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("Expected status 401, got %d", w.Code)
	}
}

// TestAuthMiddleware_ExpiredToken tests authentication with expired token
func TestAuthMiddleware_ExpiredToken(t *testing.T) {
	mockStorage := NewMockStorage()
	userID := int64(12345)
	token := "expired-token"

	// Create an expired token
	mockStorage.tokens[userID] = &storage.LoginToken{
		UserID:    userID,
		Token:     token,
		ExpiresAt: time.Now().Add(-1 * time.Hour), // Expired 1 hour ago
	}

	auth := NewAuthMiddleware(mockStorage)

	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("X-User-ID", "12345")
	req.Header.Set("X-Auth-Token", token)

	w := httptest.NewRecorder()

	auth.Authenticate(testHandler).ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("Expected status 401, got %d", w.Code)
	}
}

// TestGetUserIDFromContext tests extracting user ID from context
func TestGetUserIDFromContext(t *testing.T) {
	ctx := context.WithValue(context.Background(), "userID", int64(12345))
	req := httptest.NewRequest("GET", "/test", nil).WithContext(ctx)

	userID, ok := GetUserIDFromContext(req)
	if !ok {
		t.Error("User ID not found in context")
	}

	if userID != 12345 {
		t.Errorf("Expected user ID 12345, got %d", userID)
	}
}
