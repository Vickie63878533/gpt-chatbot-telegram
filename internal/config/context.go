package config

import (
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/tbxark/ChatGPT-Telegram-Workers/go_version/internal/storage"
)

// ShareContext contains shared context for a bot token
type ShareContext struct {
	BotToken       string
	BotID          int64
	ChatHistoryKey string
	ConfigStoreKey string
	LastMessageKey string
}

// WorkerContext contains the full context for processing a request
type WorkerContext struct {
	ShareContext      ShareContext
	UserConfig        *storage.UserConfig
	DB                storage.Storage
	Bot               interface{}            // Telegram bot API instance (tgbotapi.BotAPI)
	Config            *Config                // Bot configuration
	Context           map[string]interface{} // Request-specific context data
	PermissionChecker PermissionChecker      // Permission checker for authorization
}

// NewShareContext creates a new ShareContext from a bot token
func NewShareContext(botToken string) (*ShareContext, error) {
	// Extract bot ID from token (format: 123456:ABC-DEF...)
	botID, err := extractBotID(botToken)
	if err != nil {
		return nil, fmt.Errorf("invalid bot token format: %w", err)
	}

	return &ShareContext{
		BotToken:       botToken,
		BotID:          botID,
		ChatHistoryKey: "chat_history",
		ConfigStoreKey: "user_config",
		LastMessageKey: "last_message",
	}, nil
}

// NewWorkerContext creates a new WorkerContext
func NewWorkerContext(shareCtx ShareContext, db storage.Storage, cfg *Config) *WorkerContext {
	return &WorkerContext{
		ShareContext:      shareCtx,
		UserConfig:        &storage.UserConfig{Values: make(map[string]interface{})},
		DB:                db,
		Config:            cfg,
		Context:           make(map[string]interface{}),
		PermissionChecker: nil, // Will be set by the caller if needed
	}
}

// NewWorkerContextWithPermission creates a new WorkerContext with a permission checker
// This should be used when processing user requests to enforce ENABLE_USER_SETTING
func NewWorkerContextWithPermission(shareCtx ShareContext, db storage.Storage, cfg *Config, permChecker PermissionChecker) *WorkerContext {
	ctx := NewWorkerContext(shareCtx, db, cfg)
	ctx.PermissionChecker = permChecker
	return ctx
}

// LoadUserConfig loads user configuration from storage
func (wc *WorkerContext) LoadUserConfig(sessionCtx *storage.SessionContext) error {
	// If ENABLE_USER_SETTING is false, use global config for all users
	// Only load user config if user settings are enabled
	if !wc.Config.EnableUserSetting {
		// Initialize empty config - all users will use global config
		wc.UserConfig = &storage.UserConfig{
			DefineKeys: []string{},
			Values:     make(map[string]interface{}),
		}
		return nil
	}

	config, err := wc.DB.GetUserConfig(sessionCtx)
	if err != nil {
		return fmt.Errorf("failed to load user config: %w", err)
	}

	if config != nil {
		wc.UserConfig = config
	} else {
		// Initialize empty config if none exists
		wc.UserConfig = &storage.UserConfig{
			DefineKeys: []string{},
			Values:     make(map[string]interface{}),
		}
	}

	return nil
}

// SaveUserConfig saves user configuration to storage
func (wc *WorkerContext) SaveUserConfig(sessionCtx *storage.SessionContext) error {
	if err := wc.DB.SaveUserConfig(sessionCtx, wc.UserConfig); err != nil {
		return fmt.Errorf("failed to save user config: %w", err)
	}
	return nil
}

// SetUserConfigValue sets a user configuration value
// Requires permission check when ENABLE_USER_SETTING is false
func (wc *WorkerContext) SetUserConfigValue(key string, value interface{}, lockedKeys []string) error {
	// Check if key is locked
	for _, lockedKey := range lockedKeys {
		if key == lockedKey {
			return fmt.Errorf("configuration key '%s' is locked and cannot be modified", key)
		}
	}

	// Add to define keys if not already present
	found := false
	for _, k := range wc.UserConfig.DefineKeys {
		if k == key {
			found = true
			break
		}
	}
	if !found {
		wc.UserConfig.DefineKeys = append(wc.UserConfig.DefineKeys, key)
	}

	// Set the value
	wc.UserConfig.Values[key] = value

	return nil
}

