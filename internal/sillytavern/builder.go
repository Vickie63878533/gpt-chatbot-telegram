package sillytavern

import (
	"log"

	"github.com/tbxark/ChatGPT-Telegram-Workers/go_version/internal/storage"
)

// RequestBuilder builds AI requests by combining SillyTavern components
type RequestBuilder struct {
	characterManager *CharacterCardManager
	worldBookManager *WorldBookManager
	presetManager    *PresetManager
	regexProcessor   *RegexProcessor
}

// NewRequestBuilder creates a new RequestBuilder with all dependencies
func NewRequestBuilder(
	characterManager *CharacterCardManager,
	worldBookManager *WorldBookManager,
	presetManager *PresetManager,
	regexProcessor *RegexProcessor,
) *RequestBuilder {
	return &RequestBuilder{
		characterManager: characterManager,
		worldBookManager: worldBookManager,
		presetManager:    presetManager,
		regexProcessor:   regexProcessor,
	}
}

// BuildContext contains the context needed to build a request
type BuildContext struct {
	UserID       *int64                // User ID for loading user-specific configurations
	History      []storage.HistoryItem // Conversation history
	CurrentInput string                // Current user input
	APIType      string                // API type (e.g., "openai", "anthropic")
}

// AIRequest represents an AI request in OpenAI format (intermediate representation)
type AIRequest struct {
	Messages         []Message `json:"messages"`
	Model            string    `json:"model,omitempty"`
	Temperature      float64   `json:"temperature,omitempty"`
	MaxTokens        int       `json:"max_tokens,omitempty"`
	TopP             float64   `json:"top_p,omitempty"`
	TopK             int       `json:"top_k,omitempty"`
	PresencePenalty  float64   `json:"presence_penalty,omitempty"`
	FrequencyPenalty float64  `json:"frequency_penalty,omitempty"`
	StopSequences    []string `json:"stop,omitempty"`
}

// Message represents a single message in the conversation
type Message struct {
	Role    string `json:"role"`    // "system", "user", or "assistant"
	Content string `json:"content"` // Message content
}

// BuildRequest builds an AI request from the context
// This is the main method that orchestrates all SillyTavern components
func (b *RequestBuilder) BuildRequest(ctx *BuildContext) (*AIRequest, error) {
	log.Printf("[RequestBuilder] Starting request build for user: %v, API type: %s", ctx.UserID, ctx.APIType)

	// 1. Apply input regex transformations
	processedInput := ctx.CurrentInput
	if b.regexProcessor != nil {
		var err error
		processedInput, err = b.regexProcessor.ProcessInput(ctx.UserID, ctx.CurrentInput)
		if err != nil {
			log.Printf("[RequestBuilder] Error applying input regex: %v", err)
			// Log error but continue with original input
			processedInput = ctx.CurrentInput
		} else if processedInput != ctx.CurrentInput {
			log.Printf("[RequestBuilder] Input transformed by regex")
		}
	}

	// 2. Load active character card
	var characterData *CharacterCardV2
	if b.characterManager != nil {
		card, err := b.characterManager.GetActiveCard(ctx.UserID)
		if err == nil && card != nil {
			log.Printf("[RequestBuilder] Loaded active character card: %s (ID: %d)", card.Name, card.ID)
			// Parse character data
			characterData, _ = b.characterManager.ParseCardData(card.Data)
		} else if err != nil {
			log.Printf("[RequestBuilder] No active character card found: %v", err)
		}
	}

	// 3. Load active world book
	var worldBook *storage.WorldBook
	var triggeredEntries []*storage.WorldBookEntry
	if b.worldBookManager != nil {
		book, err := b.worldBookManager.GetActiveBook(ctx.UserID)
		if err == nil && book != nil {
			worldBook = book
			log.Printf("[RequestBuilder] Loaded active world book: %s (ID: %d)", book.Name, book.ID)
		} else if err != nil {
			log.Printf("[RequestBuilder] No active world book found: %v", err)
		}
	}

	// 4. Load active preset
	var presetData *PresetData
	if b.presetManager != nil && ctx.APIType != "" {
		p, err := b.presetManager.GetActivePreset(ctx.UserID, ctx.APIType)
		if err == nil && p != nil {
			log.Printf("[RequestBuilder] Loaded active preset: %s (ID: %d)", p.Name, p.ID)
			// Parse preset data
			presetData, _ = b.presetManager.ParsePresetData(p.Data)
		} else if err != nil {
			log.Printf("[RequestBuilder] No active preset found for API type %s: %v", ctx.APIType, err)
		}
	}

	// 5. Build the request
	request := &AIRequest{
		Messages: []Message{},
	}

	// 6. Build system prompt from character card
	systemPrompt := b.buildSystemPrompt(characterData)
	if systemPrompt != "" {
		request.Messages = append(request.Messages, Message{
			Role:    "system",
			Content: systemPrompt,
		})
		log.Printf("[RequestBuilder] Added system prompt from character card (%d chars)", len(systemPrompt))
	}

	// 7. Inject world book entries
	history := ctx.History
	if worldBook != nil {
		// Get triggered entries for logging
		triggeredEntries, _ = b.worldBookManager.TriggerEntries(worldBook.ID, ctx.History)
		if len(triggeredEntries) > 0 {
			log.Printf("[RequestBuilder] Triggered %d world book entries:", len(triggeredEntries))
			for _, entry := range triggeredEntries {
				log.Printf("  - Entry UID: %s, Position: %s, Order: %d", entry.UID, entry.Position, entry.Order)
			}
		}
		history = b.injectWorldBookEntries(worldBook, ctx.History)
	}

	// 8. Convert history to messages with role alternation
	messages := b.enforceRoleAlternation(history, processedInput)
	request.Messages = append(request.Messages, messages...)
	log.Printf("[RequestBuilder] Built %d messages with strict role alternation", len(messages))

	// 9. Apply preset parameters
	if presetData != nil {
		b.applyPresetParameters(request, presetData)
		log.Printf("[RequestBuilder] Applied preset parameters: temp=%.2f, top_p=%.2f, max_tokens=%d",
			request.Temperature, request.TopP, request.MaxTokens)
	}

	// 10. Log final message sequence
	log.Printf("[RequestBuilder] Final message sequence:")
	for i, msg := range request.Messages {
		contentPreview := msg.Content
		if len(contentPreview) > 100 {
			contentPreview = contentPreview[:100] + "..."
		}
		log.Printf("  [%d] Role: %s, Content: %s", i, msg.Role, contentPreview)
	}

	log.Printf("[RequestBuilder] Request build completed successfully")
	return request, nil
}

