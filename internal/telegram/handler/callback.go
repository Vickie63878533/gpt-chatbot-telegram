package handler

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"math"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/tbxark/ChatGPT-Telegram-Workers/go_version/internal/agent"
	"github.com/tbxark/ChatGPT-Telegram-Workers/go_version/internal/config"
	"github.com/tbxark/ChatGPT-Telegram-Workers/go_version/internal/i18n"
	"github.com/tbxark/ChatGPT-Telegram-Workers/go_version/internal/storage"
	"github.com/tbxark/ChatGPT-Telegram-Workers/go_version/internal/telegram/api"
	"github.com/tbxark/ChatGPT-Telegram-Workers/go_version/internal/telegram/command"
)

// CallbackQueryHandler processes callback queries
// Implements Requirements 2.4, 2.5
type CallbackQueryHandler struct {
	config   *config.Config
	i18n     *i18n.I18n
	handlers []CallbackHandler
}

// CallbackHandler interface for handling specific callback queries
type CallbackHandler interface {
	Prefix() string
	Handle(query *tgbotapi.CallbackQuery, data string, ctx *config.WorkerContext) error
	NeedAuth() command.AuthChecker
}

// NewCallbackQueryHandler creates a new CallbackQueryHandler
func NewCallbackQueryHandler(cfg *config.Config, i18n *i18n.I18n) *CallbackQueryHandler {
	h := &CallbackQueryHandler{
		config: cfg,
		i18n:   i18n,
	}

	// Register all callback handlers
	h.handlers = []CallbackHandler{
		NewAgentListHandler(cfg, i18n, "al:", "ca:", true),            // Chat agent list
		NewAgentListHandler(cfg, i18n, "ial:", "ica:", false),         // Image agent list
		NewModelListHandler(cfg, i18n, "ca:", "al:", "cm:", true),     // Chat model list
		NewModelListHandler(cfg, i18n, "ica:", "ial:", "icm:", false), // Image model list
		NewModelChangeHandler(cfg, i18n, "cm:", true),                 // Chat model change
		NewModelChangeHandler(cfg, i18n, "icm:", false),               // Image model change
	}

	return h
}

// Handle processes callback queries
func (h *CallbackQueryHandler) Handle(update *tgbotapi.Update, ctx *config.WorkerContext) error {
	if update.CallbackQuery == nil {
		return nil
	}

	query := update.CallbackQuery

	// Get client from context
	client, ok := ctx.Bot.(*api.Client)
	if !ok || client == nil {
		return h.answerCallbackQuery(client, query.ID, "ERROR: Bot client not available")
	}

	// Check if message exists
	if query.Message == nil {
		return nil
	}

	// Find and execute the appropriate handler
	for _, handler := range h.handlers {
		if strings.HasPrefix(query.Data, handler.Prefix()) {
			// Check permissions if needed
			authChecker := handler.NeedAuth()
			if authChecker != nil {
				// Check authorization
				authorized, err := authChecker.Check(query.Message, ctx)
				if err != nil {
					return h.answerCallbackQuery(client, query.ID, "ERROR: Authorization check failed")
				}

				if !authorized {
					return h.answerCallbackQuery(client, query.ID, "ERROR: Permission denied")
				}
			}

			// Handle the callback
			if err := handler.Handle(query, query.Data, ctx); err != nil {
				slog.Error("Failed to handle callback query", "error", err, "data", query.Data)
				return h.answerCallbackQuery(client, query.ID, fmt.Sprintf("ERROR: %v", err))
			}

			return h.answerCallbackQuery(client, query.ID, "")
		}
	}

	slog.Debug("No handler found for callback query", "data", query.Data)
	return nil
}

// answerCallbackQuery sends an answer to a callback query
func (h *CallbackQueryHandler) answerCallbackQuery(client *api.Client, callbackQueryID string, text string) error {
	if client == nil {
		return nil
	}
	return client.AnswerCallbackQuery(callbackQueryID, text)
}

// AgentListHandler handles agent list callbacks (al: and ial:)
type AgentListHandler struct {
	config            *config.Config
	i18n              *i18n.I18n
	prefix            string
	changeAgentPrefix string
	isChat            bool // true for chat agents, false for image agents
}

// NewAgentListHandler creates a new agent list handler
func NewAgentListHandler(cfg *config.Config, i18n *i18n.I18n, prefix, changeAgentPrefix string, isChat bool) *AgentListHandler {
	return &AgentListHandler{
		config:            cfg,
		i18n:              i18n,
		prefix:            prefix,
		changeAgentPrefix: changeAgentPrefix,
		isChat:            isChat,
	}
}

