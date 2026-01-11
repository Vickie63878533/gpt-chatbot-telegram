package telegraph

import (
	"testing"
)

func TestNewClient(t *testing.T) {
	// Note: This test requires network access to Telegraph API
	// Skip in CI/CD environments without network access
	if testing.Short() {
		t.Skip("Skipping network test in short mode")
	}

	client, err := NewClient()
	if err != nil {
		t.Fatalf("Failed to create Telegraph client: %v", err)
	}

	if client == nil {
		t.Fatal("Client is nil")
	}

	if client.accessToken == "" {
		t.Fatal("Access token is empty")
	}
}

func TestFormatConversation(t *testing.T) {
	messages := []ConversationMessage{
		{Role: "user", Content: "Hello, how are you?"},
		{Role: "assistant", Content: "I'm doing well, thank you!"},
		{Role: "user", Content: "What's the weather like?"},
		{Role: "assistant", Content: "I don't have access to real-time weather data."},
	}

	html := FormatConversation(messages)

	// Check that HTML contains expected elements
	if html == "" {
		t.Fatal("Formatted HTML is empty")
	}

	t.Logf("Generated HTML:\n%s", html)

	// Check for user messages
	if !contains(html, "👤 User:") {
		t.Error("HTML doesn't contain user marker")
	}

	// Check for assistant messages
	if !contains(html, "🤖 Assistant:") {
		t.Error("HTML doesn't contain assistant marker")
	}

	// Check for message content (escaped)
	if !contains(html, "Hello, how are you?") {
		t.Error("HTML doesn't contain first user message")
	}

	// The content is HTML escaped, so apostrophe becomes &#39;
	if !contains(html, "I&#39;m doing well, thank you!") {
		t.Errorf("HTML doesn't contain first assistant message (expected escaped version)")
	}
}

func TestEscapeHTML(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "Basic text",
			input:    "Hello World",
			expected: "Hello World",
		},
		{
			name:     "HTML tags",
			input:    "<script>alert('xss')</script>",
			expected: "&lt;script&gt;alert(&#39;xss&#39;)&lt;/script&gt;",
		},
		{
			name:     "Ampersand",
			input:    "Tom & Jerry",
			expected: "Tom &amp; Jerry",
		},
		{
			name:     "Quotes",
			input:    `He said "Hello"`,
			expected: "He said &quot;Hello&quot;",
		},
		{
			name:     "Newlines",
			input:    "Line 1\nLine 2",
			expected: "Line 1<br>Line 2",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := escapeHTML(tt.input)
			if result != tt.expected {
				t.Errorf("escapeHTML() = %q, want %q", result, tt.expected)
			}
		})
	}
}

func TestCreatePage(t *testing.T) {
	// Note: This test requires network access to Telegraph API
	// Skip in CI/CD environments without network access
	if testing.Short() {
		t.Skip("Skipping network test in short mode")
	}

	client, err := NewClient()
	if err != nil {
		t.Fatalf("Failed to create Telegraph client: %v", err)
	}

	title := "Test Page"
	content := "<p>This is a test page</p>"

	url, err := client.CreatePage(title, content)
	if err != nil {
		t.Fatalf("Failed to create page: %v", err)
	}

	if url == "" {
		t.Fatal("URL is empty")
	}

	// Check that URL is valid Telegraph URL
	if !contains(url, "telegra.ph") {
		t.Errorf("URL doesn't contain telegra.ph: %s", url)
	}
}

// Helper function
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > len(substr) && containsHelper(s, substr))
}

func containsHelper(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
