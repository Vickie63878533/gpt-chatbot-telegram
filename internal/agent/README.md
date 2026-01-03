# Agent Package

This package provides a unified interface for interacting with multiple AI providers (chat and image generation).

## Architecture

The agent system uses a registry pattern where each AI provider implements either the `ChatAgent` or `ImageAgent` interface and registers itself during package initialization.

## Supported Providers

### Chat Agents
- **OpenAI**: GPT-4, GPT-3.5-turbo, etc.
- **Azure OpenAI**: Azure-hosted OpenAI models
- **Google Gemini**: Gemini 1.5 Flash, Pro, etc.
- **Anthropic**: Claude 3.5 Sonnet, Haiku, Opus
- **Cloudflare Workers AI**: Llama, Qwen, Mistral models
- **Mistral AI**: Mistral Large, Medium, Small
- **Cohere**: Command R+, Command R
- **DeepSeek**: DeepSeek Chat, Coder
- **Groq**: Fast inference for Llama and Mixtral
- **XAI**: Grok 2 models

### Image Agents
- **DALL-E**: OpenAI's image generation (DALL-E 2, 3)
- **Azure DALL-E**: Azure-hosted DALL-E
- **Cloudflare Workers AI**: Flux, Stable Diffusion

## Usage

### Loading a Chat Agent

```go
import (
    "github.com/tbxark/ChatGPT-Telegram-Workers/go_version/internal/agent"
    "github.com/tbxark/ChatGPT-Telegram-Workers/go_version/internal/config"
)

// Load with auto-selection (first available provider)
cfg := &config.Config{
    AIProvider: "auto",
    OpenAIAPIKey: []string{"sk-..."},
}
chatAgent, err := agent.LoadChatLLM(cfg, nil)
if err != nil {
    // Handle error
}

// Load specific provider
cfg.AIProvider = "openai"
chatAgent, err = agent.LoadChatLLM(cfg, nil)

// Load with user config override
userConfig := &storage.UserConfig{
    Values: map[string]interface{}{
        "AI_PROVIDER": "gemini",
    },
}
chatAgent, err = agent.LoadChatLLM(cfg, userConfig)
```

### Making a Chat Request

```go
import "context"

params := &agent.LLMChatParams{
    Prompt: "You are a helpful assistant.",
    Messages: []agent.HistoryItem{
        {
            Role:    "user",
            Content: "Hello, how are you?",
        },
    },
}

// Non-streaming request
ctx := context.Background()
response, err := chatAgent.Request(ctx, params, cfg, nil)
if err != nil {
    // Handle error
}
fmt.Println(response.Messages[0].Content)

// Streaming request
streamHandler := func(text string) error {
    fmt.Print(text)
    return nil
}
response, err = chatAgent.Request(ctx, params, cfg, streamHandler)
```

### Loading an Image Agent

```go
// Load with auto-selection
imageAgent, err := agent.LoadImageGen(cfg, nil)
if err != nil {
    // Handle error
}

// Generate image
ctx := context.Background()
imageURL, err := imageAgent.Request(ctx, "A beautiful sunset over mountains", cfg)
if err != nil {
    // Handle error
}
fmt.Println("Image URL:", imageURL)
```

## Adding a New Provider

To add a new AI provider:

1. Create a new file (e.g., `newprovider.go`)
2. Implement the `ChatAgent` and/or `ImageAgent` interface
3. Register the agent in the `init()` function:

```go
func init() {
    RegisterChatAgent(&NewProviderChatAgent{})
    RegisterImageAgent(&NewProviderImageAgent{})
}
```

## Configuration

Each provider requires specific configuration fields in the `Config` struct:

- API keys
- API base URLs
- Model names
- Extra parameters

See `internal/config/config.go` for all available configuration options.

## Error Handling

All agent methods return errors that should be handled appropriately:

- Configuration errors (missing API keys)
- Network errors (API unavailable)
- API errors (rate limits, invalid requests)
- Response parsing errors

## Testing

The agent package can be tested with mock configurations:

```go
cfg := &config.Config{
    OpenAIAPIKey: []string{"test-key"},
    OpenAIChatModel: "gpt-4o-mini",
    OpenAIAPIBase: "https://api.openai.com/v1",
}

agent := &OpenAIChatAgent{}
if !agent.Enable(cfg) {
    t.Error("Agent should be enabled with valid config")
}
```

## Thread Safety

The agent registry is populated during package initialization and is read-only after that, making it safe for concurrent use. Individual agent implementations should handle their own thread safety for HTTP requests.
