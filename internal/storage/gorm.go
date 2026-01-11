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
		&CharacterCard{},
		&WorldBook{},
		&WorldBookEntry{},
		&Preset{},
		&RegexPattern{},
		&LoginToken{},
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

	// Delete expired login tokens
	if err := s.CleanupExpiredTokens(); err != nil {
		return fmt.Errorf("failed to cleanup expired tokens: %w", err)
	}

	return nil
}

// Character Card Operations

// CreateCharacterCard creates a new character card
// Uses GORM's parameterized queries to prevent SQL injection
func (s *GORMStorage) CreateCharacterCard(card *CharacterCard) error {
	result := s.db.Create(card)
	if result.Error != nil {
		return fmt.Errorf("failed to create character card: %w", result.Error)
	}
	return nil
}

// GetCharacterCard retrieves a character card by ID
// Uses GORM's parameterized queries to prevent SQL injection
func (s *GORMStorage) GetCharacterCard(id uint) (*CharacterCard, error) {
	var card CharacterCard
	result := s.db.First(&card, id)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("character card not found")
		}
		return nil, fmt.Errorf("failed to get character card: %w", result.Error)
	}
	return &card, nil
}

// ListCharacterCards lists all character cards for a user (or global if userID is nil)
// Uses GORM's parameterized queries to prevent SQL injection
func (s *GORMStorage) ListCharacterCards(userID *int64) ([]*CharacterCard, error) {
	var cards []*CharacterCard
	query := s.db

	if userID != nil {
		// Get user's cards and global cards
		query = query.Where("user_id = ? OR user_id IS NULL", *userID)
	} else {
		// Get only global cards
		query = query.Where("user_id IS NULL")
	}

	result := query.Order("created_at DESC").Find(&cards)
	if result.Error != nil {
		return nil, fmt.Errorf("failed to list character cards: %w", result.Error)
	}

	return cards, nil
}

// UpdateCharacterCard updates an existing character card
// Uses GORM's parameterized queries to prevent SQL injection
func (s *GORMStorage) UpdateCharacterCard(card *CharacterCard) error {
	result := s.db.Save(card)
	if result.Error != nil {
		return fmt.Errorf("failed to update character card: %w", result.Error)
	}
	return nil
}

// DeleteCharacterCard deletes a character card by ID
// Uses GORM's parameterized queries to prevent SQL injection
func (s *GORMStorage) DeleteCharacterCard(id uint) error {
	result := s.db.Delete(&CharacterCard{}, id)
	if result.Error != nil {
		return fmt.Errorf("failed to delete character card: %w", result.Error)
	}
	return nil
}

// GetActiveCharacterCard retrieves the active character card for a user
// Uses GORM's parameterized queries to prevent SQL injection
func (s *GORMStorage) GetActiveCharacterCard(userID *int64) (*CharacterCard, error) {
	var card CharacterCard
	query := s.db.Where("is_active = ?", true)

	if userID != nil {
		query = query.Where("user_id = ?", *userID)
	} else {
		query = query.Where("user_id IS NULL")
	}

	result := query.First(&card)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, nil // No active card
		}
		return nil, fmt.Errorf("failed to get active character card: %w", result.Error)
	}

	return &card, nil
}

// ActivateCharacterCard activates a character card and deactivates others
// Uses transaction to ensure atomicity and GORM's parameterized queries
func (s *GORMStorage) ActivateCharacterCard(userID *int64, cardID uint) error {
	return s.db.Transaction(func(tx *gorm.DB) error {
		// Deactivate all cards for this user
		query := tx.Model(&CharacterCard{})
		if userID != nil {
			query = query.Where("user_id = ?", *userID)
		} else {
			query = query.Where("user_id IS NULL")
		}

		if err := query.Update("is_active", false).Error; err != nil {
			return fmt.Errorf("failed to deactivate cards: %w", err)
		}

		// Activate the specified card
		if err := tx.Model(&CharacterCard{}).Where("id = ?", cardID).Update("is_active", true).Error; err != nil {
			return fmt.Errorf("failed to activate card: %w", err)
		}

		return nil
	})
}

// World Book Operations

// CreateWorldBook creates a new world book
// Uses GORM's parameterized queries to prevent SQL injection
func (s *GORMStorage) CreateWorldBook(book *WorldBook) error {
	result := s.db.Create(book)
	if result.Error != nil {
		return fmt.Errorf("failed to create world book: %w", result.Error)
	}
	return nil
}

// GetWorldBook retrieves a world book by ID
// Uses GORM's parameterized queries to prevent SQL injection
func (s *GORMStorage) GetWorldBook(id uint) (*WorldBook, error) {
	var book WorldBook
	result := s.db.First(&book, id)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("world book not found")
		}
		return nil, fmt.Errorf("failed to get world book: %w", result.Error)
	}
	return &book, nil
}

