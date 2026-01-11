package server

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/tbxark/ChatGPT-Telegram-Workers/go_version/internal/config"
	"github.com/tbxark/ChatGPT-Telegram-Workers/go_version/internal/i18n"
	"github.com/tbxark/ChatGPT-Telegram-Workers/go_version/internal/sillytavern"
	"github.com/tbxark/ChatGPT-Telegram-Workers/go_version/internal/telegram/api"
	"github.com/tbxark/ChatGPT-Telegram-Workers/go_version/internal/telegram/command"
)

// handleRoot handles GET / - displays welcome page
func (s *Server) handleRoot(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Only handle exact root path
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}

	html := s.generateWelcomePage()
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(html))
}

// handleInit handles GET /init - webhook binding
func (s *Server) handleInit(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	html := s.initializeWebhooks(r)
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(html))
}

// handleTelegram handles POST /telegram/:token/webhook and /telegram/:token/safehook
func (s *Server) handleTelegram(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Parse path: /telegram/:token/webhook or /telegram/:token/safehook
	path := strings.TrimPrefix(r.URL.Path, "/telegram/")
	parts := strings.Split(path, "/")

	if len(parts) != 2 {
		http.Error(w, "Invalid path", http.StatusBadRequest)
		return
	}

	token := parts[0]
	endpoint := parts[1]

	// Validate token
	if !s.isValidToken(token) {
		http.Error(w, "Invalid token", http.StatusUnauthorized)
		return
	}

	// Validate endpoint
	if endpoint != "webhook" && endpoint != "safehook" {
		http.Error(w, "Invalid endpoint", http.StatusNotFound)
		return
	}

	// TODO: Implement webhook processing
	// This will be implemented in later tasks (task 8)
	response := map[string]interface{}{
		"ok":      true,
		"message": "Webhook processing will be implemented in task 8",
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

// isValidToken checks if the token is in the list of available tokens
func (s *Server) isValidToken(token string) bool {
	for _, t := range s.config.TelegramAvailableTokens {
		if t == token {
			return true
		}
	}
	return false
}

// generateWelcomePage generates the HTML welcome page
func (s *Server) generateWelcomePage() string {
	html := fmt.Sprintf(`<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <title>Telegram Bot - Go Version</title>
    <style>
        body {
            font-family: 'Segoe UI', Tahoma, Geneva, Verdana, sans-serif;
            max-width: 800px;
            margin: 50px auto;
            padding: 20px;
            background: linear-gradient(135deg, #667eea 0%%, #764ba2 100%%);
            color: #333;
        }
        .container {
            background: white;
            border-radius: 10px;
            padding: 30px;
            box-shadow: 0 10px 40px rgba(0,0,0,0.1);
        }
        h1 {
            color: #667eea;
            margin-top: 0;
        }
        .info {
            background: #f5f5f5;
            padding: 15px;
            border-radius: 5px;
            margin: 15px 0;
        }
        .info-item {
            margin: 10px 0;
        }
        .label {
            font-weight: bold;
            color: #555;
        }
        .value {
            color: #667eea;
        }
        .status {
            display: inline-block;
            padding: 5px 15px;
            background: #4caf50;
            color: white;
            border-radius: 20px;
            font-size: 14px;
        }
        .endpoints {
            margin-top: 20px;
        }
        .endpoint {
            background: #e3f2fd;
            padding: 10px;
            margin: 10px 0;
            border-radius: 5px;
            font-family: monospace;
        }
        a {
            color: #667eea;
            text-decoration: none;
        }
        a:hover {
            text-decoration: underline;
        }
    </style>
</head>
<body>
    <div class="container">
        <h1>🤖 Telegram Bot - Go Version</h1>
        <div class="status">�?Running</div>
        
        <div class="info">
            <div class="info-item">
                <span class="label">Language:</span>
                <span class="value">%s</span>
            </div>
            <div class="info-item">
                <span class="label">Port:</span>
                <span class="value">%d</span>
            </div>
        </div>

        <div class="endpoints">
            <h2>Available Endpoints</h2>
            <div class="endpoint">
                <strong>GET /</strong> - This welcome page
            </div>
            <div class="endpoint">
                <strong>GET /init</strong> - Initialize webhook and commands
            </div>
            <div class="endpoint">
                <strong>POST /telegram/:token/webhook</strong> - Telegram webhook endpoint
            </div>
            <div class="endpoint">
                <strong>POST /telegram/:token/safehook</strong> - Telegram webhook with API guard
            </div>
        </div>

        <div style="margin-top: 30px; text-align: center; color: #999; font-size: 14px;">
            <p>Converted from Node.js to Go</p>
            <p>Original: <a href="https://github.com/TBXark/ChatGPT-Telegram-Workers" target="_blank">ChatGPT-Telegram-Workers</a></p>
        </div>
    </div>
</body>
</html>`, s.config.Language, s.config.Port)

	return html
}

// initializeWebhooks sets up webhooks and commands for all configured bot tokens
func (s *Server) initializeWebhooks(r *http.Request) string {
	result := make(map[string]map[string]interface{})
	domain := r.Host

	// Determine webhook mode (safehook if API_GUARD is configured, otherwise webhook)
	hookMode := "webhook"
	// Note: API_GUARD is not implemented in Go version yet, so always use webhook

	// Get command scopes
	commandScopes := s.getCommandScopes()

	// Process each bot token
	for _, token := range s.config.TelegramAvailableTokens {
		token = strings.TrimSpace(token)
		if token == "" {
			continue
		}

		// Extract bot ID from token (format: botID:token)
		parts := strings.Split(token, ":")
		if len(parts) < 2 {
			log.Printf("Invalid token format: %s", token)
			continue
		}
		botID := parts[0]

		// Create Telegram API client
		client, err := api.NewClient(token, s.config.TelegramAPIDomain)
		if err != nil {
			log.Printf("Failed to create Telegram client for bot %s: %v", botID, err)
			result[botID] = map[string]interface{}{
				"error": fmt.Sprintf("Failed to create client: %v", err),
			}
			continue
		}

		result[botID] = make(map[string]interface{})

		// Set webhook
		webhookURL := fmt.Sprintf("https://%s/telegram/%s/%s", domain, token, hookMode)
		err = client.SetWebhook(webhookURL)
		if err != nil {
			result[botID]["webhook"] = map[string]interface{}{
				"ok":          false,
				"description": err.Error(),
			}
			log.Printf("Failed to set webhook for bot %s: %v", botID, err)
		} else {
			result[botID]["webhook"] = map[string]interface{}{
				"ok":          true,
				"description": fmt.Sprintf("Webhook set to %s", webhookURL),
			}
			log.Printf("Webhook set successfully for bot %s: %s", botID, webhookURL)
		}

		// Set commands for each scope
		for scopeName, commands := range commandScopes {
			scope := s.createBotCommandScope(scopeName)
			config := tgbotapi.NewSetMyCommandsWithScope(scope, commands...)

			_, err := client.Request(config)
			if err != nil {
				result[botID][scopeName] = map[string]interface{}{
					"ok":          false,
					"description": err.Error(),
				}
				log.Printf("Failed to set commands for bot %s scope %s: %v", botID, scopeName, err)
			} else {
				result[botID][scopeName] = map[string]interface{}{
					"ok":          true,
					"description": fmt.Sprintf("Commands set for scope %s", scopeName),
				}
				log.Printf("Commands set successfully for bot %s scope %s", botID, scopeName)
			}
		}
	}

	return s.generateInitResultPage(domain, result)
}

// getCommandScopes returns commands organized by scope
func (s *Server) getCommandScopes() map[string][]tgbotapi.BotCommand {
	// Create command registry
	registry := command.NewRegistry(s.config)

	// Create and set permission checker
	permChecker := config.NewDefaultPermissionChecker(s.config, command.IsGroupAdmin)
	registry.SetPermissionChecker(permChecker)

	// Load i18n
	i18nInstance := i18n.LoadI18n(s.config.Language)

	// Register all system commands
	registry.RegisterAll(
		command.NewStartCommand(s.config, i18nInstance),
		command.NewHelpCommand(s.config, i18nInstance, registry),
		command.NewNewCommand(s.config, i18nInstance),
		command.NewRedoCommand(s.config, i18nInstance),
		command.NewImgCommand(s.config, i18nInstance),
		command.NewModelsCommand(s.config, i18nInstance),
		command.NewSystemCommand(s.config, i18nInstance),
	)

	// Register SillyTavern commands if context manager is available
	if s.config.SillyTavernContextManager != nil {
		contextManager, ok := s.config.SillyTavernContextManager.(*sillytavern.ContextManager)
		if ok {
			registry.RegisterAll(
				command.NewClearCommand(s.config, contextManager),
				command.NewShareCommand(s.config, contextManager),
			)
			log.Printf("SillyTavern commands registered: /clear, /share")
		}
	}

	// Register login command (doesn't require context manager)
	registry.RegisterAll(
		command.NewLoginCommand(s.config),
	)
	log.Printf("Login command registered: /login")

	// Register clear_all command (admin only)
	registry.RegisterAll(
		command.NewClearAllCommand(s.config),
	)
	log.Printf("Admin command registered: /clear_all_chat")

	// Register configuration commands with permission control
	registry.RegisterConfigCommand(command.NewSetenvCommand(s.config, i18nInstance))
	registry.RegisterConfigCommand(command.NewSetenvsCommand(s.config, i18nInstance))
	registry.RegisterConfigCommand(command.NewDelenvCommand(s.config, i18nInstance))
	registry.RegisterConfigCommand(command.NewClearenvCommand(s.config, i18nInstance))

	// Load plugins
	if err := registry.LoadPlugins(); err != nil {
		log.Printf("Failed to load plugins: %v", err)
	}

	// Get i18n for command descriptions
	lang := s.config.Language

	// Organize commands by scope
	scopeCommandMap := map[string][]tgbotapi.BotCommand{
		"all_private_chats":       {},
		"all_group_chats":         {},
		"all_chat_administrators": {},
	}

	// Get all registered commands
	for _, cmdName := range registry.List() {
		cmd, exists := registry.Get(cmdName)
		if !exists {
			continue
		}

		// Skip hidden commands
		isHidden := false
		for _, hidden := range s.config.HideCommandButtons {
			if hidden == cmdName {
				isHidden = true
				break
			}
		}
		if isHidden {
			continue
		}

		// Get command scopes
		scopes := cmd.Scopes()
		description := cmd.Description(lang)

		if description == "" {
			continue
		}

		botCommand := tgbotapi.BotCommand{
			Command:     cmdName,
			Description: description,
		}

		// Add to appropriate scopes
		for _, scope := range scopes {
			if _, exists := scopeCommandMap[scope]; exists {
				scopeCommandMap[scope] = append(scopeCommandMap[scope], botCommand)
			}
		}
	}

	return scopeCommandMap
}

// createBotCommandScope creates a BotCommandScope from a scope name
func (s *Server) createBotCommandScope(scopeName string) tgbotapi.BotCommandScope {
	switch scopeName {
	case "all_private_chats":
		return tgbotapi.BotCommandScope{Type: "all_private_chats"}
	case "all_group_chats":
		return tgbotapi.BotCommandScope{Type: "all_group_chats"}
	case "all_chat_administrators":
		return tgbotapi.BotCommandScope{Type: "all_chat_administrators"}
	default:
		return tgbotapi.BotCommandScope{Type: "default"}
	}
}

// generateInitResultPage generates the HTML result page for webhook initialization
func (s *Server) generateInitResultPage(domain string, result map[string]map[string]interface{}) string {
	helpLink := "https://github.com/TBXark/ChatGPT-Telegram-Workers/blob/master/doc/en/DEPLOY.md"
	issueLink := "https://github.com/TBXark/ChatGPT-Telegram-Workers/issues"

	html := `<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <title>Webhook Initialization - Telegram Bot</title>
    <style>
        body {
            font-family: 'Segoe UI', Tahoma, Geneva, Verdana, sans-serif;
            max-width: 900px;
            margin: 50px auto;
            padding: 20px;
            background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
            color: #333;
        }
        .container {
            background: white;
            border-radius: 10px;
            padding: 30px;
            box-shadow: 0 10px 40px rgba(0,0,0,0.1);
        }
        h1 {
            color: #667eea;
            margin-top: 0;
        }
        h2 {
            color: #555;
            border-bottom: 2px solid #667eea;
            padding-bottom: 10px;
        }
        h3 {
            color: #764ba2;
            margin-top: 20px;
        }
        .bot-section {
            background: #f9f9f9;
            padding: 20px;
            margin: 20px 0;
            border-radius: 8px;
            border-left: 4px solid #667eea;
        }
        .result-item {
            background: white;
            padding: 12px;
            margin: 10px 0;
            border-radius: 5px;
            border-left: 3px solid #ddd;
        }
        .result-item.success {
            border-left-color: #4caf50;
            background: #f1f8f4;
        }
        .result-item.error {
            border-left-color: #f44336;
            background: #fef1f0;
        }
        .status-ok {
            color: #4caf50;
            font-weight: bold;
        }
        .status-error {
            color: #f44336;
            font-weight: bold;
        }
        .scope-name {
            font-weight: bold;
            color: #667eea;
        }
        .description {
            color: #666;
            font-size: 14px;
            margin-top: 5px;
        }
        .footer {
            margin-top: 30px;
            padding-top: 20px;
            border-top: 1px solid #ddd;
            text-align: center;
            color: #999;
            font-size: 14px;
        }
        .footer a {
            color: #667eea;
            text-decoration: none;
        }
        .footer a:hover {
            text-decoration: underline;
        }
        .warning {
            background: #fff3cd;
            border: 1px solid #ffc107;
            color: #856404;
            padding: 15px;
            border-radius: 5px;
            margin: 20px 0;
        }
        code {
            background: #f5f5f5;
            padding: 2px 6px;
            border-radius: 3px;
            font-family: 'Courier New', monospace;
        }
    </style>
</head>
<body>
    <div class="container">
        <h1>🤖 Webhook Initialization</h1>
        <h2>` + domain + `</h2>`

	if len(s.config.TelegramAvailableTokens) == 0 {
		html += `
        <div class="warning">
            <strong>⚠️ Warning:</strong> No bot tokens configured. 
            Please set the <code>TELEGRAM_AVAILABLE_TOKENS</code> environment variable.
        </div>`
	} else {
		for botID, results := range result {
			html += fmt.Sprintf(`
        <div class="bot-section">
            <h3>Bot: %s</h3>`, botID)

			for scope, data := range results {
				dataMap, ok := data.(map[string]interface{})
				if !ok {
					continue
				}

				isOk := false
				if okVal, exists := dataMap["ok"]; exists {
					isOk, _ = okVal.(bool)
				}

				description := ""
				if descVal, exists := dataMap["description"]; exists {
					description, _ = descVal.(string)
				}

				statusClass := "error"
				statusText := "�?Failed"
				if isOk {
					statusClass = "success"
					statusText = "�?Success"
				}

				html += fmt.Sprintf(`
            <div class="result-item %s">
                <div><span class="scope-name">%s:</span> <span class="status-%s">%s</span></div>
                <div class="description">%s</div>
            </div>`, statusClass, scope, statusClass, statusText, description)
			}

			html += `
        </div>`
		}
	}

	html += fmt.Sprintf(`
        <div class="footer">
            <p>For more information, please visit <a href="%s" target="_blank">Deployment Guide</a></p>
            <p>If you have any questions, please visit <a href="%s" target="_blank">Issues</a></p>
            <p style="margin-top: 20px;">Converted from Node.js to Go</p>
        </div>
    </div>
</body>
</html>`, helpLink, issueLink)

	return html
}
