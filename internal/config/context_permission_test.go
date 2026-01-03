package config

import (
	"testing"

	"github.com/tbxark/ChatGPT-Telegram-Workers/go_version/internal/storage"
)

// TestSetUserConfigValueWithPermission tests the permission-aware config modification
func TestSetUserConfigValueWithPermission(t *testing.T) {
	tests := []struct {
		name              string
		enableUserSetting bool
		isAdmin           bool
		expectError       bool
	}{
		{
			name:              "ENABLE_USER_SETTING=true, regular user can modify",
			enableUserSetting: true,
			isAdmin:           false,
			expectError:       false,
		},
		{
			name:              "ENABLE_USER_SETTING=true, admin can modify",
			enableUserSetting: true,
			isAdmin:           true,
			expectError:       false,
		},
		{
			name:              "ENABLE_USER_SETTING=false, admin can modify",
			enableUserSetting: false,
			isAdmin:           true,
			expectError:       false,
		},
		{
			name:              "ENABLE_USER_SETTING=false, regular user cannot modify",
			enableUserSetting: false,
			isAdmin:           false,
			expectError:       true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup
			cfg := &Config{
				EnableUserSetting: tt.enableUserSetting,
				ChatAdminKey:      []string{"123"}, // Admin user ID
			}

			shareCtx := ShareContext{
				BotToken: "123456:ABC",
				BotID:    123456,
			}

			mockStorage := &MockStorage{}
			wc := NewWorkerContext(shareCtx, mockStorage, cfg)

			// Create permission checker
			permChecker := NewDefaultPermissionChecker(cfg, nil)
			wc.PermissionChecker = permChecker

			// Test user IDs
			var userID int64
			if tt.isAdmin {
				userID = 123 // Admin user
			} else {
				userID = 456 // Regular user
			}
			chatID := int64(-1001234567890) // Group chat

			// Test SetUserConfigValueWithPermission
			err := wc.SetUserConfigValueWithPermission("TEST_KEY", "test_value", []string{}, userID, chatID)

			if tt.expectError && err == nil {
				t.Errorf("Expected error but got none")
			}
			if !tt.expectError && err != nil {
				t.Errorf("Expected no error but got: %v", err)
			}

			// Verify value was set only if no error
			if !tt.expectError {
				if wc.UserConfig.Values["TEST_KEY"] != "test_value" {
					t.Errorf("Expected value to be set")
				}
			}
		})
	}
}

// TestDeleteUserConfigValueWithPermission tests the permission-aware config deletion
func TestDeleteUserConfigValueWithPermission(t *testing.T) {
	tests := []struct {
		name              string
		enableUserSetting bool
		isAdmin           bool
		expectError       bool
	}{
		{
			name:              "ENABLE_USER_SETTING=true, regular user can delete",
			enableUserSetting: true,
			isAdmin:           false,
			expectError:       false,
		},
		{
			name:              "ENABLE_USER_SETTING=false, admin can delete",
			enableUserSetting: false,
			isAdmin:           true,
			expectError:       false,
		},
		{
			name:              "ENABLE_USER_SETTING=false, regular user cannot delete",
			enableUserSetting: false,
			isAdmin:           false,
			expectError:       true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup
			cfg := &Config{
				EnableUserSetting: tt.enableUserSetting,
				ChatAdminKey:      []string{"123"},
			}

			shareCtx := ShareContext{
				BotToken: "123456:ABC",
				BotID:    123456,
			}

			mockStorage := &MockStorage{}
			wc := NewWorkerContext(shareCtx, mockStorage, cfg)

			// Pre-populate a key
			wc.UserConfig.DefineKeys = []string{"TEST_KEY"}
			wc.UserConfig.Values["TEST_KEY"] = "test_value"

			// Create permission checker
			permChecker := NewDefaultPermissionChecker(cfg, nil)
			wc.PermissionChecker = permChecker

			// Test user IDs
			var userID int64
			if tt.isAdmin {
				userID = 123
			} else {
				userID = 456
			}
			chatID := int64(-1001234567890)

			// Test DeleteUserConfigValueWithPermission
			err := wc.DeleteUserConfigValueWithPermission("TEST_KEY", []string{}, userID, chatID)

			if tt.expectError && err == nil {
				t.Errorf("Expected error but got none")
			}
			if !tt.expectError && err != nil {
				t.Errorf("Expected no error but got: %v", err)
			}

			// Verify value was deleted only if no error
			if !tt.expectError {
				if _, exists := wc.UserConfig.Values["TEST_KEY"]; exists {
					t.Errorf("Expected value to be deleted")
				}
			}
		})
	}
}

