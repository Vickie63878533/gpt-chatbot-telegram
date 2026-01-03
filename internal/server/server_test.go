package server

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/tbxark/ChatGPT-Telegram-Workers/go_version/internal/config"
	"github.com/tbxark/ChatGPT-Telegram-Workers/go_version/internal/storage"
)

// mockStorage is a mock implementation of storage.Storage for testing
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
	return &storage.UserConfig{}, nil
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

func (m *mockStorage) SaveDebugMessage(ctx *storage.SessionContext, message interface{}, ttl int) error {
	return nil
}

func (m *mockStorage) CleanupExpired() error {
	return nil
}

func (m *mockStorage) Close() error {
	return nil
}

func createTestServer() *Server {
	cfg := &config.Config{
		Port:                    8080,
		Language:                "en",
		TelegramAvailableTokens: []string{"test_token_123"},
	}
	return New(cfg, &mockStorage{})
}

func TestHandleRoot(t *testing.T) {
	srv := createTestServer()

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	w := httptest.NewRecorder()

	srv.router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	if !strings.Contains(w.Body.String(), "Telegram Bot") {
		t.Error("Expected welcome page to contain 'Telegram Bot'")
	}

	contentType := w.Header().Get("Content-Type")
	if !strings.Contains(contentType, "text/html") {
		t.Errorf("Expected Content-Type to contain 'text/html', got '%s'", contentType)
	}
}

func TestHandleRootMethodNotAllowed(t *testing.T) {
	srv := createTestServer()

	req := httptest.NewRequest(http.MethodPost, "/", nil)
	w := httptest.NewRecorder()

	srv.router.ServeHTTP(w, req)

	if w.Code != http.StatusMethodNotAllowed {
		t.Errorf("Expected status 405, got %d", w.Code)
	}
}

func TestHandleInit(t *testing.T) {
	srv := createTestServer()

	req := httptest.NewRequest(http.MethodGet, "/init", nil)
	req.Host = "example.com"
	w := httptest.NewRecorder()

	srv.router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	body := w.Body.String()

	// Check for key elements in the response
	expectedStrings := []string{
		"Webhook Initialization",
		"example.com",
		// Note: Actual webhook setting will fail with test token, but page should still render
	}

	for _, expected := range expectedStrings {
		if !strings.Contains(body, expected) {
			t.Errorf("Expected response to contain '%s'", expected)
		}
	}

	contentType := w.Header().Get("Content-Type")
	if !strings.Contains(contentType, "text/html") {
		t.Errorf("Expected Content-Type to contain 'text/html', got '%s'", contentType)
	}
}

func TestHandleInitMethodNotAllowed(t *testing.T) {
	srv := createTestServer()

	req := httptest.NewRequest(http.MethodPost, "/init", nil)
	w := httptest.NewRecorder()

	srv.router.ServeHTTP(w, req)

	if w.Code != http.StatusMethodNotAllowed {
		t.Errorf("Expected status 405, got %d", w.Code)
	}
}

func TestHandleTelegramWebhook(t *testing.T) {
	srv := createTestServer()

	req := httptest.NewRequest(http.MethodPost, "/telegram/test_token_123/webhook", strings.NewReader("{}"))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	srv.router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	contentType := w.Header().Get("Content-Type")
	if !strings.Contains(contentType, "application/json") {
		t.Errorf("Expected Content-Type to contain 'application/json', got '%s'", contentType)
	}
}

func TestHandleTelegramSafehook(t *testing.T) {
	srv := createTestServer()

	req := httptest.NewRequest(http.MethodPost, "/telegram/test_token_123/safehook", strings.NewReader("{}"))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	srv.router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}
}

func TestHandleTelegramInvalidToken(t *testing.T) {
	srv := createTestServer()

	req := httptest.NewRequest(http.MethodPost, "/telegram/invalid_token/webhook", strings.NewReader("{}"))
	w := httptest.NewRecorder()

	srv.router.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("Expected status 401, got %d", w.Code)
	}
}

func TestHandleTelegramInvalidEndpoint(t *testing.T) {
	srv := createTestServer()

	req := httptest.NewRequest(http.MethodPost, "/telegram/test_token_123/invalid", strings.NewReader("{}"))
	w := httptest.NewRecorder()

	srv.router.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("Expected status 404, got %d", w.Code)
	}
}

func TestHandleTelegramMethodNotAllowed(t *testing.T) {
	srv := createTestServer()

	req := httptest.NewRequest(http.MethodGet, "/telegram/test_token_123/webhook", nil)
	w := httptest.NewRecorder()

	srv.router.ServeHTTP(w, req)

	if w.Code != http.StatusMethodNotAllowed {
		t.Errorf("Expected status 405, got %d", w.Code)
	}
}

func TestHandleTelegramInvalidPath(t *testing.T) {
	srv := createTestServer()

	req := httptest.NewRequest(http.MethodPost, "/telegram/invalid", strings.NewReader("{}"))
	w := httptest.NewRecorder()

	srv.router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d", w.Code)
	}
}

func TestNotFoundRoute(t *testing.T) {
	srv := createTestServer()

	req := httptest.NewRequest(http.MethodGet, "/nonexistent", nil)
	w := httptest.NewRecorder()

	srv.router.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("Expected status 404, got %d", w.Code)
	}
}
