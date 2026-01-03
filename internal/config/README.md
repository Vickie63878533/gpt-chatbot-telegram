# Configuration Management Package

This package provides configuration management for the Go Telegram Bot.

## Overview

The config package handles:
- Loading configuration from environment variables
- Validating configuration values
- Managing user-specific configuration overrides
- Creating session contexts for storage operations

## Components

### Config Structure

The `Config` struct contains all bot configuration loaded from environment variables:

- **AI Provider Settings**: OpenAI, Azure, Gemini, Anthropic, Workers AI, Mistral, Cohere, DeepSeek, Groq, XAI
- **Telegram Settings**: API domain, tokens, parse mode, stream interval, etc.
- **Permission Settings**: Whitelists, locked config keys
- **Group Settings**: Bot name, group whitelist, share mode
- **History Settings**: Auto-trim, max length, image placeholder
- **Feature Switches**: Command buttons, reply button, extra context
- **Mode Switches**: Stream mode, safe mode, debug mode, dev mode
- **Server Settings**: Port, database path
- **Version Information**: Build timestamp and version

### ShareContext

Contains shared context for a bot token:
- `BotToken`: The Telegram bot token
- `BotID`: Extracted bot ID from token
- `ChatHistoryKey`: Key prefix for chat history
- `ConfigStoreKey`: Key prefix for user config
- `LastMessageKey`: Key prefix for last message (debug mode)

### WorkerContext

Contains the full context for processing a request:
- `ShareContext`: Shared bot context
- `UserConfig`: User-specific configuration overrides
- `DB`: Storage interface

### SessionContext

Represents the context for a chat session, used to generate storage keys:
- `ChatID`: Telegram chat ID
- `BotID`: Bot ID
- `UserID`: User ID (for non-shared group mode)
- `ThreadID`: Thread/topic ID (for forum chats)

## Usage

### Loading Configuration

```go
cfg, err := config.LoadConfig()
if err != nil {
    log.Fatalf("Failed to load config: %v", err)
}
```

### Creating Share Context

```go
shareCtx, err := config.NewShareContext(botToken)
if err != nil {
    log.Fatalf("Invalid bot token: %v", err)
}
```

### Creating Worker Context

```go
workerCtx := config.NewWorkerContext(shareCtx, db)

// Load user config from storage
sessionCtx := config.NewSessionContextFromChat(chatID, botID, isGroup, shareMode, &userID, nil)
err := workerCtx.LoadUserConfig(sessionCtx)
```

### Managing User Configuration

```go
// Set a config value
err := workerCtx.SetUserConfigValue("AI_PROVIDER", "azure", cfg.LockUserConfigKeys)

// Get a config value (checks user config first, then global)
provider := workerCtx.GetConfigString("AI_PROVIDER", cfg)

// Delete a config value
err := workerCtx.DeleteUserConfigValue("AI_PROVIDER", cfg.LockUserConfigKeys)

// Clear all user config (except locked keys)
err := workerCtx.ClearUserConfig(cfg.LockUserConfigKeys)

// Save user config to storage
err := workerCtx.SaveUserConfig(sessionCtx)
```

### Creating Session Context

```go
// For private chat
sessionCtx := config.NewSessionContextFromChat(chatID, botID, false, false, nil, nil)

// For group chat (shared mode)
sessionCtx := config.NewSessionContextFromChat(chatID, botID, true, true, nil, nil)

// For group chat (non-shared mode)
sessionCtx := config.NewSessionContextFromChat(chatID, botID, true, false, &userID, nil)

// For forum/topic chat
sessionCtx := config.NewSessionContextFromChat(chatID, botID, true, true, nil, &threadID)
```

## Environment Variables

### Required

- `TELEGRAM_AVAILABLE_TOKENS`: Comma-separated list of bot tokens

### Optional (with defaults)

See the `Config` struct for all available environment variables and their default values.

## Validation

The `Validate()` function checks:
- Required fields are present
- Port is in valid range (1-65535)
- Parse mode is "Markdown" or "HTML"
- Image transfer mode is "url" or "base64"
- Language is one of: zh-cn, en, pt, zh-hant

## Testing

Run tests with:

```bash
go test ./internal/config/...
```

Tests cover:
- Configuration loading and validation
- Share context creation
- Session context generation
- User config operations (set, delete, clear)
- Config value retrieval and merging
