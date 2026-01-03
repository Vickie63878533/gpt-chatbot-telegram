package plugin

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

// PluginRegistry holds all loaded plugins
type PluginRegistry struct {
	Plugins map[string]*PluginConfig
	Env     map[string]string
}

// NewPluginRegistry creates a new plugin registry
func NewPluginRegistry() *PluginRegistry {
	return &PluginRegistry{
		Plugins: make(map[string]*PluginConfig),
		Env:     make(map[string]string),
	}
}

// LoadFromEnvironment loads plugins from environment variables
// PLUGIN_COMMAND_<name> = template JSON or URL
// PLUGIN_DESCRIPTION_<name> = description
// PLUGIN_SCOPE_<name> = comma-separated scopes
// PLUGIN_ENV_<name> = environment variable value
func (r *PluginRegistry) LoadFromEnvironment() error {
	envVars := os.Environ()

	// First pass: collect plugin commands
	commands := make(map[string]string)
	descriptions := make(map[string]string)
	scopes := make(map[string]string)

	for _, env := range envVars {
		parts := strings.SplitN(env, "=", 2)
		if len(parts) != 2 {
			continue
		}
		key := parts[0]
		value := parts[1]

		if strings.HasPrefix(key, "PLUGIN_COMMAND_") {
			name := strings.TrimPrefix(key, "PLUGIN_COMMAND_")
			commands[name] = value
		} else if strings.HasPrefix(key, "PLUGIN_DESCRIPTION_") {
			name := strings.TrimPrefix(key, "PLUGIN_DESCRIPTION_")
			descriptions[name] = value
		} else if strings.HasPrefix(key, "PLUGIN_SCOPE_") {
			name := strings.TrimPrefix(key, "PLUGIN_SCOPE_")
			scopes[name] = value
		} else if strings.HasPrefix(key, "PLUGIN_ENV_") {
			name := strings.TrimPrefix(key, "PLUGIN_ENV_")
			r.Env[name] = value
		}
	}

	// Second pass: create plugin configs
	for name, value := range commands {
		config := &PluginConfig{
			Value:       value,
			Description: descriptions[name],
		}

		if scopeStr, ok := scopes[name]; ok && scopeStr != "" {
			config.Scope = strings.Split(scopeStr, ",")
			for i := range config.Scope {
				config.Scope[i] = strings.TrimSpace(config.Scope[i])
			}
		}

		r.Plugins["/"+name] = config
	}

	return nil
}

// LoadFromDirectory loads plugins from a directory
func (r *PluginRegistry) LoadFromDirectory(dir string) error {
	entries, err := os.ReadDir(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil // Directory doesn't exist, not an error
		}
		return fmt.Errorf("failed to read plugin directory: %w", err)
	}

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		if !strings.HasSuffix(entry.Name(), ".json") {
			continue
		}

		path := filepath.Join(dir, entry.Name())
		data, err := os.ReadFile(path)
		if err != nil {
			return fmt.Errorf("failed to read plugin file %s: %w", path, err)
		}

		// Plugin name is the filename without extension
		name := strings.TrimSuffix(entry.Name(), ".json")

		config := &PluginConfig{
			Value: string(data),
		}

		r.Plugins["/"+name] = config
	}

	return nil
}

// GetTemplate loads and parses a plugin template
func (r *PluginRegistry) GetTemplate(command string) (*RequestTemplate, error) {
	config, ok := r.Plugins[command]
	if !ok {
		return nil, fmt.Errorf("plugin not found: %s", command)
	}

	templateStr := strings.TrimSpace(config.Value)

	// If it's a URL, fetch it
	if strings.HasPrefix(templateStr, "http://") || strings.HasPrefix(templateStr, "https://") {
		resp, err := http.Get(templateStr)
		if err != nil {
			return nil, fmt.Errorf("failed to fetch plugin template: %w", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			return nil, fmt.Errorf("failed to fetch plugin template: status %d", resp.StatusCode)
		}

		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return nil, fmt.Errorf("failed to read plugin template: %w", err)
		}
		templateStr = string(body)
	}

	// Parse JSON
	var template RequestTemplate
	if err := json.Unmarshal([]byte(templateStr), &template); err != nil {
		return nil, fmt.Errorf("failed to parse plugin template: %w", err)
	}

	return &template, nil
}

// HasPlugin checks if a plugin exists
func (r *PluginRegistry) HasPlugin(command string) bool {
	_, ok := r.Plugins[command]
	return ok
}

// GetPluginConfig returns the plugin configuration
func (r *PluginRegistry) GetPluginConfig(command string) *PluginConfig {
	return r.Plugins[command]
}

// ListPlugins returns all plugin commands
func (r *PluginRegistry) ListPlugins() []string {
	commands := make([]string, 0, len(r.Plugins))
	for cmd := range r.Plugins {
		commands = append(commands, cmd)
	}
	return commands
}
