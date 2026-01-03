package storage

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/glebarez/sqlite" // Pure Go SQLite driver
	"gorm.io/driver/mysql"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// GORMStorage implements Storage interface using GORM
// All queries use GORM's parameterized query system to prevent SQL injection
type GORMStorage struct {
	db *gorm.DB
}

// NewStorage creates a new storage instance based on DSN or DB_PATH
// DSN takes priority over DB_PATH for backward compatibility
// All database operations use GORM's safe parameterized queries
func NewStorage(dsn string, dbPath string) (Storage, error) {
	var db *gorm.DB
	var err error

	if dsn != "" {
		// Use DSN if provided
		db, err = openWithDSN(dsn)
		if err != nil {
			return nil, fmt.Errorf("failed to connect with DSN: %w", err)
		}
	} else {
		// Fall back to SQLite with DB_PATH
		db, err = openSQLite(dbPath)
		if err != nil {
			return nil, fmt.Errorf("failed to connect to SQLite: %w", err)
		}
	}

	// Auto-migrate all tables
	// GORM handles schema creation safely without SQL injection risks
	if err := db.AutoMigrate(
		&ChatHistory{},
		&UserConfiguration{},
		&MessageIDs{},
		&GroupAdmins{},
	); err != nil {
		return nil, fmt.Errorf("failed to migrate database schema: %w", err)
	}

	storage := &GORMStorage{db: db}

	// Start cleanup goroutine
	go storage.cleanupLoop()

	return storage, nil
}

// openWithDSN opens a database connection based on DSN format
// Supports: mysql://, postgres://, postgresql://, sqlite://
func openWithDSN(dsn string) (*gorm.DB, error) {
	dsn = strings.TrimSpace(dsn)

	if strings.HasPrefix(dsn, "mysql://") {
		// MySQL connection
		// Remove mysql:// prefix for the driver
		connStr := strings.TrimPrefix(dsn, "mysql://")
		return gorm.Open(mysql.Open(connStr), &gorm.Config{})
	} else if strings.HasPrefix(dsn, "postgres://") || strings.HasPrefix(dsn, "postgresql://") {
		// PostgreSQL connection
		// Driver accepts the full postgres:// URL
		return gorm.Open(postgres.Open(dsn), &gorm.Config{})
	} else if strings.HasPrefix(dsn, "sqlite://") {
		// SQLite connection with explicit sqlite:// prefix
		path := strings.TrimPrefix(dsn, "sqlite://")
		return openSQLite(path)
	} else {
		return nil, fmt.Errorf("unsupported DSN format: %s (supported: mysql://, postgres://, postgresql://, sqlite://)", dsn)
	}
}

// openSQLite opens a SQLite database connection using pure Go implementation
func openSQLite(dbPath string) (*gorm.DB, error) {
	if dbPath == "" {
		dbPath = "./data/bot.db"
	}

	// Use GORM's SQLite driver which will use modernc.org/sqlite when CGO is disabled
	// The driver automatically falls back to pure Go implementation
	db, err := gorm.Open(sqlite.Open(dbPath), &gorm.Config{})
	if err != nil {
		return nil, err
	}

	return db, nil
}

// buildSessionQuery creates a GORM query for session context
// Uses GORM's Where method which automatically parameterizes all values
func (s *GORMStorage) buildSessionQuery(ctx *SessionContext) *gorm.DB {
	query := s.db.Where("chat_id = ? AND bot_id = ?", ctx.ChatID, ctx.BotID)

	if ctx.UserID != nil {
		query = query.Where("user_id = ?", *ctx.UserID)
	} else {
		query = query.Where("user_id IS NULL")
	}

	if ctx.ThreadID != nil {
		query = query.Where("thread_id = ?", *ctx.ThreadID)
	} else {
		query = query.Where("thread_id IS NULL")
	}

	return query
}