func (h *AgentListHandler) Prefix() string {
	return h.prefix
}

func (h *AgentListHandler) NeedAuth() command.AuthChecker {
	return command.ShareModeGroup
}

func (h *AgentListHandler) Handle(query *tgbotapi.CallbackQuery, data string, ctx *config.WorkerContext) error {
	// Get client from context
	client, ok := ctx.Bot.(*api.Client)
	if !ok || client == nil {
		return fmt.Errorf("bot client not available")
	}

	// Get available agents
	var names []string
	if h.isChat {
		agents := agent.GetChatAgents()
		for _, ag := range agents {
			if ag.Enable(h.config) {
				names = append(names, ag.Name())
			}
		}
	} else {
		agents := agent.GetImageAgents()
		for _, ag := range agents {
			if ag.Enable(h.config) {
				names = append(names, ag.Name())
			}
		}
	}

	// Create keyboard
	keyboard := h.createKeyboard(names)

	// Edit message
	edit := tgbotapi.NewEditMessageText(
		query.Message.Chat.ID,
		query.Message.MessageID,
		h.i18n.CallbackQuery.SelectProvider,
	)
	edit.ReplyMarkup = &keyboard

	_, err := client.BotAPI.Send(edit)
	return err
}

// createKeyboard creates an inline keyboard for agent selection
func (h *AgentListHandler) createKeyboard(names []string) tgbotapi.InlineKeyboardMarkup {
	var rows [][]tgbotapi.InlineKeyboardButton

	// Create rows with 2 buttons each
	for i := 0; i < len(names); i += 2 {
		var row []tgbotapi.InlineKeyboardButton
		for j := 0; j < 2 && i+j < len(names); j++ {
			name := names[i+j]
			// Callback data format: prefix + JSON([agent, page])
			callbackData := fmt.Sprintf("%s%s", h.changeAgentPrefix, toJSON([]interface{}{name, 0}))
			row = append(row, tgbotapi.NewInlineKeyboardButtonData(name, callbackData))
		}
		rows = append(rows, row)
	}

	return tgbotapi.NewInlineKeyboardMarkup(rows...)
}

// ModelListHandler handles model list callbacks (ca:, ica:)
type ModelListHandler struct {
	config            *config.Config
	i18n              *i18n.I18n
	prefix            string
	agentListPrefix   string
	changeModelPrefix string
	isChat            bool
}

// NewModelListHandler creates a new model list handler
func NewModelListHandler(cfg *config.Config, i18n *i18n.I18n, prefix, agentListPrefix, changeModelPrefix string, isChat bool) *ModelListHandler {
	return &ModelListHandler{
		config:            cfg,
		i18n:              i18n,
		prefix:            prefix,
		agentListPrefix:   agentListPrefix,
		changeModelPrefix: changeModelPrefix,
		isChat:            isChat,
	}
}

func (h *ModelListHandler) Prefix() string {
	return h.prefix
}

func (h *ModelListHandler) NeedAuth() command.AuthChecker {
	return command.ShareModeGroup
}

func (h *ModelListHandler) Handle(query *tgbotapi.CallbackQuery, data string, ctx *config.WorkerContext) error {
	// Get client from context
	client, ok := ctx.Bot.(*api.Client)
	if !ok || client == nil {
		return fmt.Errorf("bot client not available")
	}

	// Parse callback data: [agent, page]
	params, err := parseCallbackData(data, h.prefix)
	if err != nil {
		return fmt.Errorf("failed to parse callback data: %w", err)
	}

	if len(params) < 2 {
		return fmt.Errorf("invalid callback data format")
	}

	agentName, ok := params[0].(string)
	if !ok {
		return fmt.Errorf("invalid agent name")
	}

	page := 0
	if pageFloat, ok := params[1].(float64); ok {
		page = int(pageFloat)
	}

	// Load agent and get model list
	var models []string
	if h.isChat {
		// Create temporary user config with selected agent
		tempConfig := &storage.UserConfig{
			Values: map[string]interface{}{
				"AI_PROVIDER": agentName,
			},
		}
		ag, err := agent.LoadChatLLM(h.config, tempConfig)
		if err != nil {
			return fmt.Errorf("failed to load agent: %w", err)
		}
		models, err = ag.ModelList(h.config)
		if err != nil {
			return fmt.Errorf("failed to get model list: %w", err)
		}
	} else {
		// Create temporary user config with selected agent
		tempConfig := &storage.UserConfig{
			Values: map[string]interface{}{
				"AI_IMAGE_PROVIDER": agentName,
			},
		}
		ag, err := agent.LoadImageGen(h.config, tempConfig)
		if err != nil {
			return fmt.Errorf("failed to load agent: %w", err)
		}
		models, err = ag.ModelList(h.config)
		if err != nil {
			return fmt.Errorf("failed to get model list: %w", err)
		}
	}

	// Create keyboard
	keyboard := h.createKeyboard(models, agentName, page)

	// Edit message
	text := fmt.Sprintf("%s | %s", agentName, h.i18n.CallbackQuery.SelectModel)
	edit := tgbotapi.NewEditMessageText(
		query.Message.Chat.ID,
		query.Message.MessageID,
		text,
	)
	edit.ReplyMarkup = &keyboard

	_, err = client.BotAPI.Send(edit)
	return err
}

