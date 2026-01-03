package command

import (
	"fmt"
	"log/slog"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/tbxark/ChatGPT-Telegram-Workers/go_version/internal/config"
	"github.com/tbxark/ChatGPT-Telegram-Workers/go_version/internal/plugin"
)

// Registry manages command registration and dispatching
type Registry struct {
	commands          map[string]Command
	config            *config.Config
	pluginRegistry    *plugin.PluginRegistry
	permissionChecker config.PermissionChecker
}

// NewRegistry creates a new command registry
func NewRegistry(cfg *config.Config) *Registry {
	return &Registry{
		commands:       make(map[string]Command),
		config:         cfg,
		pluginRegistry: plugin.NewPluginRegistry(),
	}
}

// SetPermissionChecker sets the permission checker for the registry
func (r *Registry) SetPermissionChecker(checker config.PermissionChecker) {
	r.permissionChecker = checker
}

// Register registers a command
func (r *Registry) Register(cmd Command) {
	name := cmd.Name()
	if _, exists := r.commands[name]; exists {
		slog.Warn("Command already registered, overwriting", "command", name)
	}
	r.commands[name] = cmd
	slog.Debug("Command registered", "command", name)
}

// RegisterAll registers multiple commands
func (r *Registry) RegisterAll(commands ...Command) {
	for _, cmd := range commands {
		r.Register(cmd)
	}
}

// RegisterConfigCommand registers a configuration command with permission control
// If ENABLE_USER_SETTING is false, the command will require admin permission
func (r *Registry) RegisterConfigCommand(cmd Command) {
	if !r.config.EnableUserSetting && r.permissionChecker != nil {
		// Wrap the command with admin-only permission check
		wrappedCmd := NewPermissionAwareCommand(cmd, r.permissionChecker, true)
		r.Register(wrappedCmd)
		slog.Debug("Config command registered with admin-only permission", "command", cmd.Name())
	} else {
		// Register normally
		r.Register(cmd)
		slog.Debug("Config command registered", "command", cmd.Name())
	}
}

// Get retrieves a command by name
func (r *Registry) Get(name string) (Command, bool) {
	cmd, exists := r.commands[name]
	return cmd, exists
}

// Handle processes a command message
func (r *Registry) Handle(message *tgbotapi.Message, ctx *config.WorkerContext) error {
	if !message.IsCommand() {
		return nil
	}

	commandName := message.Command()
	commandArgs := message.CommandArguments()

	slog.Debug("Processing command",
		"command", commandName,
		"args", commandArgs,
		"chat_id", message.Chat.ID,
		"user_id", message.From.ID)

	// Get command handler
	cmd, exists := r.Get(commandName)
	if !exists {
		slog.Debug("Unknown command", "command", commandName)
		return fmt.Errorf("unknown command: /%s", commandName)
	}

	// Check authorization
	authChecker := cmd.NeedAuth()
	if authChecker != nil {
		authorized, err := authChecker.Check(message, ctx)
		if err != nil {
			slog.Error("Authorization check failed", "command", commandName, "error", err)
			return fmt.Errorf("authorization check failed: %w", err)
		}
		if !authorized {
			slog.Warn("Unauthorized command access",
				"command", commandName,
				"chat_id", message.Chat.ID,
				"user_id", message.From.ID)
			return fmt.Errorf("unauthorized: you don't have permission to use /%s", commandName)
		}
	}

	// Execute command
	if err := cmd.Handle(message, commandArgs, ctx); err != nil {
		slog.Error("Command execution failed",
			"command", commandName,
			"error", err)
		return fmt.Errorf("command execution failed: %w", err)
	}

	return nil
}

// GetCommandList returns a list of all registered commands for Telegram
// Commands in HIDE_COMMAND_BUTTONS will be excluded from the list
// If ENABLE_USER_SETTING is false, config commands will be excluded for non-admin users
func (r *Registry) GetCommandList(lang string) []tgbotapi.BotCommand {
	commands := make([]tgbotapi.BotCommand, 0, len(r.commands))

	// Create a map of hidden commands for quick lookup
	hiddenCommands := make(map[string]bool)
	for _, cmdName := range r.config.HideCommandButtons {
		hiddenCommands[cmdName] = true
	}

	for _, cmd := range r.commands {
		cmdName := cmd.Name()

		// Skip hidden commands
		if hiddenCommands[cmdName] {
			slog.Debug("Hiding command button", "command", cmdName)
			continue
		}

		commands = append(commands, tgbotapi.BotCommand{
			Command:     cmdName,
			Description: cmd.Description(lang),
		})
	}

	return commands
}

// GetCommandListForUser returns a list of commands visible to a specific user
// Takes into account user permissions and ENABLE_USER_SETTING configuration
func (r *Registry) GetCommandListForUser(userID int64, chatID int64, lang string, ctx *config.WorkerContext) []tgbotapi.BotCommand {
	commands := make([]tgbotapi.BotCommand, 0, len(r.commands))

	// Create a map of hidden commands for quick lookup
	hiddenCommands := make(map[string]bool)
	for _, cmdName := range r.config.HideCommandButtons {
		hiddenCommands[cmdName] = true
	}

	// If ENABLE_USER_SETTING is false, check if user is admin
	isAdmin := false
	if !r.config.EnableUserSetting && r.permissionChecker != nil {
		var err error
		isAdmin, err = r.permissionChecker.IsAdmin(userID, chatID, ctx)
		if err != nil {
			slog.Error("Failed to check admin permission", "error", err)
			// On error, assume not admin (safer)
			isAdmin = false
		}
	}

	// List of config commands that should be hidden from non-admins when ENABLE_USER_SETTING is false
	configCommands := map[string]bool{
		"setenv":   true,
		"setenvs":  true,
		"delenv":   true,
		"clearenv": true,
		"model":    true,
		"system":   true,
	}

	for _, cmd := range r.commands {
		cmdName := cmd.Name()

		// Skip hidden commands
		if hiddenCommands[cmdName] {
			slog.Debug("Hiding command button", "command", cmdName)
			continue
		}

		// If ENABLE_USER_SETTING is false and user is not admin, hide config commands
		if !r.config.EnableUserSetting && !isAdmin && configCommands[cmdName] {
			slog.Debug("Hiding config command from non-admin user", "command", cmdName, "user_id", userID)
			continue
		}

		commands = append(commands, tgbotapi.BotCommand{
			Command:     cmdName,
			Description: cmd.Description(lang),
		})
	}

	return commands
}

// GetHelpText returns formatted help text for all commands
func (r *Registry) GetHelpText(lang string) string {
	var sb strings.Builder

	for _, cmd := range r.commands {
		sb.WriteString(fmt.Sprintf("/%s - %s\n", cmd.Name(), cmd.Description(lang)))
	}

	return sb.String()
}

// List returns all registered command names
func (r *Registry) List() []string {
	names := make([]string, 0, len(r.commands))
	for name := range r.commands {
		names = append(names, name)
	}
	return names
}

// LoadPlugins loads plugins from environment variables
func (r *Registry) LoadPlugins() error {
	if err := r.pluginRegistry.LoadFromEnvironment(); err != nil {
		return fmt.Errorf("failed to load plugins from environment: %w", err)
	}

	// Register plugin commands
	for _, command := range r.pluginRegistry.ListPlugins() {
		pluginCmd := NewPluginCommand(command, r.pluginRegistry)
		r.Register(pluginCmd)
		slog.Info("Plugin command registered", "command", command)
	}

	return nil
}

// GetPluginRegistry returns the plugin registry
func (r *Registry) GetPluginRegistry() *plugin.PluginRegistry {
	return r.pluginRegistry
}
