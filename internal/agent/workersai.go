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
	RegisterChatAgent(&WorkersChatAgent{})
	RegisterImageAgent(&WorkersImageAgent{})
}

// WorkersChatAgent implements ChatAgent for Cloudflare Workers AI
type WorkersChatAgent struct{}

func (a *WorkersChatAgent) Name() string {
	return "workers"
}

func (a *WorkersChatAgent) ModelKey() string {
	return "WORKERS_CHAT_MODEL"
}

func (a *WorkersChatAgent) Enable(cfg *config.Config) bool {
	return cfg.CloudflareAccountID != "" && cfg.CloudflareToken != ""
}

func (a *WorkersChatAgent) Model(cfg *config.Config) string {
	return cfg.WorkersChatModel
}

func (a *WorkersChatAgent) ModelList(cfg *config.Config) ([]string, error) {
	if cfg.WorkersChatModelsList != "" {
		var models []string
		if err := json.Unmarshal([]byte(cfg.WorkersChatModelsList), &models); err != nil {
			return nil, fmt.Errorf("failed to parse WORKERS_CHAT_MODELS_LIST: %w", err)
		}
		return models, nil
	}
	// Default models
	return []string{
		"@cf/meta/llama-3.1-8b-instruct",
		"@cf/qwen/qwen1.5-7b-chat-awq",
		"@cf/mistral/mistral-7b-instruct-v0.1",
	}, nil
}

func (a *WorkersChatAgent) Request(ctx context.Context, params *LLMChatParams, cfg *config.Config, onStream ChatStreamTextHandler) (*ChatAgentResponse, error) {
	// Build Workers AI endpoint
	endpoint := fmt.Sprintf("https://api.cloudflare.com/client/v4/accounts/%s/ai/run/%s",
		cfg.CloudflareAccountID,
		a.Model(cfg),
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
	if cfg.WorkersChatExtraParams != nil {
		for k, v := range cfg.WorkersChatExtraParams {
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
	req.Header.Set("Authorization", "Bearer "+cfg.CloudflareToken)

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

func (a *WorkersChatAgent) handleStreamResponse(body io.Reader, onStream ChatStreamTextHandler) (*ChatAgentResponse, error) {
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

		// Workers AI streaming format
		if result, ok := line["result"].(map[string]interface{}); ok {
			if response, ok := result["response"].(string); ok {
				fullText.WriteString(response)
				if err := onStream(response); err != nil {
					return nil, fmt.Errorf("stream handler error: %w", err)
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

func (a *WorkersChatAgent) handleNonStreamResponse(body io.Reader) (*ChatAgentResponse, error) {
	var response struct {
		Result struct {
			Response string `json:"response"`
		} `json:"result"`
	}

	if err := json.NewDecoder(body).Decode(&response); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &ChatAgentResponse{
		Messages: []HistoryItem{
			{
				Role:    "assistant",
				Content: response.Result.Response,
			},
		},
	}, nil
}

// WorkersImageAgent implements ImageAgent for Cloudflare Workers AI
type WorkersImageAgent struct{}

func (a *WorkersImageAgent) Name() string {
	return "workers-image"
}

func (a *WorkersImageAgent) ModelKey() string {
	return "WORKERS_IMAGE_MODEL"
}

func (a *WorkersImageAgent) Enable(cfg *config.Config) bool {
	return cfg.CloudflareAccountID != "" && cfg.CloudflareToken != ""
}

func (a *WorkersImageAgent) Model(cfg *config.Config) string {
	return cfg.WorkersImageModel
}

func (a *WorkersImageAgent) ModelList(cfg *config.Config) ([]string, error) {
	if cfg.WorkersImageModelsList != "" {
		var models []string
		if err := json.Unmarshal([]byte(cfg.WorkersImageModelsList), &models); err != nil {
			return nil, fmt.Errorf("failed to parse WORKERS_IMAGE_MODELS_LIST: %w", err)
		}
		return models, nil
	}
	// Default models
	return []string{
		"@cf/black-forest-labs/flux-1-schnell",
		"@cf/stabilityai/stable-diffusion-xl-base-1.0",
	}, nil
}

func (a *WorkersImageAgent) Request(ctx context.Context, prompt string, cfg *config.Config) (string, error) {
	// Build Workers AI endpoint
	endpoint := fmt.Sprintf("https://api.cloudflare.com/client/v4/accounts/%s/ai/run/%s",
		cfg.CloudflareAccountID,
		a.Model(cfg),
	)

	// Build request body
	reqBody := map[string]interface{}{
		"prompt": prompt,
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
	req.Header.Set("Authorization", "Bearer "+cfg.CloudflareToken)

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

	// Workers AI returns binary image data
	imageData, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read image data: %w", err)
	}

	// Return as base64
	return string(imageData), nil
}
