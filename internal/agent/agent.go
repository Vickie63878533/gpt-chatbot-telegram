package agent

import (
	"fmt"
	"net/http"
	"time"

	"github.com/tbxark/ChatGPT-Telegram-Workers/go_version/internal/config"
	"github.com/tbxark/ChatGPT-Telegram-Workers/go_version/internal/storage"
)

// Global registry of chat agents
var chatAgents = []ChatAgent{}

// Global registry of image agents
var imageAgents = []ImageAgent{}

// RegisterChatAgent registers a chat agent in the global registry
func RegisterChatAgent(agent ChatAgent) {
	chatAgents = append(chatAgents, agent)
}

// RegisterImageAgent registers an image agent in the global registry
func RegisterImageAgent(agent ImageAgent) {
	imageAgents = append(imageAgents, agent)
}

// CreateHTTPClient creates an HTTP client with optional timeout from config
func CreateHTTPClient(cfg *config.Config) *http.Client {
	client := &http.Client{}

	// Apply timeout if configured (0 means no timeout)
	if cfg.ChatCompleteAPITimeout > 0 {
		client.Timeout = time.Duration(cfg.ChatCompleteAPITimeout) * time.Millisecond
	}

	return client
}

// LoadChatLLM loads a chat agent based on the configuration
// Priority:
// 1. User-configured AI_PROVIDER (from userConfig)
// 2. Global AI_PROVIDER (if not "auto")
// 3. First available agent (auto mode)
func LoadChatLLM(cfg *config.Config, userConfig *storage.UserConfig) (ChatAgent, error) {
	// 1. Check user configuration
	if userConfig != nil {
		if provider, ok := userConfig.Values["AI_PROVIDER"].(string); ok && provider != "" {
			for _, agent := range chatAgents {
				if agent.Name() == provider && agent.Enable(cfg) {
					return agent, nil
				}
			}
			// User specified a provider but it's not available
			return nil, fmt.Errorf("user-configured AI provider %s is not available", provider)
		}
	}

	// 2. Check global configuration (if not "auto")
	if cfg.AIProvider != "auto" && cfg.AIProvider != "" {
		for _, agent := range chatAgents {
			if agent.Name() == cfg.AIProvider && agent.Enable(cfg) {
				return agent, nil
			}
		}
		return nil, fmt.Errorf("configured AI provider %s is not available", cfg.AIProvider)
	}

	// 3. Auto-select first available agent
	for _, agent := range chatAgents {
		if agent.Enable(cfg) {
			return agent, nil
		}
	}

	return nil, fmt.Errorf("no AI chat provider available")
}

// LoadImageGen loads an image generation agent based on the configuration
// Priority:
// 1. User-configured AI_IMAGE_PROVIDER (from userConfig)
// 2. Global AI_IMAGE_PROVIDER (if not "auto")
// 3. First available agent (auto mode)
func LoadImageGen(cfg *config.Config, userConfig *storage.UserConfig) (ImageAgent, error) {
	// 1. Check user configuration
	if userConfig != nil {
		if provider, ok := userConfig.Values["AI_IMAGE_PROVIDER"].(string); ok && provider != "" {
			for _, agent := range imageAgents {
				if agent.Name() == provider && agent.Enable(cfg) {
					return agent, nil
				}
			}
			// User specified a provider but it's not available
			return nil, fmt.Errorf("user-configured image provider %s is not available", provider)
		}
	}

	// 2. Check global configuration (if not "auto")
	if cfg.AIImageProvider != "auto" && cfg.AIImageProvider != "" {
		for _, agent := range imageAgents {
			if agent.Name() == cfg.AIImageProvider && agent.Enable(cfg) {
				return agent, nil
			}
		}
		return nil, fmt.Errorf("configured image provider %s is not available", cfg.AIImageProvider)
	}

	// 3. Auto-select first available agent
	for _, agent := range imageAgents {
		if agent.Enable(cfg) {
			return agent, nil
		}
	}

	return nil, fmt.Errorf("no image generation provider available")
}

// GetChatAgents returns all registered chat agents
func GetChatAgents() []ChatAgent {
	return chatAgents
}

// GetImageAgents returns all registered image agents
func GetImageAgents() []ImageAgent {
	return imageAgents
}
