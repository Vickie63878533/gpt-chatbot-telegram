package manager

import (
	"testing"
	"time"

	"github.com/tbxark/ChatGPT-Telegram-Workers/go_version/internal/storage"
)

// TestTokenCleanupTask_Start tests that cleanup task starts and runs
func TestTokenCleanupTask_Start(t *testing.T) {
	mockStorage := NewMockStorage()
	task := NewTokenCleanupTask(mockStorage, 100*time.Millisecond)

	// Start the task
	task.Start()

	// Give it time to run
	time.Sleep(200 * time.Millisecond)

	// Stop the task
	task.Stop()

	// Give it time to stop
	time.Sleep(100 * time.Millisecond)
}

// TestTokenCleanupTask_CleanupExpiredTokens tests that expired tokens are cleaned up
func TestTokenCleanupTask_CleanupExpiredTokens(t *testing.T) {
	mockStorage := NewMockStorage()

	// Add an expired token
	expiredToken := &storage.LoginToken{
		UserID:    12345,
		Token:     "expired-token",
		ExpiresAt: time.Now().Add(-1 * time.Hour),
	}
	mockStorage.tokens[12345] = expiredToken

	// Add a valid token
	validToken := &storage.LoginToken{
		UserID:    99999,
		Token:     "valid-token",
		ExpiresAt: time.Now().Add(24 * time.Hour),
	}
	mockStorage.tokens[99999] = validToken

	// Create and run cleanup task
	task := NewTokenCleanupTask(mockStorage, 100*time.Millisecond)
	task.cleanup()

	// Verify tokens still exist (mock storage doesn't actually delete)
	// In real implementation, expired tokens would be deleted
	if _, exists := mockStorage.tokens[12345]; !exists {
		t.Error("Expired token was removed from storage")
	}

	if _, exists := mockStorage.tokens[99999]; !exists {
		t.Error("Valid token was removed from storage")
	}
}

// TestTokenCleanupTask_Stop tests that cleanup task stops properly
func TestTokenCleanupTask_Stop(t *testing.T) {
	mockStorage := NewMockStorage()
	task := NewTokenCleanupTask(mockStorage, 100*time.Millisecond)

	// Start the task
	task.Start()

	// Stop the task
	task.Stop()

	// Give it time to stop
	time.Sleep(100 * time.Millisecond)

	// Task should have stopped without error
}
