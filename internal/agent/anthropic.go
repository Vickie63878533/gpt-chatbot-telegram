package agent

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/tbxark/ChatGPT-Telegram-Workers/go_version/internal/config"
)

func init() {
	RegisterChatAgent(&AnthropicChatAgent{})
}

// AnthropicChatAgent implements ChatAgent for Anthropic Claude
type AnthropicChatAgent struct{}

func (a *AnthropicChatAgent) Name() string {
	return "anthropic"
}

func (a *AnthropicChatAgent) ModelKey() string {
	return "ANTHROPIC_CHAT_MODEL"
}

func (a *AnthropicChatAgent) Enable(cfg *config.Config) bool {
	return cfg.AnthropicAPIKey != ""
}

func (a *AnthropicChatAgent) Model(cfg *config.Config) string {
	return cfg.AnthropicChatModel
}

func (a *AnthropicChatAgent) ModelList(cfg *config.Config) ([]string, error) {
	if cfg.AnthropicChatModelsList != "" {
		var models []string
		if err := json.Unmarshal([]byte(cfg.AnthropicChatModelsList), &models); err != nil {
			return nil, fmt.Errorf("failed to parse ANTHROPIC_CHAT_MODELS_LIST: %w", err)
		}
		return models, nil
	}
	// Default models
	return []string{
		"claude-3-5-sonnet-latest",
		"claude-3-5-haiku-latest",
		"claude-3-opus-latest",
		"claude-3-sonnet-20240229",
		"claude-3-haiku-20240307",
	}, nil
}

func (a *AnthropicChatAgent) Request(ctx context.Context, params *LLMChatParams, cfg *config.Config, onStream ChatStreamTextHandler) (*ChatAgentResponse, error) {
	apiBase := cfg.AnthropicAPIBase
	if !strings.HasSuffix(apiBase, "/") {
		apiBase += "/"
	}

	endpoint := apiBase + "messages"

	// Convert messages to Anthropic format
	messages := make([]map[string]interface{}, 0)
	systemPrompt := ""

	// Extract system message
	if params.Prompt != "" {
		systemPrompt = params.Prompt
	}

	// Convert history
	for _, msg := range params.Messages {
		if msg.Role == "system" {
			// Anthropic uses a separate system parameter
			if content, ok := msg.Content.(string); ok {
				if systemPrompt != "" {
					systemPrompt += "\n"
				}
				systemPrompt += content
			}
			continue
		}

		content := msg.Content
		// Convert content to Anthropic format
		var contentArray []map[string]interface{}
		switch v := content.(type) {
		case string:
			contentArray = []map[string]interface{}{
				{
					"type": "text",
					"text": v,
				},
			}
		case []ContentPart:
			for _, part := range v {
				if part.Type == "text" {
					contentArray = append(contentArray, map[string]interface{}{
						"type": "text",
						"text": part.Text,
					})
				} else if part.Type == "image" {
					// Anthropic expects image format
					imageData := part.Image
					mediaType := "image/jpeg"

					// Check if it's a URL or base64
					if strings.HasPrefix(imageData, "http://") || strings.HasPrefix(imageData, "https://") {
						contentArray = append(contentArray, map[string]interface{}{
							"type": "image",
							"source": map[string]interface{}{
								"type": "url",
								"url":  imageData,
							},
						})
					} else {
						contentArray = append(contentArray, map[string]interface{}{
							"type": "image",
							"source": map[string]interface{}{
								"type":       "base64",
								"media_type": mediaType,
								"data":       imageData,
							},
						})
					}
				}
			}
		}

		messages = append(messages, map[string]interface{}{
			"role":    msg.Role,
			"content": contentArray,
		})
	}

	// Build request body
	reqBody := map[string]interface{}{
		"model":      a.Model(cfg),
		"messages":   messages,
		"max_tokens": 4096, // Required by Anthropic
		"stream":     onStream != nil,
	}

	if systemPrompt != "" {
		reqBody["system"] = systemPrompt
	}

	// Add extra parameters
	if cfg.AnthropicChatExtraParams != nil {
		for k, v := range cfg.AnthropicChatExtraParams {
			reqBody[k] = v
		}
	}

	bodyBytes, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	// Create HTTP request
	req, err := http.NewRequestWithContext(ctx, "POST", endpoint, bytes.NewReader(bodyBytes))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-api-key", cfg.AnthropicAPIKey)
	req.Header.Set("anthropic-version", "2023-06-01")

	// Send request
	client := CreateHTTPClient(cfg)
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(body))
	}

	// Handle streaming response
	if onStream != nil {
		return a.handleStreamResponse(resp.Body, onStream)
	}

	// Handle non-streaming response
	return a.handleNonStreamResponse(resp.Body)
}

func (a *AnthropicChatAgent) handleStreamResponse(body io.Reader, onStream ChatStreamTextHandler) (*ChatAgentResponse, error) {
	var fullText strings.Builder
	decoder := json.NewDecoder(body)

	for {
		var event map[string]interface{}
		if err := decoder.Decode(&event); err != nil {
			if err == io.EOF {
				break
			}
			return nil, fmt.Errorf("failed to decode stream: %w", err)
		}

		// Anthropic uses event-based streaming
		eventType, _ := event["type"].(string)

		if eventType == "content_block_delta" {
			if delta, ok := event["delta"].(map[string]interface{}); ok {
				if text, ok := delta["text"].(string); ok {
					fullText.WriteString(text)
					if err := onStream(text); err != nil {
						return nil, fmt.Errorf("stream handler error: %w", err)
					}
				}
			}
		}
	}

	return &ChatAgentResponse{
		Messages: []HistoryItem{
			{
				Role:    "assistant",
				Content: fullText.String(),
			},
		},
	}, nil
}

func (a *AnthropicChatAgent) handleNonStreamResponse(body io.Reader) (*ChatAgentResponse, error) {
	var response struct {
		Content []struct {
			Type string `json:"type"`
			Text string `json:"text"`
		} `json:"content"`
	}

	if err := json.NewDecoder(body).Decode(&response); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	if len(response.Content) == 0 {
		return nil, fmt.Errorf("no content in response")
	}

	// Concatenate all text blocks
	var fullText strings.Builder
	for _, block := range response.Content {
		if block.Type == "text" {
			fullText.WriteString(block.Text)
		}
	}

	return &ChatAgentResponse{
		Messages: []HistoryItem{
			{
				Role:    "assistant",
				Content: fullText.String(),
			},
		},
	}, nil
}
