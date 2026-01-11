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

	// Check SillyTavern context configuration defaults
	if cfg.MaxContextLength != 8000 {
		t.Errorf("Expected MaxContextLength to be 8000, got %d", cfg.MaxContextLength)
	}

	if cfg.SummaryThreshold != 0.8 {
		t.Errorf("Expected SummaryThreshold to be 0.8, got %f", cfg.SummaryThreshold)
	}

	if cfg.MinRecentPairs != 2 {
		t.Errorf("Expected MinRecentPairs to be 2, got %d", cfg.MinRecentPairs)
	}

	// Check Manager configuration defaults
	if cfg.ManagerPort != 8081 {
		t.Errorf("Expected ManagerPort to be 8081, got %d", cfg.ManagerPort)
	}

	if !cfg.ManagerEnabled {
		t.Error("Expected ManagerEnabled to be true")
	}

	// Check Telegraph configuration default
	if !cfg.TelegraphEnabled {
		t.Error("Expected TelegraphEnabled to be true")
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
				MaxContextLength:          8000,
				SummaryThreshold:          0.8,
				MinRecentPairs:            2,
				ManagerPort:               8081,
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
				MaxContextLength:          8000,
				SummaryThreshold:          0.8,
				MinRecentPairs:            2,
				ManagerPort:               8081,
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
				MaxContextLength:          8000,
				SummaryThreshold:          0.8,
				MinRecentPairs:            2,
				ManagerPort:               8081,
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
				MaxContextLength:          8000,
				SummaryThreshold:          0.8,
				MinRecentPairs:            2,
				ManagerPort:               8081,
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
				MaxContextLength:          8000,
				SummaryThreshold:          0.8,
				MinRecentPairs:            2,
				ManagerPort:               8081,
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
				MaxContextLength:          8000,
				SummaryThreshold:          0.8,
				MinRecentPairs:            2,
				ManagerPort:               8081,
			},
			wantErr: true,
		},
		{
			name: "invalid max context length",
			config: &Config{
				TelegramAvailableTokens:   []string{"123456:ABC"},
				Port:                      8080,
				DefaultParseMode:          "Markdown",
				TelegramImageTransferMode: "base64",
				Language:                  "zh-cn",
				MaxContextLength:          0,
				SummaryThreshold:          0.8,
				MinRecentPairs:            2,
				ManagerPort:               8081,
			},
			wantErr: true,
		},
		{
			name: "invalid summary threshold - too low",
			config: &Config{
				TelegramAvailableTokens:   []string{"123456:ABC"},
				Port:                      8080,
				DefaultParseMode:          "Markdown",
				TelegramImageTransferMode: "base64",
				Language:                  "zh-cn",
				MaxContextLength:          8000,
				SummaryThreshold:          -0.1,
				MinRecentPairs:            2,
				ManagerPort:               8081,
			},
			wantErr: true,
		},
		{
			name: "invalid summary threshold - too high",
			config: &Config{
				TelegramAvailableTokens:   []string{"123456:ABC"},
				Port:                      8080,
				DefaultParseMode:          "Markdown",
				TelegramImageTransferMode: "base64",
				Language:                  "zh-cn",
				MaxContextLength:          8000,
				SummaryThreshold:          1.5,
				MinRecentPairs:            2,
				ManagerPort:               8081,
			},
			wantErr: true,
		},
		{
			name: "invalid min recent pairs",
			config: &Config{
				TelegramAvailableTokens:   []string{"123456:ABC"},
				Port:                      8080,
				DefaultParseMode:          "Markdown",
				TelegramImageTransferMode: "base64",
				Language:                  "zh-cn",
				MaxContextLength:          8000,
				SummaryThreshold:          0.8,
				MinRecentPairs:            -1,
				ManagerPort:               8081,
			},
			wantErr: true,
		},
		{
			name: "invalid manager port - too low",
			config: &Config{
				TelegramAvailableTokens:   []string{"123456:ABC"},
				Port:                      8080,
				DefaultParseMode:          "Markdown",
				TelegramImageTransferMode: "base64",
				Language:                  "zh-cn",
				MaxContextLength:          8000,
				SummaryThreshold:          0.8,
				MinRecentPairs:            2,
				ManagerPort:               0,
			},
			wantErr: true,
		},
		{
			name: "invalid manager port - too high",
			config: &Config{
				TelegramAvailableTokens:   []string{"123456:ABC"},
				Port:                      8080,
				DefaultParseMode:          "Markdown",
				TelegramImageTransferMode: "base64",
				Language:                  "zh-cn",
				MaxContextLength:          8000,
				SummaryThreshold:          0.8,
				MinRecentPairs:            2,
				ManagerPort:               70000,
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

	// Test getEnvFloat64
	os.Setenv("TEST_FLOAT", "0.75")
	defer os.Unsetenv("TEST_FLOAT")

	if got := getEnvFloat64("TEST_FLOAT", 0.0); got != 0.75 {
		t.Errorf("getEnvFloat64() = %v, want %v", got, 0.75)
	}

	if got := getEnvFloat64("NONEXISTENT", 1.5); got != 1.5 {
		t.Errorf("getEnvFloat64() = %v, want %v", got, 1.5)
	}
}

func TestSillyTavernConfigFromEnv(t *testing.T) {
	// Set required environment variable
	os.Setenv("TELEGRAM_AVAILABLE_TOKENS", "123456:ABC-DEF")
	defer os.Unsetenv("TELEGRAM_AVAILABLE_TOKENS")

	// Set SillyTavern configuration
	os.Setenv("MAX_CONTEXT_LENGTH", "10000")
	os.Setenv("SUMMARY_THRESHOLD", "0.75")
	os.Setenv("MIN_RECENT_PAIRS", "3")
	os.Setenv("MANAGER_PORT", "9090")
	os.Setenv("MANAGER_ENABLED", "false")
	os.Setenv("TELEGRAPH_ENABLED", "false")

	defer func() {
		os.Unsetenv("MAX_CONTEXT_LENGTH")
		os.Unsetenv("SUMMARY_THRESHOLD")
		os.Unsetenv("MIN_RECENT_PAIRS")
		os.Unsetenv("MANAGER_PORT")
		os.Unsetenv("MANAGER_ENABLED")
		os.Unsetenv("TELEGRAPH_ENABLED")
	}()

	cfg, err := LoadConfig()
	if err != nil {
		t.Fatalf("LoadConfig() failed: %v", err)
	}

	// Verify custom values are loaded
	if cfg.MaxContextLength != 10000 {
		t.Errorf("Expected MaxContextLength to be 10000, got %d", cfg.MaxContextLength)
	}

	if cfg.SummaryThreshold != 0.75 {
		t.Errorf("Expected SummaryThreshold to be 0.75, got %f", cfg.SummaryThreshold)
	}

	if cfg.MinRecentPairs != 3 {
		t.Errorf("Expected MinRecentPairs to be 3, got %d", cfg.MinRecentPairs)
	}

	if cfg.ManagerPort != 9090 {
		t.Errorf("Expected ManagerPort to be 9090, got %d", cfg.ManagerPort)
	}

	if cfg.ManagerEnabled {
		t.Error("Expected ManagerEnabled to be false")
	}

	if cfg.TelegraphEnabled {
		t.Error("Expected TelegraphEnabled to be false")
	}
}
