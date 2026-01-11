package storage

import (
	"time"

	"gorm.io/gorm"
)

// ChatHistory represents the chat history table
// GORM will automatically handle SQL injection prevention through parameterized queries
type ChatHistory struct {
	ID        uint      `gorm:"primarykey"`
	CreatedAt time.Time `gorm:"index"`
	UpdatedAt time.Time
	DeletedAt gorm.DeletedAt `gorm:"index"`

	// Session identifiers - composite unique index
	ChatID   int64  `gorm:"not null;index:idx_chat_history_session,priority:1"`
	BotID    int64  `gorm:"not null;index:idx_chat_history_session,priority:2"`
	UserID   *int64 `gorm:"index:idx_chat_history_session,priority:3"` // Nullable for shared mode
	ThreadID *int64 `gorm:"index:idx_chat_history_session,priority:4"` // Nullable for non-forum chats

	// Data stored as JSON text
	History string `gorm:"type:text;not null"`
}

// TableName specifies the table name for ChatHistory
func (ChatHistory) TableName() string {
	return "chat_histories"
}

// UserConfiguration represents the user configuration table
// GORM will automatically handle SQL injection prevention through parameterized queries
type UserConfiguration struct {
	ID        uint      `gorm:"primarykey"`
	CreatedAt time.Time `gorm:"index"`
	UpdatedAt time.Time
	DeletedAt gorm.DeletedAt `gorm:"index"`

	// Session identifiers - composite unique index
	ChatID   int64  `gorm:"not null;index:idx_user_config_session,priority:1"`
	BotID    int64  `gorm:"not null;index:idx_user_config_session,priority:2"`
	UserID   *int64 `gorm:"index:idx_user_config_session,priority:3"` // Nullable for shared mode
	ThreadID *int64 `gorm:"index:idx_user_config_session,priority:4"` // Nullable for non-forum chats

	// Data stored as JSON text
	Config string `gorm:"type:text;not null"`
}

// TableName specifies the table name for UserConfiguration
func (UserConfiguration) TableName() string {
	return "user_configs"
}

// MessageIDs represents the message IDs table for SAFE_MODE
// GORM will automatically handle SQL injection prevention through parameterized queries
type MessageIDs struct {
	ID        uint      `gorm:"primarykey"`
	CreatedAt time.Time `gorm:"index"`
	UpdatedAt time.Time
	DeletedAt gorm.DeletedAt `gorm:"index"`

	// Session identifiers - composite unique index
	ChatID   int64  `gorm:"not null;index:idx_message_ids_session,priority:1"`
	BotID    int64  `gorm:"not null;index:idx_message_ids_session,priority:2"`
	UserID   *int64 `gorm:"index:idx_message_ids_session,priority:3"` // Nullable for shared mode
	ThreadID *int64 `gorm:"index:idx_message_ids_session,priority:4"` // Nullable for non-forum chats

	// Data stored as JSON text
	IDs string `gorm:"type:text;not null"`
}

// TableName specifies the table name for MessageIDs
func (MessageIDs) TableName() string {
	return "message_ids"
}

// GroupAdmins represents the group admins cache table
// GORM will automatically handle SQL injection prevention through parameterized queries
type GroupAdmins struct {
	ID        uint      `gorm:"primarykey"`
	CreatedAt time.Time `gorm:"index"`
	UpdatedAt time.Time
	DeletedAt gorm.DeletedAt `gorm:"index"`

	// Unique chat identifier
	ChatID int64 `gorm:"uniqueIndex;not null"`

	// Cached admin data as JSON
	Admins string `gorm:"type:text;not null"`

	// Expiration time for cache invalidation
	ExpiresAt time.Time `gorm:"index;not null"`
}

// TableName specifies the table name for GroupAdmins
func (GroupAdmins) TableName() string {
	return "group_admins"
}

// CharacterCard represents a SillyTavern character card
// GORM will automatically handle SQL injection prevention through parameterized queries
type CharacterCard struct {
	ID        uint      `gorm:"primarykey"`
	CreatedAt time.Time `gorm:"index"`
	UpdatedAt time.Time
	DeletedAt gorm.DeletedAt `gorm:"index"`

	// Owner information
	UserID *int64 `gorm:"index"` // nil for global cards

	// Card metadata
	Name   string `gorm:"not null;index"`
	Avatar string `gorm:"type:text"` // Avatar URL or base64

	// SillyTavern V2 format data
	Data string `gorm:"type:text;not null"` // JSON format

	// Status
	IsActive bool `gorm:"default:false;index"`
}

