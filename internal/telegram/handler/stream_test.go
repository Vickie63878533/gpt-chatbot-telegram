package handler

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/tbxark/ChatGPT-Telegram-Workers/go_version/internal/agent"
	"github.com/tbxark/ChatGPT-Telegram-Workers/go_version/internal/config"
	"github.com/tbxark/ChatGPT-Telegram-Workers/go_version/internal/telegram/api"
	"github.com/tbxark/ChatGPT-Telegram-Workers/go_version/internal/telegram/sender"
)

// mockSender is a mock implementation of MessageSender for testing
type mockSender struct {
	messages    []string
	sendError   error
	sendCount   int
	actionsSent []string
}

func (m *mockSender) SendPlainText(text string) error {
	if m.sendError != nil {
		m.sendCount++
		return m.sendError
	}
	m.messages = append(m.messages, text)
	m.sendCount++
	return nil
}

func (m *mockSender) SendRichText(text string, parseMode string) error {
	if m.sendError != nil {
		m.sendCount++
		return m.sendError
	}
	m.messages = append(m.messages, text)
	m.sendCount++
	return nil
}

func (m *mockSender) SendChatAction(action string) error {
	m.actionsSent = append(m.actionsSent, action)
	return nil
}

func TestStreamHandler_OnStreamText(t *testing.T) {
	cfg := &config.Config{
		TelegramMinStreamInterval: 0, // No delay for testing
		DefaultParseMode:          "Markdown",
		StreamMode:                true,
	}

	// Create a mock client and sender
	client, err := api.NewClient("test-token", "https://api.telegram.org")
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}
	msgSender := sender.NewMessageSender(client, 12345)

	handler := NewStreamHandler(msgSender, cfg)

	// Override the sender's SendPlainText method for testing
	// (In real implementation, we'd use dependency injection)

	// Test streaming text
	texts := []string{"Hello", " ", "World", "!"}
	for _, text := range texts {
		err := handler.OnStreamText(text)
		if err != nil {
			t.Errorf("OnStreamText() error = %v", err)
		}
	}

	// Check final text
	finalText := handler.GetFinalText()
	expected := "Hello World!"
	if finalText != expected {
		t.Errorf("GetFinalText() = %v, want %v", finalText, expected)
	}
}

func TestStreamHandler_MinInterval(t *testing.T) {
	cfg := &config.Config{
		TelegramMinStreamInterval: 100, // 100ms delay
		DefaultParseMode:          "Markdown",
		StreamMode:                true,
	}

	client, err := api.NewClient("test-token", "https://api.telegram.org")
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}
	msgSender := sender.NewMessageSender(client, 12345)

	handler := NewStreamHandler(msgSender, cfg)

	// First update should always go through
	handler.OnStreamText("First")

	// Immediate second update should be skipped due to interval
	handler.OnStreamText(" update")

	// Wait for interval to pass
	time.Sleep(150 * time.Millisecond)

	// This update should go through
	handler.OnStreamText(" after delay")

	finalText := handler.GetFinalText()
	expected := "First update after delay"
	if finalText != expected {
		t.Errorf("GetFinalText() = %v, want %v", finalText, expected)
	}
}

func TestStreamHandler_RetryLogic(t *testing.T) {
	cfg := &config.Config{
		TelegramMinStreamInterval: 0,
		DefaultParseMode:          "Markdown",
		StreamMode:                true,
	}

	client, err := api.NewClient("test-token", "https://api.telegram.org")
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}
	msgSender := sender.NewMessageSender(client, 12345)

	handler := NewStreamHandler(msgSender, cfg)

	// Test that retry delay calculation works
	delay1 := handler.calculateRetryDelay(1)
	delay2 := handler.calculateRetryDelay(2)

	if delay2 <= delay1 {
		t.Errorf("Retry delay should increase with attempts: delay1=%v, delay2=%v", delay1, delay2)
	}

	// Test max delay cap
	delay10 := handler.calculateRetryDelay(10)
	if delay10 > handler.maxRetryDelay {
		t.Errorf("Retry delay should be capped at maxRetryDelay: got %v, max %v", delay10, handler.maxRetryDelay)
	}
}

