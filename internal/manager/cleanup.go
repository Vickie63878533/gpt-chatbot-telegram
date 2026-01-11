package manager

import (
	"log"
	"time"

	"github.com/tbxark/ChatGPT-Telegram-Workers/go_version/internal/storage"
)

// TokenCleanupTask handles periodic cleanup of expired login tokens
type TokenCleanupTask struct {
	storage  storage.Storage
	interval time.Duration
	done     chan bool
}

// NewTokenCleanupTask creates a new token cleanup task
func NewTokenCleanupTask(db storage.Storage, interval time.Duration) *TokenCleanupTask {
	return &TokenCleanupTask{
		storage:  db,
		interval: interval,
		done:     make(chan bool),
	}
}

// Start begins the periodic cleanup task
func (t *TokenCleanupTask) Start() {
	go func() {
		ticker := time.NewTicker(t.interval)
		defer ticker.Stop()

		// Run cleanup immediately on start
		t.cleanup()

		for {
			select {
			case <-ticker.C:
				t.cleanup()
			case <-t.done:
				log.Println("Token cleanup task stopped")
				return
			}
		}
	}()

	log.Printf("Token cleanup task started with interval %v", t.interval)
}

// Stop stops the cleanup task
func (t *TokenCleanupTask) Stop() {
	t.done <- true
}

// cleanup performs the actual cleanup of expired tokens
func (t *TokenCleanupTask) cleanup() {
	err := t.storage.CleanupExpiredTokens()
	if err != nil {
		log.Printf("Error cleaning up expired tokens: %v", err)
		return
	}

	log.Println("Successfully cleaned up expired login tokens")
}