// SetUserConfigValueWithPermission sets a user configuration value with permission check
// This should be used by command handlers to enforce ENABLE_USER_SETTING
func (wc *WorkerContext) SetUserConfigValueWithPermission(key string, value interface{}, lockedKeys []string, userID int64, chatID int64) error {
	// Check permission if ENABLE_USER_SETTING is false
	if !wc.Config.EnableUserSetting {
		if wc.PermissionChecker == nil {
			return fmt.Errorf("permission checker not configured")
		}

		canModify, err := wc.PermissionChecker.CanModifyConfig(userID, chatID, wc)
		if err != nil {
			return fmt.Errorf("failed to check permissions: %w", err)
		}

		if !canModify {
			return fmt.Errorf("user settings are disabled, only administrators can modify configuration")
		}
	}

	return wc.SetUserConfigValue(key, value, lockedKeys)
}

// DeleteUserConfigValue deletes a user configuration value
func (wc *WorkerContext) DeleteUserConfigValue(key string, lockedKeys []string) error {
	// Check if key is locked
	for _, lockedKey := range lockedKeys {
		if key == lockedKey {
			return fmt.Errorf("configuration key '%s' is locked and cannot be deleted", key)
		}
	}

	// Remove from define keys
	newDefineKeys := []string{}
	for _, k := range wc.UserConfig.DefineKeys {
		if k != key {
			newDefineKeys = append(newDefineKeys, k)
		}
	}
	wc.UserConfig.DefineKeys = newDefineKeys

	// Delete the value
	delete(wc.UserConfig.Values, key)

	return nil
}

// DeleteUserConfigValueWithPermission deletes a user configuration value with permission check
// This should be used by command handlers to enforce ENABLE_USER_SETTING
func (wc *WorkerContext) DeleteUserConfigValueWithPermission(key string, lockedKeys []string, userID int64, chatID int64) error {
	// Check permission if ENABLE_USER_SETTING is false
	if !wc.Config.EnableUserSetting {
		if wc.PermissionChecker == nil {
			return fmt.Errorf("permission checker not configured")
		}

		canModify, err := wc.PermissionChecker.CanModifyConfig(userID, chatID, wc)
		if err != nil {
			return fmt.Errorf("failed to check permissions: %w", err)
		}

		if !canModify {
			return fmt.Errorf("user settings are disabled, only administrators can modify configuration")
		}
	}

	return wc.DeleteUserConfigValue(key, lockedKeys)
}

// ClearUserConfig clears all user configuration
func (wc *WorkerContext) ClearUserConfig(lockedKeys []string) error {
	// Keep only locked keys
	newDefineKeys := []string{}
	newValues := make(map[string]interface{})

	for _, key := range wc.UserConfig.DefineKeys {
		isLocked := false
		for _, lockedKey := range lockedKeys {
			if key == lockedKey {
				isLocked = true
				break
			}
		}
		if isLocked {
			newDefineKeys = append(newDefineKeys, key)
			if value, exists := wc.UserConfig.Values[key]; exists {
				newValues[key] = value
			}
		}
	}

	wc.UserConfig.DefineKeys = newDefineKeys
	wc.UserConfig.Values = newValues

	return nil
}

// ClearUserConfigWithPermission clears all user configuration with permission check
// This should be used by command handlers to enforce ENABLE_USER_SETTING
func (wc *WorkerContext) ClearUserConfigWithPermission(lockedKeys []string, userID int64, chatID int64) error {
	// Check permission if ENABLE_USER_SETTING is false
	if !wc.Config.EnableUserSetting {
		if wc.PermissionChecker == nil {
			return fmt.Errorf("permission checker not configured")
		}

		canModify, err := wc.PermissionChecker.CanModifyConfig(userID, chatID, wc)
		if err != nil {
			return fmt.Errorf("failed to check permissions: %w", err)
		}

		if !canModify {
			return fmt.Errorf("user settings are disabled, only administrators can modify configuration")
		}
	}

	return wc.ClearUserConfig(lockedKeys)
}

// GetConfigValue gets a configuration value, checking user config first, then global config
func (wc *WorkerContext) GetConfigValue(key string, globalConfig *Config) interface{} {
	// Check user config first
	if value, exists := wc.UserConfig.Values[key]; exists {
		return value
	}

	// Fall back to global config
	return getGlobalConfigValue(key, globalConfig)
}

// GetConfigString gets a string configuration value
func (wc *WorkerContext) GetConfigString(key string, globalConfig *Config) string {
	value := wc.GetConfigValue(key, globalConfig)
	if str, ok := value.(string); ok {
		return str
	}
	return ""
}