// ListWorldBooks lists all world books for a user (or global if userID is nil)
// Uses GORM's parameterized queries to prevent SQL injection
func (s *GORMStorage) ListWorldBooks(userID *int64) ([]*WorldBook, error) {
	var books []*WorldBook
	query := s.db

	if userID != nil {
		// Get user's books and global books
		query = query.Where("user_id = ? OR user_id IS NULL", *userID)
	} else {
		// Get only global books
		query = query.Where("user_id IS NULL")
	}

	result := query.Order("created_at DESC").Find(&books)
	if result.Error != nil {
		return nil, fmt.Errorf("failed to list world books: %w", result.Error)
	}

	return books, nil
}

// UpdateWorldBook updates an existing world book
// Uses GORM's parameterized queries to prevent SQL injection
func (s *GORMStorage) UpdateWorldBook(book *WorldBook) error {
	result := s.db.Save(book)
	if result.Error != nil {
		return fmt.Errorf("failed to update world book: %w", result.Error)
	}
	return nil
}

// DeleteWorldBook deletes a world book by ID
// Uses GORM's parameterized queries to prevent SQL injection
func (s *GORMStorage) DeleteWorldBook(id uint) error {
	result := s.db.Delete(&WorldBook{}, id)
	if result.Error != nil {
		return fmt.Errorf("failed to delete world book: %w", result.Error)
	}
	return nil
}

// GetActiveWorldBook retrieves the active world book for a user
// Uses GORM's parameterized queries to prevent SQL injection
func (s *GORMStorage) GetActiveWorldBook(userID *int64) (*WorldBook, error) {
	var book WorldBook
	query := s.db.Where("is_active = ?", true)

	if userID != nil {
		query = query.Where("user_id = ?", *userID)
	} else {
		query = query.Where("user_id IS NULL")
	}

	result := query.First(&book)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, nil // No active book
		}
		return nil, fmt.Errorf("failed to get active world book: %w", result.Error)
	}

	return &book, nil
}

// ActivateWorldBook activates a world book and deactivates others
// Uses transaction to ensure atomicity and GORM's parameterized queries
func (s *GORMStorage) ActivateWorldBook(userID *int64, bookID uint) error {
	return s.db.Transaction(func(tx *gorm.DB) error {
		// Deactivate all books for this user
		query := tx.Model(&WorldBook{})
		if userID != nil {
			query = query.Where("user_id = ?", *userID)
		} else {
			query = query.Where("user_id IS NULL")
		}

		if err := query.Update("is_active", false).Error; err != nil {
			return fmt.Errorf("failed to deactivate books: %w", err)
		}

		// Activate the specified book
		if err := tx.Model(&WorldBook{}).Where("id = ?", bookID).Update("is_active", true).Error; err != nil {
			return fmt.Errorf("failed to activate book: %w", err)
		}

		return nil
	})
}

// World Book Entry Operations

// CreateWorldBookEntry creates a new world book entry
// Uses GORM's parameterized queries to prevent SQL injection
func (s *GORMStorage) CreateWorldBookEntry(entry *WorldBookEntry) error {
	result := s.db.Create(entry)
	if result.Error != nil {
		return fmt.Errorf("failed to create world book entry: %w", result.Error)
	}
	return nil
}

// GetWorldBookEntry retrieves a world book entry by ID
// Uses GORM's parameterized queries to prevent SQL injection
func (s *GORMStorage) GetWorldBookEntry(id uint) (*WorldBookEntry, error) {
	var entry WorldBookEntry
	result := s.db.First(&entry, id)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("world book entry not found")
		}
		return nil, fmt.Errorf("failed to get world book entry: %w", result.Error)
	}
	return &entry, nil
}

// ListWorldBookEntries lists all entries for a world book
// Uses GORM's parameterized queries to prevent SQL injection
func (s *GORMStorage) ListWorldBookEntries(worldBookID uint) ([]*WorldBookEntry, error) {
	var entries []*WorldBookEntry
	result := s.db.Where("world_book_id = ?", worldBookID).Order("`order` ASC").Find(&entries)
	if result.Error != nil {
		return nil, fmt.Errorf("failed to list world book entries: %w", result.Error)
	}
	return entries, nil
}

// UpdateWorldBookEntry updates an existing world book entry
// Uses GORM's parameterized queries to prevent SQL injection
func (s *GORMStorage) UpdateWorldBookEntry(entry *WorldBookEntry) error {
	result := s.db.Save(entry)
	if result.Error != nil {
		return fmt.Errorf("failed to update world book entry: %w", result.Error)
	}
	return nil
}

