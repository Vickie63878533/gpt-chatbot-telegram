# Telegram Handler Package

This package implements the Update and Message handler chains for processing Telegram bot updates.

## Architecture

The handler system follows a chain-of-responsibility pattern with two main chains:

### Update Handler Chain

Processes Telegram updates in the following order:

1. **EnvChecker** - Verifies that required environment variables (DATABASE) are configured
2. **WhiteListFilter** - Filters updates based on whitelist configuration
3. **Update2MessageHandler** - Converts updates to messages and delegates to message handlers
4. **CallbackQueryHandler** - Processes callback queries from inline keyboards

### Message Handler Chain

Processes Telegram messages in the following order:

1. **SaveLastMessage** - Saves the last message for debugging (when DEBUG_MODE is enabled)
2. **OldMessageFilter** - Filters duplicate/old messages in SAFE_MODE
3. **MessageFilter** - Filters unsupported message types (only text, photo, and caption are supported)
4. **CommandHandler** - Processes bot commands (to be implemented in task 9)
5. **ChatHandler** - Processes chat messages (to be implemented in task 14)

## Usage

```go
import (
    "github.com/tbxark/ChatGPT-Telegram-Workers/go_version/internal/telegram/handler"
    "github.com/tbxark/ChatGPT-Telegram-Workers/go_version/internal/config"
    "github.com/tbxark/ChatGPT-Telegram-Workers/go_version/internal/i18n"
)

// Build the complete handler chain
cfg := &config.Config{...}
i18n := i18n.LoadI18n("en")
updateChain := handler.BuildUpdateHandlerChain(cfg, i18n)

// Process an update
ctx := &config.WorkerContext{...}
err := updateChain.Handle(update, ctx)
```

## Key Features

### Whitelist Filtering

- Supports private chat whitelist (CHAT_WHITE_LIST)
- Supports group chat whitelist (CHAT_GROUP_WHITE_LIST)
- Supports generous mode (I_AM_A_GENEROUS_PERSON) to allow all users

### Safe Mode

When SAFE_MODE is enabled:
- Tracks the last 100 message IDs per session
- Rejects duplicate messages to prevent replay attacks
- Stores message IDs in SQLite for persistence

### Message Type Support

Supported message types:
- Text messages
- Photo messages (with or without caption)
- Commands (starting with /)

Unsupported message types are rejected with an error.

### Edited Message Filtering

Edited messages are automatically ignored per Requirement 2.11.

## Session Context

The handler uses `SessionContext` to identify unique chat sessions:

- **Private chats**: `chat_id + bot_id`
- **Group shared mode**: `chat_id + bot_id`
- **Group non-shared mode**: `chat_id + bot_id + user_id`
- **Forum/topic mode**: `chat_id + bot_id + (user_id) + thread_id`

## Testing

Run tests with:

```bash
go test ./internal/telegram/handler/... -v
```

The test suite includes:
- EnvChecker validation
- Whitelist filtering (generous mode and strict mode)
- Edited message filtering
- Message type filtering
- Safe mode duplicate detection
- Handler chain construction

## Requirements Validated

This implementation validates the following requirements:

- **Requirement 2.11**: Edited messages are ignored
- **Requirement 2.12**: Duplicate messages are filtered in SAFE_MODE
- **Requirement 8.1**: Whitelist access control
- **Requirement 8.2**: Open mode access (generous mode)
- **Requirement 12.1**: Update handler chain order
- **Requirement 12.2**: Callback query handling
- **Requirement 12.3**: Message handler chain order
- **Requirement 12.4**: Unsupported message type rejection

## Future Work

The following handlers are placeholders and will be implemented in future tasks:

- **CommandHandler** (Task 9): Process bot commands like /start, /help, /new, etc.
- **ChatHandler** (Task 14): Process chat messages and integrate with AI agents
- **CallbackQueryHandler** (Task 17): Process callback queries for model switching

## Dependencies

- `github.com/go-telegram-bot-api/telegram-bot-api/v5` - Telegram Bot API
- `internal/config` - Configuration management
- `internal/storage` - Data persistence
- `internal/i18n` - Internationalization
