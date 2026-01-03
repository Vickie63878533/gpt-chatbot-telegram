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
