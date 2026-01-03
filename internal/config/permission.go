package config

import (
	"fmt"
	"strconv"
	"strings"
)

// GroupAdminChecker is a function type for checking if a user is a group admin
// This is injected to avoid circular imports
type GroupAdminChecker func(chatID int64, userID int64, ctx *WorkerContext) (bool, error)

// PermissionChecker defines the interface for checking user permissions
type PermissionChecker interface {
	// IsAdmin checks if a user is an administrator
	// Returns true if the user is in CHAT_ADMIN_KEY or is a Telegram group admin
	IsAdmin(userID int64, chatID int64, ctx *WorkerContext) (bool, error)

	// CanModifyConfig checks if a user can modify configuration
	// Returns true if ENABLE_USER_SETTING is true or user is an admin
	CanModifyConfig(userID int64, chatID int64, ctx *WorkerContext) (bool, error)
}

// DefaultPermissionChecker is the default implementation of PermissionChecker
type DefaultPermissionChecker struct {
	config            *Config
	groupAdminChecker GroupAdminChecker
}

// NewDefaultPermissionChecker creates a new DefaultPermissionChecker
// groupAdminChecker is a function that checks if a user is a group admin
// If nil, group admin checks will be skipped
func NewDefaultPermissionChecker(config *Config, groupAdminChecker GroupAdminChecker) *DefaultPermissionChecker {
	return &DefaultPermissionChecker{
		config:            config,
		groupAdminChecker: groupAdminChecker,
	}
}

// IsAdmin checks if a user is an administrator
// First checks CHAT_ADMIN_KEY, then checks Telegram group admin status
func (p *DefaultPermissionChecker) IsAdmin(userID int64, chatID int64, ctx *WorkerContext) (bool, error) {
	// First check if user is in CHAT_ADMIN_KEY
	userIDStr := strconv.FormatInt(userID, 10)
	for _, adminID := range p.config.ChatAdminKey {
		if strings.TrimSpace(adminID) == userIDStr {
			return true, nil
		}
	}

	// If not in CHAT_ADMIN_KEY, check if user is a Telegram group admin
	// Only check Telegram admin status for group chats
	if chatID < 0 && p.groupAdminChecker != nil { // Group chats have negative IDs
		isGroupAdmin, err := p.groupAdminChecker(chatID, userID, ctx)
		if err != nil {
			return false, fmt.Errorf("failed to check group admin status: %w", err)
		}
		return isGroupAdmin, nil
	}

	// For private chats, only CHAT_ADMIN_KEY matters
	return false, nil
}

// CanModifyConfig checks if a user can modify configuration
// If ENABLE_USER_SETTING is false, only admins can modify
// If ENABLE_USER_SETTING is true, all users can modify
func (p *DefaultPermissionChecker) CanModifyConfig(userID int64, chatID int64, ctx *WorkerContext) (bool, error) {
	// If user settings are enabled, all users can modify
	if p.config.EnableUserSetting {
		return true, nil
	}

	// If user settings are disabled, only admins can modify
	return p.IsAdmin(userID, chatID, ctx)
}
