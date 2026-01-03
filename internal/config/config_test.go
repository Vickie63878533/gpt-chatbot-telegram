package config

import (
	"os"
	"testing"
)

func TestLoadConfig(t *testing.T) {
	// Set required environment variable
	os.Setenv("TELEGRAM_AVAILABLE_TOKENS", "123456:ABC-DEF")
	defer os.Unsetenv("TELEGRAM_AVAILABLE_TOKENS")

	cfg, err := LoadConfig()
	if err != nil {
		t.Fatalf("LoadConfig() failed: %v", err)
	}

	if cfg == nil {
		t.Fatal("LoadConfig() returned nil config")
	}

	// Check required field
	if len(cfg.TelegramAvailableTokens) == 0 {
		t.Error("TelegramAvailableTokens should not be empty")
	}

	// Check default values
	if cfg.AIProvider != "auto" {
		t.Errorf("Expected AIProvider to be 'auto', got '%s'", cfg.AIProvider)
	}

	if cfg.Port != 8080 {
		t.Errorf("Expected Port to be 8080, got %d", cfg.Port)
	}

	if cfg.Language != "zh-cn" {
		t.Errorf("Expected Language to be 'zh-cn', got '%s'", cfg.Language)
	}
}

func TestValidate(t *testing.T) {
	tests := []struct {
		name    string
		config  *Config
		wantErr bool
	}{
		{
			name: "valid config",
			config: &Config{
				TelegramAvailableTokens:   []string{"123456:ABC"},
				Port:                      8080,
				DefaultParseMode:          "Markdown",
				TelegramImageTransferMode: "base64",
				Language:                  "zh-cn",
			},
			wantErr: false,
		},
		{
			name: "missing telegram tokens",
			config: &Config{
				Port:                      8080,
				DefaultParseMode:          "Markdown",
				TelegramImageTransferMode: "base64",
				Language:                  "zh-cn",
			},
			wantErr: true,
		},
		{
			name: "invalid port",
			config: &Config{
				TelegramAvailableTokens:   []string{"123456:ABC"},
				Port:                      0,
				DefaultParseMode:          "Markdown",
				TelegramImageTransferMode: "base64",
				Language:                  "zh-cn",
			},
			wantErr: true,
		},
		{
			name: "invalid parse mode",
			config: &Config{
				TelegramAvailableTokens:   []string{"123456:ABC"},
				Port:                      8080,
				DefaultParseMode:          "Invalid",
				TelegramImageTransferMode: "base64",
				Language:                  "zh-cn",
			},
			wantErr: true,
		},
		{
			name: "invalid image transfer mode",
			config: &Config{
				TelegramAvailableTokens:   []string{"123456:ABC"},
				Port:                      8080,
				DefaultParseMode:          "Markdown",
				TelegramImageTransferMode: "invalid",
				Language:                  "zh-cn",
			},
			wantErr: true,
		},
		{
			name: "invalid language",
			config: &Config{
				TelegramAvailableTokens:   []string{"123456:ABC"},
				Port:                      8080,
				DefaultParseMode:          "Markdown",
				TelegramImageTransferMode: "base64",
				Language:                  "invalid",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := Validate(tt.config)
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestGetEnvHelpers(t *testing.T) {
	// Test getEnvOrDefault
	os.Setenv("TEST_STRING", "value")
	defer os.Unsetenv("TEST_STRING")

	if got := getEnvOrDefault("TEST_STRING", "default"); got != "value" {
		t.Errorf("getEnvOrDefault() = %v, want %v", got, "value")
	}

	if got := getEnvOrDefault("NONEXISTENT", "default"); got != "default" {
		t.Errorf("getEnvOrDefault() = %v, want %v", got, "default")
	}

	// Test getEnvInt
	os.Setenv("TEST_INT", "42")
	defer os.Unsetenv("TEST_INT")

	if got := getEnvInt("TEST_INT", 0); got != 42 {
		t.Errorf("getEnvInt() = %v, want %v", got, 42)
	}

	if got := getEnvInt("NONEXISTENT", 10); got != 10 {
		t.Errorf("getEnvInt() = %v, want %v", got, 10)
	}

	// Test getEnvBool
	os.Setenv("TEST_BOOL", "true")
	defer os.Unsetenv("TEST_BOOL")

	if got := getEnvBool("TEST_BOOL", false); got != true {
		t.Errorf("getEnvBool() = %v, want %v", got, true)
	}

	if got := getEnvBool("NONEXISTENT", false); got != false {
		t.Errorf("getEnvBool() = %v, want %v", got, false)
	}

	// Test getEnvSlice
	os.Setenv("TEST_SLICE", "a,b,c")
	defer os.Unsetenv("TEST_SLICE")

	got := getEnvSlice("TEST_SLICE")
	if len(got) != 3 || got[0] != "a" || got[1] != "b" || got[2] != "c" {
		t.Errorf("getEnvSlice() = %v, want [a b c]", got)
	}

	if got := getEnvSlice("NONEXISTENT"); len(got) != 0 {
		t.Errorf("getEnvSlice() = %v, want []", got)
	}
}
