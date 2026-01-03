package config

import (
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"strings"
)

// Config holds all configuration for the bot
type Config struct {
	// General Configuration
	AIProvider        string `env:"AI_PROVIDER" default:"auto"`
	AIImageProvider   string `env:"AI_IMAGE_PROVIDER" default:"auto"`
	SystemInitMessage string `env:"SYSTEM_INIT_MESSAGE"`

	// OpenAI Configuration
	OpenAIAPIKey         []string               `env:"OPENAI_API_KEY"`
	OpenAIChatModel      string                 `env:"OPENAI_CHAT_MODEL" default:"gpt-4o-mini"`
	OpenAIAPIBase        string                 `env:"OPENAI_API_BASE" default:"https://api.openai.com/v1"`
	OpenAIAPIExtraParams map[string]interface{} `env:"OPENAI_API_EXTRA_PARAMS"`
	OpenAIChatModelsList string                 `env:"OPENAI_CHAT_MODELS_LIST"`

	// DALL-E Configuration
	DallEModel        string `env:"DALL_E_MODEL" default:"dall-e-3"`
	DallEImageSize    string `env:"DALL_E_IMAGE_SIZE" default:"1024x1024"`
	DallEImageQuality string `env:"DALL_E_IMAGE_QUALITY" default:"standard"`
	DallEImageStyle   string `env:"DALL_E_IMAGE_STYLE" default:"vivid"`
	DallEModelsList   string `env:"DALL_E_MODELS_LIST" default:"[\"dall-e-3\"]"`

	// Azure Configuration
	AzureAPIKey          string                 `env:"AZURE_API_KEY"`
	AzureResourceName    string                 `env:"AZURE_RESOURCE_NAME"`
	AzureChatModel       string                 `env:"AZURE_CHAT_MODEL" default:"gpt-4o-mini"`
	AzureImageModel      string                 `env:"AZURE_IMAGE_MODEL" default:"dall-e-3"`
	AzureAPIVersion      string                 `env:"AZURE_API_VERSION" default:"2024-06-01"`
	AzureChatModelsList  string                 `env:"AZURE_CHAT_MODELS_LIST"`
	AzureChatExtraParams map[string]interface{} `env:"AZURE_CHAT_EXTRA_PARAMS"`

	// Workers AI Configuration
	CloudflareAccountID    string                 `env:"CLOUDFLARE_ACCOUNT_ID"`
	CloudflareToken        string                 `env:"CLOUDFLARE_TOKEN"`
	WorkersChatModel       string                 `env:"WORKERS_CHAT_MODEL" default:"@cf/qwen/qwen1.5-7b-chat-awq"`
	WorkersImageModel      string                 `env:"WORKERS_IMAGE_MODEL" default:"@cf/black-forest-labs/flux-1-schnell"`
	WorkersChatModelsList  string                 `env:"WORKERS_CHAT_MODELS_LIST"`
	WorkersImageModelsList string                 `env:"WORKERS_IMAGE_MODELS_LIST"`
	WorkersChatExtraParams map[string]interface{} `env:"WORKERS_CHAT_EXTRA_PARAMS"`

	// Gemini Configuration
	GoogleAPIKey          string                 `env:"GOOGLE_API_KEY"`
	GoogleAPIBase         string                 `env:"GOOGLE_API_BASE" default:"https://generativelanguage.googleapis.com/v1beta"`
	GoogleChatModel       string                 `env:"GOOGLE_CHAT_MODEL" default:"gemini-1.5-flash"`
	GoogleChatModelsList  string                 `env:"GOOGLE_CHAT_MODELS_LIST"`
	GoogleChatExtraParams map[string]interface{} `env:"GOOGLE_CHAT_EXTRA_PARAMS"`

	// Mistral Configuration
	MistralAPIKey          string                 `env:"MISTRAL_API_KEY"`
	MistralAPIBase         string                 `env:"MISTRAL_API_BASE" default:"https://api.mistral.ai/v1"`
	MistralChatModel       string                 `env:"MISTRAL_CHAT_MODEL" default:"mistral-tiny"`
	MistralChatModelsList  string                 `env:"MISTRAL_CHAT_MODELS_LIST"`
	MistralChatExtraParams map[string]interface{} `env:"MISTRAL_CHAT_EXTRA_PARAMS"`

	// Cohere Configuration
	CohereAPIKey          string                 `env:"COHERE_API_KEY"`
	CohereAPIBase         string                 `env:"COHERE_API_BASE" default:"https://api.cohere.com/v2"`
	CohereChatModel       string                 `env:"COHERE_CHAT_MODEL" default:"command-r-plus"`
	CohereChatModelsList  string                 `env:"COHERE_CHAT_MODELS_LIST"`
	CohereChatExtraParams map[string]interface{} `env:"COHERE_CHAT_EXTRA_PARAMS"`

	// Anthropic Configuration
	AnthropicAPIKey          string                 `env:"ANTHROPIC_API_KEY"`
	AnthropicAPIBase         string                 `env:"ANTHROPIC_API_BASE" default:"https://api.anthropic.com/v1"`
	AnthropicChatModel       string                 `env:"ANTHROPIC_CHAT_MODEL" default:"claude-3-5-haiku-latest"`
	AnthropicChatModelsList  string                 `env:"ANTHROPIC_CHAT_MODELS_LIST"`
	AnthropicChatExtraParams map[string]interface{} `env:"ANTHROPIC_CHAT_EXTRA_PARAMS"`

	// DeepSeek Configuration
	DeepSeekAPIKey          string                 `env:"DEEPSEEK_API_KEY"`
	DeepSeekAPIBase         string                 `env:"DEEPSEEK_API_BASE" default:"https://api.deepseek.com"`
	DeepSeekChatModel       string                 `env:"DEEPSEEK_CHAT_MODEL" default:"deepseek-chat"`
	DeepSeekChatModelsList  string                 `env:"DEEPSEEK_CHAT_MODELS_LIST"`
	DeepSeekChatExtraParams map[string]interface{} `env:"DEEPSEEK_CHAT_EXTRA_PARAMS"`

	// Groq Configuration
	GroqAPIKey          string                 `env:"GROQ_API_KEY"`
	GroqAPIBase         string                 `env:"GROQ_API_BASE" default:"https://api.groq.com/openai/v1"`
	GroqChatModel       string                 `env:"GROQ_CHAT_MODEL" default:"groq-chat"`
	GroqChatModelsList  string                 `env:"GROQ_CHAT_MODELS_LIST"`
	GroqChatExtraParams map[string]interface{} `env:"GROQ_CHAT_EXTRA_PARAMS"`

	// XAI Configuration
	XAIAPIKey          string                 `env:"XAI_API_KEY"`
	XAIAPIBase         string                 `env:"XAI_API_BASE" default:"https://api.x.ai/v1"`
	XAIChatModel       string                 `env:"XAI_CHAT_MODEL" default:"grok-2-latest"`
	XAIChatModelsList  string                 `env:"XAI_CHAT_MODELS_LIST"`
	XAIChatExtraParams map[string]interface{} `env:"XAI_CHAT_EXTRA_PARAMS"`

	// Environment Configuration
	Language               string `env:"LANGUAGE" default:"zh-cn"`
	UpdateBranch           string `env:"UPDATE_BRANCH" default:"master"`
	ChatCompleteAPITimeout int    `env:"CHAT_COMPLETE_API_TIMEOUT" default:"0"`

	// Telegram Configuration
	TelegramAPIDomain         string   `env:"TELEGRAM_API_DOMAIN" default:"https://api.telegram.org"`
	TelegramAvailableTokens   []string `env:"TELEGRAM_AVAILABLE_TOKENS" required:"true"`
	DefaultParseMode          string   `env:"DEFAULT_PARSE_MODE" default:"Markdown"`
	TelegramMinStreamInterval int      `env:"TELEGRAM_MIN_STREAM_INTERVAL" default:"0"`
	TelegramPhotoSizeOffset   int      `env:"TELEGRAM_PHOTO_SIZE_OFFSET" default:"1"`
	TelegramImageTransferMode string   `env:"TELEGRAM_IMAGE_TRANSFER_MODE" default:"base64"`
	ModelListColumns          int      `env:"MODEL_LIST_COLUMNS" default:"1"`

	// Permission Configuration
	IAmAGenerousPerson bool     `env:"I_AM_A_GENEROUS_PERSON" default:"false"`
	ChatWhiteList      []string `env:"CHAT_WHITE_LIST"`
	LockUserConfigKeys []string `env:"LOCK_USER_CONFIG_KEYS" default:"OPENAI_API_BASE,GOOGLE_API_BASE,MISTRAL_API_BASE,COHERE_API_BASE,ANTHROPIC_API_BASE,DEEPSEEK_API_BASE,GROQ_API_BASE,XAI_API_BASE"`

	// Group Configuration
	TelegramBotName       []string `env:"TELEGRAM_BOT_NAME"`
	ChatGroupWhiteList    []string `env:"CHAT_GROUP_WHITE_LIST"`
	GroupChatBotEnable    bool     `env:"GROUP_CHAT_BOT_ENABLE" default:"true"`
	GroupChatBotShareMode bool     `env:"GROUP_CHAT_BOT_SHARE_MODE" default:"true"`

	// History Configuration
	AutoTrimHistory         bool   `env:"AUTO_TRIM_HISTORY" default:"true"`
	MaxHistoryLength        int    `env:"MAX_HISTORY_LENGTH" default:"20"`
	MaxTokenLength          int    `env:"MAX_TOKEN_LENGTH" default:"-1"`
	HistoryImagePlaceholder string `env:"HISTORY_IMAGE_PLACEHOLDER"`

	// Feature Switches
	HideCommandButtons          []string `env:"HIDE_COMMAND_BUTTONS"`
	ShowReplyButton             bool     `env:"SHOW_REPLY_BUTTON" default:"false"`
	ExtraMessageContext         bool     `env:"EXTRA_MESSAGE_CONTEXT" default:"false"`
	ExtraMessageMediaCompatible []string `env:"EXTRA_MESSAGE_MEDIA_COMPATIBLE" default:"image"`

	// Mode Switches
	StreamMode bool `env:"STREAM_MODE" default:"true"`
	SafeMode   bool `env:"SAFE_MODE" default:"true"`
	DebugMode  bool `env:"DEBUG_MODE" default:"false"`
	DevMode    bool `env:"DEV_MODE" default:"false"`

	// Server Configuration
	Port   int    `env:"PORT" default:"8080"`
	DBPath string `env:"DB_PATH" default:"./data/bot.db"`

	// Database Configuration
	DSN string `env:"DSN"`

	// User Settings Permission Control
	EnableUserSetting bool     `env:"ENABLE_USER_SETTING" default:"true"`
	ChatAdminKey      []string `env:"CHAT_ADMIN_KEY"`
}

