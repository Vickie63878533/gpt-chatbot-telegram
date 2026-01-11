package sillytavern

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/tbxark/ChatGPT-Telegram-Workers/go_version/internal/agent"
	"github.com/tbxark/ChatGPT-Telegram-Workers/go_version/internal/config"
	"github.com/tbxark/ChatGPT-Telegram-Workers/go_version/internal/storage"
)

// ContextConfig holds configuration for context management
type ContextConfig struct {
	MaxContextLength  int     // Maximum context length in tokens
	SummaryThreshold  float64 // Threshold (0.0-1.0) to trigger summary
	MinRecentPairs    int     // Minimum number of recent message pairs to keep
	TokensPerMessage  int     // Estimated tokens per message (for rough estimation)
	TokensPerChar     float64 // Estimated tokens per character (for rough estimation)
}

// DefaultContextConfig returns default configuration
func DefaultContextConfig() *ContextConfig {
	return &ContextConfig{
		MaxContextLength:  8000,
		SummaryThreshold:  0.8,
		MinRecentPairs:    2,
		TokensPerMessage:  10,   // Overhead per message
		TokensPerChar:     0.25, // Rough estimate: 4 chars per token
	}
}

// ContextManager manages conversation history, summaries, and context windows
type ContextManager struct {
	storage    storage.Storage
	config     *ContextConfig
	chatAgent  agent.ChatAgent
	botConfig  *config.Config
}

// NewContextManager creates a new context manager
func NewContextManager(
	storage storage.Storage,
	config *ContextConfig,
	chatAgent agent.ChatAgent,
	botConfig *config.Config,
) *ContextManager {
	if config == nil {
		config = DefaultContextConfig()
	}
	return &ContextManager{
		storage:   storage,
		config:    config,
		chatAgent: chatAgent,
		botConfig: botConfig,
	}
}

// EstimateTokens estimates the number of tokens in a message history
// This is a rough estimation based on character count and message overhead
func (m *ContextManager) EstimateTokens(messages []storage.HistoryItem) int {
	totalTokens := 0
	
	for _, msg := range messages {
		// Add message overhead
		totalTokens += m.config.TokensPerMessage
		
		// Estimate content tokens
		contentStr := ""
		switch v := msg.Content.(type) {
		case string:
			contentStr = v
		case []storage.ContentPart:
			for _, part := range v {
				if part.Type == "text" {
					contentStr += part.Text
				}
			}
		}
		
		// Estimate tokens from character count
		totalTokens += int(float64(len(contentStr)) * m.config.TokensPerChar)
	}
	
	return totalTokens
}

// estimateTokensForAgent converts storage.HistoryItem to agent.HistoryItem and estimates tokens
func (m *ContextManager) estimateTokensForAgent(messages []agent.HistoryItem) int {
	totalTokens := 0
	
	for _, msg := range messages {
		// Add message overhead
		totalTokens += m.config.TokensPerMessage
		
		// Estimate content tokens
		contentStr := ""
		switch v := msg.Content.(type) {
		case string:
			contentStr = v
		case []agent.ContentPart:
			for _, part := range v {
				if part.Type == "text" {
					contentStr += part.Text
				}
			}
		}
		
		// Estimate tokens from character count
		totalTokens += int(float64(len(contentStr)) * m.config.TokensPerChar)
	}
	
	return totalTokens
}

// convertToAgentHistory converts storage.HistoryItem to agent.HistoryItem
func convertToAgentHistory(items []storage.HistoryItem) []agent.HistoryItem {
	result := make([]agent.HistoryItem, 0, len(items))
	
	for _, item := range items {
		agentItem := agent.HistoryItem{
			Role: item.Role,
		}
		
		// Convert content
		switch v := item.Content.(type) {
		case string:
			agentItem.Content = v
		case []storage.ContentPart:
			parts := make([]agent.ContentPart, len(v))
			for i, part := range v {
				parts[i] = agent.ContentPart{
					Type:  part.Type,
					Text:  part.Text,
					Image: part.Image,
				}
			}
			agentItem.Content = parts
		default:
			// Try to convert to string
			if str, ok := v.(string); ok {
				agentItem.Content = str
			} else {
				// Fallback: convert to JSON string
				if jsonBytes, err := json.Marshal(v); err == nil {
					agentItem.Content = string(jsonBytes)
				} else {
					agentItem.Content = fmt.Sprintf("%v", v)
				}
			}
		}
		
		result = append(result, agentItem)
	}
	
	return result
}