// GetConfigInt gets an int configuration value
func (wc *WorkerContext) GetConfigInt(key string, globalConfig *Config) int {
	value := wc.GetConfigValue(key, globalConfig)
	switch v := value.(type) {
	case int:
		return v
	case int64:
		return int(v)
	case float64:
		return int(v)
	case string:
		if intVal, err := strconv.Atoi(v); err == nil {
			return intVal
		}
	}
	return 0
}

// GetConfigBool gets a bool configuration value
func (wc *WorkerContext) GetConfigBool(key string, globalConfig *Config) bool {
	value := wc.GetConfigValue(key, globalConfig)
	if b, ok := value.(bool); ok {
		return b
	}
	return false
}

// extractBotID extracts the bot ID from a bot token
func extractBotID(token string) (int64, error) {
	// Token format: 123456:ABC-DEF...
	// We need to extract the numeric part before the colon
	for i, c := range token {
		if c == ':' {
			if i == 0 {
				return 0, fmt.Errorf("bot ID is empty")
			}
			botIDStr := token[:i]
			botID, err := strconv.ParseInt(botIDStr, 10, 64)
			if err != nil {
				return 0, fmt.Errorf("invalid bot ID: %w", err)
			}
			return botID, nil
		}
	}
	return 0, fmt.Errorf("token does not contain ':' separator")
}

