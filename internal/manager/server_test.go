package manager

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/tbxark/ChatGPT-Telegram-Workers/go_version/internal/config"
)

// TestServer_New tests server creation
func TestServer_New(t *testing.T) {
	mockStorage := NewMockStorage()
	cfg := &config.Config{
		Port:              8080,

		EnableUserSetting: true,
	}

	server := New(cfg, mockStorage)

	if server == nil {
		t.Error("Server creation failed")
	}

	if server.config != cfg {
		t.Error("Config not set correctly")
	}

	if server.storage != mockStorage {
		t.Error("Storage not set correctly")
	}

	if server.auth == nil {
		t.Error("Auth middleware not initialized")
	}

	if server.permission == nil {
		t.Error("Permission checker not initialized")
	}
}

// TestServer_RegisterRoutes tests route registration
func TestServer_RegisterRoutes(t *testing.T) {
	mockStorage := NewMockStorage()
	cfg := &config.Config{
		Port:              8080,

		EnableUserSetting: true,
	}

	server := New(cfg, mockStorage)
	mux := http.NewServeMux()

	// Register routes
	server.RegisterRoutes(mux)

	// Test health endpoint (no auth required)
	req := httptest.NewRequest("GET", "/api/manager/health", nil)
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	if w.Header().Get("Content-Type") != "application/json" {
		t.Errorf("Expected Content-Type application/json, got %s", w.Header().Get("Content-Type"))
	}
}

// TestServer_HandleHealth tests health endpoint
func TestServer_HandleHealth(t *testing.T) {
	mockStorage := NewMockStorage()
	cfg := &config.Config{
		Port:              8080,

		EnableUserSetting: true,
	}

	server := New(cfg, mockStorage)

	req := httptest.NewRequest("GET", "/api/manager/health", nil)
	w := httptest.NewRecorder()

	server.handleHealth(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	if w.Body.String() != `{"status":"ok"}` {
		t.Errorf("Expected body {\"status\":\"ok\"}, got %s", w.Body.String())
	}
}

// TestServer_HandleCharactersList tests character list endpoint
func TestServer_HandleCharactersList(t *testing.T) {
	mockStorage := NewMockStorageWithCharacters()
	cfg := &config.Config{
		Port:              8080,

		EnableUserSetting: true,
	}

	server := New(cfg, mockStorage)
	userID := int64(12345)

	req := httptest.NewRequest("GET", "/api/manager/characters", nil)
	w := httptest.NewRecorder()

	server.handleListCharacters(w, req, userID)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}
}

// TestServer_HandleWorldBooksList tests world books list endpoint
func TestServer_HandleWorldBooksList(t *testing.T) {
	mockStorage := NewMockStorage()
	cfg := &config.Config{
		Port:              8080,

		EnableUserSetting: true,
	}

	server := New(cfg, mockStorage)
	userID := int64(12345)

	req := httptest.NewRequest("GET", "/api/manager/worldbooks", nil)
	w := httptest.NewRecorder()

	server.handleListWorldBooks(w, req, userID)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}
}

// TestServer_HandlePresetsList tests presets list endpoint
func TestServer_HandlePresetsList(t *testing.T) {
	mockStorage := NewMockStorage()
	cfg := &config.Config{
		Port:              8080,

		EnableUserSetting: true,
	}

	server := New(cfg, mockStorage)
	userID := int64(12345)

	req := httptest.NewRequest("GET", "/api/manager/presets", nil)
	w := httptest.NewRecorder()

	server.handleListPresets(w, req, userID)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}
}

// TestServer_HandleRegexList tests regex list endpoint
func TestServer_HandleRegexList(t *testing.T) {
	mockStorage := NewMockStorage()
	cfg := &config.Config{
		Port:              8080,

		EnableUserSetting: true,
	}

	server := New(cfg, mockStorage)
	userID := int64(12345)

	req := httptest.NewRequest("GET", "/api/manager/regex", nil)
	w := httptest.NewRecorder()

	server.handleListRegexPatterns(w, req, userID)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}
}