// extractTextContent extracts text content from a message for summarization
func extractTextContent(item storage.HistoryItem) string {
	switch v := item.Content.(type) {
	case string:
		return v
	case []storage.ContentPart:
		var texts []string
		for _, part := range v {
			if part.Type == "text" {
				texts = append(texts, part.Text)
			}
		}
		return strings.Join(texts, "\n")
	default:
		return fmt.Sprintf("%v", v)
	}
}

// AddMessage adds a message to the conversation history
func (m *ContextManager) AddMessage(ctx *storage.SessionContext, role string, content interface{}) error {
	// Get current history
	history, err := m.storage.GetChatHistory(ctx)
	if err != nil {
		return fmt.Errorf("failed to get chat history: %w", err)
	}
	
	// Create new message
	newMessage := storage.HistoryItem{
		Role:      role,
		Content:   content,
		Timestamp: time.Now().Unix(),
		Truncated: false,
	}
	
	// Append to history
	history = append(history, newMessage)
	
	// Save updated history
	if err := m.storage.SaveChatHistory(ctx, history); err != nil {
		return fmt.Errorf("failed to save chat history: %w", err)
	}
	
	// Check if we need to trigger summary
	buildHistory, err := m.GetBuildHistory(ctx)
	if err != nil {
		return fmt.Errorf("failed to get build history: %w", err)
	}
	
	estimatedTokens := m.EstimateTokens(buildHistory)
	threshold := int(float64(m.config.MaxContextLength) * m.config.SummaryThreshold)
	
	if estimatedTokens > threshold {
		// Trigger summary asynchronously (don't block on it)
		go func() {
			if err := m.TriggerSummary(ctx); err != nil {
				// Log error but don't fail the message addition
				fmt.Printf("Warning: failed to trigger summary: %v\n", err)
			}
		}()
	}
	
	return nil
}

// GetBuildHistory returns the history to use for building AI requests
// This applies summary logic and truncation markers
func (m *ContextManager) GetBuildHistory(ctx *storage.SessionContext) ([]storage.HistoryItem, error) {
	// Get full history
	fullHistory, err := m.storage.GetChatHistory(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get chat history: %w", err)
	}
	
	// Find the last truncation marker
	lastTruncationIndex := -1
	for i := len(fullHistory) - 1; i >= 0; i-- {
		if fullHistory[i].Truncated {
			lastTruncationIndex = i
			break
		}
	}
	
	// If there's a truncation marker, only use messages after it
	var workingHistory []storage.HistoryItem
	if lastTruncationIndex >= 0 {
		workingHistory = fullHistory[lastTruncationIndex+1:]
	} else {
		workingHistory = fullHistory
	}
	
	// Separate messages by type
	var systemMessages []storage.HistoryItem
	var summaryMessages []storage.HistoryItem
	var conversationMessages []storage.HistoryItem
	
	for _, msg := range workingHistory {
		switch msg.Role {
		case "system":
			systemMessages = append(systemMessages, msg)
		case "summary":
			summaryMessages = append(summaryMessages, msg)
		case "user", "assistant":
			conversationMessages = append(conversationMessages, msg)
		}
	}
	
	// Build final history: system + summary + recent conversation
	result := make([]storage.HistoryItem, 0)
	result = append(result, systemMessages...)
	result = append(result, summaryMessages...)
	result = append(result, conversationMessages...)
	
	return result, nil
}

// GetFullHistory returns the complete conversation history for sharing
// This includes all messages regardless of summary status or truncation
func (m *ContextManager) GetFullHistory(ctx *storage.SessionContext) ([]storage.HistoryItem, error) {
	// Get full history
	fullHistory, err := m.storage.GetChatHistory(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get chat history: %w", err)
	}
	
	// Filter to only user and assistant messages
	result := make([]storage.HistoryItem, 0)
	for _, msg := range fullHistory {
		if msg.Role == "user" || msg.Role == "assistant" {
			result = append(result, msg)
		}
	}
	
	return result, nil
}

