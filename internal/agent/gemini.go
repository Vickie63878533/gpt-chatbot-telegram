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
	RegisterChatAgent(&GeminiChatAgent{})
}

// GeminiChatAgent implements ChatAgent for Google Gemini
type GeminiChatAgent struct{}

func (a *GeminiChatAgent) Name() string {
	return "gemini"
}

func (a *GeminiChatAgent) ModelKey() string {
	return "GOOGLE_CHAT_MODEL"
}

func (a *GeminiChatAgent) Enable(cfg *config.Config) bool {
	return cfg.GoogleAPIKey != ""
}

func (a *GeminiChatAgent) Model(cfg *config.Config) string {
	return cfg.GoogleChatModel
}

func (a *GeminiChatAgent) ModelList(cfg *config.Config) ([]string, error) {
	if cfg.GoogleChatModelsList != "" {
		var models []string
		if err := json.Unmarshal([]byte(cfg.GoogleChatModelsList), &models); err != nil {
			return nil, fmt.Errorf("failed to parse GOOGLE_CHAT_MODELS_LIST: %w", err)
		}
		return models, nil
	}
	// Default models
	return []string{"gemini-1.5-flash", "gemini-1.5-pro", "gemini-pro"}, nil
}

func (a *GeminiChatAgent) Request(ctx context.Context, params *LLMChatParams, cfg *config.Config, onStream ChatStreamTextHandler) (*ChatAgentResponse, error) {
	apiBase := cfg.GoogleAPIBase
	if !strings.HasSuffix(apiBase, "/") {
		apiBase += "/"
	}

	// Build endpoint
	endpoint := fmt.Sprintf("%smodels/%s:generateContent?key=%s",
		apiBase,
		a.Model(cfg),
		cfg.GoogleAPIKey,
	)

	if onStream != nil {
		endpoint = fmt.Sprintf("%smodels/%s:streamGenerateContent?key=%s&alt=sse",
			apiBase,
			a.Model(cfg),
			cfg.GoogleAPIKey,
		)
	}

	// Convert messages to Gemini format
	contents := make([]map[string]interface{}, 0)
	systemInstruction := ""

	// Extract system message if present
	if params.Prompt != "" {
		systemInstruction = params.Prompt
	}

	// Convert history to Gemini format
	for _, msg := range params.Messages {
		role := msg.Role
		if role == "assistant" {
			role = "model"
		}
		if role == "system" {
			// Gemini doesn't support system role in contents, add to system instruction
			if content, ok := msg.Content.(string); ok {
				if systemInstruction != "" {
					systemInstruction += "\n"
				}
				systemInstruction += content
			}
			continue
		}

		content := msg.Content
		// Convert content to Gemini format
		var parts []map[string]interface{}
		switch v := content.(type) {
		case string:
			parts = []map[string]interface{}{
				{"text": v},
			}
		case []ContentPart:
			for _, part := range v {
				if part.Type == "text" {
					parts = append(parts, map[string]interface{}{
						"text": part.Text,
					})
				} else if part.Type == "image" {
					// Gemini expects inline_data format
					parts = append(parts, map[string]interface{}{
						"inline_data": map[string]interface{}{
							"mime_type": "image/jpeg",
							"data":      part.Image,
						},
					})
				}
			}
		}

		contents = append(contents, map[string]interface{}{
			"role":  role,
			"parts": parts,
		})
	}

	// Build request body
	reqBody := map[string]interface{}{
		"contents": contents,
	}

	if systemInstruction != "" {
		reqBody["systemInstruction"] = map[string]interface{}{
			"parts": []map[string]interface{}{
				{"text": systemInstruction},
			},
		}
	}

	// Add extra parameters
	if cfg.GoogleChatExtraParams != nil {
		for k, v := range cfg.GoogleChatExtraParams {
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

func (a *GeminiChatAgent) handleStreamResponse(body io.Reader, onStream ChatStreamTextHandler) (*ChatAgentResponse, error) {
	var fullText strings.Builder
	decoder := json.NewDecoder(body)

	for {
		var line map[string]interface{}
		if err := decoder.Decode(&line); err != nil {
			if err == io.EOF {
				break
			}
			return nil, fmt.Errorf("failed to decode stream: %w", err)
		}

		// Extract content from candidates
		if candidates, ok := line["candidates"].([]interface{}); ok && len(candidates) > 0 {
			if candidate, ok := candidates[0].(map[string]interface{}); ok {
				if content, ok := candidate["content"].(map[string]interface{}); ok {
					if parts, ok := content["parts"].([]interface{}); ok {
						for _, part := range parts {
							if partMap, ok := part.(map[string]interface{}); ok {
								if text, ok := partMap["text"].(string); ok {
									fullText.WriteString(text)
									if err := onStream(text); err != nil {
										return nil, fmt.Errorf("stream handler error: %w", err)
									}
								}
							}
						}
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

func (a *GeminiChatAgent) handleNonStreamResponse(body io.Reader) (*ChatAgentResponse, error) {
	var response struct {
		Candidates []struct {
			Content struct {
				Parts []struct {
					Text string `json:"text"`
				} `json:"parts"`
			} `json:"content"`
		} `json:"candidates"`
	}

	if err := json.NewDecoder(body).Decode(&response); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	if len(response.Candidates) == 0 {
		return nil, fmt.Errorf("no candidates in response")
	}

	// Concatenate all text parts
	var fullText strings.Builder
	for _, part := range response.Candidates[0].Content.Parts {
		fullText.WriteString(part.Text)
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
