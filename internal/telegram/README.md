# Telegram Package

This package provides Telegram Bot API integration for the Go version of the ChatGPT Telegram Bot.

## Structure

```
telegram/
├── api/          # Telegram API client wrapper
│   ├── client.go      # Client implementation with custom API domain support
│   └── client_test.go # Unit tests
└── sender/       # Message sender with streaming support
    ├── sender.go      # MessageSender implementation
    └── sender_test.go # Unit tests
```

## Components

### API Client (`api/client.go`)

The API client wraps the `go-telegram-bot-api` library and provides:

- **Custom API Domain Support**: Configure a custom Telegram API endpoint (useful for proxies or self-hosted instances)
- **HTTP Client Configuration**: Use custom HTTP clients with specific timeouts and transport settings
- **Convenience Methods**: Simplified methods for common operations like setting webhooks, getting chat administrators, etc.

#### Usage

```go
import "github.com/tbxark/ChatGPT-Telegram-Workers/go_version/internal/telegram/api"

// Create a client with default API domain
client, err := api.NewClient("YOUR_BOT_TOKEN", "")

// Create a client with custom API domain
client, err := api.NewClient("YOUR_BOT_TOKEN", "https://api.custom.com")

// Create a client with custom HTTP client
httpClient := api.NewDefaultHTTPClient()
client, err := api.NewClientWithHTTPClient("YOUR_BOT_TOKEN", "", httpClient)

// Use the client
me, err := client.GetMe()
err = client.SetWebhook("https://example.com/webhook")
```

#### Key Features

- **Token Validation**: Validates that a token is provided
- **API Endpoint Configuration**: Supports custom API domains via `TELEGRAM_API_DOMAIN` environment variable
- **Debug Mode**: Enable/disable debug logging
- **Webhook Management**: Set and remove webhooks
- **Command Management**: Set bot command list
- **Chat Actions**: Send typing indicators, upload_photo, etc.
- **File Operations**: Get direct URLs for files
- **Callback Queries**: Answer callback queries from inline keyboards

### Message Sender (`sender/sender.go`)

The MessageSender handles sending messages with support for streaming updates (editing the same message multiple times).

#### Features

- **Streaming Support**: Update the same message multiple times for real-time AI responses
- **Stream Interval Control**: Respect minimum intervals between updates to avoid rate limiting
- **Multiple Message Types**: Send plain text, rich text (Markdown/HTML), photos, and messages with keyboards
- **Context Storage**: Store arbitrary context data with the sender
- **Thread-Safe**: All operations are protected by mutexes

#### Usage

```go
import "github.com/tbxark/ChatGPT-Telegram-Workers/go_version/internal/telegram/sender"

// Create a sender
sender := sender.NewMessageSender(client, chatID)

// Configure streaming
sender.SetMinStreamInterval(2 * time.Second)

// Send a message (creates new message)
err := sender.SendPlainText("Hello!")

// Update the message (edits existing message)
sender.Update(messageID)
err = sender.SendPlainText("Hello, updated!")

// Send with formatting
err = sender.SendRichText("**Bold** text", "Markdown")

// Send a photo
err = sender.SendPhoto("https://example.com/image.jpg")

// Send with inline keyboard
keyboard := tgbotapi.NewInlineKeyboardMarkup(
    tgbotapi.NewInlineKeyboardRow(
        tgbotapi.NewInlineKeyboardButtonData("Button", "callback_data"),
    ),
)
err = sender.SendMessageWithKeyboard("Choose:", keyboard, "")

// Send chat action
err = sender.SendChatAction("typing")

// Reset sender state
sender.Reset()
```

#### Streaming Workflow

The streaming feature is designed for real-time AI responses:

1. **First Call**: `SendPlainText()` or `SendRichText()` creates a new message
2. **Subsequent Calls**: After calling `Update(messageID)`, the same methods edit the existing message
3. **Interval Control**: If `SetMinStreamInterval()` is configured, updates are throttled to respect the minimum interval
4. **Reset**: Call `Reset()` to clear the message ID and start fresh

Example streaming workflow:

```go
sender := sender.NewMessageSender(client, chatID)
sender.SetMinStreamInterval(1 * time.Second)

// First message
sender.SendPlainText("Thinking...")

// Stream updates as AI generates response
for chunk := range aiResponseStream {
    sender.SendPlainText(chunk) // Edits the same message
}
```

## Requirements Validation

This implementation satisfies the following requirements:

- **Requirement 2.1**: Telegram Bot integration for receiving and processing messages
- **Requirement 2.3**: Streaming response support (STREAM_MODE)
- **Requirement 13.1**: Stream mode behavior with message updates
- **Requirement 13.2**: Minimum stream interval enforcement (TELEGRAM_MIN_STREAM_INTERVAL)
- **Requirement 13.3**: Stream message reuse (editing same message)
- **Requirement 16.5**: Custom API domain support (TELEGRAM_API_DOMAIN)

## Testing

Both packages include comprehensive unit tests:

```bash
# Run all telegram package tests
go test ./internal/telegram/...

# Run with verbose output
go test -v ./internal/telegram/...

# Run specific package tests
go test ./internal/telegram/api/...
go test ./internal/telegram/sender/...
```

The tests use mock HTTP servers to avoid making real API calls during testing.

## Dependencies

- `github.com/go-telegram-bot-api/telegram-bot-api/v5`: Official Go bindings for Telegram Bot API

## Future Enhancements

Planned features for future implementation:

- **Handler Package**: Message and update handlers (task 8)
- **Command Package**: Command system implementation (task 9)
- **Chat Package**: Chat message processing (task 14)
- **Callback Query Package**: Callback query handling (task 17)
