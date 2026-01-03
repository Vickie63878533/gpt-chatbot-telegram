package config

import (
	"testing"

	"github.com/tbxark/ChatGPT-Telegram-Workers/go_version/internal/storage"
)

func TestNewShareContext(t *testing.T) {
	tests := []struct {
		name      string
		token     string
		wantBotID int64
		wantErr   bool
	}{
		{
			name:      "valid token",
			token:     "123456:ABC-DEF",
			wantBotID: 123456,
			wantErr:   false,
		},
		{
			name:      "another valid token",
			token:     "987654321:XYZ-123",
			wantBotID: 987654321,
			wantErr:   false,
		},
		{
			name:    "invalid token - no colon",
			token:   "123456ABC",
			wantErr: true,
		},
		{
			name:    "invalid token - empty bot id",
			token:   ":ABC-DEF",
			wantErr: true,
		},
		{
			name:    "invalid token - non-numeric bot id",
			token:   "ABC:DEF",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx, err := NewShareContext(tt.token)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewShareContext() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				if ctx.BotID != tt.wantBotID {
					t.Errorf("NewShareContext() BotID = %v, want %v", ctx.BotID, tt.wantBotID)
				}
				if ctx.BotToken != tt.token {
					t.Errorf("NewShareContext() BotToken = %v, want %v", ctx.BotToken, tt.token)
				}
			}
		})
	}
}

func TestNewSessionContext(t *testing.T) {
	userID := int64(111)
	threadID := int64(222)

	ctx := NewSessionContext(123, 456, &userID, &threadID)

	if ctx.ChatID != 123 {
		t.Errorf("ChatID = %v, want 123", ctx.ChatID)
	}
	if ctx.BotID != 456 {
		t.Errorf("BotID = %v, want 456", ctx.BotID)
	}
	if ctx.UserID == nil || *ctx.UserID != 111 {
		t.Errorf("UserID = %v, want 111", ctx.UserID)
	}
	if ctx.ThreadID == nil || *ctx.ThreadID != 222 {
		t.Errorf("ThreadID = %v, want 222", ctx.ThreadID)
	}
}

func TestNewSessionContextFromChat(t *testing.T) {
	userID := int64(111)
	threadID := int64(222)

	tests := []struct {
		name         string
		chatID       int64
		botID        int64
		isGroup      bool
		shareMode    bool
		userID       *int64
		threadID     *int64
		wantUserID   *int64
		wantThreadID *int64
	}{
		{
			name:         "private chat",
			chatID:       123,
			botID:        456,
			isGroup:      false,
			shareMode:    false,
			userID:       &userID,
			threadID:     nil,
			wantUserID:   nil,
			wantThreadID: nil,
		},
		{
			name:         "group chat - shared mode",
			chatID:       123,
			botID:        456,
			isGroup:      true,
			shareMode:    true,
			userID:       &userID,
			threadID:     nil,
			wantUserID:   nil,
			wantThreadID: nil,
		},
		{
			name:         "group chat - non-shared mode",
			chatID:       123,
			botID:        456,
			isGroup:      true,
			shareMode:    false,
			userID:       &userID,
			threadID:     nil,
			wantUserID:   &userID,
			wantThreadID: nil,
		},
		{
			name:         "group chat with thread",
			chatID:       123,
			botID:        456,
			isGroup:      true,
			shareMode:    true,
			userID:       &userID,
			threadID:     &threadID,
			wantUserID:   nil,
			wantThreadID: &threadID,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := NewSessionContextFromChat(tt.chatID, tt.botID, tt.isGroup, tt.shareMode, tt.userID, tt.threadID)

			if ctx.ChatID != tt.chatID {
				t.Errorf("ChatID = %v, want %v", ctx.ChatID, tt.chatID)
			}
			if ctx.BotID != tt.botID {
				t.Errorf("BotID = %v, want %v", ctx.BotID, tt.botID)
			}

			if tt.wantUserID == nil {
				if ctx.UserID != nil {
					t.Errorf("UserID = %v, want nil", ctx.UserID)
				}
			} else {
				if ctx.UserID == nil || *ctx.UserID != *tt.wantUserID {
					t.Errorf("UserID = %v, want %v", ctx.UserID, *tt.wantUserID)
				}
			}

			if tt.wantThreadID == nil {
				if ctx.ThreadID != nil {
					t.Errorf("ThreadID = %v, want nil", ctx.ThreadID)
				}
			} else {
				if ctx.ThreadID == nil || *ctx.ThreadID != *tt.wantThreadID {
					t.Errorf("ThreadID = %v, want %v", ctx.ThreadID, *tt.wantThreadID)
				}
			}
		})
	}
}

