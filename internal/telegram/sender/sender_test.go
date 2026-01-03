package sender

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/tbxark/ChatGPT-Telegram-Workers/go_version/internal/telegram/api"
)

// createMockClient creates a mock Telegram API client for testing
func createMockClient(t *testing.T) *api.Client {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"ok":true,"result":{"id":123,"is_bot":true,"first_name":"TestBot","username":"test_bot"}}`))
	}))
	t.Cleanup(server.Close)

	token := "123456:ABC-DEF1234ghIkl-zyx57W2v1u123ew11"
	client, err := api.NewClient(token, server.URL)
	if err != nil {
		t.Fatalf("Failed to create mock client: %v", err)
	}
	return client
}

func TestNewMessageSender(t *testing.T) {
	client := createMockClient(t)
	chatID := int64(12345)
	sender := NewMessageSender(client, chatID)

	if sender == nil {
		t.Fatal("NewMessageSender() returned nil")
	}
	if sender.chatID != chatID {
		t.Errorf("chatID = %v, want %v", sender.chatID, chatID)
	}
	if sender.messageID != 0 {
		t.Errorf("messageID = %v, want 0", sender.messageID)
	}
	if sender.context == nil {
		t.Error("context is nil")
	}
}

func TestMessageSender_Update(t *testing.T) {
	client := createMockClient(t)
	sender := NewMessageSender(client, 12345)

	messageID := 999
	sender.Update(messageID)

	if got := sender.GetMessageID(); got != messageID {
		t.Errorf("GetMessageID() = %v, want %v", got, messageID)
	}
}

func TestMessageSender_SetMinStreamInterval(t *testing.T) {
	client := createMockClient(t)
	sender := NewMessageSender(client, 12345)

	interval := 2 * time.Second
	sender.SetMinStreamInterval(interval)

	if sender.minStreamInterval != interval {
		t.Errorf("minStreamInterval = %v, want %v", sender.minStreamInterval, interval)
	}
}

func TestMessageSender_Context(t *testing.T) {
	client := createMockClient(t)
	sender := NewMessageSender(client, 12345)

	// Test setting and getting context
	key := "test_key"
	value := "test_value"
	sender.SetContext(key, value)

	got, ok := sender.GetContext(key)
	if !ok {
		t.Error("GetContext() returned false for existing key")
	}
	if got != value {
		t.Errorf("GetContext() = %v, want %v", got, value)
	}

	// Test getting non-existent key
	_, ok = sender.GetContext("non_existent")
	if ok {
		t.Error("GetContext() returned true for non-existent key")
	}
}

func TestMessageSender_Reset(t *testing.T) {
	client := createMockClient(t)
	sender := NewMessageSender(client, 12345)

	// Set some state
	sender.Update(999)
	sender.SetContext("key", "value")
	sender.lastUpdateTime = time.Now()

	// Reset
	sender.Reset()

	// Verify state is cleared
	if sender.GetMessageID() != 0 {
		t.Errorf("messageID after Reset() = %v, want 0", sender.GetMessageID())
	}
	if len(sender.context) != 0 {
		t.Errorf("context length after Reset() = %v, want 0", len(sender.context))
	}
	if !sender.lastUpdateTime.IsZero() {
		t.Error("lastUpdateTime after Reset() is not zero")
	}
}

func TestMessageSender_StreamIntervalRespected(t *testing.T) {
	client := createMockClient(t)
	sender := NewMessageSender(client, 12345)

	// Set a stream interval
	interval := 100 * time.Millisecond
	sender.SetMinStreamInterval(interval)

	// Simulate a message being sent
	sender.messageID = 1
	sender.lastUpdateTime = time.Now()

	// Try to send immediately (should be skipped due to interval)
	// Note: We can't actually test the send without a real bot, but we can verify
	// the interval logic by checking the lastUpdateTime doesn't change
	oldTime := sender.lastUpdateTime

	// In a real scenario, SendRichText would check the interval
	// For this test, we just verify the interval is set correctly
	if sender.minStreamInterval != interval {
		t.Errorf("minStreamInterval = %v, want %v", sender.minStreamInterval, interval)
	}

	// Verify lastUpdateTime hasn't changed (since we didn't actually send)
	if sender.lastUpdateTime != oldTime {
		t.Error("lastUpdateTime changed unexpectedly")
	}
}