// TestClearUserConfigWithPermission tests the permission-aware config clearing
func TestClearUserConfigWithPermission(t *testing.T) {
	tests := []struct {
		name              string
		enableUserSetting bool
		isAdmin           bool
		expectError       bool
	}{
		{
			name:              "ENABLE_USER_SETTING=true, regular user can clear",
			enableUserSetting: true,
			isAdmin:           false,
			expectError:       false,
		},
		{
			name:              "ENABLE_USER_SETTING=false, admin can clear",
			enableUserSetting: false,
			isAdmin:           true,
			expectError:       false,
		},
		{
			name:              "ENABLE_USER_SETTING=false, regular user cannot clear",
			enableUserSetting: false,
			isAdmin:           false,
			expectError:       true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup
			cfg := &Config{
				EnableUserSetting: tt.enableUserSetting,
				ChatAdminKey:      []string{"123"},
			}

			shareCtx := ShareContext{
				BotToken: "123456:ABC",
				BotID:    123456,
			}

			mockStorage := &MockStorage{}
			wc := NewWorkerContext(shareCtx, mockStorage, cfg)

			// Pre-populate keys
			wc.UserConfig.DefineKeys = []string{"KEY1", "KEY2", "LOCKED_KEY"}
			wc.UserConfig.Values["KEY1"] = "value1"
			wc.UserConfig.Values["KEY2"] = "value2"
			wc.UserConfig.Values["LOCKED_KEY"] = "locked_value"

			// Create permission checker
			permChecker := NewDefaultPermissionChecker(cfg, nil)
			wc.PermissionChecker = permChecker

			// Test user IDs
			var userID int64
			if tt.isAdmin {
				userID = 123
			} else {
				userID = 456
			}
			chatID := int64(-1001234567890)

			// Test ClearUserConfigWithPermission
			lockedKeys := []string{"LOCKED_KEY"}
			err := wc.ClearUserConfigWithPermission(lockedKeys, userID, chatID)

			if tt.expectError && err == nil {
				t.Errorf("Expected error but got none")
			}
			if !tt.expectError && err != nil {
				t.Errorf("Expected no error but got: %v", err)
			}

			// Verify config was cleared only if no error
			if !tt.expectError {
				// Locked key should remain
				if _, exists := wc.UserConfig.Values["LOCKED_KEY"]; !exists {
					t.Errorf("Expected locked key to remain")
				}
				// Other keys should be cleared
				if _, exists := wc.UserConfig.Values["KEY1"]; exists {
					t.Errorf("Expected KEY1 to be cleared")
				}
				if _, exists := wc.UserConfig.Values["KEY2"]; exists {
					t.Errorf("Expected KEY2 to be cleared")
				}
			}
		})
	}
}

// TestLoadUserConfig_WithEnableUserSettingFalse tests that LoadUserConfig returns empty config when ENABLE_USER_SETTING=false
func TestLoadUserConfig_WithEnableUserSettingFalse(t *testing.T) {
	cfg := &Config{
		EnableUserSetting: false,
	}

	shareCtx := ShareContext{
		BotToken: "123456:ABC",
		BotID:    123456,
	}

	mockStorage := &MockStorage{
		userConfigs: map[string]*storage.UserConfig{
			"test_session": {
				DefineKeys: []string{"KEY1"},
				Values:     map[string]interface{}{"KEY1": "value1"},
			},
		},
	}

	wc := NewWorkerContext(shareCtx, mockStorage, cfg)

	sessionCtx := &storage.SessionContext{
		ChatID: 123,
		BotID:  123456,
	}

	// Load user config
	err := wc.LoadUserConfig(sessionCtx)
	if err != nil {
		t.Fatalf("LoadUserConfig failed: %v", err)
	}

	// Verify that user config is empty (global config will be used)
	if len(wc.UserConfig.DefineKeys) != 0 {
		t.Errorf("Expected empty DefineKeys when ENABLE_USER_SETTING=false, got %v", wc.UserConfig.DefineKeys)
	}
	if len(wc.UserConfig.Values) != 0 {
		t.Errorf("Expected empty Values when ENABLE_USER_SETTING=false, got %v", wc.UserConfig.Values)
	}
}

// TestLoadUserConfig_WithEnableUserSettingTrue tests that LoadUserConfig loads config when ENABLE_USER_SETTING=true
func TestLoadUserConfig_WithEnableUserSettingTrue(t *testing.T) {
	cfg := &Config{
		EnableUserSetting: true,
	}

	shareCtx := ShareContext{
		BotToken: "123456:ABC",
		BotID:    123456,
	}

	expectedConfig := &storage.UserConfig{
		DefineKeys: []string{"KEY1"},
		Values:     map[string]interface{}{"KEY1": "value1"},
	}

	mockStorage := &MockStorage{
		userConfigs: map[string]*storage.UserConfig{
			"test_session": expectedConfig,
		},
	}

	wc := NewWorkerContext(shareCtx, mockStorage, cfg)

	sessionCtx := &storage.SessionContext{
		ChatID: 123,
		BotID:  123456,
	}

	// Load user config
	err := wc.LoadUserConfig(sessionCtx)
	if err != nil {
		t.Fatalf("LoadUserConfig failed: %v", err)
	}

	// Verify that user config is loaded
	if len(wc.UserConfig.DefineKeys) != 1 || wc.UserConfig.DefineKeys[0] != "KEY1" {
		t.Errorf("Expected DefineKeys to be loaded, got %v", wc.UserConfig.DefineKeys)
	}
	if wc.UserConfig.Values["KEY1"] != "value1" {
		t.Errorf("Expected Values to be loaded, got %v", wc.UserConfig.Values)
	}
}
