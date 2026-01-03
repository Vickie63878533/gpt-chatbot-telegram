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
	RegisterChatAgent(NewMistralChatAgent())
	RegisterChatAgent(NewCohereChatAgent())
	RegisterChatAgent(NewDeepSeekChatAgent())
	RegisterChatAgent(NewGroqChatAgent())
	RegisterChatAgent(NewXAIChatAgent())
}

// OpenAICompatibleAgent is a generic agent for OpenAI-compatible APIs
type OpenAICompatibleAgent struct {
	name           string
	modelKey       string
	enableCheck    func(*config.Config) bool
	getModel       func(*config.Config) string
	getAPIBase     func(*config.Config) string
	getAPIKey      func(*config.Config) string
	getModelsList  func(*config.Config) string
	getExtraParams func(*config.Config) map[string]interface{}
	defaultModels  []string
}

func (a *OpenAICompatibleAgent) Name() string {
	return a.name
}

func (a *OpenAICompatibleAgent) ModelKey() string {
	return a.modelKey
}

func (a *OpenAICompatibleAgent) Enable(cfg *config.Config) bool {
	return a.enableCheck(cfg)
}

func (a *OpenAICompatibleAgent) Model(cfg *config.Config) string {
	return a.getModel(cfg)
}

func (a *OpenAICompatibleAgent) ModelList(cfg *config.Config) ([]string, error) {
	modelsList := a.getModelsList(cfg)
	if modelsList != "" {
		var models []string
		if err := json.Unmarshal([]byte(modelsList), &models); err != nil {
			return nil, fmt.Errorf("failed to parse models list: %w", err)
		}
		return models, nil
	}
	return a.defaultModels, nil
}

func (a *OpenAICompatibleAgent) Request(ctx context.Context, params *LLMChatParams, cfg *config.Config, onStream ChatStreamTextHandler) (*ChatAgentResponse, error) {
	apiKey := a.getAPIKey(cfg)
	apiBase := a.getAPIBase(cfg)
	if !strings.HasSuffix(apiBase, "/") {
		apiBase += "/"
	}

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
		"model":    a.Model(cfg),
		"messages": messages,
		"stream":   onStream != nil,
	}

	// Add extra parameters
	extraParams := a.getExtraParams(cfg)
	if extraParams != nil {
		for k, v := range extraParams {
			reqBody[k] = v
		}
	}

	bodyBytes, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	// Create HTTP request
	req, err := http.NewRequestWithContext(ctx, "POST", apiBase+"chat/completions", bytes.NewReader(bodyBytes))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+apiKey)

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