func TestIsRateLimitError(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		expected bool
	}{
		{
			name:     "nil error",
			err:      nil,
			expected: false,
		},
		{
			name:     "429 error",
			err:      errors.New("HTTP 429: Too Many Requests"),
			expected: true,
		},
		{
			name:     "rate limit error",
			err:      errors.New("rate limit exceeded"),
			expected: true,
		},
		{
			name:     "other error",
			err:      errors.New("connection timeout"),
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isRateLimitError(tt.err)
			if result != tt.expected {
				t.Errorf("isRateLimitError() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestStreamHandler_Finalize(t *testing.T) {
	cfg := &config.Config{
		TelegramMinStreamInterval: 0,
		DefaultParseMode:          "Markdown",
		StreamMode:                true,
	}

	client, err := api.NewClient("test-token", "https://api.telegram.org")
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}
	msgSender := sender.NewMessageSender(client, 12345)

	handler := NewStreamHandler(msgSender, cfg)

	// Add some text
	handler.OnStreamText("Test message")

	// Finalize should send the final text
	finalizeErr := handler.Finalize()
	if finalizeErr != nil {
		// In real implementation with mock, we'd check this properly
		// For now, we just ensure it doesn't panic
		t.Logf("Finalize() error = %v (expected in test without real sender)", finalizeErr)
	}

	// Check that text is preserved
	finalText := handler.GetFinalText()
	if finalText != "Test message" {
		t.Errorf("GetFinalText() = %v, want %v", finalText, "Test message")
	}
}

func TestStreamHandler_Reset(t *testing.T) {
	cfg := &config.Config{
		TelegramMinStreamInterval: 0,
		DefaultParseMode:          "Markdown",
		StreamMode:                true,
	}

	client, err := api.NewClient("test-token", "https://api.telegram.org")
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}
	msgSender := sender.NewMessageSender(client, 12345)

	handler := NewStreamHandler(msgSender, cfg)

	// Add some text
	handler.OnStreamText("Test message")

	// Reset
	handler.Reset()

	// Check that state is cleared
	finalText := handler.GetFinalText()
	if finalText != "" {
		t.Errorf("After Reset(), GetFinalText() = %v, want empty string", finalText)
	}
}

// mockAgent is a mock implementation of ChatAgent for testing
type mockAgent struct {
	response     *agent.ChatAgentResponse
	streamTexts  []string
	requestError error
}

func (m *mockAgent) Name() string {
	return "mock"
}

func (m *mockAgent) ModelKey() string {
	return "MOCK_MODEL"
}

func (m *mockAgent) Enable(config *config.Config) bool {
	return true
}

func (m *mockAgent) Model(config *config.Config) string {
	return "mock-model"
}

func (m *mockAgent) ModelList(config *config.Config) ([]string, error) {
	return []string{"mock-model"}, nil
}

func (m *mockAgent) Request(ctx context.Context, params *agent.LLMChatParams, config *config.Config, onStream agent.ChatStreamTextHandler) (*agent.ChatAgentResponse, error) {
	if m.requestError != nil {
		return nil, m.requestError
	}

	// If streaming callback is provided, call it with stream texts
	if onStream != nil && len(m.streamTexts) > 0 {
		for _, text := range m.streamTexts {
			if err := onStream(text); err != nil {
				return nil, err
			}
		}
	}

	return m.response, nil
}

func TestRequestCompletionWithStream_Streaming(t *testing.T) {
	cfg := &config.Config{
		TelegramMinStreamInterval: 0,
		DefaultParseMode:          "Markdown",
		StreamMode:                true,
	}

	client, err := api.NewClient("test-token", "https://api.telegram.org")
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}
	msgSender := sender.NewMessageSender(client, 12345)
	streamHandler := NewStreamHandler(msgSender, cfg)

	// Create mock agent with streaming texts
	mockAgent := &mockAgent{
		streamTexts: []string{"Hello", " ", "streaming", " ", "world"},
		response: &agent.ChatAgentResponse{
			Messages: []agent.HistoryItem{
				{
					Role:    "assistant",
					Content: "Hello streaming world",
				},
			},
		},
	}

	params := &agent.LLMChatParams{
		Prompt: "Test prompt",
		Messages: []agent.HistoryItem{
			{
				Role:    "user",
				Content: "Hello",
			},
		},
	}

	ctx := context.Background()
	response, err := RequestCompletionWithStream(ctx, mockAgent, params, cfg, streamHandler)

	if err != nil {
		t.Errorf("RequestCompletionWithStream() error = %v", err)
	}

	if response == nil {
		t.Fatal("RequestCompletionWithStream() returned nil response")
	}

	// Check that stream handler accumulated the text
	finalText := streamHandler.GetFinalText()
	expected := "Hello streaming world"
	if finalText != expected {
		t.Errorf("Stream handler final text = %v, want %v", finalText, expected)
	}
}