// DeleteWorldBookEntry deletes a world book entry by ID
// Uses GORM's parameterized queries to prevent SQL injection
func (s *GORMStorage) DeleteWorldBookEntry(id uint) error {
	result := s.db.Delete(&WorldBookEntry{}, id)
	if result.Error != nil {
		return fmt.Errorf("failed to delete world book entry: %w", result.Error)
	}
	return nil
}

// Preset Operations

// CreatePreset creates a new preset
// Uses GORM's parameterized queries to prevent SQL injection
func (s *GORMStorage) CreatePreset(preset *Preset) error {
	result := s.db.Create(preset)
	if result.Error != nil {
		return fmt.Errorf("failed to create preset: %w", result.Error)
	}
	return nil
}

// GetPreset retrieves a preset by ID
// Uses GORM's parameterized queries to prevent SQL injection
func (s *GORMStorage) GetPreset(id uint) (*Preset, error) {
	var preset Preset
	result := s.db.First(&preset, id)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("preset not found")
		}
		return nil, fmt.Errorf("failed to get preset: %w", result.Error)
	}
	return &preset, nil
}

// ListPresets lists all presets for a user and API type (or global if userID is nil)
// Uses GORM's parameterized queries to prevent SQL injection
func (s *GORMStorage) ListPresets(userID *int64, apiType string) ([]*Preset, error) {
	var presets []*Preset
	query := s.db

	if userID != nil {
		// Get user's presets and global presets
		query = query.Where("user_id = ? OR user_id IS NULL", *userID)
	} else {
		// Get only global presets
		query = query.Where("user_id IS NULL")
	}

	// Filter by API type if specified
	if apiType != "" {
		query = query.Where("api_type = ?", apiType)
	}

	result := query.Order("created_at DESC").Find(&presets)
	if result.Error != nil {
		return nil, fmt.Errorf("failed to list presets: %w", result.Error)
	}

	return presets, nil
}

// UpdatePreset updates an existing preset
// Uses GORM's parameterized queries to prevent SQL injection
func (s *GORMStorage) UpdatePreset(preset *Preset) error {
	result := s.db.Save(preset)
	if result.Error != nil {
		return fmt.Errorf("failed to update preset: %w", result.Error)
	}
	return nil
}

// DeletePreset deletes a preset by ID
// Uses GORM's parameterized queries to prevent SQL injection
func (s *GORMStorage) DeletePreset(id uint) error {
	result := s.db.Delete(&Preset{}, id)
	if result.Error != nil {
		return fmt.Errorf("failed to delete preset: %w", result.Error)
	}
	return nil
}

// GetActivePreset retrieves the active preset for a user and API type
// Uses GORM's parameterized queries to prevent SQL injection
func (s *GORMStorage) GetActivePreset(userID *int64, apiType string) (*Preset, error) {
	var preset Preset
	query := s.db.Where("is_active = ? AND api_type = ?", true, apiType)

	if userID != nil {
		query = query.Where("user_id = ?", *userID)
	} else {
		query = query.Where("user_id IS NULL")
	}

	result := query.First(&preset)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, nil // No active preset
		}
		return nil, fmt.Errorf("failed to get active preset: %w", result.Error)
	}

	return &preset, nil
}

// ActivatePreset activates a preset and deactivates others of the same API type
// Uses transaction to ensure atomicity and GORM's parameterized queries
func (s *GORMStorage) ActivatePreset(userID *int64, presetID uint) error {
	return s.db.Transaction(func(tx *gorm.DB) error {
		// Get the preset to find its API type
		var preset Preset
		if err := tx.First(&preset, presetID).Error; err != nil {
			return fmt.Errorf("failed to find preset: %w", err)
		}

		// Deactivate all presets for this user and API type
		query := tx.Model(&Preset{}).Where("api_type = ?", preset.APIType)
		if userID != nil {
			query = query.Where("user_id = ?", *userID)
		} else {
			query = query.Where("user_id IS NULL")
		}

		if err := query.Update("is_active", false).Error; err != nil {
			return fmt.Errorf("failed to deactivate presets: %w", err)
		}

		// Activate the specified preset
		if err := tx.Model(&Preset{}).Where("id = ?", presetID).Update("is_active", true).Error; err != nil {
			return fmt.Errorf("failed to activate preset: %w", err)
		}

		return nil
	})
}

// Regex Pattern Operations

// CreateRegexPattern creates a new regex pattern
// Uses GORM's parameterized queries to prevent SQL injection
func (s *GORMStorage) CreateRegexPattern(pattern *RegexPattern) error {
	result := s.db.Create(pattern)
	if result.Error != nil {
		return fmt.Errorf("failed to create regex pattern: %w", result.Error)
	}
	return nil
}

