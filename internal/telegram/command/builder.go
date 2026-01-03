package command

import (
	"log/slog"

	"github.com/tbxark/ChatGPT-Telegram-Workers/go_version/internal/config"
	"github.com/tbxark/ChatGPT-Telegram-Workers/go_version/internal/i18n"
)

// BuildCommandRegistry creates and registers all commands
func BuildCommandRegistry(cfg *config.Config, i18n *i18n.I18n) *Registry {
	registry := NewRegistry(cfg)

	// Register basic commands
	registry.Register(NewStartCommand(cfg, i18n))
	registry.Register(NewNewCommand(cfg, i18n))
	registry.Register(NewRedoCommand(cfg, i18n))

	// Register help command (needs registry reference)
	helpCmd := NewHelpCommand(cfg, i18n, registry)
	registry.Register(helpCmd)

	// Register configuration commands
	registry.Register(NewSetenvCommand(cfg, i18n))
	registry.Register(NewSetenvsCommand(cfg, i18n))
	registry.Register(NewDelenvCommand(cfg, i18n))
	registry.Register(NewClearenvCommand(cfg, i18n))

	// Register system/debug commands
	registry.Register(NewSystemCommand(cfg, i18n))
	registry.Register(NewEchoCommand(cfg, i18n))

	// Register image commands
	registry.Register(NewImgCommand(cfg, i18n))
	registry.Register(NewModelsCommand(cfg, i18n))

	// Load and register plugin commands
	if err := registry.LoadPlugins(); err != nil {
		slog.Warn("Failed to load plugins", "error", err)
	}

	return registry
}
