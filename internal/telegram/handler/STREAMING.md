# Streaming Response Implementation

This document describes the streaming response implementation for the Telegram bot.

## Overview

The streaming implementation allows the bot to send incremental updates to users as the AI generates responses, providing a better user experience with real-time feedback.

## Components

### StreamHandler

The `StreamHandler` manages streaming responses with the following features:

1. **Text Accumulation**: Accumulates streaming text chunks
2. **Interval Control**: Respects minimum interval between updates (configurable via `TELEGRAM_MIN_STREAM_INTERVAL`)
3. **Retry Logic**: Implements exponential backoff for failed updates
4. **Rate Limit Handling**: Detects and retries on 429 (Too Many Requests) errors
5. **Finalization**: Ensures the final text is sent even if streaming stops

### Key Features

#### Minimum Stream Interval

Configure the minimum time between stream updates to avoid hitting Telegram's rate limits:

```go
cfg := &config.Config{
    TelegramMinStreamInterval: 500, // 500ms between updates
}
```

#### Retry Logic with Exponential Backoff

The handler automatically retries failed updates with exponential backoff:

- Initial retry delay: 1 second
- Backoff factor: 2.0
- Maximum retry delay: 30 seconds
- Maximum retries: 3

#### Rate Limit Detection

The handler detects rate limit errors (429) and automatically retries:

```go
func isRateLimitError(err error) bool {
    errStr := err.Error()
    return strings.Contains(errStr, "429") || 
           strings.Contains(errStr, "Too Many Requests") ||
           strings.Contains(errStr, "rate limit")
}
```

## Usage Example

### Basic Streaming

```go
import (
    "context"
    "github.com/tbxark/ChatGPT-Telegram-Workers/go_version/internal/agent"
    "github.com/tbxark/ChatGPT-Telegram-Workers/go_version/internal/config"
    "github.com/tbxark/ChatGPT-Telegram-Workers/go_version/internal/telegram/handler"
    "github.com/tbxark/ChatGPT-Telegram-Workers/go_version/internal/telegram/sender"
)

// Create a message sender
msgSender := sender.NewMessageSender(client, chatID)

// Create a stream handler
streamHandler := handler.NewStreamHandler(msgSender, cfg)

// Set parse mode if needed
streamHandler.SetParseMode("Markdown")

// Load the AI agent
chatAgent := agent.LoadChatLLM(cfg, userConfig)

// Prepare the request parameters
params := &agent.LLMChatParams{
    Prompt:   "You are a helpful assistant",
    Messages: historyItems,
}

// Request completion with streaming
ctx := context.Background()
response, err := handler.RequestCompletionWithStream(
    ctx,
    chatAgent,
    params,
    cfg,
    streamHandler,
)

if err != nil {
    // Handle error
    return err
}

// Get the final text
finalText := streamHandler.GetFinalText()
```

### Manual Streaming Control

For more control over the streaming process:

```go
// Create stream handler
streamHandler := handler.NewStreamHandler(msgSender, cfg)

// Send typing action
msgSender.SendChatAction("typing")

// Create streaming callback
streamCallback := func(text string) error {
    return streamHandler.OnStreamText(text)
}

// Request with streaming
response, err := chatAgent.Request(ctx, params, cfg, streamCallback)
if err != nil {
    return err
}

// Finalize the stream (send any pending updates)
if err := streamHandler.Finalize(); err != nil {
    log.Printf("Failed to finalize stream: %v", err)
}

// Get final text
finalText := streamHandler.GetFinalText()
```

### Non-Streaming Mode

When `STREAM_MODE` is disabled, the same code works without streaming:

```go
cfg := &config.Config{
    StreamMode: false, // Disable streaming
}

// Same code as above - will send a single message instead of streaming
response, err := handler.RequestCompletionWithStream(
    ctx,
    chatAgent,
    params,
    cfg,
    streamHandler,
)
```

## Configuration

### Environment Variables

- `STREAM_MODE`: Enable/disable streaming (default: `true`)
- `TELEGRAM_MIN_STREAM_INTERVAL`: Minimum milliseconds between stream updates (default: `0`)

### Example Configuration

```bash
# Enable streaming with 500ms minimum interval
export STREAM_MODE=true
export TELEGRAM_MIN_STREAM_INTERVAL=500

# Disable streaming
export STREAM_MODE=false
```

## Error Handling

The streaming implementation handles errors gracefully:

1. **Network Errors**: Logged but don't stop the stream
2. **Rate Limit Errors (429)**: Automatically retried with backoff
3. **Other Errors**: Logged and returned to caller

### Error Logging

All errors are logged with structured logging:

```go
slog.Warn("Failed to send stream update", "error", err)
slog.Warn("Rate limit hit, will retry", "attempt", attempt, "error", err)
slog.Warn("Failed to finalize stream", "error", err)
```

## Testing

The implementation includes comprehensive tests:

- `TestStreamHandler_OnStreamText`: Text accumulation
- `TestStreamHandler_MinInterval`: Interval enforcement
- `TestStreamHandler_RetryLogic`: Exponential backoff
- `TestIsRateLimitError`: Rate limit detection
- `TestStreamHandler_Finalize`: Finalization
- `TestStreamHandler_Reset`: State reset
- `TestRequestCompletionWithStream_Streaming`: Streaming mode
- `TestRequestCompletionWithStream_NonStreaming`: Non-streaming mode
- `TestRequestCompletionWithStream_Error`: Error handling

Run tests:

```bash
go test -v ./internal/telegram/handler -run "TestStreamHandler|TestRequestCompletion"
```

## Performance Considerations

1. **Rate Limits**: Use `TELEGRAM_MIN_STREAM_INTERVAL` to avoid hitting Telegram's rate limits
2. **Message Updates**: Telegram has limits on how frequently you can edit messages
3. **Retry Delays**: Exponential backoff prevents overwhelming the API during issues
4. **Buffer Management**: Text is accumulated in memory - consider limits for very long responses

## Future Enhancements

Potential improvements:

1. Configurable retry parameters (max retries, backoff factor)
2. Adaptive interval based on response time
3. Metrics collection (update count, retry count, etc.)
4. Circuit breaker pattern for persistent failures
5. Batch updates for very rapid streaming