// GetChatHistory retrieves the chat history for a session
// Uses GORM's parameterized queries to prevent SQL injection
func (s *GORMStorage) GetChatHistory(ctx *SessionContext) ([]HistoryItem, error) {
	var record ChatHistory

	result := s.buildSessionQuery(ctx).First(&record)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return []HistoryItem{}, nil
		}
		return nil, fmt.Errorf("failed to get chat history: %w", result.Error)
	}

	var history []HistoryItem
	if err := json.Unmarshal([]byte(record.History), &history); err != nil {
		return nil, fmt.Errorf("failed to unmarshal history: %w", err)
	}

	return history, nil
}

// SaveChatHistory saves the chat history for a session
// Uses GORM's parameterized queries to prevent SQL injection
func (s *GORMStorage) SaveChatHistory(ctx *SessionContext, history []HistoryItem) error {
	historyJSON, err := json.Marshal(history)
	if err != nil {
		return fmt.Errorf("failed to marshal history: %w", err)
	}

	record := ChatHistory{
		ChatID:   ctx.ChatID,
		BotID:    ctx.BotID,
		UserID:   ctx.UserID,
		ThreadID: ctx.ThreadID,
		History:  string(historyJSON),
	}

	// Use GORM's Assign + FirstOrCreate for upsert behavior
	// This is safe from SQL injection as GORM parameterizes all values
	result := s.buildSessionQuery(ctx).Assign(ChatHistory{
		History:   string(historyJSON),
		UpdatedAt: time.Now(),
	}).FirstOrCreate(&record)

	if result.Error != nil {
		return fmt.Errorf("failed to save chat history: %w", result.Error)
	}

	return nil
}

// DeleteChatHistory deletes the chat history for a session
// Uses GORM's parameterized queries to prevent SQL injection
func (s *GORMStorage) DeleteChatHistory(ctx *SessionContext) error {
	result := s.buildSessionQuery(ctx).Delete(&ChatHistory{})
	if result.Error != nil {
		return fmt.Errorf("failed to delete chat history: %w", result.Error)
	}
	return nil
}

// GetUserConfig retrieves the user configuration for a session
// Uses GORM's parameterized queries to prevent SQL injection
func (s *GORMStorage) GetUserConfig(ctx *SessionContext) (*UserConfig, error) {
	var record UserConfiguration

	result := s.buildSessionQuery(ctx).First(&record)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return &UserConfig{
				DefineKeys: []string{},
				Values:     make(map[string]interface{}),
			}, nil
		}
		return nil, fmt.Errorf("failed to get user config: %w", result.Error)
	}

	var config UserConfig
	if err := json.Unmarshal([]byte(record.Config), &config); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	// Ensure Values map is initialized
	if config.Values == nil {
		config.Values = make(map[string]interface{})
	}

	return &config, nil
}

// SaveUserConfig saves the user configuration for a session
// Uses GORM's parameterized queries to prevent SQL injection
func (s *GORMStorage) SaveUserConfig(ctx *SessionContext, config *UserConfig) error {
	configJSON, err := json.Marshal(config)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	record := UserConfiguration{
		ChatID:   ctx.ChatID,
		BotID:    ctx.BotID,
		UserID:   ctx.UserID,
		ThreadID: ctx.ThreadID,
		Config:   string(configJSON),
	}

	// Use GORM's Assign + FirstOrCreate for upsert behavior
	// This is safe from SQL injection as GORM parameterizes all values
	result := s.buildSessionQuery(ctx).Assign(UserConfiguration{
		Config:    string(configJSON),
		UpdatedAt: time.Now(),
	}).FirstOrCreate(&record)

	if result.Error != nil {
		return fmt.Errorf("failed to save user config: %w", result.Error)
	}

	return nil
}

// GetMessageIDs retrieves the message IDs for a session
// Uses GORM's parameterized queries to prevent SQL injection
func (s *GORMStorage) GetMessageIDs(ctx *SessionContext) ([]int, error) {
	var record MessageIDs

	result := s.buildSessionQuery(ctx).First(&record)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return []int{}, nil
		}
		return nil, fmt.Errorf("failed to get message IDs: %w", result.Error)
	}

	var ids []int
	if err := json.Unmarshal([]byte(record.IDs), &ids); err != nil {
		return nil, fmt.Errorf("failed to unmarshal message IDs: %w", err)
	}

	return ids, nil
}

