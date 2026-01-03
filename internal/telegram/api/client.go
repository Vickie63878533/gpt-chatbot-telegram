package api

import (
	"fmt"
	"net/http"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// Client wraps the Telegram Bot API client with custom configuration
type Client struct {
	*tgbotapi.BotAPI
	apiDomain string
}

// NewClient creates a new Telegram API client with custom API domain support
func NewClient(token string, apiDomain string) (*Client, error) {
	if token == "" {
		return nil, fmt.Errorf("telegram bot token is required")
	}

	// Create the bot API instance without validation
	bot := &tgbotapi.BotAPI{
		Token:  token,
		Client: &http.Client{},
		Buffer: 100,
	}

	// Set the API endpoint
	if apiDomain != "" && apiDomain != "https://api.telegram.org" {
		bot.SetAPIEndpoint(apiDomain + "/bot%s/%s")
	} else {
		bot.SetAPIEndpoint(tgbotapi.APIEndpoint)
	}

	client := &Client{
		BotAPI:    bot,
		apiDomain: apiDomain,
	}

	return client, nil
}

// NewClientWithHTTPClient creates a new Telegram API client with a custom HTTP client
func NewClientWithHTTPClient(token string, apiDomain string, httpClient *http.Client) (*Client, error) {
	if token == "" {
		return nil, fmt.Errorf("telegram bot token is required")
	}

	// Create the bot API instance with custom HTTP client
	bot := &tgbotapi.BotAPI{
		Token:  token,
		Client: httpClient,
		Buffer: 100,
	}

	// Set the API endpoint
	if apiDomain != "" && apiDomain != "https://api.telegram.org" {
		bot.SetAPIEndpoint(apiDomain + "/bot%s/%s")
	} else {
		bot.SetAPIEndpoint(tgbotapi.APIEndpoint)
	}

	client := &Client{
		BotAPI:    bot,
		apiDomain: apiDomain,
	}

	return client, nil
}

// GetAPIDomain returns the configured API domain
func (c *Client) GetAPIDomain() string {
	return c.apiDomain
}

// SetDebug enables or disables debug mode
func (c *Client) SetDebug(debug bool) {
	c.BotAPI.Debug = debug
}

// GetMe returns information about the bot
func (c *Client) GetMe() (tgbotapi.User, error) {
	return c.BotAPI.GetMe()
}

// Send sends a Chattable to Telegram
func (c *Client) Send(config tgbotapi.Chattable) (tgbotapi.Message, error) {
	return c.BotAPI.Send(config)
}

// Request makes a request to the Telegram API
func (c *Client) Request(config tgbotapi.Chattable) (*tgbotapi.APIResponse, error) {
	return c.BotAPI.Request(config)
}

// SetWebhook sets the webhook URL for the bot
func (c *Client) SetWebhook(webhookURL string) error {
	webhookConfig, _ := tgbotapi.NewWebhook(webhookURL)
	_, err := c.BotAPI.Request(webhookConfig)
	return err
}

// RemoveWebhook removes the webhook
func (c *Client) RemoveWebhook() error {
	_, err := c.BotAPI.Request(tgbotapi.DeleteWebhookConfig{})
	return err
}

// SetMyCommands sets the bot's command list
func (c *Client) SetMyCommands(commands []tgbotapi.BotCommand) error {
	config := tgbotapi.NewSetMyCommands(commands...)
	_, err := c.BotAPI.Request(config)
	return err
}

// GetChatAdministrators gets the list of administrators in a chat
func (c *Client) GetChatAdministrators(chatID int64) ([]tgbotapi.ChatMember, error) {
	config := tgbotapi.ChatAdministratorsConfig{
		ChatConfig: tgbotapi.ChatConfig{
			ChatID: chatID,
		},
	}
	return c.BotAPI.GetChatAdministrators(config)
}

// SendChatAction sends a chat action (typing, upload_photo, etc.)
func (c *Client) SendChatAction(chatID int64, action string) error {
	config := tgbotapi.NewChatAction(chatID, action)
	_, err := c.BotAPI.Request(config)
	return err
}

// GetFileDirectURL gets the direct URL for a file
func (c *Client) GetFileDirectURL(fileID string) (string, error) {
	file, err := c.BotAPI.GetFile(tgbotapi.FileConfig{FileID: fileID})
	if err != nil {
		return "", err
	}
	return file.Link(c.BotAPI.Token), nil
}

// AnswerCallbackQuery sends an answer to a callback query
func (c *Client) AnswerCallbackQuery(callbackQueryID string, text string) error {
	config := tgbotapi.NewCallback(callbackQueryID, text)
	_, err := c.BotAPI.Request(config)
	return err
}

// NewDefaultHTTPClient creates a default HTTP client with reasonable timeouts
func NewDefaultHTTPClient() *http.Client {
	return &http.Client{
		Timeout: 30 * time.Second,
		Transport: &http.Transport{
			MaxIdleConns:        100,
			MaxIdleConnsPerHost: 10,
			IdleConnTimeout:     90 * time.Second,
		},
	}
}
