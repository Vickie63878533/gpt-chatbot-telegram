package handler

import (
	"context"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/tbxark/ChatGPT-Telegram-Workers/go_version/internal/agent"
	"github.com/tbxark/ChatGPT-Telegram-Workers/go_version/internal/config"
	"github.com/tbxark/ChatGPT-Telegram-Workers/go_version/internal/telegram/sender"
)

// StreamHandler handles streaming responses from AI agents
type StreamHandler struct {
	sender             *sender.MessageSender
	config             *config.Config
	buffer             strings.Builder
	lastSentText       string
	minInterval        time.Duration
	lastUpdateTime     time.Time
	parseMode          string
	retryCount         int
	maxRetries         int
	retryDelay         time.Duration
	maxRetryDelay      time.Duration
	retryBackoffFactor float64
}

// NewStreamHandler creates a new StreamHandler
func NewStreamHandler(sender *sender.MessageSender, cfg *config.Config) *StreamHandler {
	minInterval := time.Duration(cfg.TelegramMinStreamInterval) * time.Millisecond

	return &StreamHandler{
		sender:             sender,
		config:             cfg,
		minInterval:        minInterval,
		parseMode:          cfg.DefaultParseMode,
		retryCount:         0,
		maxRetries:         3,
		retryDelay:         time.Second,
		maxRetryDelay:      30 * time.Second,
		retryBackoffFactor: 2.0,
	}
}

// SetParseMode sets the parse mode for messages
func (h *StreamHandler) SetParseMode(mode string) {
	h.parseMode = mode
}

// OnStreamText is the callback function for streaming text
// It accumulates text and sends updates respecting the minimum interval
func (h *StreamHandler) OnStreamText(text string) error {
	// Append to buffer
	h.buffer.WriteString(text)
	currentText := h.buffer.String()

	// Check if we should send an update
	if h.shouldSendUpdate(currentText) {
		if err := h.sendUpdate(currentText); err != nil {
			// Log error but don't stop streaming
			slog.Warn("Failed to send stream update", "error", err)
			// Don't return error to avoid stopping the stream
		}
	}

	return nil
}

// shouldSendUpdate determines if we should send an update based on timing and content
func (h *StreamHandler) shouldSendUpdate(currentText string) bool {
	// Always send if this is the first update
	if h.lastSentText == "" {
		return true
	}

	// Don't send if text hasn't changed
	if currentText == h.lastSentText {
		return false
	}

	// Check minimum interval
	if h.minInterval > 0 {
		elapsed := time.Since(h.lastUpdateTime)
		if elapsed < h.minInterval {
			return false
		}
	}

	return true
}

// sendUpdate sends the current text to Telegram with retry logic
func (h *StreamHandler) sendUpdate(text string) error {
	var err error

	for attempt := 0; attempt <= h.maxRetries; attempt++ {
		if attempt > 0 {
			// Calculate exponential backoff delay
			delay := h.calculateRetryDelay(attempt)
			slog.Debug("Retrying stream update", "attempt", attempt, "delay", delay)
			time.Sleep(delay)
		}

		// Try to send the update
		if h.parseMode != "" {
			err = h.sender.SendRichText(text, h.parseMode)
		} else {
			err = h.sender.SendPlainText(text)
		}

		if err == nil {
			// Success
			h.lastSentText = text
			h.lastUpdateTime = time.Now()
			h.retryCount = 0
			return nil
		}

		// Check if it's a rate limit error (429)
		if isRateLimitError(err) {
			slog.Warn("Rate limit hit, will retry", "attempt", attempt, "error", err)
			continue
		}

		// For other errors, log and return
		slog.Warn("Failed to send stream update", "attempt", attempt, "error", err)
		return err
	}

	// Max retries exceeded
	return fmt.Errorf("max retries exceeded: %w", err)
}

// calculateRetryDelay calculates the delay for the next retry using exponential backoff
func (h *StreamHandler) calculateRetryDelay(attempt int) time.Duration {
	delay := float64(h.retryDelay) * float64(attempt) * h.retryBackoffFactor

	if delay > float64(h.maxRetryDelay) {
		delay = float64(h.maxRetryDelay)
	}

	return time.Duration(delay)
}

// Finalize sends the final text if it hasn't been sent yet
func (h *StreamHandler) Finalize() error {
	finalText := h.buffer.String()

	// If we haven't sent anything yet, or the text has changed, send it
	if finalText != h.lastSentText && finalText != "" {
		return h.sendUpdate(finalText)
	}

	return nil
}

// GetFinalText returns the accumulated text
func (h *StreamHandler) GetFinalText() string {
	return h.buffer.String()
}

// Reset resets the stream handler state
func (h *StreamHandler) Reset() {
	h.buffer.Reset()
	h.lastSentText = ""
	h.lastUpdateTime = time.Time{}
	h.retryCount = 0
}

// isRateLimitError checks if an error is a rate limit error (429)
func isRateLimitError(err error) bool {
	if err == nil {
		return false
	}

	// Check if the error message contains "429" or "Too Many Requests"
	errStr := err.Error()
	return strings.Contains(errStr, "429") ||
		strings.Contains(errStr, "Too Many Requests") ||
		strings.Contains(errStr, "rate limit")
}

// RequestCompletionWithStream requests a completion from an AI agent with streaming support
func RequestCompletionWithStream(
	ctx context.Context,
	agent agent.ChatAgent,
	params *agent.LLMChatParams,
	cfg *config.Config,
	streamHandler *StreamHandler,
) (*agent.ChatAgentResponse, error) {
	// If streaming is enabled and we have a stream handler, use it
	if cfg.StreamMode && streamHandler != nil {
		// Send typing action
		if err := streamHandler.sender.SendChatAction("typing"); err != nil {
			slog.Warn("Failed to send typing action", "error", err)
		}

		// Create the stream callback
		streamCallback := func(text string) error {
			return streamHandler.OnStreamText(text)
		}

		// Request with streaming
		response, err := agent.Request(ctx, params, cfg, streamCallback)
		if err != nil {
			return nil, fmt.Errorf("streaming request failed: %w", err)
		}

		// Finalize the stream (send any pending updates)
		if err := streamHandler.Finalize(); err != nil {
			slog.Warn("Failed to finalize stream", "error", err)
		}

		return response, nil
	}

	// Non-streaming mode
	response, err := agent.Request(ctx, params, cfg, nil)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}

	return response, nil
}
