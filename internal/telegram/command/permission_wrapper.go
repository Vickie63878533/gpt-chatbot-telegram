package command

import (
	"fmt"
	"log/slog"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/tbxark/ChatGPT-Telegram-Workers/go_version/internal/config"
)

// PermissionAwareCommand wraps a command with permission checking
type PermissionAwareCommand struct {
	command           Command
	permissionChecker config.PermissionChecker
	requireAdminOnly  bool
}

// NewPermissionAwareCommand creates a new permission-aware command wrapper
func NewPermissionAwareCommand(cmd Command, permChecker config.PermissionChecker, requireAdminOnly bool) *PermissionAwareCommand {
	return &PermissionAwareCommand{
		command:           cmd,
		permissionChecker: permChecker,
		requireAdminOnly:  requireAdminOnly,
	}
}

// Name returns the command name
func (p *PermissionAwareCommand) Name() string {
	return p.command.Name()
}

// Description returns the command description
func (p *PermissionAwareCommand) Description(lang string) string {
	return p.command.Description(lang)
}

// Scopes returns the command scopes
func (p *PermissionAwareCommand) Scopes() []string {
	return p.command.Scopes()
}

// NeedAuth returns the auth checker
// If requireAdminOnly is true, we need to check admin permissions
func (p *PermissionAwareCommand) NeedAuth() AuthChecker {
	if !p.requireAdminOnly {
		// No additional permission check needed
		return p.command.NeedAuth()
	}

	// Create a combined auth checker that checks both original auth and admin permission
	return AuthCheckerFunc(func(message *tgbotapi.Message, ctx *config.WorkerContext) (bool, error) {
		// First check original auth
		originalAuth := p.command.NeedAuth()
		if originalAuth != nil {
			authorized, err := originalAuth.Check(message, ctx)
			if err != nil {
				return false, err
			}
			if !authorized {
				return false, nil
			}
		}

		// Then check admin permission
		isAdmin, err := p.permissionChecker.IsAdmin(message.From.ID, message.Chat.ID, ctx)
		if err != nil {
			slog.Error("Failed to check admin permission", "error", err)
			return false, fmt.Errorf("failed to check admin permission: %w", err)
		}

		return isAdmin, nil
	})
}

// Handle processes the command
func (p *PermissionAwareCommand) Handle(message *tgbotapi.Message, args string, ctx *config.WorkerContext) error {
	return p.command.Handle(message, args, ctx)
}
