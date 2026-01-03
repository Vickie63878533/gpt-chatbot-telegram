package config_test

import (
	"fmt"
	"os"

	"github.com/tbxark/ChatGPT-Telegram-Workers/go_version/internal/config"
)

func ExampleLoadConfig() {
	// Set required environment variable
	os.Setenv("TELEGRAM_AVAILABLE_TOKENS", "123456:ABC-DEF")
	defer os.Unsetenv("TELEGRAM_AVAILABLE_TOKENS")

	// Load configuration
	cfg, err := config.LoadConfig()
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	fmt.Printf("AI Provider: %s\n", cfg.AIProvider)
	fmt.Printf("Port: %d\n", cfg.Port)
	fmt.Printf("Language: %s\n", cfg.Language)

	// Output:
	// AI Provider: auto
	// Port: 8080
	// Language: zh-cn
}

func ExampleNewShareContext() {
	// Create share context from bot token
	shareCtx, err := config.NewShareContext("123456:ABC-DEF")
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	fmt.Printf("Bot ID: %d\n", shareCtx.BotID)
	fmt.Printf("Bot Token: %s\n", shareCtx.BotToken)

	// Output:
	// Bot ID: 123456
	// Bot Token: 123456:ABC-DEF
}

func ExampleNewSessionContextFromChat() {
	userID := int64(789)

	// Private chat
	ctx1 := config.NewSessionContextFromChat(123, 456, false, false, nil, nil)
	fmt.Printf("Private chat - ChatID: %d, BotID: %d, UserID: %v\n",
		ctx1.ChatID, ctx1.BotID, ctx1.UserID)

	// Group chat (shared mode)
	ctx2 := config.NewSessionContextFromChat(123, 456, true, true, &userID, nil)
	fmt.Printf("Group shared - ChatID: %d, BotID: %d, UserID: %v\n",
		ctx2.ChatID, ctx2.BotID, ctx2.UserID)

	// Group chat (non-shared mode)
	ctx3 := config.NewSessionContextFromChat(123, 456, true, false, &userID, nil)
	fmt.Printf("Group non-shared - ChatID: %d, BotID: %d, UserID: %v\n",
		ctx3.ChatID, ctx3.BotID, *ctx3.UserID)

	// Output:
	// Private chat - ChatID: 123, BotID: 456, UserID: <nil>
	// Group shared - ChatID: 123, BotID: 456, UserID: <nil>
	// Group non-shared - ChatID: 123, BotID: 456, UserID: 789
}

func ExampleWorkerContext_SetUserConfigValue() {
	shareCtx := config.ShareContext{
		BotToken: "123456:ABC",
		BotID:    123456,
	}
	cfg := &config.Config{}
	wc := config.NewWorkerContext(shareCtx, nil, cfg)

	lockedKeys := []string{"OPENAI_API_BASE"}

	// Set a normal config value
	err := wc.SetUserConfigValue("AI_PROVIDER", "azure", lockedKeys)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}
	fmt.Println("Set AI_PROVIDER to azure")

	// Try to set a locked key
	err = wc.SetUserConfigValue("OPENAI_API_BASE", "https://custom.api", lockedKeys)
	if err != nil {
		fmt.Printf("Cannot set locked key: %v\n", err)
	}

	// Output:
	// Set AI_PROVIDER to azure
	// Cannot set locked key: configuration key 'OPENAI_API_BASE' is locked and cannot be modified
}
