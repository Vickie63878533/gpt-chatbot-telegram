package storage

// SessionContext represents the context for a chat session
type SessionContext struct {
	ChatID   int64
	BotID    int64
	UserID   *int64 // For non-shared group mode
	ThreadID *int64 // For forum/topic mode
}

// HistoryItem represents a single message in the conversation history
type HistoryItem struct {
	Role    string      `json:"role"`    // "user", "assistant", "system"
	Content interface{} `json:"content"` // string or []ContentPart
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

	// Maintenance
	CleanupExpired() error
	Close() error
}
