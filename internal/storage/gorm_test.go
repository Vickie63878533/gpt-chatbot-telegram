package storage

import (
	"os"
	"testing"
	"time"
)

// TestNewStorage_SQLite tests SQLite database initialization
func TestNewStorage_SQLite(t *testing.T) {
	// Create temporary database file
	tmpFile := "./test_gorm.db"
	defer os.Remove(tmpFile)

	storage, err := NewStorage("", tmpFile)
	if err != nil {
		t.Fatalf("Failed to create SQLite storage: %v", err)
	}
	defer storage.Close()

	if storage == nil {
		t.Fatal("Storage should not be nil")
	}
}

// TestNewStorage_DSN_SQLite tests SQLite with DSN format
func TestNewStorage_DSN_SQLite(t *testing.T) {
	tmpFile := "./test_gorm_dsn.db"
	defer os.Remove(tmpFile)

	dsn := "sqlite://" + tmpFile
	storage, err := NewStorage(dsn, "")
	if err != nil {
		t.Fatalf("Failed to create SQLite storage with DSN: %v", err)
	}
	defer storage.Close()

	if storage == nil {
		t.Fatal("Storage should not be nil")
	}
}

// TestNewStorage_InvalidDSN tests error handling for invalid DSN
func TestNewStorage_InvalidDSN(t *testing.T) {
	dsn := "invalid://connection"
	_, err := NewStorage(dsn, "")
	if err == nil {
		t.Fatal("Expected error for invalid DSN, got nil")
	}

	if err.Error() == "" {
		t.Fatal("Error message should not be empty")
	}
}

// TestGORMStorage_ChatHistory tests chat history operations
func TestGORMStorage_ChatHistory(t *testing.T) {
	tmpFile := "./test_history.db"
	defer os.Remove(tmpFile)

	storage, err := NewStorage("", tmpFile)
	if err != nil {
		t.Fatalf("Failed to create storage: %v", err)
	}
	defer storage.Close()

	ctx := &SessionContext{
		ChatID: 123,
		BotID:  456,
	}

	// Test empty history
	history, err := storage.GetChatHistory(ctx)
	if err != nil {
		t.Fatalf("Failed to get empty history: %v", err)
	}
	if len(history) != 0 {
		t.Errorf("Expected empty history, got %d items", len(history))
	}

	// Test save history
	testHistory := []HistoryItem{
		{Role: "user", Content: "Hello"},
		{Role: "assistant", Content: "Hi there!"},
	}
	err = storage.SaveChatHistory(ctx, testHistory)
	if err != nil {
		t.Fatalf("Failed to save history: %v", err)
	}

	// Test retrieve history
	retrieved, err := storage.GetChatHistory(ctx)
	if err != nil {
		t.Fatalf("Failed to retrieve history: %v", err)
	}
	if len(retrieved) != 2 {
		t.Errorf("Expected 2 history items, got %d", len(retrieved))
	}

	// Test delete history
	err = storage.DeleteChatHistory(ctx)
	if err != nil {
		t.Fatalf("Failed to delete history: %v", err)
	}

	// Verify deletion
	history, err = storage.GetChatHistory(ctx)
	if err != nil {
		t.Fatalf("Failed to get history after deletion: %v", err)
	}
	if len(history) != 0 {
		t.Errorf("Expected empty history after deletion, got %d items", len(history))
	}
}

// TestGORMStorage_UserConfig tests user configuration operations
func TestGORMStorage_UserConfig(t *testing.T) {
	tmpFile := "./test_config.db"
	defer os.Remove(tmpFile)

	storage, err := NewStorage("", tmpFile)
	if err != nil {
		t.Fatalf("Failed to create storage: %v", err)
	}
	defer storage.Close()

	ctx := &SessionContext{
		ChatID: 123,
		BotID:  456,
	}

	// Test default config
	config, err := storage.GetUserConfig(ctx)
	if err != nil {
		t.Fatalf("Failed to get default config: %v", err)
	}
	if config.Values == nil {
		t.Error("Config Values map should be initialized")
	}

	// Test save config
	testConfig := &UserConfig{
		DefineKeys: []string{"key1", "key2"},
		Values:     map[string]interface{}{"key1": "value1"},
	}
	err = storage.SaveUserConfig(ctx, testConfig)
	if err != nil {
		t.Fatalf("Failed to save config: %v", err)
	}

	// Test retrieve config
	retrieved, err := storage.GetUserConfig(ctx)
	if err != nil {
		t.Fatalf("Failed to retrieve config: %v", err)
	}
	if len(retrieved.DefineKeys) != 2 {
		t.Errorf("Expected 2 define keys, got %d", len(retrieved.DefineKeys))
	}
}