// SaveMessageIDs saves the message IDs for a session
// Uses GORM's parameterized queries to prevent SQL injection
func (s *GORMStorage) SaveMessageIDs(ctx *SessionContext, ids []int) error {
	idsJSON, err := json.Marshal(ids)
	if err != nil {
		return fmt.Errorf("failed to marshal message IDs: %w", err)
	}

	record := MessageIDs{
		ChatID:   ctx.ChatID,
		BotID:    ctx.BotID,
		UserID:   ctx.UserID,
		ThreadID: ctx.ThreadID,
		IDs:      string(idsJSON),
	}

	// Use GORM's Assign + FirstOrCreate for upsert behavior
	// This is safe from SQL injection as GORM parameterizes all values
	result := s.buildSessionQuery(ctx).Assign(MessageIDs{
		IDs:       string(idsJSON),
		UpdatedAt: time.Now(),
	}).FirstOrCreate(&record)

	if result.Error != nil {
		return fmt.Errorf("failed to save message IDs: %w", result.Error)
	}

	return nil
}

// GetGroupAdmins retrieves the cached group admins
// Uses GORM's parameterized queries to prevent SQL injection
func (s *GORMStorage) GetGroupAdmins(chatID int64) ([]ChatMember, error) {
	var record GroupAdmins

	// Query with expiration check - GORM parameterizes all values
	result := s.db.Where("chat_id = ? AND expires_at > ?", chatID, time.Now()).First(&record)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, nil // No cached admins or expired
		}
		return nil, fmt.Errorf("failed to get group admins: %w", result.Error)
	}

	var admins []ChatMember
	if err := json.Unmarshal([]byte(record.Admins), &admins); err != nil {
		return nil, fmt.Errorf("failed to unmarshal group admins: %w", err)
	}

	return admins, nil
}

// SaveGroupAdmins saves the group admins with TTL
// Uses GORM's parameterized queries to prevent SQL injection
func (s *GORMStorage) SaveGroupAdmins(chatID int64, admins []ChatMember, ttl int) error {
	adminsJSON, err := json.Marshal(admins)
	if err != nil {
		return fmt.Errorf("failed to marshal group admins: %w", err)
	}

	expiresAt := time.Now().Add(time.Duration(ttl) * time.Second)

	record := GroupAdmins{
		ChatID:    chatID,
		Admins:    string(adminsJSON),
		ExpiresAt: expiresAt,
	}

	// Use GORM's Save for upsert behavior with unique constraint
	// This is safe from SQL injection as GORM parameterizes all values
	result := s.db.Where("chat_id = ?", chatID).Assign(GroupAdmins{
		Admins:    string(adminsJSON),
		ExpiresAt: expiresAt,
		UpdatedAt: time.Now(),
	}).FirstOrCreate(&record)

	if result.Error != nil {
		return fmt.Errorf("failed to save group admins: %w", result.Error)
	}

	return nil
}

// CleanupExpired removes expired data from the database
// Uses GORM's parameterized queries to prevent SQL injection
func (s *GORMStorage) CleanupExpired() error {
	// Delete expired group admins - GORM parameterizes the time value
	result := s.db.Where("expires_at < ?", time.Now()).Delete(&GroupAdmins{})
	if result.Error != nil {
		return fmt.Errorf("failed to cleanup expired data: %w", result.Error)
	}
	return nil
}

// Close closes the database connection
func (s *GORMStorage) Close() error {
	sqlDB, err := s.db.DB()
	if err != nil {
		return err
	}
	return sqlDB.Close()
}

// cleanupLoop runs a background goroutine that periodically cleans up expired data
func (s *GORMStorage) cleanupLoop() {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		if err := s.CleanupExpired(); err != nil {
			// Log error but continue running
			// Note: In production, consider using a proper logging framework
			_ = err
		}
	}
}