// LoadConfig loads configuration from environment variables
func LoadConfig() (*Config, error) {
	cfg := &Config{}

	// Load string fields
	cfg.AIProvider = getEnvOrDefault("AI_PROVIDER", "auto")
	cfg.AIImageProvider = getEnvOrDefault("AI_IMAGE_PROVIDER", "auto")
	cfg.SystemInitMessage = os.Getenv("SYSTEM_INIT_MESSAGE")

	// OpenAI
	cfg.OpenAIAPIKey = getEnvSlice("OPENAI_API_KEY")
	cfg.OpenAIChatModel = getEnvOrDefault("OPENAI_CHAT_MODEL", "gpt-4o-mini")
	cfg.OpenAIAPIBase = getEnvOrDefault("OPENAI_API_BASE", "https://api.openai.com/v1")
	cfg.OpenAIAPIExtraParams = getEnvJSON("OPENAI_API_EXTRA_PARAMS")
	cfg.OpenAIChatModelsList = os.Getenv("OPENAI_CHAT_MODELS_LIST")

	// DALL-E
	cfg.DallEModel = getEnvOrDefault("DALL_E_MODEL", "dall-e-3")
	cfg.DallEImageSize = getEnvOrDefault("DALL_E_IMAGE_SIZE", "1024x1024")
	cfg.DallEImageQuality = getEnvOrDefault("DALL_E_IMAGE_QUALITY", "standard")
	cfg.DallEImageStyle = getEnvOrDefault("DALL_E_IMAGE_STYLE", "vivid")
	cfg.DallEModelsList = getEnvOrDefault("DALL_E_MODELS_LIST", "[\"dall-e-3\"]")

	// Azure
	cfg.AzureAPIKey = os.Getenv("AZURE_API_KEY")
	cfg.AzureResourceName = os.Getenv("AZURE_RESOURCE_NAME")
	cfg.AzureChatModel = getEnvOrDefault("AZURE_CHAT_MODEL", "gpt-4o-mini")
	cfg.AzureImageModel = getEnvOrDefault("AZURE_IMAGE_MODEL", "dall-e-3")
	cfg.AzureAPIVersion = getEnvOrDefault("AZURE_API_VERSION", "2024-06-01")
	cfg.AzureChatModelsList = os.Getenv("AZURE_CHAT_MODELS_LIST")
	cfg.AzureChatExtraParams = getEnvJSON("AZURE_CHAT_EXTRA_PARAMS")

	// Workers AI
	cfg.CloudflareAccountID = os.Getenv("CLOUDFLARE_ACCOUNT_ID")
	cfg.CloudflareToken = os.Getenv("CLOUDFLARE_TOKEN")
	cfg.WorkersChatModel = getEnvOrDefault("WORKERS_CHAT_MODEL", "@cf/qwen/qwen1.5-7b-chat-awq")
	cfg.WorkersImageModel = getEnvOrDefault("WORKERS_IMAGE_MODEL", "@cf/black-forest-labs/flux-1-schnell")
	cfg.WorkersChatModelsList = os.Getenv("WORKERS_CHAT_MODELS_LIST")
	cfg.WorkersImageModelsList = os.Getenv("WORKERS_IMAGE_MODELS_LIST")
	cfg.WorkersChatExtraParams = getEnvJSON("WORKERS_CHAT_EXTRA_PARAMS")

	// Gemini
	cfg.GoogleAPIKey = os.Getenv("GOOGLE_API_KEY")
	cfg.GoogleAPIBase = getEnvOrDefault("GOOGLE_API_BASE", "https://generativelanguage.googleapis.com/v1beta")
	cfg.GoogleChatModel = getEnvOrDefault("GOOGLE_CHAT_MODEL", "gemini-1.5-flash")
	cfg.GoogleChatModelsList = os.Getenv("GOOGLE_CHAT_MODELS_LIST")
	cfg.GoogleChatExtraParams = getEnvJSON("GOOGLE_CHAT_EXTRA_PARAMS")

	// Mistral
	cfg.MistralAPIKey = os.Getenv("MISTRAL_API_KEY")
	cfg.MistralAPIBase = getEnvOrDefault("MISTRAL_API_BASE", "https://api.mistral.ai/v1")
	cfg.MistralChatModel = getEnvOrDefault("MISTRAL_CHAT_MODEL", "mistral-tiny")
	cfg.MistralChatModelsList = os.Getenv("MISTRAL_CHAT_MODELS_LIST")
	cfg.MistralChatExtraParams = getEnvJSON("MISTRAL_CHAT_EXTRA_PARAMS")

	// Cohere
	cfg.CohereAPIKey = os.Getenv("COHERE_API_KEY")
	cfg.CohereAPIBase = getEnvOrDefault("COHERE_API_BASE", "https://api.cohere.com/v2")
	cfg.CohereChatModel = getEnvOrDefault("COHERE_CHAT_MODEL", "command-r-plus")
	cfg.CohereChatModelsList = os.Getenv("COHERE_CHAT_MODELS_LIST")
	cfg.CohereChatExtraParams = getEnvJSON("COHERE_CHAT_EXTRA_PARAMS")

	// Anthropic
	cfg.AnthropicAPIKey = os.Getenv("ANTHROPIC_API_KEY")
	cfg.AnthropicAPIBase = getEnvOrDefault("ANTHROPIC_API_BASE", "https://api.anthropic.com/v1")
	cfg.AnthropicChatModel = getEnvOrDefault("ANTHROPIC_CHAT_MODEL", "claude-3-5-haiku-latest")
	cfg.AnthropicChatModelsList = os.Getenv("ANTHROPIC_CHAT_MODELS_LIST")
	cfg.AnthropicChatExtraParams = getEnvJSON("ANTHROPIC_CHAT_EXTRA_PARAMS")

	// DeepSeek
	cfg.DeepSeekAPIKey = os.Getenv("DEEPSEEK_API_KEY")
	cfg.DeepSeekAPIBase = getEnvOrDefault("DEEPSEEK_API_BASE", "https://api.deepseek.com")
	cfg.DeepSeekChatModel = getEnvOrDefault("DEEPSEEK_CHAT_MODEL", "deepseek-chat")
	cfg.DeepSeekChatModelsList = os.Getenv("DEEPSEEK_CHAT_MODELS_LIST")
	cfg.DeepSeekChatExtraParams = getEnvJSON("DEEPSEEK_CHAT_EXTRA_PARAMS")

	// Groq
	cfg.GroqAPIKey = os.Getenv("GROQ_API_KEY")
	cfg.GroqAPIBase = getEnvOrDefault("GROQ_API_BASE", "https://api.groq.com/openai/v1")
	cfg.GroqChatModel = getEnvOrDefault("GROQ_CHAT_MODEL", "groq-chat")
	cfg.GroqChatModelsList = os.Getenv("GROQ_CHAT_MODELS_LIST")
	cfg.GroqChatExtraParams = getEnvJSON("GROQ_CHAT_EXTRA_PARAMS")

	// XAI
	cfg.XAIAPIKey = os.Getenv("XAI_API_KEY")
	cfg.XAIAPIBase = getEnvOrDefault("XAI_API_BASE", "https://api.x.ai/v1")
	cfg.XAIChatModel = getEnvOrDefault("XAI_CHAT_MODEL", "grok-2-latest")
	cfg.XAIChatModelsList = os.Getenv("XAI_CHAT_MODELS_LIST")
	cfg.XAIChatExtraParams = getEnvJSON("XAI_CHAT_EXTRA_PARAMS")

	// Environment
	cfg.Language = getEnvOrDefault("LANGUAGE", "zh-cn")
	cfg.UpdateBranch = getEnvOrDefault("UPDATE_BRANCH", "master")
	cfg.ChatCompleteAPITimeout = getEnvInt("CHAT_COMPLETE_API_TIMEOUT", 0)

	// Telegram
	cfg.TelegramAPIDomain = getEnvOrDefault("TELEGRAM_API_DOMAIN", "https://api.telegram.org")
	cfg.TelegramAvailableTokens = getEnvSlice("TELEGRAM_AVAILABLE_TOKENS")
	cfg.DefaultParseMode = getEnvOrDefault("DEFAULT_PARSE_MODE", "Markdown")
	cfg.TelegramMinStreamInterval = getEnvInt("TELEGRAM_MIN_STREAM_INTERVAL", 0)
	cfg.TelegramPhotoSizeOffset = getEnvInt("TELEGRAM_PHOTO_SIZE_OFFSET", 1)
	cfg.TelegramImageTransferMode = getEnvOrDefault("TELEGRAM_IMAGE_TRANSFER_MODE", "base64")
	cfg.ModelListColumns = getEnvInt("MODEL_LIST_COLUMNS", 1)

	// Permissions
	cfg.IAmAGenerousPerson = getEnvBool("I_AM_A_GENEROUS_PERSON", false)
	cfg.ChatWhiteList = getEnvSlice("CHAT_WHITE_LIST")
	cfg.LockUserConfigKeys = getEnvSliceOrDefault("LOCK_USER_CONFIG_KEYS", []string{
		"OPENAI_API_BASE", "GOOGLE_API_BASE", "MISTRAL_API_BASE", "COHERE_API_BASE",
		"ANTHROPIC_API_BASE", "DEEPSEEK_API_BASE", "GROQ_API_BASE", "XAI_API_BASE",
	})

	// Group
	cfg.TelegramBotName = getEnvSlice("TELEGRAM_BOT_NAME")
	cfg.ChatGroupWhiteList = getEnvSlice("CHAT_GROUP_WHITE_LIST")
	cfg.GroupChatBotEnable = getEnvBool("GROUP_CHAT_BOT_ENABLE", true)
	cfg.GroupChatBotShareMode = getEnvBool("GROUP_CHAT_BOT_SHARE_MODE", true)

	// History
	cfg.AutoTrimHistory = getEnvBool("AUTO_TRIM_HISTORY", true)
	cfg.MaxHistoryLength = getEnvInt("MAX_HISTORY_LENGTH", 20)
	cfg.MaxTokenLength = getEnvInt("MAX_TOKEN_LENGTH", -1)
	cfg.HistoryImagePlaceholder = os.Getenv("HISTORY_IMAGE_PLACEHOLDER")

	// Features
	cfg.HideCommandButtons = getEnvSlice("HIDE_COMMAND_BUTTONS")
	cfg.ShowReplyButton = getEnvBool("SHOW_REPLY_BUTTON", false)
	cfg.ExtraMessageContext = getEnvBool("EXTRA_MESSAGE_CONTEXT", false)
	cfg.ExtraMessageMediaCompatible = getEnvSliceOrDefault("EXTRA_MESSAGE_MEDIA_COMPATIBLE", []string{"image"})

	// Modes
	cfg.StreamMode = getEnvBool("STREAM_MODE", true)
	cfg.SafeMode = getEnvBool("SAFE_MODE", true)
	cfg.DebugMode = getEnvBool("DEBUG_MODE", false)
	cfg.DevMode = getEnvBool("DEV_MODE", false)

	// Server
	cfg.Port = getEnvInt("PORT", 8080)
	cfg.DBPath = getEnvOrDefault("DB_PATH", "./data/bot.db")

	// Database
	cfg.DSN = os.Getenv("DSN")

	// User Settings Permission Control
	cfg.EnableUserSetting = getEnvBool("ENABLE_USER_SETTING", true)
	cfg.ChatAdminKey = getEnvSlice("CHAT_ADMIN_KEY")

	// Validate configuration
	if err := Validate(cfg); err != nil {
		return nil, err
	}

	return cfg, nil
}

