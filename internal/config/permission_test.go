package config

import (
	"os"
	"testing"

	"github.com/tbxark/ChatGPT-Telegram-Workers/go_version/internal/storage"
)

// MockStorage is a mock implementation of storage.Storage for testing
type MockStorage struct {
	groupAdmins map[int64][]storage.ChatMember
	userConfigs map[string]*storage.UserConfig
}

func (m *MockStorage) GetChatHistory(ctx *storage.SessionContext) ([]storage.HistoryItem, error) {
	return []storage.HistoryItem{}, nil
}

func (m *MockStorage) SaveChatHistory(ctx *storage.SessionContext, history []storage.HistoryItem) error {
	return nil
}

func (m *MockStorage) DeleteChatHistory(ctx *storage.SessionContext) error {
	return nil
}

func (m *MockStorage) GetUserConfig(ctx *storage.SessionContext) (*storage.UserConfig, error) {
	if m.userConfigs != nil {
		if config, ok := m.userConfigs["test_session"]; ok {
			return config, nil
		}
	}
	return &storage.UserConfig{Values: make(map[string]interface{})}, nil
}

func (m *MockStorage) SaveUserConfig(ctx *storage.SessionContext, config *storage.UserConfig) error {
	return nil
}

func (m *MockStorage) GetMessageIDs(ctx *storage.SessionContext) ([]int, error) {
	return []int{}, nil
}

func (m *MockStorage) SaveMessageIDs(ctx *storage.SessionContext, ids []int) error {
	return nil
}

func (m *MockStorage) GetGroupAdmins(chatID int64) ([]storage.ChatMember, error) {
	if admins, ok := m.groupAdmins[chatID]; ok {
		return admins, nil
	}
	return []storage.ChatMember{}, nil
}

func (m *MockStorage) SaveGroupAdmins(chatID int64, admins []storage.ChatMember, ttl int) error {
	m.groupAdmins[chatID] = admins
	return nil
}

func (m *MockStorage) CleanupExpired() error {
	return nil
}

func (m *MockStorage) Close() error {
	return nil
}

// MockBotAPI is a mock implementation of the bot API for testing
type MockBotAPI struct {
	admins map[int64][]storage.ChatMember
}