// getGlobalConfigValue gets a value from the global config by key name
func getGlobalConfigValue(key string, cfg *Config) interface{} {
	switch key {
	// General
	case "AI_PROVIDER":
		return cfg.AIProvider
	case "AI_IMAGE_PROVIDER":
		return cfg.AIImageProvider
	case "SYSTEM_INIT_MESSAGE":
		return cfg.SystemInitMessage

	// OpenAI
	case "OPENAI_API_KEY":
		return cfg.OpenAIAPIKey
	case "OPENAI_CHAT_MODEL":
		return cfg.OpenAIChatModel
	case "OPENAI_API_BASE":
		return cfg.OpenAIAPIBase
	case "OPENAI_API_EXTRA_PARAMS":
		return cfg.OpenAIAPIExtraParams
	case "OPENAI_CHAT_MODELS_LIST":
		return cfg.OpenAIChatModelsList

	// DALL-E
	case "DALL_E_MODEL":
		return cfg.DallEModel
	case "DALL_E_IMAGE_SIZE":
		return cfg.DallEImageSize
	case "DALL_E_IMAGE_QUALITY":
		return cfg.DallEImageQuality
	case "DALL_E_IMAGE_STYLE":
		return cfg.DallEImageStyle
	case "DALL_E_MODELS_LIST":
		return cfg.DallEModelsList

	// Azure
	case "AZURE_API_KEY":
		return cfg.AzureAPIKey
	case "AZURE_RESOURCE_NAME":
		return cfg.AzureResourceName
	case "AZURE_CHAT_MODEL":
		return cfg.AzureChatModel
	case "AZURE_IMAGE_MODEL":
		return cfg.AzureImageModel
	case "AZURE_API_VERSION":
		return cfg.AzureAPIVersion
	case "AZURE_CHAT_MODELS_LIST":
		return cfg.AzureChatModelsList
	case "AZURE_CHAT_EXTRA_PARAMS":
		return cfg.AzureChatExtraParams

	// Workers AI
	case "CLOUDFLARE_ACCOUNT_ID":
		return cfg.CloudflareAccountID
	case "CLOUDFLARE_TOKEN":
		return cfg.CloudflareToken
	case "WORKERS_CHAT_MODEL":
		return cfg.WorkersChatModel
	case "WORKERS_IMAGE_MODEL":
		return cfg.WorkersImageModel
	case "WORKERS_CHAT_MODELS_LIST":
		return cfg.WorkersChatModelsList
	case "WORKERS_IMAGE_MODELS_LIST":
		return cfg.WorkersImageModelsList
	case "WORKERS_CHAT_EXTRA_PARAMS":
		return cfg.WorkersChatExtraParams

	// Gemini
	case "GOOGLE_API_KEY":
		return cfg.GoogleAPIKey
	case "GOOGLE_API_BASE":
		return cfg.GoogleAPIBase
	case "GOOGLE_CHAT_MODEL":
		return cfg.GoogleChatModel
	case "GOOGLE_CHAT_MODELS_LIST":
		return cfg.GoogleChatModelsList
	case "GOOGLE_CHAT_EXTRA_PARAMS":
		return cfg.GoogleChatExtraParams

	// Mistral
	case "MISTRAL_API_KEY":
		return cfg.MistralAPIKey
	case "MISTRAL_API_BASE":
		return cfg.MistralAPIBase
	case "MISTRAL_CHAT_MODEL":
		return cfg.MistralChatModel
	case "MISTRAL_CHAT_MODELS_LIST":
		return cfg.MistralChatModelsList
	case "MISTRAL_CHAT_EXTRA_PARAMS":
		return cfg.MistralChatExtraParams

	// Cohere
	case "COHERE_API_KEY":
		return cfg.CohereAPIKey
	case "COHERE_API_BASE":
		return cfg.CohereAPIBase
	case "COHERE_CHAT_MODEL":
		return cfg.CohereChatModel
	case "COHERE_CHAT_MODELS_LIST":
		return cfg.CohereChatModelsList
	case "COHERE_CHAT_EXTRA_PARAMS":
		return cfg.CohereChatExtraParams

	// Anthropic
	case "ANTHROPIC_API_KEY":
		return cfg.AnthropicAPIKey
	case "ANTHROPIC_API_BASE":
		return cfg.AnthropicAPIBase
	case "ANTHROPIC_CHAT_MODEL":
		return cfg.AnthropicChatModel
	case "ANTHROPIC_CHAT_MODELS_LIST":
		return cfg.AnthropicChatModelsList
	case "ANTHROPIC_CHAT_EXTRA_PARAMS":
		return cfg.AnthropicChatExtraParams

	// DeepSeek
	case "DEEPSEEK_API_KEY":
		return cfg.DeepSeekAPIKey
	case "DEEPSEEK_API_BASE":
		return cfg.DeepSeekAPIBase
	case "DEEPSEEK_CHAT_MODEL":
		return cfg.DeepSeekChatModel
	case "DEEPSEEK_CHAT_MODELS_LIST":
		return cfg.DeepSeekChatModelsList
	case "DEEPSEEK_CHAT_EXTRA_PARAMS":
		return cfg.DeepSeekChatExtraParams

	// Groq
	case "GROQ_API_KEY":
		return cfg.GroqAPIKey
	case "GROQ_API_BASE":
		return cfg.GroqAPIBase
	case "GROQ_CHAT_MODEL":
		return cfg.GroqChatModel
	case "GROQ_CHAT_MODELS_LIST":
		return cfg.GroqChatModelsList
	case "GROQ_CHAT_EXTRA_PARAMS":
		return cfg.GroqChatExtraParams

	// XAI
	case "XAI_API_KEY":
		return cfg.XAIAPIKey
	case "XAI_API_BASE":
		return cfg.XAIAPIBase
	case "XAI_CHAT_MODEL":
		return cfg.XAIChatModel
	case "XAI_CHAT_MODELS_LIST":
		return cfg.XAIChatModelsList
	case "XAI_CHAT_EXTRA_PARAMS":
		return cfg.XAIChatExtraParams

	// Environment
	case "LANGUAGE":
		return cfg.Language
	case "UPDATE_BRANCH":
		return cfg.UpdateBranch
	case "CHAT_COMPLETE_API_TIMEOUT":
		return cfg.ChatCompleteAPITimeout

	// Telegram
	case "TELEGRAM_API_DOMAIN":
		return cfg.TelegramAPIDomain
	case "TELEGRAM_AVAILABLE_TOKENS":
		return cfg.TelegramAvailableTokens
	case "DEFAULT_PARSE_MODE":
		return cfg.DefaultParseMode
	case "TELEGRAM_MIN_STREAM_INTERVAL":
		return cfg.TelegramMinStreamInterval
	case "TELEGRAM_PHOTO_SIZE_OFFSET":
		return cfg.TelegramPhotoSizeOffset
	case "TELEGRAM_IMAGE_TRANSFER_MODE":
		return cfg.TelegramImageTransferMode
	case "MODEL_LIST_COLUMNS":
		return cfg.ModelListColumns

	// Permissions
	case "I_AM_A_GENEROUS_PERSON":
		return cfg.IAmAGenerousPerson
	case "CHAT_WHITE_LIST":
		return cfg.ChatWhiteList
	case "LOCK_USER_CONFIG_KEYS":
		return cfg.LockUserConfigKeys

	// Group
	case "TELEGRAM_BOT_NAME":
		return cfg.TelegramBotName
	case "CHAT_GROUP_WHITE_LIST":
		return cfg.ChatGroupWhiteList
	case "GROUP_CHAT_BOT_ENABLE":
		return cfg.GroupChatBotEnable
	case "GROUP_CHAT_BOT_SHARE_MODE":
		return cfg.GroupChatBotShareMode

	// History
	case "AUTO_TRIM_HISTORY":
		return cfg.AutoTrimHistory
	case "MAX_HISTORY_LENGTH":
		return cfg.MaxHistoryLength
	case "MAX_TOKEN_LENGTH":
		return cfg.MaxTokenLength
	case "HISTORY_IMAGE_PLACEHOLDER":
		return cfg.HistoryImagePlaceholder

	// Features
	case "HIDE_COMMAND_BUTTONS":
		return cfg.HideCommandButtons
	case "SHOW_REPLY_BUTTON":
		return cfg.ShowReplyButton
	case "EXTRA_MESSAGE_CONTEXT":
		return cfg.ExtraMessageContext
	case "EXTRA_MESSAGE_MEDIA_COMPATIBLE":
		return cfg.ExtraMessageMediaCompatible

	// Modes
	case "STREAM_MODE":
		return cfg.StreamMode
	case "SAFE_MODE":
		return cfg.SafeMode
	case "DEBUG_MODE":
		return cfg.DebugMode
	case "DEV_MODE":
		return cfg.DevMode

	// Server
	case "PORT":
		return cfg.Port
	case "DB_PATH":
		return cfg.DBPath

	default:
		return nil
	}
}

