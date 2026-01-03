package i18n

import (
	"testing"
)

func TestLoadI18n(t *testing.T) {
	tests := []struct {
		name     string
		lang     string
		expected string // Expected SystemInitMessage
	}{
		{
			name:     "English",
			lang:     "en",
			expected: "You are a helpful assistant",
		},
		{
			name:     "English US",
			lang:     "en-us",
			expected: "You are a helpful assistant",
		},
		{
			name:     "Simplified Chinese",
			lang:     "zh-cn",
			expected: "你是一个得力的助手",
		},
		{
			name:     "Simplified Chinese (hans)",
			lang:     "zh-hans",
			expected: "你是一个得力的助手",
		},
		{
			name:     "Traditional Chinese",
			lang:     "zh-hant",
			expected: "你是一個得力的助手",
		},
		{
			name:     "Traditional Chinese (tw)",
			lang:     "zh-tw",
			expected: "你是一個得力的助手",
		},
		{
			name:     "Portuguese",
			lang:     "pt",
			expected: "Você é um assistente útil",
		},
		{
			name:     "Portuguese Brazil",
			lang:     "pt-br",
			expected: "Você é um assistente útil",
		},
		{
			name:     "Default (unknown language)",
			lang:     "unknown",
			expected: "You are a helpful assistant",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			i18n := LoadI18n(tt.lang)
			if i18n.Env.SystemInitMessage != tt.expected {
				t.Errorf("LoadI18n(%q).Env.SystemInitMessage = %q, want %q",
					tt.lang, i18n.Env.SystemInitMessage, tt.expected)
			}
		})
	}
}

func TestAllLanguagesHaveRequiredFields(t *testing.T) {
	languages := []string{"en", "zh-hans", "zh-hant", "pt"}

	for _, lang := range languages {
		t.Run(lang, func(t *testing.T) {
			i18n := LoadI18n(lang)

			// Check Env fields
			if i18n.Env.SystemInitMessage == "" {
				t.Error("SystemInitMessage is empty")
			}

			// Check Command.Help fields
			if i18n.Command.Help.Summary == "" {
				t.Error("Command.Help.Summary is empty")
			}
			if i18n.Command.Help.Help == "" {
				t.Error("Command.Help.Help is empty")
			}
			if i18n.Command.Help.New == "" {
				t.Error("Command.Help.New is empty")
			}
			if i18n.Command.Help.Start == "" {
				t.Error("Command.Help.Start is empty")
			}
			if i18n.Command.Help.Img == "" {
				t.Error("Command.Help.Img is empty")
			}
			if i18n.Command.Help.Version == "" {
				t.Error("Command.Help.Version is empty")
			}
			if i18n.Command.Help.Setenv == "" {
				t.Error("Command.Help.Setenv is empty")
			}
			if i18n.Command.Help.Setenvs == "" {
				t.Error("Command.Help.Setenvs is empty")
			}
			if i18n.Command.Help.Delenv == "" {
				t.Error("Command.Help.Delenv is empty")
			}
			if i18n.Command.Help.Clearenv == "" {
				t.Error("Command.Help.Clearenv is empty")
			}
			if i18n.Command.Help.System == "" {
				t.Error("Command.Help.System is empty")
			}
			if i18n.Command.Help.Redo == "" {
				t.Error("Command.Help.Redo is empty")
			}
			if i18n.Command.Help.Models == "" {
				t.Error("Command.Help.Models is empty")
			}
			if i18n.Command.Help.Echo == "" {
				t.Error("Command.Help.Echo is empty")
			}

			// Check Command.New fields
			if i18n.Command.New.NewChatStart == "" {
				t.Error("Command.New.NewChatStart is empty")
			}

			// Check CallbackQuery fields
			if i18n.CallbackQuery.OpenModelList == "" {
				t.Error("CallbackQuery.OpenModelList is empty")
			}
			if i18n.CallbackQuery.SelectProvider == "" {
				t.Error("CallbackQuery.SelectProvider is empty")
			}
			if i18n.CallbackQuery.SelectModel == "" {
				t.Error("CallbackQuery.SelectModel is empty")
			}
			if i18n.CallbackQuery.ChangeModel == "" {
				t.Error("CallbackQuery.ChangeModel is empty")
			}
		})
	}
}

func TestCommandDescriptions(t *testing.T) {
	// Test that each language has unique command descriptions
	languages := map[string]string{
		"en":      "Start a new conversation",
		"zh-hans": "发起新的对话",
		"zh-hant": "開始一個新對話",
		"pt":      "Iniciar uma nova conversa",
	}

	for lang, expected := range languages {
		t.Run(lang, func(t *testing.T) {
			i18n := LoadI18n(lang)
			if i18n.Command.Help.New != expected {
				t.Errorf("LoadI18n(%q).Command.Help.New = %q, want %q",
					lang, i18n.Command.Help.New, expected)
			}
		})
	}
}