// TestGORMStorage_GroupAdmins tests group admins caching
func TestGORMStorage_GroupAdmins(t *testing.T) {
	tmpFile := "./test_admins.db"
	defer os.Remove(tmpFile)

	storage, err := NewStorage("", tmpFile)
	if err != nil {
		t.Fatalf("Failed to create storage: %v", err)
	}
	defer storage.Close()

	chatID := int64(789)

	// Test empty admins
	admins, err := storage.GetGroupAdmins(chatID)
	if err != nil {
		t.Fatalf("Failed to get empty admins: %v", err)
	}
	if admins != nil {
		t.Error("Expected nil for non-existent admins")
	}

	// Test save admins
	testAdmins := []ChatMember{
		{User: User{ID: 111}, Status: "administrator"},
		{User: User{ID: 222}, Status: "creator"},
	}
	err = storage.SaveGroupAdmins(chatID, testAdmins, 3600)
	if err != nil {
		t.Fatalf("Failed to save admins: %v", err)
	}

	// Test retrieve admins
	retrieved, err := storage.GetGroupAdmins(chatID)
	if err != nil {
		t.Fatalf("Failed to retrieve admins: %v", err)
	}
	if len(retrieved) != 2 {
		t.Errorf("Expected 2 admins, got %d", len(retrieved))
	}
}

// TestGORMStorage_CleanupExpired tests cleanup of expired data
func TestGORMStorage_CleanupExpired(t *testing.T) {
	tmpFile := "./test_cleanup.db"
	defer os.Remove(tmpFile)

	storage, err := NewStorage("", tmpFile)
	if err != nil {
		t.Fatalf("Failed to create storage: %v", err)
	}
	defer storage.Close()

	chatID := int64(999)

	// Save admins with very short TTL
	testAdmins := []ChatMember{
		{User: User{ID: 333}, Status: "administrator"},
	}
	err = storage.SaveGroupAdmins(chatID, testAdmins, 1) // 1 second TTL
	if err != nil {
		t.Fatalf("Failed to save admins: %v", err)
	}

	// Wait for expiration
	time.Sleep(2 * time.Second)

	// Run cleanup
	err = storage.CleanupExpired()
	if err != nil {
		t.Fatalf("Failed to cleanup: %v", err)
	}

	// Verify admins are gone
	admins, err := storage.GetGroupAdmins(chatID)
	if err != nil {
		t.Fatalf("Failed to get admins after cleanup: %v", err)
	}
	if admins != nil {
		t.Error("Expected nil for expired admins after cleanup")
	}
}

// TestGORMStorage_SessionContext tests session context handling
func TestGORMStorage_SessionContext(t *testing.T) {
	tmpFile := "./test_session.db"
	defer os.Remove(tmpFile)

	storage, err := NewStorage("", tmpFile)
	if err != nil {
		t.Fatalf("Failed to create storage: %v", err)
	}
	defer storage.Close()

	userID := int64(111)
	threadID := int64(222)

	// Test with UserID
	ctx1 := &SessionContext{
		ChatID: 123,
		BotID:  456,
		UserID: &userID,
	}

	history1 := []HistoryItem{{Role: "user", Content: "Test 1"}}
	err = storage.SaveChatHistory(ctx1, history1)
	if err != nil {
		t.Fatalf("Failed to save history with UserID: %v", err)
	}

	// Test with ThreadID
	ctx2 := &SessionContext{
		ChatID:   123,
		BotID:    456,
		ThreadID: &threadID,
	}

	history2 := []HistoryItem{{Role: "user", Content: "Test 2"}}
	err = storage.SaveChatHistory(ctx2, history2)
	if err != nil {
		t.Fatalf("Failed to save history with ThreadID: %v", err)
	}

	// Verify they are separate
	retrieved1, err := storage.GetChatHistory(ctx1)
	if err != nil {
		t.Fatalf("Failed to retrieve history 1: %v", err)
	}

	retrieved2, err := storage.GetChatHistory(ctx2)
	if err != nil {
		t.Fatalf("Failed to retrieve history 2: %v", err)
	}

	if len(retrieved1) != 1 || len(retrieved2) != 1 {
		t.Error("Expected separate histories for different contexts")
	}
}