// TableName specifies the table name for CharacterCard
func (CharacterCard) TableName() string {
	return "character_cards"
}

// WorldBook represents a SillyTavern world book
// GORM will automatically handle SQL injection prevention through parameterized queries
type WorldBook struct {
	ID        uint      `gorm:"primarykey"`
	CreatedAt time.Time `gorm:"index"`
	UpdatedAt time.Time
	DeletedAt gorm.DeletedAt `gorm:"index"`

	// Owner information
	UserID *int64 `gorm:"index"` // nil for global books

	// Book metadata
	Name string `gorm:"not null;index"`

	// World book data
	Data string `gorm:"type:text;not null"` // JSON format

	// Status
	IsActive bool `gorm:"default:false;index"`
}

// TableName specifies the table name for WorldBook
func (WorldBook) TableName() string {
	return "world_books"
}

// WorldBookEntry represents an entry in a world book
// GORM will automatically handle SQL injection prevention through parameterized queries
type WorldBookEntry struct {
	ID        uint      `gorm:"primarykey"`
	CreatedAt time.Time `gorm:"index"`
	UpdatedAt time.Time
	DeletedAt gorm.DeletedAt `gorm:"index"`

	// Association
	WorldBookID uint `gorm:"not null;index"`

	// Entry data
	UID           string `gorm:"not null;uniqueIndex"`
	Keys          string `gorm:"type:text;not null"` // JSON array
	SecondaryKeys string `gorm:"type:text"`          // JSON array
	Content       string `gorm:"type:text;not null"`
	Comment       string `gorm:"type:text"`

	// Configuration
	Constant  bool   `gorm:"default:false"`
	Selective bool   `gorm:"default:false"`
	Order     int    `gorm:"default:100"`
	Position  string `gorm:"default:'after_char'"` // before_char, after_char
	Enabled   bool   `gorm:"default:true;index"`

	// Extensions
	Extensions string `gorm:"type:text"` // JSON format
}

// TableName specifies the table name for WorldBookEntry
func (WorldBookEntry) TableName() string {
	return "world_book_entries"
}

// Preset represents a SillyTavern preset
// GORM will automatically handle SQL injection prevention through parameterized queries
type Preset struct {
	ID        uint      `gorm:"primarykey"`
	CreatedAt time.Time `gorm:"index"`
	UpdatedAt time.Time
	DeletedAt gorm.DeletedAt `gorm:"index"`

	// Owner information
	UserID *int64 `gorm:"index"` // nil for global presets

	// Preset metadata
	Name    string `gorm:"not null;index"`
	APIType string `gorm:"not null;index"` // openai, anthropic, etc.

	// Preset data
	Data string `gorm:"type:text;not null"` // JSON format

	// Status
	IsActive bool `gorm:"default:false;index"`
}

// TableName specifies the table name for Preset
func (Preset) TableName() string {
	return "presets"
}

// RegexPattern represents a regex transformation pattern
// GORM will automatically handle SQL injection prevention through parameterized queries
type RegexPattern struct {
	ID        uint      `gorm:"primarykey"`
	CreatedAt time.Time `gorm:"index"`
	UpdatedAt time.Time
	DeletedAt gorm.DeletedAt `gorm:"index"`

	// Owner information
	UserID *int64 `gorm:"index"` // nil for global patterns

	// Pattern metadata
	Name string `gorm:"not null;index"`

	// Pattern configuration
	Pattern string `gorm:"type:text;not null"` // Regex pattern
	Replace string `gorm:"type:text;not null"` // Replacement text
	Type    string `gorm:"not null"`           // input, output
	Order   int    `gorm:"default:100"`
	Enabled bool   `gorm:"default:true;index"`
}

// TableName specifies the table name for RegexPattern
func (RegexPattern) TableName() string {
	return "regex_patterns"
}

// LoginToken represents a temporary login token for web manager
// GORM will automatically handle SQL injection prevention through parameterized queries
type LoginToken struct {
	ID        uint      `gorm:"primarykey"`
	CreatedAt time.Time `gorm:"index"`
	UpdatedAt time.Time

	// User information
	UserID int64 `gorm:"not null;uniqueIndex"`

	// Token
	Token string `gorm:"not null;uniqueIndex"`

	// Expiration
	ExpiresAt time.Time `gorm:"not null;index"`
}

// TableName specifies the table name for LoginToken
func (LoginToken) TableName() string {
	return "login_tokens"
}
