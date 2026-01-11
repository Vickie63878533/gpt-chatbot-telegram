package manager

import (
	"testing"

	"github.com/tbxark/ChatGPT-Telegram-Workers/go_version/internal/config"
)

// TestPermissionChecker_CanModifyGlobal tests global modification permissions
func TestPermissionChecker_CanModifyGlobal(t *testing.T) {
	tests := []struct {
		name         string
		userID       int64
		adminKeys    []string
		wantModify   bool
	}{
		{
			name:       "Admin can modify global",
			userID:     12345,
			adminKeys:  []string{"12345"},
			wantModify: true,
		},
		{
			name:       "Non-admin cannot modify global",
			userID:     12345,
			adminKeys:  []string{"99999"},
			wantModify: false,
		},
		{
			name:       "No admins configured",
			userID:     12345,
			adminKeys:  []string{},
			wantModify: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &config.Config{
				ChatAdminKey: tt.adminKeys,
			}
			checker := NewPermissionChecker(cfg)

			got := checker.CanModifyGlobal(tt.userID)
			if got != tt.wantModify {
				t.Errorf("CanModifyGlobal(%d) = %v, want %v", tt.userID, got, tt.wantModify)
			}
		})
	}
}

// TestPermissionChecker_CanModifyPersonal tests personal modification permissions
func TestPermissionChecker_CanModifyPersonal(t *testing.T) {
	tests := []struct {
		name              string
		userID            int64
		enableUserSetting bool
		adminKeys         []string
		wantModify        bool
	}{
		{
			name:              "User can modify personal when enabled",
			userID:            12345,
			enableUserSetting: true,
			adminKeys:         []string{},
			wantModify:        true,
		},
		{
			name:              "User cannot modify personal when disabled",
			userID:            12345,
			enableUserSetting: false,
			adminKeys:         []string{},
			wantModify:        false,
		},
		{
			name:              "Admin can modify personal when disabled",
			userID:            12345,
			enableUserSetting: false,
			adminKeys:         []string{"12345"},
			wantModify:        true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &config.Config{
				EnableUserSetting: tt.enableUserSetting,
				ChatAdminKey:      tt.adminKeys,
			}
			checker := NewPermissionChecker(cfg)

			got := checker.CanModifyPersonal(tt.userID)
			if got != tt.wantModify {
				t.Errorf("CanModifyPersonal(%d) = %v, want %v", tt.userID, got, tt.wantModify)
			}
		})
	}
}

// TestPermissionChecker_CanAccessResource tests resource access permissions
func TestPermissionChecker_CanAccessResource(t *testing.T) {
	userID := int64(12345)
	otherUserID := int64(99999)

	tests := []struct {
		name            string
		userID          int64
		resourceUserID  *int64
		adminKeys       []string
		wantAccess      bool
	}{
		{
			name:           "Global resource accessible to all",
			userID:         12345,
			resourceUserID: nil,
			adminKeys:      []string{},
			wantAccess:     true,
		},
		{
			name:           "User can access own resource",
			userID:         12345,
			resourceUserID: &userID,
			adminKeys:      []string{},
			wantAccess:     true,
		},
		{
			name:           "User cannot access other's resource",
			userID:         12345,
			resourceUserID: &otherUserID,
			adminKeys:      []string{},
			wantAccess:     false,
		},
		{
			name:           "Admin can access other's resource",
			userID:         12345,
			resourceUserID: &otherUserID,
			adminKeys:      []string{"12345"},
			wantAccess:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &config.Config{
				ChatAdminKey: tt.adminKeys,
			}
			checker := NewPermissionChecker(cfg)

			got := checker.CanAccessResource(tt.userID, tt.resourceUserID)
			if got != tt.wantAccess {
				t.Errorf("CanAccessResource(%d, %v) = %v, want %v", tt.userID, tt.resourceUserID, got, tt.wantAccess)
			}
		})
	}
}

// TestPermissionChecker_CanModifyResource tests resource modification permissions
func TestPermissionChecker_CanModifyResource(t *testing.T) {
	userID := int64(12345)
	otherUserID := int64(99999)

	tests := []struct {
		name              string
		userID            int64
		resourceUserID    *int64
		enableUserSetting bool
		adminKeys         []string
		wantModify        bool
	}{
		{
			name:              "Global resource only modifiable by admin",
			userID:            12345,
			resourceUserID:    nil,
			enableUserSetting: true,
			adminKeys:         []string{},
			wantModify:        false,
		},
		{
			name:              "Admin can modify global resource",
			userID:            12345,
			resourceUserID:    nil,
			enableUserSetting: true,
			adminKeys:         []string{"12345"},
			wantModify:        true,
		},
		{
			name:              "User can modify own resource when enabled",
			userID:            12345,
			resourceUserID:    &userID,
			enableUserSetting: true,
			adminKeys:         []string{},
			wantModify:        true,
		},
		{
			name:              "User cannot modify own resource when disabled",
			userID:            12345,
			resourceUserID:    &userID,
			enableUserSetting: false,
			adminKeys:         []string{},
			wantModify:        false,
		},
		{
			name:              "User cannot modify other's resource",
			userID:            12345,
			resourceUserID:    &otherUserID,
			enableUserSetting: true,
			adminKeys:         []string{},
			wantModify:        false,
		},
		{
			name:              "Admin can modify other's resource",
			userID:            12345,
			resourceUserID:    &otherUserID,
			enableUserSetting: true,
			adminKeys:         []string{"12345"},
			wantModify:        true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &config.Config{
				EnableUserSetting: tt.enableUserSetting,
				ChatAdminKey:      tt.adminKeys,
			}
			checker := NewPermissionChecker(cfg)

			got := checker.CanModifyResource(tt.userID, tt.resourceUserID)
			if got != tt.wantModify {
				t.Errorf("CanModifyResource(%d, %v) = %v, want %v", tt.userID, tt.resourceUserID, got, tt.wantModify)
			}
		})
	}
}