func TestWorkerContextConfigOperations(t *testing.T) {
	// Create a mock storage (we'll use nil for this test since we're not actually calling DB methods)
	shareCtx := ShareContext{
		BotToken: "123456:ABC",
		BotID:    123456,
	}
	cfg := &Config{}
	wc := NewWorkerContext(shareCtx, nil, cfg)

	// Test SetUserConfigValue
	lockedKeys := []string{"LOCKED_KEY"}

	// Set a normal key
	err := wc.SetUserConfigValue("TEST_KEY", "test_value", lockedKeys)
	if err != nil {
		t.Errorf("SetUserConfigValue() failed: %v", err)
	}

	if wc.UserConfig.Values["TEST_KEY"] != "test_value" {
		t.Errorf("Value not set correctly")
	}

	if len(wc.UserConfig.DefineKeys) != 1 || wc.UserConfig.DefineKeys[0] != "TEST_KEY" {
		t.Errorf("DefineKeys not updated correctly")
	}

	// Try to set a locked key
	err = wc.SetUserConfigValue("LOCKED_KEY", "value", lockedKeys)
	if err == nil {
		t.Error("Expected error when setting locked key")
	}

	// Test DeleteUserConfigValue
	err = wc.DeleteUserConfigValue("TEST_KEY", lockedKeys)
	if err != nil {
		t.Errorf("DeleteUserConfigValue() failed: %v", err)
	}

	if _, exists := wc.UserConfig.Values["TEST_KEY"]; exists {
		t.Error("Value should be deleted")
	}

	if len(wc.UserConfig.DefineKeys) != 0 {
		t.Error("DefineKeys should be empty")
	}

	// Try to delete a locked key
	wc.SetUserConfigValue("LOCKED_KEY", "value", []string{})
	err = wc.DeleteUserConfigValue("LOCKED_KEY", lockedKeys)
	if err == nil {
		t.Error("Expected error when deleting locked key")
	}
}

func TestWorkerContextClearUserConfig(t *testing.T) {
	shareCtx := ShareContext{
		BotToken: "123456:ABC",
		BotID:    123456,
	}
	cfg := &Config{}
	wc := NewWorkerContext(shareCtx, nil, cfg)

	lockedKeys := []string{"LOCKED_KEY"}

	// Set some keys
	wc.SetUserConfigValue("KEY1", "value1", []string{})
	wc.SetUserConfigValue("KEY2", "value2", []string{})
	wc.SetUserConfigValue("LOCKED_KEY", "locked_value", []string{})

	// Clear config
	err := wc.ClearUserConfig(lockedKeys)
	if err != nil {
		t.Errorf("ClearUserConfig() failed: %v", err)
	}

	// Check that only locked key remains
	if len(wc.UserConfig.DefineKeys) != 1 || wc.UserConfig.DefineKeys[0] != "LOCKED_KEY" {
		t.Errorf("DefineKeys = %v, want [LOCKED_KEY]", wc.UserConfig.DefineKeys)
	}

	if len(wc.UserConfig.Values) != 1 {
		t.Errorf("Values length = %v, want 1", len(wc.UserConfig.Values))
	}

	if wc.UserConfig.Values["LOCKED_KEY"] != "locked_value" {
		t.Error("Locked key value should be preserved")
	}
}

func TestGetConfigValue(t *testing.T) {
	globalConfig := &Config{
		AIProvider: "openai",
		StreamMode: true,
		Port:       8080,
	}

	shareCtx := ShareContext{
		BotToken: "123456:ABC",
		BotID:    123456,
	}
	cfg := &Config{}
	wc := NewWorkerContext(shareCtx, nil, cfg)

	// Test getting from global config
	if val := wc.GetConfigString("AI_PROVIDER", globalConfig); val != "openai" {
		t.Errorf("GetConfigString() = %v, want openai", val)
	}

	// Set user config override
	wc.SetUserConfigValue("AI_PROVIDER", "azure", []string{})

	// Test getting from user config (should override global)
	if val := wc.GetConfigString("AI_PROVIDER", globalConfig); val != "azure" {
		t.Errorf("GetConfigString() = %v, want azure", val)
	}

	// Test GetConfigInt
	if val := wc.GetConfigInt("PORT", globalConfig); val != 8080 {
		t.Errorf("GetConfigInt() = %v, want 8080", val)
	}

	// Test GetConfigBool
	if val := wc.GetConfigBool("STREAM_MODE", globalConfig); val != true {
		t.Errorf("GetConfigBool() = %v, want true", val)
	}
}

func TestMergeUserConfig(t *testing.T) {
	globalConfig := &Config{
		AIProvider:      "openai",
		OpenAIChatModel: "gpt-4",
		StreamMode:      true,
	}

	userConfig := &storage.UserConfig{
		DefineKeys: []string{"AI_PROVIDER", "STREAM_MODE"},
		Values: map[string]interface{}{
			"AI_PROVIDER": "azure",
			"STREAM_MODE": false,
		},
	}

	merged := MergeUserConfig(globalConfig, userConfig)

	if merged.AIProvider != "azure" {
		t.Errorf("AIProvider = %v, want azure", merged.AIProvider)
	}

	if merged.StreamMode != false {
		t.Errorf("StreamMode = %v, want false", merged.StreamMode)
	}

	// Check that unmodified values remain
	if merged.OpenAIChatModel != "gpt-4" {
		t.Errorf("OpenAIChatModel = %v, want gpt-4", merged.OpenAIChatModel)
	}
}
