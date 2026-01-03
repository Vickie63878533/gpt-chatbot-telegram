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
	RegisterChatAgent(&AzureChatAgent{})
	RegisterImageAgent(&AzureImageAgent{})
}

// AzureChatAgent implements ChatAgent for Azure OpenAI
type AzureChatAgent struct{}

func (a *AzureChatAgent) Name() string {
	return "azure"
}

func (a *AzureChatAgent) ModelKey() string {
	return "AZURE_CHAT_MODEL"
}

func (a *AzureChatAgent) Enable(cfg *config.Config) bool {
	return cfg.AzureAPIKey != "" && cfg.AzureResourceName != ""
}

func (a *AzureChatAgent) Model(cfg *config.Config) string {
	return cfg.AzureChatModel
}

func (a *AzureChatAgent) ModelList(cfg *config.Config) ([]string, error) {
	if cfg.AzureChatModelsList != "" {
		var models []string
		if err := json.Unmarshal([]byte(cfg.AzureChatModelsList), &models); err != nil {
			return nil, fmt.Errorf("failed to parse AZURE_CHAT_MODELS_LIST: %w", err)
		}
		return models, nil
	}
	// Default models
	return []string{"gpt-4o", "gpt-4o-mini", "gpt-4-turbo", "gpt-35-turbo"}, nil
}

func (a *AzureChatAgent) Request(ctx context.Context, params *LLMChatParams, cfg *config.Config, onStream ChatStreamTextHandler) (*ChatAgentResponse, error) {
	// Build Azure OpenAI endpoint
	// Format: https://{resource-name}.openai.azure.com/openai/deployments/{deployment-id}/chat/completions?api-version={api-version}
	endpoint := fmt.Sprintf("https://%s.openai.azure.com/openai/deployments/%s/chat/completions?api-version=%s",
		cfg.AzureResourceName,
		a.Model(cfg),
		cfg.AzureAPIVersion,
	)

	// Build messages
	messages := make([]map[string]interface{}, 0, len(params.Messages)+1)
	if params.Prompt != "" {
		messages = append(messages, map[string]interface{}{
			"role":    "system",
			"content": params.Prompt,
		})
	}
	for _, msg := range params.Messages {
		messages = append(messages, map[string]interface{}{
			"role":    msg.Role,
			"content": msg.Content,
		})
	}

	// Build request body
	reqBody := map[string]interface{}{
		"messages": messages,
		"stream":   onStream != nil,
	}

	// Add extra parameters
	if cfg.AzureChatExtraParams != nil {
		for k, v := range cfg.AzureChatExtraParams {
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
	req.Header.Set("api-key", cfg.AzureAPIKey)

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

func (a *AzureChatAgent) handleStreamResponse(body io.Reader, onStream ChatStreamTextHandler) (*ChatAgentResponse, error) {
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

		// Extract delta content
		if choices, ok := line["choices"].([]interface{}); ok && len(choices) > 0 {
			if choice, ok := choices[0].(map[string]interface{}); ok {
				if delta, ok := choice["delta"].(map[string]interface{}); ok {
					if content, ok := delta["content"].(string); ok {
						fullText.WriteString(content)
						if err := onStream(content); err != nil {
							return nil, fmt.Errorf("stream handler error: %w", err)
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

func (a *AzureChatAgent) handleNonStreamResponse(body io.Reader) (*ChatAgentResponse, error) {
	var response struct {
		Choices []struct {
			Message struct {
				Role    string `json:"role"`
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
	}

	if err := json.NewDecoder(body).Decode(&response); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	if len(response.Choices) == 0 {
		return nil, fmt.Errorf("no choices in response")
	}

	return &ChatAgentResponse{
		Messages: []HistoryItem{
			{
				Role:    response.Choices[0].Message.Role,
				Content: response.Choices[0].Message.Content,
			},
		},
	}, nil
}

// AzureImageAgent implements ImageAgent for Azure DALL-E
type AzureImageAgent struct{}

func (a *AzureImageAgent) Name() string {
	return "azure-dalle"
}

func (a *AzureImageAgent) ModelKey() string {
	return "AZURE_IMAGE_MODEL"
}

func (a *AzureImageAgent) Enable(cfg *config.Config) bool {
	return cfg.AzureAPIKey != "" && cfg.AzureResourceName != ""
}

func (a *AzureImageAgent) Model(cfg *config.Config) string {
	return cfg.AzureImageModel
}

func (a *AzureImageAgent) ModelList(cfg *config.Config) ([]string, error) {
	// Azure typically uses deployment names, not model names
	return []string{"dall-e-3", "dall-e-2"}, nil
}

func (a *AzureImageAgent) Request(ctx context.Context, prompt string, cfg *config.Config) (string, error) {
	// Build Azure OpenAI endpoint for image generation
	endpoint := fmt.Sprintf("https://%s.openai.azure.com/openai/deployments/%s/images/generations?api-version=%s",
		cfg.AzureResourceName,
		a.Model(cfg),
		cfg.AzureAPIVersion,
	)

	// Build request body
	reqBody := map[string]interface{}{
		"prompt": prompt,
		"n":      1,
		"size":   cfg.DallEImageSize,
	}

	// DALL-E 3 specific parameters
	if strings.Contains(a.Model(cfg), "dall-e-3") {
		reqBody["quality"] = cfg.DallEImageQuality
		reqBody["style"] = cfg.DallEImageStyle
	}

	bodyBytes, err := json.Marshal(reqBody)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request: %w", err)
	}

	// Create HTTP request
	req, err := http.NewRequestWithContext(ctx, "POST", endpoint, bytes.NewReader(bodyBytes))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("api-key", cfg.AzureAPIKey)

	// Send request
	client := CreateHTTPClient(cfg)
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(body))
	}

	// Parse response
	var response struct {
		Data []struct {
			URL     string `json:"url"`
			B64JSON string `json:"b64_json"`
		} `json:"data"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return "", fmt.Errorf("failed to decode response: %w", err)
	}

	if len(response.Data) == 0 {
		return "", fmt.Errorf("no image data in response")
	}

	// Return URL or base64 data
	if response.Data[0].URL != "" {
		return response.Data[0].URL, nil
	}
	return response.Data[0].B64JSON, nil
}