// GetRegexPattern retrieves a regex pattern by ID
// Uses GORM's parameterized queries to prevent SQL injection
func (s *GORMStorage) GetRegexPattern(id uint) (*RegexPattern, error) {
	var pattern RegexPattern
	result := s.db.First(&pattern, id)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("regex pattern not found")
		}
		return nil, fmt.Errorf("failed to get regex pattern: %w", result.Error)
	}
	return &pattern, nil
}

// ListRegexPatterns lists all regex patterns for a user and type (or global if userID is nil)
// Uses GORM's parameterized queries to prevent SQL injection
func (s *GORMStorage) ListRegexPatterns(userID *int64, patternType string) ([]*RegexPattern, error) {
	var patterns []*RegexPattern
	query := s.db

	if userID != nil {
		// Get user's patterns and global patterns
		query = query.Where("user_id = ? OR user_id IS NULL", *userID)
	} else {
		// Get only global patterns
		query = query.Where("user_id IS NULL")
	}

	// Filter by type if specified
	if patternType != "" {
		query = query.Where("type = ?", patternType)
	}

	result := query.Order("`order` ASC").Find(&patterns)
	if result.Error != nil {
		return nil, fmt.Errorf("failed to list regex patterns: %w", result.Error)
	}

	return patterns, nil
}

// UpdateRegexPattern updates an existing regex pattern
// Uses GORM's parameterized queries to prevent SQL injection
func (s *GORMStorage) UpdateRegexPattern(pattern *RegexPattern) error {
	result := s.db.Save(pattern)
	if result.Error != nil {
		return fmt.Errorf("failed to update regex pattern: %w", result.Error)
	}
	return nil
}

// DeleteRegexPattern deletes a regex pattern by ID
// Uses GORM's parameterized queries to prevent SQL injection
func (s *GORMStorage) DeleteRegexPattern(id uint) error {
	result := s.db.Delete(&RegexPattern{}, id)
	if result.Error != nil {
		return fmt.Errorf("failed to delete regex pattern: %w", result.Error)
	}
	return nil
}

// Login Token Operations

// CreateLoginToken creates a new login token
// Uses GORM's parameterized queries to prevent SQL injection
func (s *GORMStorage) CreateLoginToken(token *LoginToken) error {
	// Delete any existing token for this user first
	if err := s.DeleteLoginToken(token.UserID); err != nil {
		return fmt.Errorf("failed to delete existing token: %w", err)
	}

	result := s.db.Create(token)
	if result.Error != nil {
		return fmt.Errorf("failed to create login token: %w", result.Error)
	}
	return nil
}

// ValidateLoginToken validates a login token for a user
// Uses GORM's parameterized queries to prevent SQL injection
func (s *GORMStorage) ValidateLoginToken(userID int64, token string) (bool, error) {
	var loginToken LoginToken
	result := s.db.Where("user_id = ? AND token = ? AND expires_at > ?", userID, token, time.Now()).First(&loginToken)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return false, nil // Token not found or expired
		}
		return false, fmt.Errorf("failed to validate login token: %w", result.Error)
	}
	return true, nil
}

// DeleteLoginToken deletes a login token for a user
// Uses GORM's parameterized queries to prevent SQL injection
func (s *GORMStorage) DeleteLoginToken(userID int64) error {
	result := s.db.Where("user_id = ?", userID).Delete(&LoginToken{})
	if result.Error != nil {
		return fmt.Errorf("failed to delete login token: %w", result.Error)
	}
	return nil
}

// CleanupExpiredTokens removes expired login tokens
// Uses GORM's parameterized queries to prevent SQL injection
func (s *GORMStorage) CleanupExpiredTokens() error {
	result := s.db.Where("expires_at < ?", time.Now()).Delete(&LoginToken{})
	if result.Error != nil {
		return fmt.Errorf("failed to cleanup expired tokens: %w", result.Error)
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

// UpdateRegexPatternStatus updates the enabled status of a regex pattern
func (s *GORMStorage) UpdateRegexPatternStatus(id uint, enabled bool) error {
	result := s.db.Model(&RegexPattern{}).Where("id = ?", id).Update("enabled", enabled)
	if result.Error != nil {
		return fmt.Errorf("failed to update regex pattern status: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return ErrNotFound
	}
	return nil
}

// UpdateWorldBookEntryStatus updates the enabled status of a world book entry
func (s *GORMStorage) UpdateWorldBookEntryStatus(id uint, enabled bool) error {
	result := s.db.Model(&WorldBookEntry{}).Where("id = ?", id).Update("enabled", enabled)
	if result.Error != nil {
		return fmt.Errorf("failed to update world book entry status: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return ErrNotFound
	}
	return nil
}