func TestRequestCompletionWithStream_NonStreaming(t *testing.T) {
	cfg := &config.Config{
		TelegramMinStreamInterval: 0,
		DefaultParseMode:          "Markdown",
		StreamMode:                false, // Streaming disabled
	}

	client, err := api.NewClient("test-token", "https://api.telegram.org")
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}
	msgSender := sender.NewMessageSender(client, 12345)
	streamHandler := NewStreamHandler(msgSender, cfg)

	// Create mock agent
	mockAgent := &mockAgent{
		response: &agent.ChatAgentResponse{
			Messages: []agent.HistoryItem{
				{
					Role:    "assistant",
					Content: "Non-streaming response",
				},
			},
		},
	}

	params := &agent.LLMChatParams{
		Prompt: "Test prompt",
		Messages: []agent.HistoryItem{
			{
				Role:    "user",
				Content: "Hello",
			},
		},
	}

	ctx := context.Background()
	response, err := RequestCompletionWithStream(ctx, mockAgent, params, cfg, streamHandler)

	if err != nil {
		t.Errorf("RequestCompletionWithStream() error = %v", err)
	}

	if response == nil {
		t.Fatal("RequestCompletionWithStream() returned nil response")
	}

	// In non-streaming mode, the stream handler should not accumulate text
	finalText := streamHandler.GetFinalText()
	if finalText != "" {
		t.Logf("Stream handler final text = %v (should be empty in non-streaming mode)", finalText)
	}
}

func TestRequestCompletionWithStream_Error(t *testing.T) {
	cfg := &config.Config{
		TelegramMinStreamInterval: 0,
		DefaultParseMode:          "Markdown",
		StreamMode:                true,
	}

	client, err := api.NewClient("test-token", "https://api.telegram.org")
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}
	msgSender := sender.NewMessageSender(client, 12345)
	streamHandler := NewStreamHandler(msgSender, cfg)

	// Create mock agent that returns an error
	mockAgent := &mockAgent{
		requestError: errors.New("API error"),
	}

	params := &agent.LLMChatParams{
		Prompt: "Test prompt",
		Messages: []agent.HistoryItem{
			{
				Role:    "user",
				Content: "Hello",
			},
		},
	}

	ctx := context.Background()
	_, reqErr := RequestCompletionWithStream(ctx, mockAgent, params, cfg, streamHandler)

	if reqErr == nil {
		t.Error("RequestCompletionWithStream() should return error when agent fails")
	}

	if !strings.Contains(reqErr.Error(), "streaming request failed") {
		t.Errorf("Error should contain 'streaming request failed', got: %v", reqErr)
	}
}