// Validate checks if the configuration is valid
func Validate(cfg *Config) error {
	// Check required fields
	if len(cfg.TelegramAvailableTokens) == 0 {
		return fmt.Errorf("TELEGRAM_AVAILABLE_TOKENS is required")
	}

	// Validate port
	if cfg.Port < 1 || cfg.Port > 65535 {
		return fmt.Errorf("PORT must be between 1 and 65535, got %d", cfg.Port)
	}

	// Validate parse mode
	if cfg.DefaultParseMode != "Markdown" && cfg.DefaultParseMode != "HTML" && cfg.DefaultParseMode != "" {
		return fmt.Errorf("DEFAULT_PARSE_MODE must be 'Markdown' or 'HTML', got '%s'", cfg.DefaultParseMode)
	}

	// Validate image transfer mode
	if cfg.TelegramImageTransferMode != "url" && cfg.TelegramImageTransferMode != "base64" {
		return fmt.Errorf("TELEGRAM_IMAGE_TRANSFER_MODE must be 'url' or 'base64', got '%s'", cfg.TelegramImageTransferMode)
	}

	// Validate language
	validLanguages := map[string]bool{
		"zh-cn":   true,
		"en":      true,
		"pt":      true,
		"zh-hant": true,
	}
	if !validLanguages[cfg.Language] {
		return fmt.Errorf("LANGUAGE must be one of: zh-cn, en, pt, zh-hant, got '%s'", cfg.Language)
	}

	return nil
}

// Helper functions

func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}

func getEnvInt64(key string, defaultValue int64) int64 {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.ParseInt(value, 10, 64); err == nil {
			return intValue
		}
	}
	return defaultValue
}

func getEnvBool(key string, defaultValue bool) bool {
	if value := os.Getenv(key); value != "" {
		if boolValue, err := strconv.ParseBool(value); err == nil {
			return boolValue
		}
	}
	return defaultValue
}

func getEnvSlice(key string) []string {
	value := os.Getenv(key)
	if value == "" {
		return []string{}
	}
	return strings.Split(value, ",")
}

func getEnvSliceOrDefault(key string, defaultValue []string) []string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return strings.Split(value, ",")
}

func getEnvJSON(key string) map[string]interface{} {
	value := os.Getenv(key)
	if value == "" {
		return nil
	}
	var result map[string]interface{}
	if err := json.Unmarshal([]byte(value), &result); err != nil {
		return nil
	}
	return result
}
