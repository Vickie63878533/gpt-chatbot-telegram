package agent

import (
	"context"

	"github.com/tbxark/ChatGPT-Telegram-Workers/go_version/internal/config"
)

// HistoryItem represents a single message in the conversation history
type HistoryItem struct {
	Role    string      `json:"role"`    // "user", "assistant", "system"
	Content interface{} `json:"content"` // string or []ContentPart
}

// ContentPart represents a part of a message (text or image)
type ContentPart struct {
	Type  string `json:"type"`            // "text" or "image"
	Text  string `json:"text,omitempty"`  // Text content
	Image string `json:"image,omitempty"` // Image URL or base64 data
}

// LLMChatParams contains parameters for a chat completion request
type LLMChatParams struct {
	Prompt   string        // System prompt or initial message
	Messages []HistoryItem // Conversation history
}

// ChatAgentResponse represents the response from a chat agent
type ChatAgentResponse struct {
	Messages []HistoryItem // Response messages
}

// ChatStreamTextHandler is a callback function for streaming text responses
type ChatStreamTextHandler func(text string) error

// ChatAgent defines the interface for AI chat providers
type ChatAgent interface {
	// Name returns the unique identifier for this agent (e.g., "openai", "azure")
	Name() string

	// ModelKey returns the configuration key for the model (e.g., "OPENAI_CHAT_MODEL")
	ModelKey() string

	// Enable checks if this agent is enabled based on the configuration
	Enable(config *config.Config) bool

	// Model returns the current model name from the configuration
	Model(config *config.Config) string

	// ModelList returns the list of available models for this agent
	ModelList(config *config.Config) ([]string, error)

	// Request sends a chat completion request
	// If onStream is provided, the response will be streamed through the callback
	Request(ctx context.Context, params *LLMChatParams, config *config.Config, onStream ChatStreamTextHandler) (*ChatAgentResponse, error)
}

// ImageAgent defines the interface for AI image generation providers
type ImageAgent interface {
	// Name returns the unique identifier for this agent (e.g., "dall-e", "flux")
	Name() string

	// ModelKey returns the configuration key for the model (e.g., "DALL_E_MODEL")
	ModelKey() string

	// Enable checks if this agent is enabled based on the configuration
	Enable(config *config.Config) bool

	// Model returns the current model name from the configuration
	Model(config *config.Config) string

	// ModelList returns the list of available models for this agent
	ModelList(config *config.Config) ([]string, error)

	// Request generates an image based on the prompt
	// Returns either an image URL or base64-encoded image data
	Request(ctx context.Context, prompt string, config *config.Config) (string, error)
}