// TriggerSummary generates a summary of older messages and marks them as summarized
func (m *ContextManager) TriggerSummary(ctx *storage.SessionContext) error {
	// Get current build history
	buildHistory, err := m.GetBuildHistory(ctx)
	if err != nil {
		return fmt.Errorf("failed to get build history: %w", err)
	}
	
	// Separate system, summary, and conversation messages
	var systemMessages []storage.HistoryItem
	var existingSummaries []storage.HistoryItem
	var conversationMessages []storage.HistoryItem
	
	for _, msg := range buildHistory {
		switch msg.Role {
		case "system":
			systemMessages = append(systemMessages, msg)
		case "summary":
			existingSummaries = append(existingSummaries, msg)
		case "user", "assistant":
			conversationMessages = append(conversationMessages, msg)
		}
	}
	
	// Calculate how many recent pairs to keep
	recentPairsToKeep := m.config.MinRecentPairs * 2 // user + assistant = 2 messages per pair
	
	// If we don't have enough messages to summarize, skip
	if len(conversationMessages) <= recentPairsToKeep {
		return nil
	}
	
	// Split into messages to summarize and recent messages to keep
	messagesToSummarize := conversationMessages[:len(conversationMessages)-recentPairsToKeep]
	recentMessages := conversationMessages[len(conversationMessages)-recentPairsToKeep:]
	
	// Build summary prompt
	summaryPrompt := "Please provide a concise summary of the following conversation. Focus on key points, decisions, and important information:\n\n"
	for _, msg := range messagesToSummarize {
		content := extractTextContent(msg)
		summaryPrompt += fmt.Sprintf("%s: %s\n", msg.Role, content)
	}
	
	// Call AI to generate summary
	agentMessages := []agent.HistoryItem{
		{
			Role:    "user",
			Content: summaryPrompt,
		},
	}
	
	response, err := m.chatAgent.Request(
		context.Background(),
		&agent.LLMChatParams{
			Messages: agentMessages,
		},
		m.botConfig,
		nil, // No streaming for summary
	)
	
	if err != nil {
		return fmt.Errorf("failed to generate summary: %w", err)
	}
	
	// Extract summary text from response
	summaryText := ""
	if len(response.Messages) > 0 {
		switch v := response.Messages[0].Content.(type) {
		case string:
			summaryText = v
		default:
			summaryText = fmt.Sprintf("%v", v)
		}
	}
	
	if summaryText == "" {
		return fmt.Errorf("received empty summary from AI")
	}
	
	// Create summary message
	summaryMessage := storage.HistoryItem{
		Role:      "summary",
		Content:   fmt.Sprintf("Previous conversation summary: %s", summaryText),
		Timestamp: time.Now().Unix(),
		Truncated: false,
	}
	
	// Rebuild history: system + old summaries + new summary + recent messages
	newHistory := make([]storage.HistoryItem, 0)
	newHistory = append(newHistory, systemMessages...)
	newHistory = append(newHistory, existingSummaries...)
	newHistory = append(newHistory, summaryMessage)
	newHistory = append(newHistory, recentMessages...)
	
	// Get full history to preserve truncation markers and other metadata
	fullHistory, err := m.storage.GetChatHistory(ctx)
	if err != nil {
		return fmt.Errorf("failed to get full history: %w", err)
	}
	
	// Find where the build history starts in full history
	// We need to preserve any truncation markers
	var finalHistory []storage.HistoryItem
	lastTruncationIndex := -1
	for i := len(fullHistory) - 1; i >= 0; i-- {
		if fullHistory[i].Truncated {
			lastTruncationIndex = i
			break
		}
	}
	
	if lastTruncationIndex >= 0 {
		// Keep everything up to and including the truncation marker
		finalHistory = append(finalHistory, fullHistory[:lastTruncationIndex+1]...)
	}
	
	// Add the new summarized history
	finalHistory = append(finalHistory, newHistory...)
	
	// Save updated history
	if err := m.storage.SaveChatHistory(ctx, finalHistory); err != nil {
		return fmt.Errorf("failed to save summarized history: %w", err)
	}
	
	return nil
}

// ClearHistory creates a truncation marker without deleting history
func (m *ContextManager) ClearHistory(ctx *storage.SessionContext) error {
	// Get current history
	history, err := m.storage.GetChatHistory(ctx)
	if err != nil {
		return fmt.Errorf("failed to get chat history: %w", err)
	}
	
	// Create truncation marker
	truncationMarker := storage.HistoryItem{
		Role:      "system",
		Content:   "[Conversation cleared by user]",
		Timestamp: time.Now().Unix(),
		Truncated: true,
	}
	
	// Append truncation marker to history
	history = append(history, truncationMarker)
	
	// Save updated history
	if err := m.storage.SaveChatHistory(ctx, history); err != nil {
		return fmt.Errorf("failed to save history with truncation marker: %w", err)
	}
	
	return nil
}
