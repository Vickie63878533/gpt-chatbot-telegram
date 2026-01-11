package telegraph

import (
	"fmt"
	"strings"
)

// ConversationMessage represents a message in the conversation
type ConversationMessage struct {
	Role    string
	Content string
}

// FormatConversation formats a conversation history as HTML for Telegraph
// It converts messages to Telegraph-compatible HTML format with proper escaping
func FormatConversation(history []ConversationMessage) string {
	var sb strings.Builder

	for _, msg := range history {
		// Escape HTML special characters
		content := escapeHTML(msg.Content)

		// Format based on role
		switch msg.Role {
		case "user":
			sb.WriteString(fmt.Sprintf("<p><strong>👤 User:</strong></p><p>%s</p>", content))
		case "assistant":
			sb.WriteString(fmt.Sprintf("<p><strong>🤖 Assistant:</strong></p><p>%s</p>", content))
		}

		sb.WriteString("<hr>")
	}

	return sb.String()
}

// escapeHTML escapes HTML special characters to prevent XSS attacks
// It handles all common HTML entities and preserves newlines as <br> tags
func escapeHTML(s string) string {
	// Order matters: & must be escaped first to avoid double-escaping
	s = strings.ReplaceAll(s, "&", "&amp;")
	s = strings.ReplaceAll(s, "<", "&lt;")
	s = strings.ReplaceAll(s, ">", "&gt;")
	s = strings.ReplaceAll(s, "\"", "&quot;")
	s = strings.ReplaceAll(s, "'", "&#39;")
	// Preserve newlines as <br>
	s = strings.ReplaceAll(s, "\n", "<br>")
	return s
}