func TestDefaultPermissionChecker_IsAdmin_WithChatAdminKey(t *testing.T) {
	// Setup
	cfg := &Config{
		EnableUserSetting: true,
		ChatAdminKey:      []string{"123", "456", "789"},
	}
	checker := NewDefaultPermissionChecker(cfg, nil)

	// Create mock context
	mockStorage := &MockStorage{groupAdmins: make(map[int64][]storage.ChatMember)}
	ctx := &WorkerContext{
		DB: mockStorage,
	}

	tests := []struct {
		name    string
		userID  int64
		chatID  int64
		want    bool
		wantErr bool
	}{
		{
			name:    "user in CHAT_ADMIN_KEY",
			userID:  123,
			chatID:  -1, // group chat
			want:    true,
			wantErr: false,
		},
		{
			name:    "user in CHAT_ADMIN_KEY with spaces",
			userID:  456,
			chatID:  -1,
			want:    true,
			wantErr: false,
		},
		{
			name:    "user not in CHAT_ADMIN_KEY",
			userID:  999,
			chatID:  -1,
			want:    false,
			wantErr: false,
		},
		{
			name:    "private chat, user in CHAT_ADMIN_KEY",
			userID:  123,
			chatID:  1, // private chat
			want:    true,
			wantErr: false,
		},
		{
			name:    "private chat, user not in CHAT_ADMIN_KEY",
			userID:  999,
			chatID:  1,
			want:    false,
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := checker.IsAdmin(tt.userID, tt.chatID, ctx)
			if (err != nil) != tt.wantErr {
				t.Errorf("IsAdmin() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("IsAdmin() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestDefaultPermissionChecker_IsAdmin_WithSpaces(t *testing.T) {
	// Test that spaces in CHAT_ADMIN_KEY are trimmed
	cfg := &Config{
		EnableUserSetting: true,
		ChatAdminKey:      []string{" 123 ", "456", " 789"},
	}
	checker := NewDefaultPermissionChecker(cfg, nil)

	mockStorage := &MockStorage{groupAdmins: make(map[int64][]storage.ChatMember)}
	ctx := &WorkerContext{
		DB: mockStorage,
	}

	got, err := checker.IsAdmin(123, 1, ctx)
	if err != nil {
		t.Errorf("IsAdmin() error = %v", err)
	}
	if !got {
		t.Errorf("IsAdmin() = %v, want true (spaces should be trimmed)", got)
	}
}

func TestDefaultPermissionChecker_CanModifyConfig_EnableUserSetting(t *testing.T) {
	// Test with ENABLE_USER_SETTING = true
	cfg := &Config{
		EnableUserSetting: true,
		ChatAdminKey:      []string{"123"},
	}
	checker := NewDefaultPermissionChecker(cfg, nil)

	mockStorage := &MockStorage{groupAdmins: make(map[int64][]storage.ChatMember)}
	ctx := &WorkerContext{
		DB: mockStorage,
	}

	tests := []struct {
		name   string
		userID int64
		chatID int64
		want   bool
	}{
		{
			name:   "admin user can modify",
			userID: 123,
			chatID: 1,
			want:   true,
		},
		{
			name:   "regular user can modify when ENABLE_USER_SETTING=true",
			userID: 999,
			chatID: 1,
			want:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := checker.CanModifyConfig(tt.userID, tt.chatID, ctx)
			if err != nil {
				t.Errorf("CanModifyConfig() error = %v", err)
			}
			if got != tt.want {
				t.Errorf("CanModifyConfig() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestDefaultPermissionChecker_CanModifyConfig_DisableUserSetting(t *testing.T) {
	// Test with ENABLE_USER_SETTING = false
	cfg := &Config{
		EnableUserSetting: false,
		ChatAdminKey:      []string{"123"},
	}
	checker := NewDefaultPermissionChecker(cfg, nil)

	mockStorage := &MockStorage{groupAdmins: make(map[int64][]storage.ChatMember)}
	ctx := &WorkerContext{
		DB: mockStorage,
	}

	tests := []struct {
		name   string
		userID int64
		chatID int64
		want   bool
	}{
		{
			name:   "admin user can modify",
			userID: 123,
			chatID: 1,
			want:   true,
		},
		{
			name:   "regular user cannot modify when ENABLE_USER_SETTING=false",
			userID: 999,
			chatID: 1,
			want:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := checker.CanModifyConfig(tt.userID, tt.chatID, ctx)
			if err != nil {
				t.Errorf("CanModifyConfig() error = %v", err)
			}
			if got != tt.want {
				t.Errorf("CanModifyConfig() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestDefaultPermissionChecker_EmptyChatAdminKey(t *testing.T) {
	// Test with empty CHAT_ADMIN_KEY
	cfg := &Config{
		EnableUserSetting: true,
		ChatAdminKey:      []string{},
	}
	checker := NewDefaultPermissionChecker(cfg, nil)

	mockStorage := &MockStorage{groupAdmins: make(map[int64][]storage.ChatMember)}
	ctx := &WorkerContext{
		DB: mockStorage,
	}

	got, err := checker.IsAdmin(123, 1, ctx)
	if err != nil {
		t.Errorf("IsAdmin() error = %v", err)
	}
	if got {
		t.Errorf("IsAdmin() = %v, want false (no admins configured)", got)
	}
}

func TestLoadConfigWithPermissionFields(t *testing.T) {
	// Test that permission fields are loaded correctly
	os.Setenv("TELEGRAM_AVAILABLE_TOKENS", "123456:ABC-DEF")
	os.Setenv("ENABLE_USER_SETTING", "false")
	os.Setenv("CHAT_ADMIN_KEY", "123,456,789")
	defer func() {
		os.Unsetenv("TELEGRAM_AVAILABLE_TOKENS")
		os.Unsetenv("ENABLE_USER_SETTING")
		os.Unsetenv("CHAT_ADMIN_KEY")
	}()

	cfg, err := LoadConfig()
	if err != nil {
		t.Fatalf("LoadConfig() failed: %v", err)
	}

	if cfg.EnableUserSetting != false {
		t.Errorf("Expected EnableUserSetting to be false, got %v", cfg.EnableUserSetting)
	}

	if len(cfg.ChatAdminKey) != 3 {
		t.Errorf("Expected ChatAdminKey to have 3 items, got %d", len(cfg.ChatAdminKey))
	}

	if cfg.ChatAdminKey[0] != "123" || cfg.ChatAdminKey[1] != "456" || cfg.ChatAdminKey[2] != "789" {
		t.Errorf("ChatAdminKey values don't match: %v", cfg.ChatAdminKey)
	}
}

func TestLoadConfigPermissionDefaults(t *testing.T) {
	// Test default values for permission fields
	os.Setenv("TELEGRAM_AVAILABLE_TOKENS", "123456:ABC-DEF")
	defer os.Unsetenv("TELEGRAM_AVAILABLE_TOKENS")

	cfg, err := LoadConfig()
	if err != nil {
		t.Fatalf("LoadConfig() failed: %v", err)
	}

	if cfg.EnableUserSetting != true {
		t.Errorf("Expected EnableUserSetting default to be true, got %v", cfg.EnableUserSetting)
	}

	if len(cfg.ChatAdminKey) != 0 {
		t.Errorf("Expected ChatAdminKey default to be empty, got %v", cfg.ChatAdminKey)
	}
}