// createKeyboard creates an inline keyboard for model selection with pagination
func (h *ModelListHandler) createKeyboard(models []string, agentName string, page int) tgbotapi.InlineKeyboardMarkup {
	var rows [][]tgbotapi.InlineKeyboardButton

	maxRow := 10
	maxCol := int(math.Max(1, math.Min(5, float64(h.config.ModelListColumns))))
	maxPage := int(math.Ceil(float64(len(models)) / float64(maxRow*maxCol)))

	// Add model buttons
	start := page * maxRow * maxCol
	end := int(math.Min(float64(len(models)), float64((page+1)*maxRow*maxCol)))

	var currentRow []tgbotapi.InlineKeyboardButton
	for i := start; i < end; i++ {
		model := models[i]
		callbackData := fmt.Sprintf("%s%s", h.changeModelPrefix, toJSON([]interface{}{agentName, model}))
		currentRow = append(currentRow, tgbotapi.NewInlineKeyboardButtonData(model, callbackData))

		if len(currentRow) >= maxCol {
			rows = append(rows, currentRow)
			currentRow = []tgbotapi.InlineKeyboardButton{}
		}

		if len(rows) >= maxRow {
			break
		}
	}

	if len(currentRow) > 0 {
		rows = append(rows, currentRow)
	}

	// Add navigation row
	navRow := []tgbotapi.InlineKeyboardButton{
		tgbotapi.NewInlineKeyboardButtonData("<", fmt.Sprintf("%s%s", h.prefix, toJSON([]interface{}{agentName, int(math.Max(0, float64(page-1)))}))),
		tgbotapi.NewInlineKeyboardButtonData(fmt.Sprintf("%d/%d", page+1, maxPage), fmt.Sprintf("%s%s", h.prefix, toJSON([]interface{}{agentName, page}))),
		tgbotapi.NewInlineKeyboardButtonData(">", fmt.Sprintf("%s%s", h.prefix, toJSON([]interface{}{agentName, int(math.Min(float64(page+1), float64(maxPage-1)))}))),
		tgbotapi.NewInlineKeyboardButtonData("⇤", h.agentListPrefix),
	}
	rows = append(rows, navRow)

	return tgbotapi.NewInlineKeyboardMarkup(rows...)
}

// ModelChangeHandler handles model change callbacks (cm:, icm:)
type ModelChangeHandler struct {
	config *config.Config
	i18n   *i18n.I18n
	prefix string
	isChat bool
}

// NewModelChangeHandler creates a new model change handler
func NewModelChangeHandler(cfg *config.Config, i18n *i18n.I18n, prefix string, isChat bool) *ModelChangeHandler {
	return &ModelChangeHandler{
		config: cfg,
		i18n:   i18n,
		prefix: prefix,
		isChat: isChat,
	}
}

func (h *ModelChangeHandler) Prefix() string {
	return h.prefix
}

func (h *ModelChangeHandler) NeedAuth() command.AuthChecker {
	return command.ShareModeGroup
}