// buildSystemPrompt constructs the system prompt from character card data
func (b *RequestBuilder) buildSystemPrompt(characterData *CharacterCardV2) string {
	if characterData == nil {
		return ""
	}

	var prompt string

	// Use system_prompt if available
	if characterData.Data.SystemPrompt != "" {
		prompt = characterData.Data.SystemPrompt
	} else {
		// Build from character description, personality, and scenario
		if characterData.Data.Description != "" {
			prompt += characterData.Data.Description + "\n\n"
		}
		if characterData.Data.Personality != "" {
			prompt += "Personality: " + characterData.Data.Personality + "\n\n"
		}
		if characterData.Data.Scenario != "" {
			prompt += "Scenario: " + characterData.Data.Scenario + "\n\n"
		}
	}

	// Add post history instructions if available
	if characterData.Data.PostHistoryInstructions != "" {
		prompt += "\n" + characterData.Data.PostHistoryInstructions
	}

	return prompt
}

// injectWorldBookEntries injects world book entries into the history
// This scans the history for trigger keywords and injects matching entries
func (b *RequestBuilder) injectWorldBookEntries(worldBook *storage.WorldBook, history []storage.HistoryItem) []storage.HistoryItem {
	// Get triggered entries
	triggeredEntries, err := b.worldBookManager.TriggerEntries(worldBook.ID, history)
	if err != nil || len(triggeredEntries) == 0 {
		return history
	}

	// Separate entries by position
	beforeCharEntries := make([]*storage.WorldBookEntry, 0)
	afterCharEntries := make([]*storage.WorldBookEntry, 0)

	for _, entry := range triggeredEntries {
		if entry.Position == "before_char" {
			beforeCharEntries = append(beforeCharEntries, entry)
		} else {
			afterCharEntries = append(afterCharEntries, entry)
		}
	}

	// Build new history with injected entries
	newHistory := make([]storage.HistoryItem, 0, len(history)+len(triggeredEntries))

	// Add before_char entries at the beginning (after system messages)
	systemMessageCount := 0
	for i, item := range history {
		if item.Role == "system" || item.Role == "summary" {
			newHistory = append(newHistory, item)
			systemMessageCount = i + 1
		} else {
			break
		}
	}

	// Inject before_char entries
	for _, entry := range beforeCharEntries {
		newHistory = append(newHistory, storage.HistoryItem{
			Role:    "system",
			Content: entry.Content,
		})
	}

	// Add the rest of the history
	newHistory = append(newHistory, history[systemMessageCount:]...)

	// Inject after_char entries at the end (before the current user message)
	if len(afterCharEntries) > 0 {
		// Find the last position before user messages
		insertPos := len(newHistory)
		for i := len(newHistory) - 1; i >= 0; i-- {
			if newHistory[i].Role != "user" {
				insertPos = i + 1
				break
			}
		}

		// Insert after_char entries
		afterEntries := make([]storage.HistoryItem, 0, len(afterCharEntries))
		for _, entry := range afterCharEntries {
			afterEntries = append(afterEntries, storage.HistoryItem{
				Role:    "system",
				Content: entry.Content,
			})
		}

		// Rebuild history with injected entries
		result := make([]storage.HistoryItem, 0, len(newHistory)+len(afterEntries))
		result = append(result, newHistory[:insertPos]...)
		result = append(result, afterEntries...)
		result = append(result, newHistory[insertPos:]...)
		newHistory = result
	}

	return newHistory
}

