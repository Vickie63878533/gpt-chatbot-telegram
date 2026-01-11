package storage

import "errors"

// Common errors
var (
	ErrNotFound = errors.New("resource not found")
)

// SessionContext represents the context for a chat session
type SessionContext struct {
	ChatID   int64
	BotID    int64
	UserID   *int64 // For non-shared group mode
	ThreadID *int64 // For forum/topic mode
}

// HistoryItem represents a single message in the conversation history
type HistoryItem struct {
	Role      string      `json:"role"`               // "user", "assistant", "system", "summary"
	Content   interface{} `json:"content"`            // string or []ContentPart
	Timestamp int64       `json:"timestamp,omitempty"` // Unix timestamp
	Truncated bool        `json:"truncated,omitempty"` // Marks if this is a truncation point from /clear
}

// ContentPart represents a part of a message (text or image)
type ContentPart struct {
	Type  string `json:"type"`            // "text" or "image"
	Text  string `json:"text,omitempty"`  // Text content
	Image string `json:"image,omitempty"` // URL or base64
}

// UserConfig represents user-specific configuration
type UserConfig struct {
	DefineKeys []string               `json:"DEFINE_KEYS"`
	Values     map[string]interface{} `json:"values"`
}

// ChatMember represents a Telegram chat member
type ChatMember struct {
	User   User   `json:"user"`
	Status string `json:"status"`
}

// User represents a Telegram user
type User struct {
	ID int64 `json:"id"`
}

// Storage defines the interface for data persistence
type Storage interface {
	// Chat History Operations
	GetChatHistory(ctx *SessionContext) ([]HistoryItem, error)
	SaveChatHistory(ctx *SessionContext, history []HistoryItem) error
	DeleteChatHistory(ctx *SessionContext) error

	// User Config Operations
	GetUserConfig(ctx *SessionContext) (*UserConfig, error)
	SaveUserConfig(ctx *SessionContext, config *UserConfig) error

	// Message IDs Operations (for SAFE_MODE)
	GetMessageIDs(ctx *SessionContext) ([]int, error)
	SaveMessageIDs(ctx *SessionContext, ids []int) error

	// Group Admins Operations
	GetGroupAdmins(chatID int64) ([]ChatMember, error)
	SaveGroupAdmins(chatID int64, admins []ChatMember, ttl int) error

	// Character Card Operations
	CreateCharacterCard(card *CharacterCard) error
	GetCharacterCard(id uint) (*CharacterCard, error)
	ListCharacterCards(userID *int64) ([]*CharacterCard, error)
	UpdateCharacterCard(card *CharacterCard) error
	DeleteCharacterCard(id uint) error
	GetActiveCharacterCard(userID *int64) (*CharacterCard, error)
	ActivateCharacterCard(userID *int64, cardID uint) error

	// World Book Operations
	CreateWorldBook(book *WorldBook) error
	GetWorldBook(id uint) (*WorldBook, error)
	ListWorldBooks(userID *int64) ([]*WorldBook, error)
	UpdateWorldBook(book *WorldBook) error
	DeleteWorldBook(id uint) error
	GetActiveWorldBook(userID *int64) (*WorldBook, error)
	ActivateWorldBook(userID *int64, bookID uint) error

	// World Book Entry Operations
	CreateWorldBookEntry(entry *WorldBookEntry) error
	GetWorldBookEntry(id uint) (*WorldBookEntry, error)
	ListWorldBookEntries(worldBookID uint) ([]*WorldBookEntry, error)
	UpdateWorldBookEntry(entry *WorldBookEntry) error
	DeleteWorldBookEntry(id uint) error

	// Preset Operations
	CreatePreset(preset *Preset) error
	GetPreset(id uint) (*Preset, error)
	ListPresets(userID *int64, apiType string) ([]*Preset, error)
	UpdatePreset(preset *Preset) error
	DeletePreset(id uint) error
	GetActivePreset(userID *int64, apiType string) (*Preset, error)
	ActivatePreset(userID *int64, presetID uint) error

	// Regex Pattern Operations
	CreateRegexPattern(pattern *RegexPattern) error
	GetRegexPattern(id uint) (*RegexPattern, error)
	ListRegexPatterns(userID *int64, patternType string) ([]*RegexPattern, error)
	UpdateRegexPattern(pattern *RegexPattern) error
	DeleteRegexPattern(id uint) error
	UpdateRegexPatternStatus(id uint, enabled bool) error

	// Login Token Operations
	CreateLoginToken(token *LoginToken) error
	ValidateLoginToken(userID int64, token string) (bool, error)
	DeleteLoginToken(userID int64) error
	CleanupExpiredTokens() error
	UpdateWorldBookEntryStatus(id uint, enabled bool) error

	// Maintenance
	CleanupExpired() error
	Close() error
}