func (h *ModelChangeHandler) Handle(query *tgbotapi.CallbackQuery, data string, ctx *config.WorkerContext) error {
	// Get client from context
	client, ok := ctx.Bot.(*api.Client)
	if !ok || client == nil {
		return fmt.Errorf("bot client not available")
	}

	// Parse callback data: [agent, model]
	params, err := parseCallbackData(data, h.prefix)
	if err != nil {
		return fmt.Errorf("failed to parse callback data: %w", err)
	}

	if len(params) < 2 {
		return fmt.Errorf("invalid callback data format")
	}

	agentName, ok := params[0].(string)
	if !ok {
		return fmt.Errorf("invalid agent name")
	}

	modelName, ok := params[1].(string)
	if !ok {
		return fmt.Errorf("invalid model name")
	}

	// Load agent to get model key
	var modelKey string
	if h.isChat {
		tempConfig := &storage.UserConfig{
			Values: map[string]interface{}{
				"AI_PROVIDER": agentName,
			},
		}
		ag, err := agent.LoadChatLLM(h.config, tempConfig)
		if err != nil {
			return fmt.Errorf("failed to load agent: %w", err)
		}
		modelKey = ag.ModelKey()
	} else {
		tempConfig := &storage.UserConfig{
			Values: map[string]interface{}{
				"AI_IMAGE_PROVIDER": agentName,
			},
		}
		ag, err := agent.LoadImageGen(h.config, tempConfig)
		if err != nil {
			return fmt.Errorf("failed to load agent: %w", err)
		}
		modelKey = ag.ModelKey()
	}

	// Get session context
	sessionCtx := command.NewSessionContext(query.Message, ctx.ShareContext.BotID, h.config.GroupChatBotShareMode)

	// Load current user config
	userConfig, err := ctx.DB.GetUserConfig(sessionCtx)
	if err != nil {
		return fmt.Errorf("failed to load user config: %w", err)
	}

	// Check permission if ENABLE_USER_SETTING is false
	if !h.config.EnableUserSetting {
		if ctx.PermissionChecker == nil {
			return fmt.Errorf("permission checker not configured")
		}

		canModify, err := ctx.PermissionChecker.CanModifyConfig(int64(query.From.ID), query.Message.Chat.ID, ctx)
		if err != nil {
			return fmt.Errorf("failed to check permissions: %w", err)
		}

		if !canModify {
			// Send error message
			text := "❌ User settings are disabled. Only administrators can modify configuration."
			edit := tgbotapi.NewEditMessageText(
				query.Message.Chat.ID,
				query.Message.MessageID,
				text,
			)
			_, _ = client.BotAPI.Send(edit)
			return fmt.Errorf("user settings are disabled, only administrators can modify configuration")
		}
	}

	// Update user config
	if userConfig.Values == nil {
		userConfig.Values = make(map[string]interface{})
	}

	if h.isChat {
		userConfig.Values["AI_PROVIDER"] = agentName
	} else {
		userConfig.Values["AI_IMAGE_PROVIDER"] = agentName
	}
	userConfig.Values[modelKey] = modelName

	// Update DEFINE_KEYS
	keys := map[string]bool{}
	for _, key := range userConfig.DefineKeys {
		keys[key] = true
	}
	if h.isChat {
		keys["AI_PROVIDER"] = true
	} else {
		keys["AI_IMAGE_PROVIDER"] = true
	}
	keys[modelKey] = true

	userConfig.DefineKeys = []string{}
	for key := range keys {
		userConfig.DefineKeys = append(userConfig.DefineKeys, key)
	}

	// Save user config
	if err := ctx.DB.SaveUserConfig(sessionCtx, userConfig); err != nil {
		return fmt.Errorf("failed to save user config: %w", err)
	}

	slog.Info("Model changed", "agent", agentName, "model", modelName)

	// Edit message
	text := fmt.Sprintf("%s %s > %s", h.i18n.CallbackQuery.ChangeModel, agentName, modelName)
	edit := tgbotapi.NewEditMessageText(
		query.Message.Chat.ID,
		query.Message.MessageID,
		text,
	)

	_, err = client.BotAPI.Send(edit)
	return err
}

// Helper functions

// toJSON converts a value to JSON string
func toJSON(v interface{}) string {
	data, _ := json.Marshal(v)
	return string(data)
}

// parseCallbackData parses callback data by removing prefix and parsing JSON
func parseCallbackData(data, prefix string) ([]interface{}, error) {
	if !strings.HasPrefix(data, prefix) {
		return nil, fmt.Errorf("data does not start with prefix")
	}

	jsonStr := strings.TrimPrefix(data, prefix)
	var params []interface{}
	if err := json.Unmarshal([]byte(jsonStr), &params); err != nil {
		return nil, err
	}

	return params, nil
}
