package manager

import (
	"log"
	"strconv"
	"strings"

	"github.com/tbxark/ChatGPT-Telegram-Workers/go_version/internal/config"
)

// PermissionChecker handles permission checks for manager operations
type PermissionChecker struct {
	config *config.Config
}

// NewPermissionChecker creates a new permission checker
func NewPermissionChecker(cfg *config.Config) *PermissionChecker {
	return &PermissionChecker{
		config: cfg,
	}
}

// CanModifyGlobal checks if a user can modify global settings
func (p *PermissionChecker) CanModifyGlobal(userID int64) bool {
	return p.isAdmin(userID)
}

// CanModifyPersonal checks if a user can modify personal settings
func (p *PermissionChecker) CanModifyPersonal(userID int64) bool {
	// If user settings are enabled, all users can modify their own settings
	if p.config.EnableUserSetting {
		return true
	}

	// Otherwise, only admins can modify settings
	return p.isAdmin(userID)
}

// CanAccessResource checks if a user can access a specific resource
func (p *PermissionChecker) CanAccessResource(userID int64, resourceUserID *int64) bool {
	// Global resources (resourceUserID == nil) are accessible to everyone
	if resourceUserID == nil {
		return true
	}

	// User can access their own resources
	if *resourceUserID == userID {
		return true
	}

	// Admins can access all resources
	if p.isAdmin(userID) {
		return true
	}

	return false
}

// CanModifyResource checks if a user can modify a specific resource
func (p *PermissionChecker) CanModifyResource(userID int64, resourceUserID *int64) bool {
	// Global resources can only be modified by admins
	if resourceUserID == nil {
		return p.isAdmin(userID)
	}

	// User can modify their own resources if personal modification is allowed
	if *resourceUserID == userID {
		return p.CanModifyPersonal(userID)
	}

	// Admins can modify all resources
	if p.isAdmin(userID) {
		return true
	}

	return false
}

// isAdmin checks if a user is an admin
func (p *PermissionChecker) isAdmin(userID int64) bool {
	if len(p.config.ChatAdminKey) == 0 {
		return false
	}

	userIDStr := strconv.FormatInt(userID, 10)

	for _, adminKey := range p.config.ChatAdminKey {
		// Handle comma-separated list
		admins := strings.Split(adminKey, ",")
		for _, admin := range admins {
			admin = strings.TrimSpace(admin)
			if admin == userIDStr {
				log.Printf("User %d is an admin", userID)
				return true
			}
		}
	}

	return false
}
