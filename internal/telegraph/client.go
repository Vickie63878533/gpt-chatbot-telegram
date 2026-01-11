package telegraph

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// Client represents a Telegraph API client
type Client struct {
	accessToken string
	httpClient  *http.Client
}

// NewClient creates a new Telegraph client
// It automatically creates a Telegraph account
func NewClient() (*Client, error) {
	client := &Client{
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}

	// Create Telegraph account
	if err := client.createAccount(); err != nil {
		return nil, fmt.Errorf("failed to create Telegraph account: %w", err)
	}

	return client, nil
}

// createAccount creates a new Telegraph account
func (c *Client) createAccount() error {
	reqBody := map[string]interface{}{
		"short_name": "TelegramBot",
		"author_name": "Telegram Bot",
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return err
	}

	resp, err := c.httpClient.Post(
		"https://api.telegra.ph/createAccount",
		"application/json",
		bytes.NewBuffer(jsonData),
	)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	var result struct {
		OK     bool `json:"ok"`
		Result struct {
			AccessToken string `json:"access_token"`
		} `json:"result"`
		Error string `json:"error"`
	}

	if err := json.Unmarshal(body, &result); err != nil {
		return err
	}

	if !result.OK {
		return fmt.Errorf("Telegraph API error: %s", result.Error)
	}

	c.accessToken = result.Result.AccessToken
	return nil
}

// CreatePage creates a new Telegraph page
func (c *Client) CreatePage(title string, content string) (string, error) {
	// Convert HTML content to Telegraph format
	// Telegraph expects content as an array of Node objects
	// For simplicity, we'll use a single text node with HTML
	reqBody := map[string]interface{}{
		"access_token": c.accessToken,
		"title":        title,
		"content":      []interface{}{content},
		"return_content": false,
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return "", err
	}

	resp, err := c.httpClient.Post(
		"https://api.telegra.ph/createPage",
		"application/json",
		bytes.NewBuffer(jsonData),
	)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	var result struct {
		OK     bool `json:"ok"`
		Result struct {
			Path string `json:"path"`
			URL  string `json:"url"`
		} `json:"result"`
		Error string `json:"error"`
	}

	if err := json.Unmarshal(body, &result); err != nil {
		return "", err
	}

	if !result.OK {
		return "", fmt.Errorf("Telegraph API error: %s", result.Error)
	}

	return result.Result.URL, nil
}