// MergeUserConfig merges user config values into a new Config instance
func MergeUserConfig(globalConfig *Config, userConfig *storage.UserConfig) *Config {
	// Create a copy of the global config
	merged := *globalConfig

	// Override with user config values
	for key, value := range userConfig.Values {
		switch key {
		case "AI_PROVIDER":
			if str, ok := value.(string); ok {
				merged.AIProvider = str
			}
		case "AI_IMAGE_PROVIDER":
			if str, ok := value.(string); ok {
				merged.AIImageProvider = str
			}
		case "OPENAI_CHAT_MODEL":
			if str, ok := value.(string); ok {
				merged.OpenAIChatModel = str
			}
		case "AZURE_CHAT_MODEL":
			if str, ok := value.(string); ok {
				merged.AzureChatModel = str
			}
		case "GOOGLE_CHAT_MODEL":
			if str, ok := value.(string); ok {
				merged.GoogleChatModel = str
			}
		case "STREAM_MODE":
			if b, ok := value.(bool); ok {
				merged.StreamMode = b
			}
			// Add more cases as needed for other configurable fields
		}
	}

	return &merged
}

// ParseUserConfigFromJSON parses user config from JSON string
func ParseUserConfigFromJSON(jsonStr string) (*storage.UserConfig, error) {
	var config storage.UserConfig
	if err := json.Unmarshal([]byte(jsonStr), &config); err != nil {
		return nil, fmt.Errorf("failed to parse user config: %w", err)
	}
	if config.Values == nil {
		config.Values = make(map[string]interface{})
	}
	return &config, nil
}

// ToJSON converts user config to JSON string
func ToJSON(config *storage.UserConfig) (string, error) {
	data, err := json.Marshal(config)
	if err != nil {
		return "", fmt.Errorf("failed to marshal user config: %w", err)
	}
	return string(data), nil
}

// NewSessionContext creates a SessionContext from chat information
// This is used to generate the appropriate storage key based on chat type and mode
func NewSessionContext(chatID, botID int64, userID *int64, threadID *int64) *storage.SessionContext {
	return &storage.SessionContext{
		ChatID:   chatID,
		BotID:    botID,
		UserID:   userID,
		ThreadID: threadID,
	}
}

// NewSessionContextFromChat creates a SessionContext for a specific chat
// Parameters:
//   - chatID: The Telegram chat ID
//   - botID: The bot ID
//   - isGroup: Whether this is a group chat
//   - shareMode: Whether group share mode is enabled (only relevant for groups)
//   - userID: The user ID (used in non-shared group mode)
//   - threadID: The thread/topic ID (for forum chats)
func NewSessionContextFromChat(chatID, botID int64, isGroup, shareMode bool, userID, threadID *int64) *storage.SessionContext {
	ctx := &storage.SessionContext{
		ChatID: chatID,
		BotID:  botID,
	}

	// In group non-shared mode, include user ID
	if isGroup && !shareMode && userID != nil {
		ctx.UserID = userID
	}

	// Include thread ID if present
	if threadID != nil {
		ctx.ThreadID = threadID
	}

	return ctx
}