func (a *OpenAICompatibleAgent) handleStreamResponse(body io.Reader, onStream ChatStreamTextHandler) (*ChatAgentResponse, error) {
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

func (a *OpenAICompatibleAgent) handleNonStreamResponse(body io.Reader) (*ChatAgentResponse, error) {
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

// MistralChatAgent implements ChatAgent for Mistral AI
type MistralChatAgent struct {
	OpenAICompatibleAgent
}

func NewMistralChatAgent() *MistralChatAgent {
	return &MistralChatAgent{
		OpenAICompatibleAgent: OpenAICompatibleAgent{
			name:     "mistral",
			modelKey: "MISTRAL_CHAT_MODEL",
			enableCheck: func(cfg *config.Config) bool {
				return cfg.MistralAPIKey != ""
			},
			getModel: func(cfg *config.Config) string {
				return cfg.MistralChatModel
			},
			getAPIBase: func(cfg *config.Config) string {
				return cfg.MistralAPIBase
			},
			getAPIKey: func(cfg *config.Config) string {
				return cfg.MistralAPIKey
			},
			getModelsList: func(cfg *config.Config) string {
				return cfg.MistralChatModelsList
			},
			getExtraParams: func(cfg *config.Config) map[string]interface{} {
				return cfg.MistralChatExtraParams
			},
			defaultModels: []string{"mistral-large-latest", "mistral-medium-latest", "mistral-small-latest"},
		},
	}
}

// CohereChatAgent implements ChatAgent for Cohere
type CohereChatAgent struct {
	OpenAICompatibleAgent
}

func NewCohereChatAgent() *CohereChatAgent {
	return &CohereChatAgent{
		OpenAICompatibleAgent: OpenAICompatibleAgent{
			name:     "cohere",
			modelKey: "COHERE_CHAT_MODEL",
			enableCheck: func(cfg *config.Config) bool {
				return cfg.CohereAPIKey != ""
			},
			getModel: func(cfg *config.Config) string {
				return cfg.CohereChatModel
			},
			getAPIBase: func(cfg *config.Config) string {
				return cfg.CohereAPIBase
			},
			getAPIKey: func(cfg *config.Config) string {
				return cfg.CohereAPIKey
			},
			getModelsList: func(cfg *config.Config) string {
				return cfg.CohereChatModelsList
			},
			getExtraParams: func(cfg *config.Config) map[string]interface{} {
				return cfg.CohereChatExtraParams
			},
			defaultModels: []string{"command-r-plus", "command-r", "command"},
		},
	}
}

// DeepSeekChatAgent implements ChatAgent for DeepSeek
type DeepSeekChatAgent struct {
	OpenAICompatibleAgent
}

func NewDeepSeekChatAgent() *DeepSeekChatAgent {
	return &DeepSeekChatAgent{
		OpenAICompatibleAgent: OpenAICompatibleAgent{
			name:     "deepseek",
			modelKey: "DEEPSEEK_CHAT_MODEL",
			enableCheck: func(cfg *config.Config) bool {
				return cfg.DeepSeekAPIKey != ""
			},
			getModel: func(cfg *config.Config) string {
				return cfg.DeepSeekChatModel
			},
			getAPIBase: func(cfg *config.Config) string {
				return cfg.DeepSeekAPIBase
			},
			getAPIKey: func(cfg *config.Config) string {
				return cfg.DeepSeekAPIKey
			},
			getModelsList: func(cfg *config.Config) string {
				return cfg.DeepSeekChatModelsList
			},
			getExtraParams: func(cfg *config.Config) map[string]interface{} {
				return cfg.DeepSeekChatExtraParams
			},
			defaultModels: []string{"deepseek-chat", "deepseek-coder"},
		},
	}
}

// GroqChatAgent implements ChatAgent for Groq
type GroqChatAgent struct {
	OpenAICompatibleAgent
}

func NewGroqChatAgent() *GroqChatAgent {
	return &GroqChatAgent{
		OpenAICompatibleAgent: OpenAICompatibleAgent{
			name:     "groq",
			modelKey: "GROQ_CHAT_MODEL",
			enableCheck: func(cfg *config.Config) bool {
				return cfg.GroqAPIKey != ""
			},
			getModel: func(cfg *config.Config) string {
				return cfg.GroqChatModel
			},
			getAPIBase: func(cfg *config.Config) string {
				return cfg.GroqAPIBase
			},
			getAPIKey: func(cfg *config.Config) string {
				return cfg.GroqAPIKey
			},
			getModelsList: func(cfg *config.Config) string {
				return cfg.GroqChatModelsList
			},
			getExtraParams: func(cfg *config.Config) map[string]interface{} {
				return cfg.GroqChatExtraParams
			},
			defaultModels: []string{"llama-3.1-70b-versatile", "llama-3.1-8b-instant", "mixtral-8x7b-32768"},
		},
	}
}

// XAIChatAgent implements ChatAgent for XAI (Grok)
type XAIChatAgent struct {
	OpenAICompatibleAgent
}

func NewXAIChatAgent() *XAIChatAgent {
	return &XAIChatAgent{
		OpenAICompatibleAgent: OpenAICompatibleAgent{
			name:     "xai",
			modelKey: "XAI_CHAT_MODEL",
			enableCheck: func(cfg *config.Config) bool {
				return cfg.XAIAPIKey != ""
			},
			getModel: func(cfg *config.Config) string {
				return cfg.XAIChatModel
			},
			getAPIBase: func(cfg *config.Config) string {
				return cfg.XAIAPIBase
			},
			getAPIKey: func(cfg *config.Config) string {
				return cfg.XAIAPIKey
			},
			getModelsList: func(cfg *config.Config) string {
				return cfg.XAIChatModelsList
			},
			getExtraParams: func(cfg *config.Config) map[string]interface{} {
				return cfg.XAIChatExtraParams
			},
			defaultModels: []string{"grok-2-latest", "grok-2-vision-latest"},
		},
	}
}