// enforceRoleAlternation ensures strict user/assistant role alternation
// This merges consecutive messages with the same role and handles edge cases
func (b *RequestBuilder) enforceRoleAlternation(history []storage.HistoryItem, currentInput string) []Message {
	messages := make([]Message, 0)

	// Convert history items to messages, skipping system messages (already added)
	for _, item := range history {
		// Skip system messages and summaries (they're already in the request)
		if item.Role == "system" || item.Role == "summary" {
			continue
		}

		// Skip truncated markers
		if item.Truncated {
			continue
		}

		// Convert content to string
		contentStr := ""
		switch v := item.Content.(type) {
		case string:
			contentStr = v
		case []storage.ContentPart:
			// Concatenate text parts
			for _, part := range v {
				if part.Type == "text" {
					contentStr += part.Text
				}
			}
		}

		if contentStr == "" {
			continue
		}

		messages = append(messages, Message{
			Role:    item.Role,
			Content: contentStr,
		})
	}

	// Add current input as user message
	if currentInput != "" {
		messages = append(messages, Message{
			Role:    "user",
			Content: currentInput,
		})
	}

	// Now enforce strict alternation
	if len(messages) == 0 {
		return messages
	}

	result := make([]Message, 0, len(messages))
	
	for i, msg := range messages {
		// Only keep user and assistant messages
		if msg.Role != "user" && msg.Role != "assistant" {
			continue
		}

		// If this is the first message
		if len(result) == 0 {
			// First message must be from user
			if msg.Role != "user" {
				// Add empty user message
				result = append(result, Message{
					Role:    "user",
					Content: "[conversation start]",
				})
			}
			result = append(result, msg)
			continue
		}

		lastMsg := &result[len(result)-1]

		// If same role as previous, merge the content
		if msg.Role == lastMsg.Role {
			lastMsg.Content += "\n\n" + msg.Content
		} else {
			// Different role, add as new message
			result = append(result, msg)
		}

		// Ensure we don't have more than one consecutive message of the same role
		if i == len(messages)-1 {
			// Last message must be from user
			if result[len(result)-1].Role != "user" {
				result = append(result, Message{
					Role:    "user",
					Content: currentInput,
				})
			}
		}
	}

	// Final validation: ensure alternation
	validated := make([]Message, 0, len(result))
	for i, msg := range result {
		if i == 0 {
			// First must be user
			if msg.Role != "user" {
				validated = append(validated, Message{
					Role:    "user",
					Content: "[conversation start]",
				})
			}
			validated = append(validated, msg)
		} else {
			// Check alternation
			if validated[len(validated)-1].Role == msg.Role {
				// Merge with previous
				validated[len(validated)-1].Content += "\n\n" + msg.Content
			} else {
				validated = append(validated, msg)
			}
		}
	}

	return validated
}

// applyPresetParameters applies preset parameters to the request
func (b *RequestBuilder) applyPresetParameters(request *AIRequest, presetData *PresetData) {
	if presetData == nil {
		return
	}

	// Apply temperature
	if presetData.Temperature > 0 {
		request.Temperature = presetData.Temperature
	}

	// Apply top_p
	if presetData.TopP > 0 {
		request.TopP = presetData.TopP
	}

	// Apply top_k
	if presetData.TopK > 0 {
		request.TopK = presetData.TopK
	}

	// Apply max_tokens
	if presetData.MaxTokens > 0 {
		request.MaxTokens = presetData.MaxTokens
	}

	// Apply presence_penalty
	if presetData.PresencePenalty != 0 {
		request.PresencePenalty = presetData.PresencePenalty
	}

	// Apply frequency_penalty
	if presetData.FrequencyPenalty != 0 {
		request.FrequencyPenalty = presetData.FrequencyPenalty
	}

	// Apply stop sequences
	if len(presetData.StopSequences) > 0 {
		request.StopSequences = presetData.StopSequences
	}
}
