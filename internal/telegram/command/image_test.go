package command

import (
	"testing"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/tbxark/ChatGPT-Telegram-Workers/go_version/internal/config"
	"github.com/tbxark/ChatGPT-Telegram-Workers/go_version/internal/i18n"
)

func TestImgCommand_Name(t *testing.T) {
	cfg := &config.Config{}
	i18n := i18n.LoadI18n("en")
	cmd := NewImgCommand(cfg, i18n)

	if cmd.Name() != "img" {
		t.Errorf("Expected command name 'img', got '%s'", cmd.Name())
	}
}

func TestImgCommand_Description(t *testing.T) {
	cfg := &config.Config{}
	i18n := i18n.LoadI18n("en")
	cmd := NewImgCommand(cfg, i18n)

	desc := cmd.Description("en")
	if desc == "" {
		t.Error("Expected non-empty description")
	}
}

func TestImgCommand_Scopes(t *testing.T) {
	cfg := &config.Config{}
	i18n := i18n.LoadI18n("en")
	cmd := NewImgCommand(cfg, i18n)

	scopes := cmd.Scopes()
	if len(scopes) == 0 {
		t.Error("Expected non-empty scopes")
	}

	expectedScopes := []string{"all_private_chats", "all_chat_administrators"}
	if len(scopes) != len(expectedScopes) {
		t.Errorf("Expected %d scopes, got %d", len(expectedScopes), len(scopes))
	}
}

func TestModelsCommand_Name(t *testing.T) {
	cfg := &config.Config{}
	i18n := i18n.LoadI18n("en")
	cmd := NewModelsCommand(cfg, i18n)

	if cmd.Name() != "models" {
		t.Errorf("Expected command name 'models', got '%s'", cmd.Name())
	}
}

func TestModelsCommand_Description(t *testing.T) {
	cfg := &config.Config{}
	i18n := i18n.LoadI18n("en")
	cmd := NewModelsCommand(cfg, i18n)

	desc := cmd.Description("en")
	if desc == "" {
		t.Error("Expected non-empty description")
	}
}

func TestModelsCommand_Scopes(t *testing.T) {
	cfg := &config.Config{}
	i18n := i18n.LoadI18n("en")
	cmd := NewModelsCommand(cfg, i18n)

	scopes := cmd.Scopes()
	if len(scopes) == 0 {
		t.Error("Expected non-empty scopes")
	}

	expectedScopes := []string{"all_private_chats", "all_group_chats", "all_chat_administrators"}
	if len(scopes) != len(expectedScopes) {
		t.Errorf("Expected %d scopes, got %d", len(expectedScopes), len(scopes))
	}
}

func TestModelsCommand_BuildProviderKeyboard(t *testing.T) {
	cfg := &config.Config{
		ModelListColumns: 2,
	}
	i18n := i18n.LoadI18n("en")
	cmd := NewModelsCommand(cfg, i18n)

	// Test with empty agents
	keyboard := cmd.buildProviderKeyboard(nil)
	if len(keyboard.InlineKeyboard) != 0 {
		t.Errorf("Expected empty keyboard for nil agents, got %d rows", len(keyboard.InlineKeyboard))
	}
}

func TestImgCommand_Handle_EmptyPrompt(t *testing.T) {
	cfg := &config.Config{}
	i18n := i18n.LoadI18n("en")
	cmd := NewImgCommand(cfg, i18n)

	message := &tgbotapi.Message{
		Chat: &tgbotapi.Chat{ID: 123},
		From: &tgbotapi.User{ID: 456},
	}

	ctx := &config.WorkerContext{}

	// Test with empty prompt
	err := cmd.Handle(message, "", ctx)
	if err == nil {
		t.Error("Expected error for empty prompt, got nil")
	}
}
