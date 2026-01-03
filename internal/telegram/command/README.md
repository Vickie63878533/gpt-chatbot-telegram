# Command Package

This package implements the command system for the Telegram bot, providing a flexible and extensible framework for handling bot commands.

## Architecture

### Command Interface

The `Command` interface defines the contract for all bot commands:

```go
type Command interface {
    Name() string                                                    // Command name (without /)
    Description(lang string) string                                  // Localized description
    Scopes() []string                                               // Command scopes
    NeedAuth() AuthChecker                                          // Authorization checker
    Handle(message *tgbotapi.Message, args string, ctx *config.WorkerContext) error
}
```

### Authorization

Commands can specify authorization requirements using `AuthChecker`:

- `NoAuthRequired`: No authorization needed (default)
- `AdminOnly`: Requires admin permissions in groups
- `ShareModeGroup`: Requires admin permissions in groups (for config commands)

### Registry

The `Registry` manages command registration and dispatching:

- Registers commands by name
- Handles command routing
- Performs authorization checks
- Generates command lists for Telegram

## Implemented Commands

### Basic Commands

- `/start` - Show welcome message and chat ID, start new conversation
- `/help` - Display help text with all available commands
- `/new` - Start a new conversation (clear history)

### Configuration Commands

- `/setenv KEY=VALUE` - Set a user configuration value
- `/setenvs {"KEY1":"VALUE1","KEY2":"VALUE2"}` - Batch set configuration values
- `/delenv KEY` - Delete a user configuration value
- `/clearenv` - Clear all user configuration (preserves locked keys)

### System/Debug Commands

- `/system` - Display system information (runtime, configuration)
- `/echo` - Echo the message in JSON format (for debugging)

## Usage

### Building the Registry

```go
import (
    "github.com/tbxark/ChatGPT-Telegram-Workers/go_version/internal/telegram/command"
)

// Create and populate the command registry
registry := command.BuildCommandRegistry(cfg, i18n)
```

### Integrating with Handler

```go
// Set the registry in the command handler
cmdHandler := handler.NewCommandHandler(cfg)
cmdHandler.SetRegistry(registry)
```

### Adding New Commands

1. Create a new command struct implementing the `Command` interface
2. Implement all required methods
3. Register the command in `builder.go`:

```go
func BuildCommandRegistry(cfg *config.Config, i18n *i18n.I18n) *Registry {
    registry := NewRegistry(cfg)
    
    // ... existing commands ...
    
    // Register your new command
    registry.Register(NewYourCommand(cfg, i18n))
    
    return registry
}
```

## Authorization Flow

1. User sends a command message
2. Registry extracts command name and arguments
3. Registry looks up the command handler
4. Registry checks authorization using the command's `AuthChecker`
5. If authorized, command's `Handle` method is called
6. Command processes the request and sends response

## Configuration Management

Configuration commands allow users to customize bot behavior per-session:

- User configs are stored in SQLite per session
- Locked keys (defined in `LOCK_USER_CONFIG_KEYS`) cannot be modified
- User configs override global configuration
- Configs are loaded/saved using `WorkerContext` methods

## Future Enhancements

The following commands will be implemented in later tasks:

- `/redo` - Regenerate the last response (Task 16)
- `/img` - Generate images (Task 15)
- `/models` - Switch AI models (Task 15)

## Testing

The command system integrates with the existing handler test suite. Mock implementations can be created for testing:

```go
type mockCommandRegistry struct{}

func (m *mockCommandRegistry) Handle(message *tgbotapi.Message, ctx *config.WorkerContext) error {
    return nil
}
```

## Dependencies

- `github.com/go-telegram-bot-api/telegram-bot-api/v5` - Telegram Bot API
- `internal/config` - Configuration management
- `internal/i18n` - Internationalization
- `internal/storage` - Data persistence
