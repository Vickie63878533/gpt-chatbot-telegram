package plugin

import (
	"os"
	"testing"
)

func TestLoadFromEnvironment(t *testing.T) {
	// Set up test environment variables
	os.Setenv("PLUGIN_COMMAND_test", `{"url":"https://example.com","method":"GET"}`)
	os.Setenv("PLUGIN_DESCRIPTION_test", "Test plugin")
	os.Setenv("PLUGIN_SCOPE_test", "all_private_chats,all_chat_administrators")
	os.Setenv("PLUGIN_ENV_API_KEY", "test-key-123")
	defer func() {
		os.Unsetenv("PLUGIN_COMMAND_test")
		os.Unsetenv("PLUGIN_DESCRIPTION_test")
		os.Unsetenv("PLUGIN_SCOPE_test")
		os.Unsetenv("PLUGIN_ENV_API_KEY")
	}()

	registry := NewPluginRegistry()
	err := registry.LoadFromEnvironment()
	if err != nil {
		t.Fatalf("Failed to load from environment: %v", err)
	}

	// Check plugin was loaded
	if !registry.HasPlugin("/test") {
		t.Error("Plugin /test was not loaded")
	}

	// Check plugin config
	config := registry.GetPluginConfig("/test")
	if config == nil {
		t.Fatal("Plugin config is nil")
	}

	if config.Description != "Test plugin" {
		t.Errorf("Expected description 'Test plugin', got %q", config.Description)
	}

	if len(config.Scope) != 2 {
		t.Errorf("Expected 2 scopes, got %d", len(config.Scope))
	}

	// Check environment variable
	if registry.Env["API_KEY"] != "test-key-123" {
		t.Errorf("Expected env API_KEY='test-key-123', got %q", registry.Env["API_KEY"])
	}
}

func TestListPlugins(t *testing.T) {
	registry := NewPluginRegistry()
	registry.Plugins["/test1"] = &PluginConfig{Value: "{}"}
	registry.Plugins["/test2"] = &PluginConfig{Value: "{}"}

	plugins := registry.ListPlugins()
	if len(plugins) != 2 {
		t.Errorf("Expected 2 plugins, got %d", len(plugins))
	}
}

func TestGetTemplate(t *testing.T) {
	registry := NewPluginRegistry()
	templateJSON := `{
		"url": "https://api.example.com/{{DATA}}",
		"method": "GET",
		"input": {
			"type": "text",
			"required": true
		},
		"response": {
			"content": {
				"input_type": "json",
				"output_type": "text",
				"output": "Result: {{result}}"
			},
			"error": {
				"input_type": "text",
				"output_type": "text",
				"output": "Error: {{.}}"
			}
		}
	}`

	registry.Plugins["/test"] = &PluginConfig{
		Value: templateJSON,
	}

	template, err := registry.GetTemplate("/test")
	if err != nil {
		t.Fatalf("Failed to get template: %v", err)
	}

	if template.URL != "https://api.example.com/{{DATA}}" {
		t.Errorf("Expected URL with placeholder, got %q", template.URL)
	}

	if template.Method != "GET" {
		t.Errorf("Expected method GET, got %q", template.Method)
	}

	if template.Input.Type != InputTypeText {
		t.Errorf("Expected input type text, got %q", template.Input.Type)
	}
}
